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
	var promoCode string
	var remise int64
	var zoneID int64
	var frais int64
	var payeur, adresse, zoneNom, zoneType string
	var paiementStatut string
	err := a.DB.QueryRow(`SELECT o.statut,o.moyen_paiement,o.client_json,o.created_at,o.total,o.user_id,COALESCE(o.promo_code,''),COALESCE(o.remise,0),COALESCE(o.livraison_zone_id,0),COALESCE(o.livraison_frais,0),COALESCE(o.numero_payeur,''),COALESCE(o.adresse,''),COALESCE(z.nom,''),COALESCE(z.type,''),COALESCE(o.paiement_statut,'en_attente') FROM orders o LEFT JOIN livraison_zones z ON z.id=o.livraison_zone_id WHERE o.id=?`, id).Scan(&st, &moy, &client, &at, &tot, &uid, &promoCode, &remise, &zoneID, &frais, &payeur, &adresse, &zoneNom, &zoneType, &paiementStatut)
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
		// Compare-and-swap : la transition ne s'applique que si la commande
		// est toujours dans l'état lu en début de requête. Deux admins
		// qui valident simultanément : un seul gagne, l'autre reçoit une
		// erreur honnête au lieu d'un double événement contradictoire.
		resUpd, err := tx.Exec(`UPDATE orders SET statut=? WHERE id=? AND statut=?`, vers, id, st)
		if err != nil {
			http.Error(w, "impossible", 500)
			return
		}
		if n, _ := resUpd.RowsAffected(); n != 1 {
			http.Redirect(w, r, "/admin/commandes/"+strconv.FormatInt(id, 10)+"?err=Commande+modifi%C3%A9e+entre-temps+(rechargez+la+page)", http.StatusSeeOther)
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
	// Suivi paiement : « Paiement : <Statut> — <total> via <moyen>
	// (payeur) » + bouton d'action suivante (états non terminaux).
	// COALESCE au SELECT : les lignes d'avant la colonne se lisent
	// « En attente », jamais vide.
	paiementBloc := `<p>Paiement : ` + render.Esc(mxLabel(mxPaiementStatutLabels, paiementStatut)) + ` — ` + t + ` via ` + render.Esc(mxLabel(mxPaiementLabels, moy))
	if payeur != "" {
		paiementBloc += ` (` + render.Esc(payeur) + `)`
	}
	paiementBloc += `</p>`
	switch paiementStatut {
	case "en_attente":
		paiementBloc += `<form method="post" action="/admin/commandes/` + strconv.FormatInt(id, 10) + `/paiement"><input type="hidden" name="csrf" value="` + a.adminCSRF() + `"><input type="hidden" name="action" value="marquer_recu"><button class="mx-btn primary">Marquer reçu</button></form>`
	case "a_encaisser":
		paiementBloc += `<form method="post" action="/admin/commandes/` + strconv.FormatInt(id, 10) + `/paiement"><input type="hidden" name="csrf" value="` + a.adminCSRF() + `"><input type="hidden" name="action" value="marquer_encaisse"><button class="mx-btn primary">Marquer encaissé</button></form>`
	}
	promoLigne := ""
	if promoCode != "" && remise > 0 {
		promoLigne = `<br>Promo : <strong>` + render.Esc(promoCode) + `</strong> (−` + fmtMoney(remise) + ` F CFA)`
	}
	livraisonLigne := ""
	if zoneNom != "" {
		suffixe := ""
		if frais > 0 {
			suffixe = " (+" + fmtMoney(frais) + " F CFA)"
		} else if zoneType != "retrait" {
			suffixe = " (Offerte)"
		}
		livraisonLigne = `<br>Livraison : ` + render.Esc(zoneNom) + suffixe
	} else if frais > 0 || zoneID != 0 {
		livraisonLigne = `<br>Livraison : (zone supprimée) (+` + render.Esc(fmtMoney(frais)) + ` F CFA)`
	}
	if payeur != "" {
		livraisonLigne += `<br>N° payeur : ` + render.Esc(payeur)
	}
	if adresse != "" {
		livraisonLigne += `<br>Adresse : ` + render.Esc(adresse)
	}
	a.write(w, 200, a.adminPage(fmt.Sprintf("Commande #%d", id), "commandes",
		stepper.String()+`
<div class="mx-cols"><div>
<div class="mx-postbox"><h2>Lignes</h2><div class="in"><div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>Qté</th><th>Prix unit.</th><th>Total ligne</th></tr>`+items.String()+`</table></div>
<p style="text-align:right">Total des lignes : <strong>`+fmtMoney(lignesTotal)+` F CFA</strong></p></div></div>
<div class="mx-postbox"><h2>Historique</h2><div class="in"><ul class="mx-timeline">`+tl.String()+`</ul></div></div>
</div><div><div class="mx-postbox mx-pub"><h2>Résumé</h2><div class="in"><p class="mx-big-total">`+t+`</p><p>Moyen : `+render.Esc(mxLabel(mxPaiementLabels, moy))+promoLigne+livraisonLigne+`<br>Client : `+render.Esc(clientLabel(client))+`<br>Compte : `+cli+`<br><span class="mx-muted">Créée le `+render.Esc(at)+`</span></p></div></div>
<div class="mx-postbox"><h2>Paiement</h2><div class="in">`+paiementBloc+`</div></div>
<div class="mx-postbox mx-pub"><h2>Statut</h2><div class="in"><p>Actuel : `+statutBadge(st)+`</p>`+btns.String()+`</div></div></div></div>`, r))
}

