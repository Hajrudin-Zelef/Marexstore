package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// clientLabel : client_json {"nom":"…","tel":"…"} → « Nom · tél » lisible.
func clientLabel(j string) string {
	var m map[string]string
	if err := json.Unmarshal([]byte(j), &m); err == nil {
		nom, tel := strings.TrimSpace(m["nom"]), strings.TrimSpace(m["tel"])
		switch {
		case nom != "" && tel != "":
			return nom + " · " + tel
		case nom != "":
			return nom
		case tel != "":
			return tel
		}
	}
	if len(j) > 60 {
		return j[:60] + "…"
	}
	return j
}

// Transitions autorisées (machine à états, jamais de saut libre).
var orderNext = map[string][]string{
	"devis":         {"whatsapp_recu", "annule"},
	"whatsapp_recu": {"attente_momo", "annule"},
	"attente_momo":  {"paye", "annule"},
	"paye":          {"prepare", "annule"},
	"prepare":       {"livre", "annule"},
	"livre":         {},
	"annule":        {},
}

// AdminCommande : détail + transitions + timeline.
func (a *App) AdminCommande(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var st, moy, client, at string
	var tot sql.NullInt64
	var uid sql.NullInt64
	err := a.DB.QueryRow(`SELECT statut,moyen_paiement,client_json,created_at,total,user_id FROM orders WHERE id=?`, id).Scan(&st, &moy, &client, &at, &tot, &uid)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		vers := r.FormValue("statut")
		ok := false
		for _, n := range orderNext[st] {
			if n == vers {
				ok = true
			}
		}
		if !ok {
			http.Redirect(w, r, fmt.Sprintf("/admin/commandes/%d?err=Transition+interdite", id), http.StatusSeeOther)
			return
		}
		// T1 — backend dur : transition + événement + stock en UNE SEULE
		// transaction. Toute erreur annule tout (jamais de statut sans
		// événement, jamais de stock décrémenté à moitié).
		tx, err := a.DB.Begin()
		if err != nil {
			http.Error(w, "impossible", 500)
			return
		}
		defer tx.Rollback()
		if _, err := tx.Exec(`UPDATE orders SET statut=? WHERE id=?`, vers, id); err != nil {
			http.Error(w, "impossible", 500)
			return
		}
		if _, err := tx.Exec(`INSERT INTO order_events(order_id,de_statut,vers_statut,par) VALUES(?,?,?,?)`, id, st, vers, "admin"); err != nil {
			http.Error(w, "impossible", 500)
			return
		}
		if vers == "paye" {
			// Décrémente le stock des lignes (plancher 0, jamais négatif).
			// On collecte d'abord les lignes, puis on met à jour : une
			// erreur sur une ligne annule toute la transition.
			type ligne struct{ pid, q int64 }
			var lignes []ligne
			rows, err := tx.Query(`SELECT product_id,qty FROM order_items WHERE order_id=?`, id)
			if err != nil {
				http.Error(w, "impossible", 500)
				return
			}
			for rows.Next() {
				var l ligne
				if err := rows.Scan(&l.pid, &l.q); err != nil {
					rows.Close()
					http.Error(w, "impossible", 500)
					return
				}
				lignes = append(lignes, l)
			}
			rows.Close()
			for _, l := range lignes {
				if _, err := tx.Exec(`UPDATE products SET stock=CASE WHEN stock>=? THEN stock-? ELSE 0 END WHERE id=? AND stock IS NOT NULL`, l.q, l.q, l.pid); err != nil {
					http.Error(w, "impossible", 500)
					return
				}
			}
		}
		if err := tx.Commit(); err != nil {
			http.Error(w, "impossible", 500)
			return
		}
		a.auditAdmin(r, "commande.statut", fmt.Sprintf("#%d %s -> %s", id, st, vers))
		a.reload()
		http.Redirect(w, r, "/admin/commandes/"+strconv.FormatInt(id, 10)+"?msg=Commande+enregistr%C3%A9e", http.StatusSeeOther)
		return
	}
	t := "devis"
	if tot.Valid {
		t = fmtMoney(tot.Int64) + " F CFA"
	}
	// Lignes (+ total ligne = qté × prix unit.).
	var items strings.Builder
	var lignesTotal int64
	ir, _ := a.DB.Query(`SELECT p.nom,p.slug,oi.qty,oi.prix_unitaire FROM order_items oi JOIN products p ON p.id=oi.product_id WHERE oi.order_id=?`, id)
	if ir != nil {
		defer ir.Close()
		for ir.Next() {
			var nom, slug string
			var q int64
			var pu sql.NullInt64
			_ = ir.Scan(&nom, &slug, &q, &pu)
			pus, lts := "—", "—"
			if pu.Valid {
				pus = fmtMoney(pu.Int64) + " F CFA"
				lignesTotal += q * pu.Int64
				lts = fmtMoney(q*pu.Int64) + " F CFA"
			}
			fmt.Fprintf(&items, `<tr><td><a href="/admin/produits/%s">%s</a></td><td>%d</td><td>%s</td><td><strong>%s</strong></td></tr>`, slug, render.Esc(nom), q, pus, lts)
		}
	}
	// Stepper de la machine à états.
	steps := []string{"devis", "whatsapp_recu", "attente_momo", "paye", "prepare", "livre"}
	stepLabels := map[string]string{"devis": "Devis", "whatsapp_recu": "WhatsApp reçu", "attente_momo": "Attente MoMo", "paye": "Payée", "prepare": "Préparée", "livre": "Livrée"}
	var stepper strings.Builder
	if st == "annule" {
		stepper.WriteString(`<div class="mx-notice err" style="margin-bottom:16px">Commande annulée.</div>`)
	} else {
		cur := 0
		for i, s := range steps {
			if s == st {
				cur = i
			}
		}
		stepper.WriteString(`<ol class="mx-steps">`)
		for i, s := range steps {
			cls := "todo"
			if i < cur {
				cls = "done"
			} else if i == cur {
				cls = "cur"
			}
			fmt.Fprintf(&stepper, `<li class="%s"><span class="dot">%d</span>%s</li>`, cls, i+1, render.Esc(stepLabels[s]))
		}
		stepper.WriteString(`</ol>`)
	}
	// Timeline.
	var tl strings.Builder
	fmt.Fprintf(&tl, `<li><strong>Création</strong> — statut <em>devis</em><div class="t">%s</div></li>`, render.Esc(at))
	er, _ := a.DB.Query(`SELECT de_statut,vers_statut,par,cree_le FROM order_events WHERE order_id=? ORDER BY id`, id)
	if er != nil {
		defer er.Close()
		for er.Next() {
			var de, vers, par, at2 string
			_ = er.Scan(&de, &vers, &par, &at2)
			fmt.Fprintf(&tl, `<li><strong>%s</strong> → %s <span class="t">par %s · %s</span></li>`, render.Esc(de), statutBadge(vers), render.Esc(par), render.Esc(at2))
		}
	}
	// Boutons de transition.
	var btns strings.Builder
	for _, n := range orderNext[st] {
		fmt.Fprintf(&btns, `<form method="post" style="display:inline"><input type="hidden" name="csrf" value="%s"><input type="hidden" name="statut" value="%s"><button class="mx-btn">→ %s</button></form> `, a.adminCSRF(), n, render.Esc(n))
	}
	if len(orderNext[st]) == 0 {
		btns.WriteString(`<span class="mx-badge">statut final</span>`)
	}
	cli := "invité"
	if uid.Valid {
		var em, nm string
		if err := a.DB.QueryRow(`SELECT email,nom FROM users WHERE id=?`, uid.Int64).Scan(&em, &nm); err == nil {
			cli = fmt.Sprintf(`<a href="/admin/clients/%d">%s</a> (%s)`, uid.Int64, render.Esc(nm), render.Esc(em))
		}
	}
	a.write(w, 200, a.adminPage(fmt.Sprintf("Commande #%d", id), "commandes",
		stepper.String()+`
<div class="mx-cols"><div>
<div class="mx-postbox"><h2>Lignes</h2><div class="in"><div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>Qté</th><th>Prix unit.</th><th>Total ligne</th></tr>`+items.String()+`</table></div>
<p style="text-align:right">Total des lignes : <strong>`+fmtMoney(lignesTotal)+` F CFA</strong></p></div></div>
<div class="mx-postbox"><h2>Historique</h2><div class="in"><ul class="mx-timeline">`+tl.String()+`</ul></div></div>
</div><div><div class="mx-postbox mx-pub"><h2>Résumé</h2><div class="in"><p class="mx-big-total">`+t+`</p><p>Moyen : `+render.Esc(moy)+`<br>Client : `+render.Esc(clientLabel(client))+`<br>Compte : `+cli+`<br><span class="mx-muted">Créée le `+render.Esc(at)+`</span></p></div></div>
<div class="mx-postbox mx-pub"><h2>Statut</h2><div class="in"><p>Actuel : `+statutBadge(st)+`</p>`+btns.String()+`</div></div></div></div>`, r))
}

