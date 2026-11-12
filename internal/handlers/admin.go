package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
	"github.com/go-chi/chi/v5"
)

// /admin dans le style du site (même CSS : .wrap, .sec, table.specs, .btn).
func (a *App) adminAuth(w http.ResponseWriter, r *http.Request) bool {
	user, pass, ok := r.BasicAuth()
	if !ok || user != a.Cfg.AdminUser || a.Cfg.AdminPassHash == "" ||
		!auth.CheckPassword(a.Cfg.AdminPassHash, pass) {
		w.Header().Set("WWW-Authenticate", `Basic realm="admin"`)
		http.Error(w, "Non autorisé.", http.StatusUnauthorized)
		return false
	}
	return true
}

func (a *App) adminPage(title, inner string) string {
	body := render.Header(a.Store, "", "") + `<div class="wrap">
<h1 style="font-size:24px">` + title + `</h1>
<p><a class="btn ghost" href="/admin">← Tableau de bord</a> <a class="btn ghost" href="/admin/produits">Produits</a> <a class="btn ghost" href="/admin/commandes">Commandes</a></p>
` + inner + `</div>` + render.Footer(a.Store, "")
	s := render.Doc("Admin — "+title, body, "", "")
	return strings.ReplaceAll(s, "{{CSRF}}", "")
}

func (a *App) Admin(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	var nProd, nUsers, nOrders int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&nProd)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&nUsers)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&nOrders)
	a.write(w, 200, a.adminPage("Tableau de bord",
		fmt.Sprintf(`<div class="sec"><div class="widgets"><div><h3>Produits</h3><div class="pr">%d</div></div><div><h3>Comptes</h3><div class="pr">%d</div></div><div><h3>Commandes</h3><div class="pr">%d</div></div></div></div>`, nProd, nUsers, nOrders)))
}

func (a *App) AdminProduits(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	rows, err := a.DB.Query(`SELECT slug,nom,prix,dispo,stock,live_top FROM products WHERE nom LIKE ? OR slug LIKE ? ORDER BY nom LIMIT 200`, "%"+q+"%", "%"+q+"%")
	if err != nil {
		http.Error(w, "indisponible", 500)
		return
	}
	defer rows.Close()
	var tr strings.Builder
	for rows.Next() {
		var slug, nom, dispo, top string
		var prix, stock sql.NullInt64
		_ = rows.Scan(&slug, &nom, &prix, &dispo, &stock, &top)
		pp := "devis"
		if prix.Valid {
			pp = fmt.Sprint(prix.Int64)
		}
		fmt.Fprintf(&tr, `<tr><td><a href="/produits/%s.html">%s</a></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td><a class="linklike" href="/admin/produits/%s">Modifier</a></td></tr>`,
			slug, render.Esc(nom), pp, dispo, nullStr(stock), top, slug)
	}
	a.write(w, 200, a.adminPage("Produits",
		`<form method="get" action="/admin/produits"><input type="search" name="q" value="`+render.Esc(q)+`" placeholder="Rechercher…"> <button class="btn">OK</button></form>
<table class="specs"><tr><td><strong>Nom</strong></td><td><strong>Prix</strong></td><td><strong>Dispo</strong></td><td><strong>Stock</strong></td><td><strong>Store</strong></td><td></td></tr>`+tr.String()+`</table>`))
}

func nullStr(v sql.NullInt64) string {
	if v.Valid {
		return fmt.Sprint(v.Int64)
	}
	return "—"
}

func (a *App) reload() {
	if st, err := site.LoadStore(a.DB); err == nil {
		a.Store = st
	}
}