// AdminCommandePaiement : POST /admin/commandes/{id}/paiement {action}.
// Machine à états stricte : en_attente→recu (marquer_recu),
// a_encaisser→encaisse (marquer_encaisse). Répéter l'action sur un état
// déjà atteint est idempotent (303, rien ne change) ; tout le reste = 400.
// COALESCE : les lignes d'avant la colonne se lisent « en_attente ».
func (a *App) AdminCommandePaiement(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "méthode interdite", http.StatusMethodNotAllowed)
		return
	}
	if !a.checkAdminCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var cur string
	if err := a.DB.QueryRow(`SELECT COALESCE(paiement_statut,'en_attente') FROM orders WHERE id=?`, id).Scan(&cur); err != nil {
		http.NotFound(w, r)
		return
	}
	action := strings.TrimSpace(r.FormValue("action"))
	var cible string
	switch {
	case action == "marquer_recu" && cur == "en_attente":
		cible = "recu"
	case action == "marquer_encaisse" && cur == "a_encaisser":
		cible = "encaisse"
	case (action == "marquer_recu" && cur == "recu") ||
		(action == "marquer_encaisse" && cur == "encaisse"):
		http.Redirect(w, r, "/admin/commandes/"+strconv.FormatInt(id, 10)+"?msg=Paiement+d%C3%A9j%C3%A0+enregistr%C3%A9", http.StatusSeeOther)
		return
	default:
		http.Error(w, "Transition paiement interdite", http.StatusBadRequest)
		return
	}
	// Compare-and-swap comme les transitions de statut : deux validations
	// concurrentes ne produisent pas un état incohérent.
	res, err := a.DB.Exec(`UPDATE orders SET paiement_statut=? WHERE id=? AND COALESCE(paiement_statut,'en_attente')=?`, cible, id, cur)
	if err != nil {
		http.Error(w, "impossible", 500)
		return
	}
	if n, _ := res.RowsAffected(); n != 1 {
		http.Redirect(w, r, "/admin/commandes/"+strconv.FormatInt(id, 10)+"?err=Commande+modifi%C3%A9e+entre-temps+(rechargez+la+page)", http.StatusSeeOther)
		return
	}
	a.auditAdmin(r, "commande.paiement", fmt.Sprintf("#%d %s -> %s", id, cur, cible))
	http.Redirect(w, r, "/admin/commandes/"+strconv.FormatInt(id, 10)+"?msg=Paiement+enregistr%C3%A9", http.StatusSeeOther)
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
</div><div><div class="mx-preview mx-client"><div class="pv-top"><span class="mx-av-big">`+render.Esc(ini)+`</span><div class="pv-nom">`+render.Esc(nm)+`</div><span class="pv-badge">`+render.Esc(em)+`</span></div><div class="pv-body"><div class="pv-info"><div class="pv-row"><span>Téléphone</span><strong>`+render.Esc(tel)+`</strong></div><div class="pv-row"><span>Inscription</span><strong>`+render.Esc(at)+`</strong></div><div class="pv-row"><span>Commandes</span><strong>`+fmt.Sprint(nbCmd)+`</strong></div><div class="pv-row"><span>Total dépensé</span><strong class="pv-prix">`+fmtMoney(totDep)+` F CFA</strong></div></div></div></div>`+a.clientActionsHTML(r, id, em, nm, tel)+`</div></div>`, r))
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
	// Références en base : principale + galerie (images_json), sinon les
	// extras passent pour orphelins et peuvent être supprimés à tort.
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
	if jr, err := a.DB.Query(`SELECT images_json FROM products WHERE images_json<>'' AND images_json<>'[]'`); err == nil && jr != nil {
		defer jr.Close()
		for jr.Next() {
			var js string
			_ = jr.Scan(&js)
			var extras []string
			if _ = json.Unmarshal([]byte(js), &extras); extras != nil {
				for _, e := range extras {
					if e != "" {
						used[e] = true
					}
				}
			}
		}
	}
	if r.Method == http.MethodPost && r.FormValue("supprimer") != "" {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		rel := r.FormValue("supprimer")
		if !used[rel] && !strings.Contains(rel, "..") && strings.HasPrefix(rel, "images/") {
			if err := os.Remove(rel); err != nil {
				http.Redirect(w, r, "/admin/medias?err=Suppression+impossible+("+url.QueryEscape(err.Error())+")", http.StatusSeeOther)
				return
			}
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

// splitType : valeur de select "top|sub" → (top, sub).
func splitType(v string) (string, string) {
	if i := strings.Index(v, "|"); i >= 0 {
		return v[:i], v[i+1:]
	}
	return "", v
}

// fuseTypes : déplace les produits live_sub src→dst sous srcTop, en
// transaction (rollback complet en cas d'erreur). La source déclarée est
// retirée (elle serait vide et fantôme). Retourne les slugs déplacés.
func (a *App) fuseTypes(srcTop, srcSub, dstSub string) ([]string, error) {
	tx, err := a.DB.Begin()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()
	res, err := tx.Exec(`UPDATE products SET live_sub=? WHERE live_top=? AND live_sub=?`, dstSub, srcTop, srcSub)
	if err != nil {
		return nil, err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return nil, nil
	}
	var moved []string
	rows, err := tx.Query(`SELECT slug FROM products WHERE live_top=? AND live_sub=?`, srcTop, dstSub)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var s string
		_ = rows.Scan(&s)
		moved = append(moved, s)
	}
	rows.Close()
	if _, err := tx.Exec(`DELETE FROM declared_types WHERE top=? AND sub=?`, srcTop, srcSub); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return moved, nil
}

// adminTypesAction : gestion des catégories (créer/renommer/fusionner/
// supprimer/creer_top). Chaque succès : reload + audit + msg ; chaque
// échec : err honnête, rien à moitié appliqué.
func (a *App) adminTypesAction(w http.ResponseWriter, r *http.Request, act string) {
	switch act {
	case "creer":
		top := strings.TrimSpace(r.FormValue("top"))
		sub := strings.TrimSpace(r.FormValue("sub"))
		if top == "" || sub == "" || strings.Contains(top, "|") || strings.Contains(sub, "|") {
			http.Redirect(w, r, "/admin/types?err=Nom+de+cat%C3%A9gorie+invalide", http.StatusSeeOther)
			return
		}
		var n, nd int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top=? AND live_sub=?`, top, sub).Scan(&n)
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top=? AND sub=?`, top, sub).Scan(&nd)
		if n+nd > 0 {
			http.Redirect(w, r, "/admin/types?err=Cat%C3%A9gorie+d%C3%A9j%C3%A0+existante", http.StatusSeeOther)
			return
		}
		if _, err := a.DB.Exec(`INSERT INTO declared_types(top, sub) VALUES(?,?)`, top, sub); err != nil {
			http.Redirect(w, r, "/admin/types?err=Cr%C3%A9ation+impossible", http.StatusSeeOther)
			return
		}
		a.reload()
		a.auditAdmin(r, "types.creer", top+"|"+sub)
		http.Redirect(w, r, "/admin/types?msg=Cat%C3%A9gorie+cr%C3%A9%C3%A9e", http.StatusSeeOther)
	case "creer_top":
		top := strings.TrimSpace(r.FormValue("top"))
		if top == "" || strings.Contains(top, "|") {
			http.Redirect(w, r, "/admin/types?err=Nom+de+store+invalide", http.StatusSeeOther)
			return
		}
		var n, nd int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top=?`, top).Scan(&n)
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top=?`, top).Scan(&nd)
		if n+nd > 0 {
			http.Redirect(w, r, "/admin/types?err=Store+d%C3%A9j%C3%A0+existant", http.StatusSeeOther)
			return
		}
		if _, err := a.DB.Exec(`INSERT INTO declared_types(top, sub) VALUES(?,'')`, top); err != nil {
			http.Redirect(w, r, "/admin/types?err=Cr%C3%A9ation+impossible", http.StatusSeeOther)
			return
		}
		a.reload()
		a.auditAdmin(r, "types.creer_top", top)
		http.Redirect(w, r, "/admin/types?msg=Store+cr%C3%A9%C3%A9", http.StatusSeeOther)
	case "renommer":
		top := strings.TrimSpace(r.FormValue("top"))
		ancien := strings.TrimSpace(r.FormValue("ancien"))
		nouveau := strings.TrimSpace(r.FormValue("nouveau"))
		if top == "" || ancien == "" || nouveau == "" || ancien == nouveau ||
			strings.Contains(top, "|") || strings.Contains(ancien, "|") || strings.Contains(nouveau, "|") {
			http.Redirect(w, r, "/admin/types?err=Renommage+impossible", http.StatusSeeOther)
			return
		}
		var pris, prisD int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top=? AND live_sub=?`, top, nouveau).Scan(&pris)
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top=? AND sub=?`, top, nouveau).Scan(&prisD)
		if pris+prisD > 0 {
			http.Redirect(w, r, "/admin/types?err=Cible+d%C3%A9j%C3%A0+prise", http.StatusSeeOther)
			return
		}
		tx, err := a.DB.Begin()
		if err != nil {
			http.Redirect(w, r, "/admin/types?err=Renommage+impossible", http.StatusSeeOther)
			return
		}
		defer tx.Rollback()
		if _, err := tx.Exec(`UPDATE products SET live_sub=? WHERE live_top=? AND live_sub=?`, nouveau, top, ancien); err != nil {
			http.Redirect(w, r, "/admin/types?err=Renommage+impossible", http.StatusSeeOther)
			return
		}
		if _, err := tx.Exec(`UPDATE declared_types SET sub=? WHERE top=? AND sub=?`, nouveau, top, ancien); err != nil {
			http.Redirect(w, r, "/admin/types?err=Renommage+impossible", http.StatusSeeOther)
			return
		}
		if err := tx.Commit(); err != nil {
			http.Redirect(w, r, "/admin/types?err=Renommage+impossible", http.StatusSeeOther)
			return
		}
		var renommes []string
		if slugs, qerr := a.DB.Query(`SELECT slug FROM products WHERE live_top=? AND live_sub=?`, top, nouveau); qerr == nil && slugs != nil {
			for slugs.Next() {
				var s string
				_ = slugs.Scan(&s)
				renommes = append(renommes, s)
			}
			slugs.Close()
		}
		for _, s := range renommes {
			db.SyncFTS(a.DB, s)
		}
		a.reload()
		a.auditAdmin(r, "types.renommer", top+"|"+ancien+" -> "+nouveau)
		http.Redirect(w, r, "/admin/types?msg=Cat%C3%A9gorie+renomm%C3%A9e", http.StatusSeeOther)
	case "fusionner":
		srcTop, srcSub := splitType(strings.TrimSpace(r.FormValue("src")))
		_, dstSub := splitType(strings.TrimSpace(r.FormValue("dst")))
		dstTop, _ := splitType(strings.TrimSpace(r.FormValue("dst")))
		if srcTop == "" || srcSub == "" || dstSub == "" || (srcTop == dstTop && srcSub == dstSub) {
			http.Redirect(w, r, "/admin/types?err=Fusion+impossible", http.StatusSeeOther)
			return
		}
		moved, err := a.fuseTypes(srcTop, srcSub, dstSub)
		if err != nil {
			http.Redirect(w, r, "/admin/types?err=Fusion+impossible", http.StatusSeeOther)
			return
		}
		if len(moved) == 0 {
			http.Redirect(w, r, "/admin/types?err=Rien+%C3%A0+fusionner+(type+source+vide)", http.StatusSeeOther)
			return
		}
		for _, s := range moved {
			db.SyncFTS(a.DB, s)
		}
		a.reload()
		a.auditAdmin(r, "types.fusionner", srcTop+"|"+srcSub+" -> "+dstSub)
		http.Redirect(w, r, "/admin/types?msg=Types+fusionn%C3%A9s", http.StatusSeeOther)
	case "supprimer":
		top := strings.TrimSpace(r.FormValue("top"))
		sub := strings.TrimSpace(r.FormValue("sub"))
		if top == "" || sub == "" || strings.Contains(top, "|") || strings.Contains(sub, "|") {
			http.Redirect(w, r, "/admin/types?err=Suppression+impossible", http.StatusSeeOther)
			return
		}
		var nprod int
		_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top=? AND live_sub=?`, top, sub).Scan(&nprod)
		if nprod > 0 {
			http.Redirect(w, r, "/admin/types?err=Type+non+vide+("+fmt.Sprint(nprod)+"+produit(s)+%3A+fusionner+d%27abord)", http.StatusSeeOther)
			return
		}
		res, err := a.DB.Exec(`DELETE FROM declared_types WHERE top=? AND sub=?`, top, sub)
		if err != nil {
			http.Redirect(w, r, "/admin/types?err=Suppression+impossible", http.StatusSeeOther)
			return
		}
		if n, _ := res.RowsAffected(); n == 0 {
			http.Redirect(w, r, "/admin/types?err=Type+introuvable", http.StatusSeeOther)
			return
		}
		a.reload()
		a.auditAdmin(r, "types.supprimer", top+"|"+sub)
		http.Redirect(w, r, "/admin/types?msg=Type+supprim%C3%A9", http.StatusSeeOther)
	default:
		http.Redirect(w, r, "/admin/types?err=Action+inconnue", http.StatusSeeOther)
	}
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
		// Gestion des catégories : action explicite (créer/renommer/
		// fusionner/supprimer). Sans action : fusion legacy (formulaire
		// historique top/src/dst) pour compatibilité.
		if act := strings.TrimSpace(r.FormValue("action")); act != "" {
			a.adminTypesAction(w, r, act)
			return
		}
		// Les selects portent "top|sub" : le store se déduit de la source
		// côté serveur (aucun JS inline, doctrine CSP script-src 'self').
		// Même moteur que l'action "fusionner" (transaction + declared).
		srcTop, srcSub := splitType(r.FormValue("src"))
		dstTop, dstSub := splitType(r.FormValue("dst"))
		identique := srcTop == dstTop && srcSub == dstSub
		if srcTop != "" && srcSub != "" && dstSub != "" && !identique {
			moved, err := a.fuseTypes(srcTop, srcSub, dstSub)
			if err != nil {
				http.Redirect(w, r, "/admin/types?err=Fusion+impossible", http.StatusSeeOther)
				return
			}
			if len(moved) > 0 {
				for _, s := range moved {
					db.SyncFTS(a.DB, s)
				}
				a.reload()
				a.auditAdmin(r, "types.fusionner", srcTop+"|"+srcSub+" -> "+dstSub)
				http.Redirect(w, r, "/admin/types?msg=Types+fusionn%C3%A9s", http.StatusSeeOther)
				return
			}
			http.Redirect(w, r, "/admin/types?err=Rien+%C3%A0+fusionner+(type+source+vide)", http.StatusSeeOther)
			return
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
	seen := map[string]bool{}
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
			seen[top+"\x00"+sub] = true
			topTotal[top] += n
			nTypes++
			nProd += n
		}
	}
	// Déclarées (même vides) : en bas de leur carte, par création. Un store
	// déclaré sans type (sub='') ne fait que lister la carte vide.
	if dr, err := a.DB.Query(`SELECT top, sub FROM declared_types ORDER BY cree_le, rowid`); err == nil && dr != nil {
		for dr.Next() {
			var top, sub string
			_ = dr.Scan(&top, &sub)
			if sub == "" {
				if _, ok := byTop[top]; !ok {
					tops = append(tops, top)
					byTop[top] = []subRow{}
				}
				continue
			}
			if seen[top+"\x00"+sub] {
				continue
			}
			seen[top+"\x00"+sub] = true
			if _, ok := byTop[top]; !ok {
				tops = append(tops, top)
			}
			byTop[top] = append(byTop[top], subRow{sub, 0})
			nTypes++
		}
		dr.Close()
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
		fmt.Fprintf(&cards, `<section class="mx-card mx-type" data-types-card data-top="%s"><div class="mx-type-head"><h2>%s</h2><details class="mx-type-add" data-types-add><summary title="Ajouter une sous-catégorie">+</summary><form class="mx-form" method="post"><input type="hidden" name="csrf" value="%s"><input type="hidden" name="action" value="creer"><input type="hidden" name="top" value="%s"><input name="sub" placeholder="Nouvelle catégorie…" maxlength="80" required><button class="mx-btn primary">Créer</button></form></details></div><span class="mx-type-total">%d produit(s)</span><ul class="mx-type-list">`,
			render.Esc(top), render.Esc(lab), a.adminCSRF(), render.Esc(top), topTotal[top])
		for _, s := range byTop[top] {
			lab := s.sub
			href := fmt.Sprintf(`/categories/%s.html?type=%s`, render.Esc(top), render.Esc(s.sub))
			if lab == "" {
				lab = "Sans type"
				href = fmt.Sprintf(`/categories/%s.html`, render.Esc(top))
			}
			var dstOpts strings.Builder
			for _, o := range byTop[top] {
				if o.sub == s.sub {
					continue
				}
				olab := o.sub
				if olab == "" {
					olab = "Sans type"
				}
				fmt.Fprintf(&dstOpts, `<option value="%s|%s">%s</option>`, render.Esc(top), render.Esc(o.sub), render.Esc(olab))
			}
			fmt.Fprintf(&cards, `<li data-types-row data-top="%s" data-sub="%s"><a href="%s">%s</a><span class="vl">%d</span>`+
				`<details class="mx-type-menu" data-types-menu><summary title="Actions">⋯</summary><div class="mx-type-menu-list">`+
				`<button type="button" data-types-act="renommer">Renommer</button>`+
				`<button type="button" data-types-act="ajouter">Ajouter</button>`+
				`<details class="mx-type-fuse"><summary>Fusionner…</summary>`+
				`<form class="mx-form" method="post"><input type="hidden" name="csrf" value="%s"><input type="hidden" name="action" value="fusionner"><input type="hidden" name="src" value="%s|%s">`+
				`<label>Vers <select name="dst">%s</select></label><button class="mx-btn primary" data-mx-confirm="Fusionner « %s » vers la cible ? Irréversible.">Fusionner</button></form></details>`+
				`<a href="/admin/produits?cat=%s">Voir les produits</a>`+
				`<form class="mx-form" method="post"><input type="hidden" name="csrf" value="%s"><input type="hidden" name="action" value="supprimer"><input type="hidden" name="top" value="%s"><input type="hidden" name="sub" value="%s">`+
				`<button data-mx-confirm="Supprimer définitivement ce type ?">Supprimer</button></form>`+
				`</div></details></li>`,
				render.Esc(top), render.Esc(s.sub), href, render.Esc(lab), s.n,
				a.adminCSRF(), render.Esc(top), render.Esc(s.sub), dstOpts.String(), render.Esc(lab),
				render.Esc(top),
				a.adminCSRF(), render.Esc(top), render.Esc(s.sub))
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
<p><details class="mx-type-addtop" data-types-addtop><summary class="mx-btn">+ Nouveau store</summary><form class="mx-form" method="post"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`"><input type="hidden" name="action" value="creer_top"><input name="top" placeholder="Nom du store…" maxlength="80" required><button class="mx-btn primary">Créer</button></form></details></p>
<div class="mx-grid mx-types">`+cards.String()+`</div>`, r))
}