// AdminClients : liste des comptes.
func (a *App) AdminClients(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const per = 30
	var total int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE email LIKE ? OR nom LIKE ?`, "%"+q+"%", "%"+q+"%").Scan(&total)
	pages := (total + per - 1) / per
	if pages < 1 {
		pages = 1
	}
	// Mini-stats : avec / sans commande, total commandes.
	var avecCmd, totCmd int64
	_ = a.DB.QueryRow(`SELECT COUNT(DISTINCT user_id) FROM orders WHERE user_id IS NOT NULL`).Scan(&avecCmd)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&totCmd)
	var totClients int64
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&totClients)
	rows, _ := a.DB.Query(`SELECT u.id,u.email,u.nom,u.tel,u.created_at,(SELECT COUNT(*) FROM orders WHERE user_id=u.id) FROM users u WHERE u.email LIKE ? OR u.nom LIKE ? ORDER BY u.id DESC LIMIT ? OFFSET ?`, "%"+q+"%", "%"+q+"%", per, (page-1)*per)
	var tr strings.Builder
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id, nb int64
			var em, nm, tel, at string
			_ = rows.Scan(&id, &em, &nm, &tel, &at, &nb)
			ini := "?"
			if nm != "" {
				ini = strings.ToUpper(string([]rune(nm)[0]))
			} else if em != "" {
				ini = strings.ToUpper(string([]rune(em)[0]))
			}
			cmdBadge := `<span class="mx-badge">0 commande</span>`
			if nb > 0 {
				cmdBadge = fmt.Sprintf(`<span class="mx-badge ok">%d commande(s)</span>`, nb)
			}
			fmt.Fprintf(&tr, `<tr><td><a href="/admin/clients/%d"><span class="mx-av">%s</span><strong>%s</strong></a><div class="row-actions"><a href="/admin/clients/%d">Détail</a></div></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				id, render.Esc(ini), render.Esc(em), id, render.Esc(nm), render.Esc(tel), render.Esc(at), cmdBadge)
		}
	}
	a.write(w, 200, a.adminPage("Clients", "clients",
		`<div class="mx-stats">`+
			miniStat("/admin/clients", "Clients", fmt.Sprint(totClients), "#2e5136")+
			miniStat("/admin/clients", "Avec commande", fmt.Sprint(avecCmd), "#2271b1")+
			miniStat("/admin/clients", "Sans commande", fmt.Sprint(totClients-avecCmd), "#dba617")+
			miniStat("/admin/commandes", "Commandes", fmt.Sprint(totCmd), "#6fa82e")+
			`</div>
<form class="mx-tools" method="get" action="/admin/clients"><input type="search" name="q" value="`+render.Esc(q)+`" placeholder="Rechercher e-mail, nom…"><button class="mx-btn">OK</button></form>
<div class="mx-tablewrap"><table class="mx-table"><tr><th>E-mail</th><th>Nom</th><th>Tél</th><th>Inscription</th><th>Commandes</th></tr>`+tr.String()+`</table></div>`+render.MxPager("/admin/clients?q="+url.QueryEscape(q), page, pages), r))
}

