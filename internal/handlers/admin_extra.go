package handlers

import (
	"database/sql"
	"fmt"
	"github.com/go-chi/chi/v5"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

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
		tx, _ := a.DB.Begin()
		if _, err := tx.Exec(`UPDATE orders SET statut=? WHERE id=?`, vers, id); err != nil {
			_ = tx.Rollback()
			http.Error(w, "impossible", 500)
			return
		}
		if _, err := tx.Exec(`INSERT INTO order_events(order_id,de_statut,vers_statut,par) VALUES(?,?,?,?)`, id, st, vers, "admin"); err != nil {
			_ = tx.Rollback()
			http.Error(w, "impossible", 500)
			return
		}
		if vers == "paye" {
			// Décrémente le stock des lignes (plancher 0, jamais négatif).
			rows, _ := tx.Query(`SELECT product_id,qty FROM order_items WHERE order_id=?`, id)
			if rows != nil {
				for rows.Next() {
					var pid, q int64
					_ = rows.Scan(&pid, &q)
					_, _ = tx.Exec(`UPDATE products SET stock=CASE WHEN stock>=? THEN stock-? ELSE 0 END WHERE id=? AND stock IS NOT NULL`, q, q, pid)
				}
				rows.Close()
			}
		}
		_ = tx.Commit()
		a.reload()
		http.Redirect(w, r, "/admin/commandes/"+strconv.FormatInt(id, 10)+"?msg=Commande+enregistr%C3%A9e", http.StatusSeeOther)
		return
	}
	t := "devis"
	if tot.Valid {
		t = fmt.Sprint(tot.Int64)
	}
	// Lignes.
	var items strings.Builder
	ir, _ := a.DB.Query(`SELECT p.nom,p.slug,oi.qty,oi.prix_unitaire FROM order_items oi JOIN products p ON p.id=oi.product_id WHERE oi.order_id=?`, id)
	if ir != nil {
		defer ir.Close()
		for ir.Next() {
			var nom, slug string
			var q int64
			var pu sql.NullInt64
			_ = ir.Scan(&nom, &slug, &q, &pu)
			pus := "—"
			if pu.Valid {
				pus = fmt.Sprint(pu.Int64)
			}
			fmt.Fprintf(&items, `<tr><td><a href="/admin/produits/%s">%s</a></td><td>%d</td><td>%s</td></tr>`, slug, render.Esc(nom), q, pus)
		}
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
		`<div class="mx-cols"><div>
<div class="mx-postbox"><h2>Lignes</h2><div class="in"><div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>Qté</th><th>Prix unit.</th></tr>`+items.String()+`</table></div>
<p>Total : <strong>`+t+` F CFA</strong> · Moyen : `+render.Esc(moy)+` · Client : `+render.Esc(client)+` · Compte : `+cli+`</p></div></div>
<div class="mx-postbox"><h2>Historique</h2><div class="in"><ul class="mx-timeline">`+tl.String()+`</ul></div></div>
</div><div><div class="mx-postbox mx-pub"><h2>Statut</h2><div class="in"><p>Actuel : `+statutBadge(st)+`</p>`+btns.String()+`</div></div></div></div>`, r))
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
	rows, _ := a.DB.Query(`SELECT u.id,u.email,u.nom,u.tel,u.created_at,(SELECT COUNT(*) FROM orders WHERE user_id=u.id) FROM users u WHERE u.email LIKE ? OR u.nom LIKE ? ORDER BY u.id DESC LIMIT ? OFFSET ?`, "%"+q+"%", "%"+q+"%", per, (page-1)*per)
	var tr strings.Builder
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id, nb int64
			var em, nm, tel, at string
			_ = rows.Scan(&id, &em, &nm, &tel, &at, &nb)
			fmt.Fprintf(&tr, `<tr><td><a href="/admin/clients/%d"><strong>%s</strong></a><div class="row-actions"><a href="/admin/clients/%d">Détail</a></div></td><td>%s</td><td>%s</td><td>%s</td><td>%d commande(s)</td></tr>`,
				id, render.Esc(em), id, render.Esc(nm), render.Esc(tel), render.Esc(at), nb)
		}
	}
	a.write(w, 200, a.adminPage("Clients", "clients",
		`<form class="mx-tools" method="get" action="/admin/clients"><input type="search" name="q" value="`+render.Esc(q)+`" placeholder="Rechercher e-mail, nom…"><button class="mx-btn">OK</button></form>
<div class="mx-tablewrap"><table class="mx-table"><tr><th>E-mail</th><th>Nom</th><th>Tél</th><th>Inscription</th><th>Commandes</th></tr>`+tr.String()+`</table>`+render.MxPager("/admin/clients?q="+q, page, pages)+`</div>`, r))
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
	var tr strings.Builder
	rows, _ := a.DB.Query(`SELECT id,statut,total,created_at FROM orders WHERE user_id=? ORDER BY id DESC`, id)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var oid int64
			var st, at2 string
			var tot sql.NullInt64
			_ = rows.Scan(&oid, &st, &tot, &at2)
			t := "devis"
			if tot.Valid {
				t = fmt.Sprint(tot.Int64)
			}
			fmt.Fprintf(&tr, `<tr><td><a href="/admin/commandes/%d">#%d</a></td><td>%s</td><td>%s</td><td>%s</td></tr>`, oid, oid, render.Esc(at2), t, statutBadge(st))
		}
	}
	a.write(w, 200, a.adminPage("Client — "+nm, "clients",
		`<div class="mx-cols"><div><div class="mx-postbox"><h2>Commandes</h2><div class="in"><div class="mx-tablewrap"><table class="mx-table"><tr><th>#</th><th>Date</th><th>Total</th><th>Statut</th></tr>`+tr.String()+`</table></div></div></div>
</div><div><div class="mx-postbox mx-pub"><h2>Compte</h2><div class="in"><p><strong>`+render.Esc(nm)+`</strong><br>`+render.Esc(em)+`<br>`+render.Esc(tel)+`<br><span class="mx-badge info">inscrit le `+render.Esc(at)+`</span></p></div></div></div></div>`, r))
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
	for _, e := range entries {
		if e.IsDir() {
			continue
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
	views := render.MxViews([]render.MxView{
		{Key: "all", Href: "/admin/medias?vue=all", Label: "Toutes", Count: nAll},
		{Key: "orphelins", Href: "/admin/medias?vue=orphelins", Label: "Orphelins", Count: nOrph},
	}, vue)
	a.write(w, 200, a.adminPage("Médias", "medias", views+`<div class="mx-media">`+grid+`</div>`, r))
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
	var tr strings.Builder
	byTop := map[string][]string{}
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var top, sub string
			var n int
			_ = rows.Scan(&top, &sub, &n)
			byTop[top] = append(byTop[top], sub)
			fmt.Fprintf(&tr, `<tr><td>%s</td><td>%s</td><td>%d</td><td><a class="linklike" href="/categories/%s.html?type=%s">Voir</a></td></tr>`,
				render.Esc(top), render.Esc(sub), n, render.Esc(top), render.Esc(sub))
		}
	}
	var opts strings.Builder
	for top, subs := range byTop {
		for _, s := range subs {
			fmt.Fprintf(&opts, `<option value="%s|%s">%s › %s</option>`, render.Esc(top), render.Esc(s), render.Esc(top), render.Esc(s))
		}
	}
	a.write(w, 200, a.adminPage("Types d'équipements", "types",
		`<div class="mx-tablewrap"><table class="mx-table"><tr><th>Store</th><th>Type</th><th>Produits</th><th></th></tr>`+tr.String()+`</table></div>
<br><div class="mx-postbox" style="max-width:640px"><h2>Fusionner deux types</h2><div class="in"><form class="mx-form" method="post"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<label>De (sera vidé) <select name="src">`+opts.String()+`</select></label>
<label>Vers <select name="dst">`+opts.String()+`</select></label>
<input type="hidden" name="top" id="mx-top"><button class="mx-btn">Fusionner</button>
<p class="hint">Le store est déduit du type source (« top|sub »). Irréversible : les produits du type source rejoignent le type cible.</p>
</form></div></div>
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
		http.Redirect(w, r, "/admin/reglages?msg=R%C3%A9glages+enregistr%C3%A9s", http.StatusSeeOther)
		return
	}
	maint := get("maintenance", "0")
	mmsg := get("maintenance_message", "Boutique en maintenance, retour très vite.")
	wa := get("whatsapp", a.Cfg.WhatsApp)
	chk := ""
	if maint == "1" {
		chk = " checked"
	}
	a.write(w, 200, a.adminPage("Réglages", "reglages",
		`<div class="mx-postbox" style="max-width:640px"><h2>Boutique</h2><div class="in"><form class="mx-form" method="post"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<label><input type="checkbox" name="maintenance" value="1"`+chk+`> Mode maintenance (la boutique affiche un message, /admin reste accessible)</label>
<label>Message de maintenance <input name="maintenance_message" value="`+render.Esc(mmsg)+`"></label>
<label>WhatsApp (appliqué au redémarrage du conteneur) <input name="whatsapp" value="`+render.Esc(wa)+`"></label>
<button class="mx-btn">Enregistrer</button></form></div></div>`, r))
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