// AdminReglages : 4 onglets (?vue=boutique|contact|paiement|notifications),
// chacun avec son formulaire et son enregistrement. Un onglet ne touche
// jamais les clés des autres (pas d'écrasement par champ absent).
func (a *App) AdminReglages(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	vue := r.URL.Query().Get("vue")
	switch vue {
	case "contact", "paiement", "notifications":
	default:
		vue = "boutique"
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
		detail := vue
		switch vue {
		case "boutique":
			maint := "0"
			if r.FormValue("maintenance") == "1" {
				maint = "1"
			}
			set("maintenance", maint)
			set("maintenance_message", strings.TrimSpace(r.FormValue("maintenance_message")))
			detail = "maintenance=" + maint
		case "contact":
			set("whatsapp", strings.TrimSpace(r.FormValue("whatsapp")))
			detail = "whatsapp"
		case "paiement":
			payOn := func(name string) string {
				if r.FormValue(name) == "1" {
					return "1"
				}
				return "0"
			}
			set("paiement_orange_on", payOn("paiement_orange_on"))
			set("paiement_orange_numero", strings.TrimSpace(r.FormValue("paiement_orange_numero")))
			set("paiement_mtn_on", payOn("paiement_mtn_on"))
			set("paiement_mtn_numero", strings.TrimSpace(r.FormValue("paiement_mtn_numero")))
			set("paiement_wave_on", payOn("paiement_wave_on"))
			set("paiement_wave_numero", strings.TrimSpace(r.FormValue("paiement_wave_numero")))
			set("paiement_moov_on", payOn("paiement_moov_on"))
			set("paiement_moov_numero", strings.TrimSpace(r.FormValue("paiement_moov_numero")))
			set("paiement_cod_on", payOn("paiement_cod_on"))
			detail = "paiement"
		case "notifications":
			set("notif_numero", strings.TrimSpace(r.FormValue("notif_numero")))
			detail = "notif_numero"
		}

		a.auditAdmin(r, "reglages.enregistrer", detail)
		http.Redirect(w, r, "/admin/reglages?vue="+vue+"&msg=R%C3%A9glages+enregistr%C3%A9s", http.StatusSeeOther)
		return
	}

	maint := get("maintenance", "0")
	mmsg := get("maintenance_message", "Boutique en maintenance, retour très vite.")
	wa := get("whatsapp", a.Cfg.WhatsApp)

	payOrangeOn := get("paiement_orange_on", "0")
	payOrangeNum := get("paiement_orange_numero", "")
	payMtnOn := get("paiement_mtn_on", "0")
	payMtnNum := get("paiement_mtn_numero", "")
	payWaveOn := get("paiement_wave_on", "0")
	payWaveNum := get("paiement_wave_numero", "")
	payMoovOn := get("paiement_moov_on", "0")
	payMoovNum := get("paiement_moov_numero", "")
	payCodOn := get("paiement_cod_on", "0")

	notifNumero := get("notif_numero", wa)
	stockReportLink := a.buildStockReportWaLink(notifNumero)
	stockReportTg := a.buildStockReportTgLink()

	status := `<div class="mx-notice" style="margin-bottom:16px">Boutique <strong>en ligne</strong> — les visiteurs voient le catalogue normalement.</div>`
	if maint == "1" {
		status = `<div class="mx-notice warn" style="margin-bottom:16px">Mode maintenance <strong>actif</strong> — la boutique affiche un message aux visiteurs, /admin reste accessible.</div>`
	}

	var b strings.Builder
	b.WriteString(reglagesTabs(vue))
	formFor := func(inner string) string {
		return `<form method="post" action="/admin/reglages?vue=` + vue + `"><input type="hidden" name="csrf" value="` + a.adminCSRF() + `">` + inner +
			`<div class="mx-form-actions"><button class="mx-btn primary">Enregistrer</button></div></form>`
	}
	switch vue {
	case "contact":
		b.WriteString(formFor(reglagesSectionContact(wa)))
	case "paiement":
		b.WriteString(formFor(reglagesSectionPaiement(payOrangeOn, payOrangeNum, payMtnOn, payMtnNum, payWaveOn, payWaveNum, payMoovOn, payMoovNum, payCodOn)))
	case "notifications":
		b.WriteString(reglagesSectionNotifications(notifNumero, stockReportLink, stockReportTg, a.adminCSRF(), vue))
	default:
		b.WriteString(status)
		b.WriteString(formFor(reglagesSectionBoutique(maint, mmsg)))
	}

	a.write(w, 200, a.adminPage("Réglages", "reglages", b.String(), r))
}