// AdminClient : détail + commandes du compte.
func (a *App) AdminClient(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var em, nm, tel, at string
	if err := a.DB.QueryRow(`SELECT email,nom,tel,created_at FROM users WHERE id=?`, id).Scan(&em, &nm, &tel, &at); err != nil {
		http.NotFound(w, r)
		return
	}
	var nbCmd, totDep int64
	_ = a.DB.QueryRow(`SELECT COUNT(*),COALESCE(SUM(total),0) FROM orders WHERE user_id=?`, id).Scan(&nbCmd, &totDep)
	ini := "?"
	if nm != "" {
		ini = strings.ToUpper(string([]rune(nm)[0]))
	} else if em != "" {
		ini = strings.ToUpper(string([]rune(em)[0]))
	}
	var tr strings.Builder
	rows, _ := a.DB.Query(`SELECT id,statut,total,created_at FROM orders WHERE user_id=? ORDER BY id DESC`, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var oid int64
			var st, at2 string
			var tot sql.NullInt64
			_ = rows.Scan(&oid, &st, &tot, &at2)
			t := `<span class="mx-muted">devis</span>`
			if tot.Valid {
				t = `<span class="mx-price">` + fmtMoney(tot.Int64) + ` F CFA</span>`
			}
			fmt.Fprintf(&tr, `<tr><td><a href="/admin/commandes/%d">#%d</a></td><td>%s</td><td>%s</td><td>%s</td></tr>`, oid, oid, render.Esc(at2), t, statutBadge(st))
		}
	}
	a.write(w, 200, a.adminPage("Client — "+nm, "clients",
		`<div class="mx-cols"><div><div class="mx-postbox"><h2>Commandes</h2><div class="in"><div class="mx-tablewrap"><table class="mx-table"><tr><th>#</th><th>Date</th><th>Total</th><th>Statut</th></tr>`+tr.String()+`</table></div></div></div>
</div><div><div class="mx-preview mx-client"><div class="pv-top"><span class="mx-av-big">`+render.Esc(ini)+`</span><div class="pv-nom">`+render.Esc(nm)+`</div><span class="pv-badge">`+render.Esc(em)+`</span></div><div class="pv-body"><div class="pv-info"><div class="pv-row"><span>Téléphone</span><strong>`+render.Esc(tel)+`</strong></div><div class="pv-row"><span>Inscription</span><strong>`+render.Esc(at)+`</strong></div><div class="pv-row"><span>Commandes</span><strong>`+fmt.Sprint(nbCmd)+`</strong></div><div class="pv-row"><span>Total dépensé</span><strong class="pv-prix">`+fmtMoney(totDep)+` F CFA</strong></div></div></div></div></div></div>`, r))
}