func (a *App) AdminProduitEdit(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	slug := chi.URLParam(r, "slug")
	var p catalog.RawProduct
	var prix, pb, pa, stock sql.NullInt64
	err := a.DB.QueryRow(`SELECT slug,nom,marque,prix,prix_boutique,prix_avant,dispo,accroche,description,image,ref_fabricant,stock,main,sub,live_top,live_sub
		FROM products WHERE slug=?`, slug).Scan(&p.Slug, &p.Nom, &p.Marque, &prix, &pb, &pa, &p.Dispo, &p.Accroche, &p.Description, &p.Image, &p.RefFabricant, &stock, &p.Main, &p.Sub, &p.LiveTop, &p.LiveSub)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if r.Method == http.MethodPost {
		nn := func(k string) string { return strings.TrimSpace(r.FormValue(k)) }
		ni := func(k string) any {
			v := strings.TrimSpace(r.FormValue(k))
			if v == "" {
				return nil
			}
			var n int64
			fmt.Sscan(v, &n)
			return n
		}
		_, err = a.DB.Exec(`UPDATE products SET nom=?,marque=?,prix=?,prix_boutique=?,prix_avant=?,dispo=?,accroche=?,description=?,image=?,stock=?,main=?,sub=?,live_top=?,live_sub=? WHERE slug=?`,
			nn("nom"), nn("marque"), ni("prix"), ni("prix_boutique"), ni("prix_avant"), nn("dispo"), nn("accroche"), nn("description"), nn("image"), ni("stock"), nn("main"), nn("sub"), nn("live_top"), nn("live_sub"), slug)
		if err != nil {
			http.Error(w, "enregistrement impossible", 500)
			return
		}
		a.reload()
		http.Redirect(w, r, "/admin/produits/"+slug, http.StatusSeeOther)
		return
	}
	str := func(v sql.NullInt64) string {
		if v.Valid {
			return fmt.Sprint(v.Int64)
		}
		return ""
	}
	a.write(w, 200, a.adminPage("Modifier — "+p.Nom,
		`<form class="acct-form" method="post">
<label class="full">Nom * <input name="nom" required value="`+render.Esc(p.Nom)+`"></label>
<label>Marque <input name="marque" value="`+render.Esc(p.Marque)+`"></label>
<label>Prix (vide = devis) <input name="prix" inputmode="numeric" value="`+str(prix)+`"></label>
<label>Prix boutique <input name="prix_boutique" inputmode="numeric" value="`+str(pb)+`"></label>
<label>Prix avant <input name="prix_avant" inputmode="numeric" value="`+str(pa)+`"></label>
<label>Dispo <select name="dispo"><option>en stock</option><option>rupture de stock</option><option>retour</option></select></label>
<label>Stock <input name="stock" inputmode="numeric" value="`+str(stock)+`"></label>
<label>Image <input name="image" value="`+render.Esc(p.Image)+`"></label>
<label class="full">Accroche <input name="accroche" value="`+render.Esc(p.Accroche)+`"></label>
<label class="full">Description <textarea name="description" rows="4">`+render.Esc(p.Description)+`</textarea></label>
<label>Catégorie <input name="main" value="`+render.Esc(p.Main)+`"></label>
<label>Sous-catégorie <input name="sub" value="`+render.Esc(p.Sub)+`"></label>
<label>Store (live_top) <input name="live_top" value="`+render.Esc(p.LiveTop)+`"></label>
<label>Sous-store (live_sub) <input name="live_sub" value="`+render.Esc(p.LiveSub)+`"></label>
<div class="full"><button class="btn">Enregistrer</button> <a class="btn ghost" href="/produits/`+p.Slug+`.html">Voir la fiche</a></div>
</form>`))
}

func (a *App) AdminCommandes(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodPost {
		_, _ = a.DB.Exec(`UPDATE orders SET statut=? WHERE id=?`, r.FormValue("statut"), r.FormValue("id"))
		http.Redirect(w, r, "/admin/commandes", http.StatusSeeOther)
		return
	}
	rows, _ := a.DB.Query(`SELECT id,statut,moyen_paiement,total,client_json,created_at FROM orders ORDER BY id DESC LIMIT 100`)
	var tr strings.Builder
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var statut, moyen, client, at string
			var total sql.NullInt64
			_ = rows.Scan(&id, &statut, &moyen, &total, &client, &at)
			t := "devis"
			if total.Valid {
				t = fmt.Sprint(total.Int64)
			}
			fmt.Fprintf(&tr, `<tr><td>#%d</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td><form method="post" action="/admin/commandes"><input type="hidden" name="id" value="%d"><select name="statut"><option>devis</option><option>whatsapp_recu</option><option>attente_momo</option><option>paye</option><option>prepare</option><option>livre</option><option>annule</option></select> <button class="btn">OK</button></form></td></tr>`,
				id, render.Esc(at), render.Esc(moyen), t, render.Esc(client), id)
		}
	}
	a.write(w, 200, a.adminPage("Commandes",
		`<table class="specs"><tr><td><strong>#</strong></td><td><strong>Date</strong></td><td><strong>Moyen</strong></td><td><strong>Total</strong></td><td><strong>Client</strong></td><td><strong>Statut</strong></td></tr>`+tr.String()+`</table>`))
}