// reglagesTabs : onglets des 4 blocs (même style que mx-views, sans compteurs).
func reglagesTabs(cur string) string {
	tabs := []struct{ key, label string }{
		{"boutique", "Boutique"},
		{"contact", "Contact"},
		{"paiement", "Paiement"},
		{"notifications", "Notifications"},
	}
	var b strings.Builder
	b.WriteString(`<ul class="mx-views">`)
	for i, t := range tabs {
		if i > 0 {
			b.WriteString(`<li aria-hidden="true">|</li>`)
		}
		cls := ""
		if t.key == cur {
			cls = ` class="on"`
		}
		fmt.Fprintf(&b, `<li><a href="/admin/reglages?vue=%s"%s>%s</a></li>`, t.key, cls, t.label)
	}
	return b.String() + `</ul>`
}

// reglagesSectionBoutique : bloc "Boutique" (mode maintenance + message).
func reglagesSectionBoutique(maint, mmsg string) string {
	chk := ""
	if maint == "1" {
		chk = " checked"
	}
	return `<div class="mx-fsec mx-fsec-premium" style="max-width:640px">` + reglagesHead(mxIcoShop, "Boutique", "Visibilité de la boutique et message affiché aux visiteurs.", "") + `
<label class="mx-switch-row"><span class="mx-switch-tx"><strong>Mode maintenance</strong><span class="hint">La boutique affiche un message, /admin reste accessible.</span></span><span class="mx-switch"><input type="checkbox" name="maintenance" value="1"` + chk + `><span class="mx-switch-sl"></span></span></label>
<div class="mx-fgrid"><label class="span2">Message de maintenance <input name="maintenance_message" value="` + render.Esc(mmsg) + `"></label></div>
</div>`
}