// AdminMedias : grille + détection d'orphelins.
func (a *App) AdminMedias(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	vue := r.URL.Query().Get("vue")
	if vue == "" {
		vue = "all"
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	// Références en base.
	used := map[string]bool{}
	ur, _ := a.DB.Query(`SELECT DISTINCT image FROM products WHERE image<>''`)
	if ur != nil {
		defer ur.Close()
		for ur.Next() {
			var im string
			_ = ur.Scan(&im)
			used[im] = true
		}
	}
	if r.Method == http.MethodPost && r.FormValue("supprimer") != "" {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		rel := r.FormValue("supprimer")
		if !used[rel] && !strings.Contains(rel, "..") && strings.HasPrefix(rel, "images/") {
			_ = os.Remove(rel)
			http.Redirect(w, r, "/admin/medias?vue=orphelins&msg=Fichier+supprim%C3%A9", http.StatusSeeOther)
			return
		}
		http.Redirect(w, r, "/admin/medias?err=Suppression+refus%C3%A9e+(image+utilis%C3%A9e)", http.StatusSeeOther)
		return
	}
	entries, _ := os.ReadDir("images")
	var orph, all strings.Builder
	nAll, nOrph := 0, 0
	var sizeTotal int64
	ql := strings.ToLower(q)
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if ql != "" && !strings.Contains(strings.ToLower(e.Name()), ql) {
			continue
		}
		if info, err := e.Info(); err == nil {
			sizeTotal += info.Size()
		}
		rel := "images/" + e.Name()
		oc := ""
		if !used[rel] {
			oc = ` <span class="mx-badge warn">orphelin</span>`
			nOrph++
			fmt.Fprintf(&orph, `<figure><img src="/%s" alt="" loading="lazy"><figcaption>%s%s<form method="post"><input type="hidden" name="csrf" value="%s"><input type="hidden" name="supprimer" value="%s"><button class="mx-btn danger" data-mx-confirm="Supprimer définitivement ce fichier ?">Supprimer</button></form></figcaption></figure>`, rel, e.Name(), oc, a.adminCSRF(), rel)
		}
		nAll++
		fmt.Fprintf(&all, `<figure><img src="/%s" alt="" loading="lazy"><figcaption>%s%s</figcaption></figure>`, rel, e.Name(), oc)
	}
	grid := all.String()
	if vue == "orphelins" {
		grid = orph.String()
		if nOrph == 0 {
			grid = `<p>Aucun orphelin. Toutes les images sont référencées.</p>`
		}
	}
	sizeTxt := fmt.Sprintf("%.1f Mo", float64(sizeTotal)/1024/1024)
	qp := ""
	if q != "" {
		qp = "&q=" + url.QueryEscape(q)
	}
	views := render.MxViews([]render.MxView{
		{Key: "all", Href: "/admin/medias?vue=all" + qp, Label: "Toutes", Count: nAll},
		{Key: "orphelins", Href: "/admin/medias?vue=orphelins" + qp, Label: "Orphelins", Count: nOrph},
	}, vue)
	a.write(w, 200, a.adminPage("Médias", "medias",
		`<div class="mx-stats">`+
			miniStat("/admin/medias?vue=all", "Images", fmt.Sprint(nAll), "#2e5136")+
			miniStat("/admin/medias?vue=all", "Utilisées", fmt.Sprint(nAll-nOrph), "#2271b1")+
			miniStat("/admin/medias?vue=orphelins", "Orphelines", fmt.Sprint(nOrph), "#dba617")+
			miniStat("/admin/medias?vue=all", "Taille totale", sizeTxt, "#6fa82e")+
			`</div>
<form class="mx-tools" method="get" action="/admin/medias"><input type="hidden" name="vue" value="`+render.Esc(vue)+`"><input type="search" name="q" value="`+render.Esc(q)+`" placeholder="Rechercher un fichier…"><button class="mx-btn">OK</button></form>`+
			views+`<div class="mx-media">`+grid+`</div>`, r))
}