// reglagesSectionContact : bloc "Contact" (WhatsApp).
func reglagesSectionContact(wa string) string {
	return `<div class="mx-fsec mx-fsec-premium" style="max-width:640px">` + reglagesHead(mxIcoChat, "Contact", "Numéro WhatsApp affiché aux clients et utilisé pour les commandes.", "") + `
<div class="mx-fgrid"><label class="span2">WhatsApp <input name="whatsapp" value="` + render.Esc(wa) + `"></label></div>
<p class="hint">Appliqué au redémarrage du conteneur.</p>
</div>`
}

// escQForWa : réplique l'échappement utilisé par render.qesc pour les liens
// wa.me (espaces en %20, slash non échappé), afin de rester cohérent avec
// les liens WhatsApp déjà générés ailleurs sur le site.
func escQForWa(s string) string {
	q := url.QueryEscape(s)
	q = strings.ReplaceAll(q, "+", "%20")
	return strings.ReplaceAll(q, "%2F", "/")
}

// buildStockReportText : message listant les produits en rupture et en
// stock faible (≤3). Le canal (WhatsApp / Telegram) est ajouté par l'appelant.
func (a *App) buildStockReportText() string {
	var b strings.Builder
	b.WriteString("📋 Rapport stock — Marexstore\n")

	rrows, _ := a.DB.Query(`SELECT nom FROM products WHERE dispo='rupture de stock' OR (stock IS NOT NULL AND stock<=0) ORDER BY nom LIMIT 30`)
	var ruptures []string
	if rrows != nil {
		defer rrows.Close()
		for rrows.Next() {
			var nom string
			_ = rrows.Scan(&nom)
			ruptures = append(ruptures, nom)
		}
	}
	b.WriteString(fmt.Sprintf("\n🔴 Rupture (%d) :\n", len(ruptures)))
	if len(ruptures) == 0 {
		b.WriteString("Aucune\n")
	}
	for _, nom := range ruptures {
		b.WriteString("- " + nom + "\n")
	}

	frows, _ := a.DB.Query(`SELECT nom, stock FROM products WHERE stock IS NOT NULL AND stock>0 AND stock<=3 ORDER BY stock, nom LIMIT 30`)
	type faible struct {
		nom   string
		stock int
	}
	var faibles []faible
	if frows != nil {
		defer frows.Close()
		for frows.Next() {
			var nom string
			var st int
			_ = frows.Scan(&nom, &st)
			faibles = append(faibles, faible{nom, st})
		}
	}
	b.WriteString(fmt.Sprintf("\n🟡 Stock faible (%d) :\n", len(faibles)))
	if len(faibles) == 0 {
		b.WriteString("Aucun\n")
	}
	for _, f := range faibles {
		b.WriteString(fmt.Sprintf("- %s (%d)\n", f.nom, f.stock))
	}
	return b.String()
}