// AdminTypes : regroupements (live_top/live_sub) + compteurs + fusion.
func (a *App) AdminTypes(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		top, src, dst := r.FormValue("top"), r.FormValue("src"), r.FormValue("dst")
		if top != "" && src != "" && dst != "" && src != dst {
			res, err := a.DB.Exec(`UPDATE products SET live_sub=? WHERE live_top=? AND live_sub=?`, dst, top, src)
			if err == nil {
				if n, _ := res.RowsAffected(); n > 0 {
					slugs, _ := a.DB.Query(`SELECT slug FROM products WHERE live_top=? AND live_sub=?`, top, dst)
					if slugs != nil {
						for slugs.Next() {
							var s string
							_ = slugs.Scan(&s)
							db.SyncFTS(a.DB, s)
						}
						slugs.Close()
					}
					a.reload()
				}
				http.Redirect(w, r, "/admin/types?msg=Types+fusionn%C3%A9s", http.StatusSeeOther)
				return
			}
		}
		http.Redirect(w, r, "/admin/types?err=Fusion+impossible", http.StatusSeeOther)
		return
	}
	rows, _ := a.DB.Query(`SELECT live_top,live_sub,COUNT(*) FROM products GROUP BY live_top,live_sub ORDER BY live_top,live_sub`)
	type subRow struct {
		sub string
		n   int
	}
	byTop := map[string][]subRow{}
	var tops []string
	topTotal := map[string]int{}
	nTypes, nProd := 0, 0
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var top, sub string
			var n int
			_ = rows.Scan(&top, &sub, &n)
			if _, ok := byTop[top]; !ok {
				tops = append(tops, top)
			}
			byTop[top] = append(byTop[top], subRow{sub, n})
			topTotal[top] += n
			nTypes++
			nProd += n
		}
	}
	var opts strings.Builder
	for _, top := range tops {
		for _, s := range byTop[top] {
			fmt.Fprintf(&opts, `<option value="%s|%s">%s › %s</option>`, render.Esc(top), render.Esc(s.sub), render.Esc(top), render.Esc(s.sub))
		}
	}
	// Cartes par top (libellé français, total, liste des types avec compteurs).
	var cards strings.Builder
	bigTop, bigN := "", 0
	for _, top := range tops {
		if topTotal[top] > bigN {
			bigTop, bigN = top, topTotal[top]
		}
	}
	bigLab := topLabels[bigTop]
	if bigLab == "" {
		bigLab = bigTop
	}
	for _, top := range tops {
		lab := topLabels[top]
		if lab == "" {
			lab = top
		}
		fmt.Fprintf(&cards, `<section class="mx-card mx-type"><h2>%s</h2><span class="mx-type-total">%d produit(s)</span><ul class="mx-type-list">`, render.Esc(lab), topTotal[top])
		for _, s := range byTop[top] {
			lab := s.sub
			href := fmt.Sprintf(`/categories/%s.html?type=%s`, render.Esc(top), render.Esc(s.sub))
			if lab == "" {
				lab = "Sans type"
				href = fmt.Sprintf(`/categories/%s.html`, render.Esc(top))
			}
			fmt.Fprintf(&cards, `<li><a href="%s">%s</a><span class="vl">%d</span></li>`, href, render.Esc(lab), s.n)
		}
		cards.WriteString(`</ul></section>`)
	}
	a.write(w, 200, a.adminPage("Types d'équipements", "types",
		`<div class="mx-stats">`+
			miniStat("/admin/types", "Tops", fmt.Sprint(len(tops)), "#2e5136")+
			miniStat("/admin/types", "Types", fmt.Sprint(nTypes), "#2271b1")+
			miniStat("/admin/produits", "Produits", fmt.Sprint(nProd), "#6fa82e")+
			miniStat("/admin/types", "Plus gros top", fmt.Sprint(bigN)+" · "+render.Esc(bigLab), "#dba617")+
			`</div>
<div class="mx-grid mx-types">`+cards.String()+`</div>
<div class="mx-fsec" style="max-width:640px"><h3>Fusionner deux types</h3><form class="mx-form" method="post"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<div class="mx-fgrid"><label>De (sera vidé) <select name="src">`+opts.String()+`</select></label>
<label>Vers <select name="dst">`+opts.String()+`</select></label></div>
<input type="hidden" name="top" id="mx-top"><div class="mx-form-actions"><button class="mx-btn primary">Fusionner</button></div>
<p class="hint">Le store est déduit du type source (« top|sub »). Irréversible : les produits du type source rejoignent le type cible.</p>
</form></div>
<script>document.querySelector('select[name=src]').addEventListener('change',function(e){document.getElementById('mx-top').value=e.target.value.split('|')[0];});document.getElementById('mx-top').value=document.querySelector('select[name=src]').value.split('|')[0];</script>`, r))
}