// buildStockReportWaLink : lien wa.me prérempli avec le rapport de stock.
func (a *App) buildStockReportWaLink(numero string) string {
	// numero vient d'un champ admin : échappé pour l'attribut href.
	return "https://wa.me/" + render.Esc(strings.TrimSpace(numero)) + "?text=" + escQForWa(a.buildStockReportText())
}

// buildStockReportTgLink : même rapport via le partage Telegram (sans bot,
// sans numéro : le texte est prérempli, l'utilisateur choisit le destinataire).
func (a *App) buildStockReportTgLink() string {
	return "https://t.me/share/url?url=&text=" + escQForWa(a.buildStockReportText())
}

// reglagesSectionNotifications : bloc "Notifications" (rapport stock bas),
// avec son propre formulaire (le champ doit être envoyé, sinon le numéro
// ne persiste jamais).
func reglagesSectionNotifications(numero, waLink, tgLink, csrf, vue string) string {
	return `<form method="post" action="/admin/reglages?vue=` + vue + `"><input type="hidden" name="csrf" value="` + csrf + `">` +
		`<div class="mx-fsec mx-fsec-premium" style="max-width:640px">` + reglagesHead(mxIcoBell, "Notifications", "Alertes de stock envoyées en un clic, sur WhatsApp ou Telegram.", "") + `
<div class="mx-fgrid"><label class="span2">Numéro destinataire des alertes <input name="notif_numero" value="` + render.Esc(numero) + `"></label></div>
<p class="hint">Par défaut, le numéro WhatsApp de Contact. Modifie-le ici pour envoyer les alertes WhatsApp à un autre numéro. Telegram n'a besoin d'aucun numéro : le texte est prérempli, tu choisis le destinataire.</p>
<p class="hint">Génère un message listant les produits en rupture et en stock faible (≤3), à envoyer manuellement.</p>
<div class="mx-form-actions"><button class="mx-btn primary">Enregistrer</button>
<a class="mx-btn ghost" target="_blank" rel="noopener" href="` + waLink + `"><img class="mx-btn-logo" src="/static/logos/partage-whatsapp.png" alt="WhatsApp"> WhatsApp</a>
<a class="mx-btn ghost" target="_blank" rel="noopener" href="` + tgLink + `"><img class="mx-btn-logo" src="/static/logos/partage-telegram.jpg" alt="Telegram"> Telegram</a></div>
</div></form>`
}