// AdminReglages : maintenance + WhatsApp (appliqué au redémarrage).
func (a *App) AdminReglages(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	get := func(k, def string) string {
		var v string
		if err := a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom=?`, k).Scan(&v); err != nil {
			return def
		}
		return v
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		set := func(k, v string) {
			_, _ = a.DB.Exec(`INSERT INTO settings(nom,valeur) VALUES(?,?) ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`, k, v)
		}
		maint := "0"
		if r.FormValue("maintenance") == "1" {
			maint = "1"
		}
		set("maintenance", maint)
		set("maintenance_message", strings.TrimSpace(r.FormValue("maintenance_message")))
		set("whatsapp", strings.TrimSpace(r.FormValue("whatsapp")))
		a.auditAdmin(r, "reglages.enregistrer", "maintenance="+maint)
		http.Redirect(w, r, "/admin/reglages?msg=R%C3%A9glages+enregistr%C3%A9s", http.StatusSeeOther)
		return
	}
	maint := get("maintenance", "0")
	mmsg := get("maintenance_message", "Boutique en maintenance, retour très vite.")
	wa := get("whatsapp", a.Cfg.WhatsApp)
	chk := ""
	status := `<div class="mx-notice" style="margin-bottom:16px">Boutique <strong>en ligne</strong> — les visiteurs voient le catalogue normalement.</div>`
	if maint == "1" {
		chk = " checked"
		status = `<div class="mx-notice warn" style="margin-bottom:16px">Mode maintenance <strong>actif</strong> — la boutique affiche un message aux visiteurs, /admin reste accessible.</div>`
	}
	a.write(w, 200, a.adminPage("Réglages", "reglages", status+
		`<form method="post"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<div class="mx-fsec" style="max-width:640px"><h3>Boutique</h3>
<label class="mx-switch-row"><span class="mx-switch-tx"><strong>Mode maintenance</strong><span class="hint">La boutique affiche un message, /admin reste accessible.</span></span><span class="mx-switch"><input type="checkbox" name="maintenance" value="1"`+chk+`><span class="mx-switch-sl"></span></span></label>
<div class="mx-fgrid"><label class="span2">Message de maintenance <input name="maintenance_message" value="`+render.Esc(mmsg)+`"></label></div>
</div>
<div class="mx-fsec" style="max-width:640px"><h3>Contact</h3>
<div class="mx-fgrid"><label class="span2">WhatsApp <input name="whatsapp" value="`+render.Esc(wa)+`"></label></div>
<p class="hint">Appliqué au redémarrage du conteneur.</p>
</div>
<div class="mx-form-actions"><button class="mx-btn primary">Enregistrer</button></div></form>`, r))
}

// saveUpload : photo produit (JPG/PNG/WebP ≤ 5 Mo) → images/<slug>.<ext>.
func (a *App) saveUpload(r *http.Request, slug string) (string, error) {
	f, h, err := r.FormFile("photo")
	if err != nil || h == nil {
		return "", nil // pas de fichier : on garde l'existant
	}
	defer f.Close()
	if h.Size > 5<<20 {
		return "", fmt.Errorf("fichier trop lourd (max 5 Mo)")
	}
	ext := strings.ToLower(filepath.Ext(h.Filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".webp":
		if ext == ".jpeg" {
			ext = ".jpg"
		}
	default:
		return "", fmt.Errorf("format refusé (JPG/PNG/WebP uniquement)")
	}
	rel := "images/" + slug + ext
	dst, err := os.Create(rel)
	if err != nil {
		return "", err
	}
	defer dst.Close()
	if _, err := io.Copy(dst, f); err != nil {
		return "", err
	}
	return rel, nil
}