// reglagesHead : en-tête de section premium (icône + titre + chapeau).
// `extra` (ex. pastille de compteur) s'affiche en bout de ligne.
func reglagesHead(icon, title, lead, extra string) string {
	return `<div class="mx-sec-head"><span class="mx-sec-ic">` + icon + `</span><div><h3>` + title + `</h3><p>` + lead + `</p></div>` + extra + `</div>`
}

const (
	mxIcoShop = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M3 9l9-7 9 7v11a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/><path d="M9 22V12h6v10"/></svg>`
	mxIcoChat = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M21 15a2 2 0 0 1-2 2H7l-4 4V5a2 2 0 0 1 2-2h14a2 2 0 0 1 2 2z"/></svg>`
	mxIcoCard = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="4" width="22" height="16" rx="2"/><line x1="1" y1="10" x2="23" y2="10"/></svg>`
	mxIcoBell = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 8A6 6 0 0 0 6 8c0 7-3 9-3 9h18s-3-2-3-9"/><path d="M13.7 21a2 2 0 0 1-3.4 0"/></svg>`
	mxIcoCash = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><rect x="1" y="4" width="22" height="16" rx="2"/><circle cx="12" cy="12" r="3"/><path d="M6 12h.01M18 12h.01"/></svg>`
)

// reglagesSectionPaiement : cartes fournisseurs avec logos (grille 2
// colonnes, interrupteur conservé : même markup, même nom de champ).
func reglagesSectionPaiement(orangeOn, orangeNum, mtnOn, mtnNum, waveOn, waveNum, moovOn, moovNum, codOn string) string {
	actifs := 0
	for _, on := range []string{orangeOn, mtnOn, waveOn, moovOn, codOn} {
		if on == "1" {
			actifs++
		}
	}
	card := func(name, label, hint, numVal, on, logo string, showNumero, wide bool) string {
		chk := ""
		if on == "1" {
			chk = " checked"
		}
		num := ""
		if showNumero {
			num = `<label class="mx-pay-num">Numéro marchand <input name="` + name + `_numero" value="` + render.Esc(numVal) + `" placeholder="Ex. 0700000000" inputmode="tel" autocomplete="off"></label>`
		}
		wideCls := ""
		if wide {
			wideCls = " wide"
		}
		return `<div class="mx-pay-card` + wideCls + `"><span class="mx-pay-logo">` + logo + `</span>` +
			`<div class="mx-pay-tx"><strong>` + label + `</strong><span>` + hint + `</span></div>` +
			`<span class="mx-switch"><input type="checkbox" name="` + name + `_on" value="1"` + chk + `><span class="mx-switch-sl"></span></span>` +
			num + `</div>`
	}
	img := func(src, alt string) string {
		return `<img src="/static/logos/` + src + `" alt="` + alt + `" loading="lazy">`
	}
	return `<div class="mx-fsec mx-fsec-premium" style="max-width:720px">` +
		reglagesHead(mxIcoCard, "Paiement", "Moyens acceptés sur la boutique. Le client choisit au panier ; le marchand reçoit sur le numéro indiqué.", `<span class="mx-pill">`+fmt.Sprint(actifs)+` actif(s)</span>`) +
		`<div class="mx-pay-grid">` +
		card("paiement_orange", "Orange Money", "Paiement via Orange Money", orangeNum, orangeOn, img("paiement-orange.png", "Orange Money"), true, false) +
		card("paiement_mtn", "MTN Mobile Money", "Paiement via MTN MoMo", mtnNum, mtnOn, img("paiement-mtn.jpg", "MTN MoMo"), true, false) +
		card("paiement_wave", "Wave", "Paiement via Wave", waveNum, waveOn, img("paiement-wave.jpg", "Wave"), true, false) +
		card("paiement_moov", "Moov Money", "Paiement via Moov Money", moovNum, moovOn, img("paiement-moov.png", "Moov Money"), true, false) +
		card("paiement_cod", "Paiement à la livraison", "Le client paie en espèces à la réception", "", codOn, mxIcoCash, false, true) +
		`</div></div>`
}

// mergeGallery : fusionne les extras existants (moins les suppressions
// cochées) avec les nouveaux uploads : ordre conservé, sans doublon ni
// vide, plafond 4 extras (5 photos avec la principale).
func mergeGallery(kept []string, deleted map[string]bool, uploaded []string) []string {
	out := []string{}
	seen := map[string]bool{}
	add := func(s string) {
		if s == "" || seen[s] {
			return
		}
		seen[s] = true
		out = append(out, s)
	}
	for _, s := range kept {
		if !deleted[s] {
			add(s)
		}
	}
	for _, s := range uploaded {
		add(s)
	}
	if len(out) > 4 {
		out = out[:4]
	}
	return out
}

// saveGallery : applique les suppressions cochées (cases del_img) et les
// nouveaux uploads (input multiple "photos"), persiste images_json et
// supprime du disque les fichiers retirés (best-effort). Retourne un
// message d'erreur affichable, ou "" si tout est OK.
func (a *App) saveGallery(r *http.Request, slug string, existing []string) string {
	del := map[string]bool{}
	for _, d := range r.Form["del_img"] {
		if d = strings.TrimSpace(d); d != "" {
			del[d] = true
		}
	}
	var kept []string
	for _, s := range existing {
		if !del[s] {
			kept = append(kept, s)
		}
	}
	up, err := a.saveUploads(r, slug, kept, 4)
	if err != nil {
		return err.Error()
	}
	// Plafond atteint : les fichiers surnuméraires sont écartés sans
	// erreur — on le dit honnêtement au lieu de les jeter en silence.
	avert := ""
	if r.MultipartForm != nil {
		valides := 0
		for _, h := range r.MultipartForm.File["photos"] {
			if h != nil && strings.TrimSpace(h.Filename) != "" {
				valides++
			}
		}
		if reste := valides - len(up); reste > 0 {
			avert = fmt.Sprintf("%d photo(s) ignorée(s), 4 maximum", reste)
		}
	}
	merged := mergeGallery(kept, map[string]bool{}, up)
	b, _ := json.Marshal(merged)
	if _, err := a.DB.Exec(`UPDATE products SET images_json=? WHERE slug=?`, string(b), slug); err != nil {
		return "enregistrement galerie impossible"
	}
	inMerged := map[string]bool{}
	for _, s := range merged {
		inMerged[s] = true
	}
	for _, s := range existing {
		if del[s] && !inMerged[s] && strings.HasPrefix(s, "images/") && !strings.Contains(s, "..") {
			os.Remove(s) // best-effort
		}
	}
	return avert
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

// saveUploads : photos supplémentaires de la galerie (input multiple
// "photos", JPG/PNG/WebP ≤ 5 Mo chacun). Noms stables
// images/<slug>-<k>.<ext>, k = premier créneau libre 2..5 (jamais de
// renommage, jamais d'orphelin : on s'arrête à maxTotal extras).
func (a *App) saveUploads(r *http.Request, slug string, kept []string, maxTotal int) ([]string, error) {
	if r.MultipartForm == nil {
		if err := r.ParseMultipartForm(32 << 20); err != nil && err != http.ErrNotMultipart {
			return nil, err
		}
	}
	if r.MultipartForm == nil {
		return nil, nil
	}
	used := map[string]bool{}
	slotUsed := map[int]bool{}
	for _, s := range kept {
		used[s] = true
		if stem := strings.TrimSuffix(filepath.Base(s), filepath.Ext(s)); strings.HasPrefix(stem, slug+"-") {
			if k, err := strconv.Atoi(strings.TrimPrefix(stem, slug+"-")); err == nil {
				slotUsed[k] = true
			}
		}
	}
	var out []string
	for _, h := range r.MultipartForm.File["photos"] {
		if h == nil || h.Filename == "" {
			continue
		}
		if len(kept)+len(out) >= maxTotal {
			break
		}
		if h.Size > 5<<20 {
			return nil, fmt.Errorf("fichier trop lourd (max 5 Mo) : %s", h.Filename)
		}
		ext := strings.ToLower(filepath.Ext(h.Filename))
		switch ext {
		case ".jpg", ".jpeg", ".png", ".webp":
			if ext == ".jpeg" {
				ext = ".jpg"
			}
		default:
			return nil, fmt.Errorf("format refusé (JPG/PNG/WebP uniquement) : %s", h.Filename)
		}
		rel := ""
		for k := 2; k <= 5; k++ {
			if slotUsed[k] {
				continue
			}
			rel = fmt.Sprintf("images/%s-%d%s", slug, k, ext)
			slotUsed[k] = true
			break
		}
		if rel == "" {
			break
		}
		src, err := h.Open()
		if err != nil {
			return nil, err
		}
		dst, err := os.Create(rel)
		if err != nil {
			src.Close()
			return nil, err
		}
		_, err = io.Copy(dst, src)
		dst.Close()
		src.Close()
		if err != nil {
			os.Remove(rel)
			return nil, err
		}
		used[rel] = true
		out = append(out, rel)
	}
	return out, nil
}
