package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
	"github.com/go-chi/chi/v5"
)

// /admin : coquille dédiée (style WordPress), même langue visuelle que la boutique.
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

// CSRF admin : token dérivé du secret serveur + compte admin (l'admin
// Basic Auth n'a pas de cookie de session, donc pas de sessionID).
func (a *App) adminCSRF() string {
	return auth.CSRFToken(a.Cfg.SessionSecret, "admin:"+a.Cfg.AdminUser)
}

func (a *App) checkAdminCSRF(r *http.Request) bool {
	return auth.CheckCSRF(a.Cfg.SessionSecret, "admin:"+a.Cfg.AdminUser, r.FormValue("csrf"))
}

func (a *App) counts() map[string]int {
	c := map[string]int{}
	var p, u, o int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&p)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&u)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&o)
	c["produits"], c["clients"], c["commandes"] = p, u, o
	return c
}

func (a *App) adminPage(title, active, inner string, r *http.Request) string {
	msg := ""
	if r != nil {
		msg = render.MxNotice(r.URL.Query().Get("msg"), r.URL.Query().Get("err"))
	}
	return render.MxDoc(title, active, a.counts(), msg+inner)
}

func (a *App) Admin(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	var nProd, nUsers, nOrders int
	var ca sql.NullInt64
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&nProd)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&nUsers)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&nOrders)
	_ = a.DB.QueryRow(`SELECT SUM(total) FROM orders WHERE statut IN ('paye','prepare','livre')`).Scan(&ca)
	var rupt int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE dispo='rupture de stock' OR (stock IS NOT NULL AND stock<=0)`).Scan(&rupt)
	caStr := "—"
	if ca.Valid {
		caStr = fmt.Sprint(ca.Int64)
	}
	// Dernières commandes.
	var last strings.Builder
	rows, _ := a.DB.Query(`SELECT id,statut,moyen_paiement,total,created_at FROM orders ORDER BY id DESC LIMIT 5`)
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var st, moy, at string
			var tot sql.NullInt64
			_ = rows.Scan(&id, &st, &moy, &tot, &at)
			t := "devis"
			if tot.Valid {
				t = fmt.Sprint(tot.Int64)
			}
			fmt.Fprintf(&last, `<tr><td><a href="/admin/commandes/%d">#%d</a></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				id, id, render.Esc(at), render.Esc(moy), t, statutBadge(st))
		}
	}
	// Produits en rupture (top 5).
	var low strings.Builder
	lrows, _ := a.DB.Query(`SELECT slug,nom FROM products WHERE dispo='rupture de stock' OR (stock IS NOT NULL AND stock<=0) ORDER BY nom LIMIT 5`)
	if lrows != nil {
		defer lrows.Close()
		for lrows.Next() {
			var slug, nom string
			_ = lrows.Scan(&slug, &nom)
			fmt.Fprintf(&low, `<tr><td><a href="/admin/produits/%s">%s</a></td><td>%s</td></tr>`, slug, render.Esc(nom), statutBadge("rupture"))
		}
	}
	a.write(w, 200, a.adminPage("Tableau de bord", "", `<div class="mx-grid kpi">
<div class="mx-card"><h3>Chiffre encaissé</h3><div class="big">`+caStr+`</div><div class="sub">F CFA (payé/préparé/livré)</div></div>
<div class="mx-card"><h3>Commandes</h3><div class="big">`+fmt.Sprint(nOrders)+`</div><div class="sub"><a href="/admin/commandes">Tout voir →</a></div></div>
<div class="mx-card"><h3>Produits</h3><div class="big">`+fmt.Sprint(nProd)+`</div><div class="sub"><a href="/admin/produits">Gérer →</a></div></div>
<div class="mx-card"><h3>Ruptures</h3><div class="big">`+fmt.Sprint(rupt)+`</div><div class="sub"><a href="/admin/produits?vue=rupture">Voir →</a></div></div>
<div class="mx-card"><h3>Comptes</h3><div class="big">`+fmt.Sprint(nUsers)+`</div><div class="sub"><a href="/admin/clients">Voir →</a></div></div>
</div><br><div class="mx-grid two">
<div class="mx-card"><h2>Dernières commandes</h2><div class="mx-tablewrap"><table class="mx-table"><tr><th>#</th><th>Date</th><th>Moyen</th><th>Total</th><th>Statut</th></tr>`+last.String()+`</table></div></div>
<div class="mx-card"><h2>À réassortir</h2><div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>État</th></tr>`+low.String()+`</table></div></div>
</div>`, r))
}

func statutBadge(st string) string {
	cls := ""
	switch st {
	case "paye", "livre", "en stock":
		cls = "ok"
	case "attente_momo", "prepare", "whatsapp_recu", "retour":
		cls = "warn"
	case "annule", "rupture", "rupture de stock":
		cls = "bad"
	case "devis":
		cls = "info"
	}
	return `<span class="mx-badge ` + cls + `">` + render.Esc(st) + `</span>`
}

func (a *App) AdminProduits(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	vue := r.URL.Query().Get("vue")
	if vue == "" {
		vue = "all"
	}
	tri := r.URL.Query().Get("tri")
	if tri == "" {
		tri = "nom"
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const per = 50
	cond := ""
	switch vue {
	case "stock":
		cond = ` AND dispo='en stock'`
	case "rupture":
		cond = ` AND (dispo='rupture de stock' OR (stock IS NOT NULL AND stock<=0))`
	case "sansprix":
		cond = ` AND prix IS NULL AND prix_boutique IS NULL`
	}
	order := "nom"
	switch tri {
	case "prix":
		order = "COALESCE(prix_boutique,prix,999999999)"
	case "stock":
		order = "stock"
	}
	var total int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE (nom LIKE ? OR slug LIKE ?)`+cond, "%"+q+"%", "%"+q+"%").Scan(&total)
	pages := (total + per - 1) / per
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}
	// Compteurs des vues.
	var cAll, cStock, cRupt, cSP int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&cAll)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE dispo='en stock'`).Scan(&cStock)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE dispo='rupture de stock' OR (stock IS NOT NULL AND stock<=0)`).Scan(&cRupt)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE prix IS NULL AND prix_boutique IS NULL`).Scan(&cSP)
	base := "/admin/produits?q=" + q + "&tri=" + tri
	views := render.MxViews([]render.MxView{
		render.MxView{Key: "all", Href: base + "&vue=all", Label: "Tous", Count: cAll},
		render.MxView{Key: "stock", Href: base + "&vue=stock", Label: "En stock", Count: cStock},
		render.MxView{Key: "rupture", Href: base + "&vue=rupture", Label: "Rupture", Count: cRupt},
		render.MxView{Key: "sansprix", Href: base + "&vue=sansprix", Label: "Sans prix", Count: cSP},
	}, vue)
	rows, err := a.DB.Query(`SELECT slug,nom,marque,prix,prix_boutique,dispo,stock,image FROM products WHERE (nom LIKE ? OR slug LIKE ?)`+cond+` ORDER BY `+order+` LIMIT ? OFFSET ?`, "%"+q+"%", "%"+q+"%", per, (page-1)*per)
	if err != nil {
		http.Error(w, "indisponible", 500)
		return
	}
	defer rows.Close()
	var tr strings.Builder
	for rows.Next() {
		var slug, nom, marque, dispo, img string
		var prix, pb, stock sql.NullInt64
		_ = rows.Scan(&slug, &nom, &marque, &prix, &pb, &dispo, &stock, &img)
		pp := "devis"
		if pb.Valid {
			pp = fmt.Sprint(pb.Int64)
		} else if prix.Valid {
			pp = fmt.Sprint(prix.Int64)
		}
		st := "—"
		if stock.Valid {
			st = fmt.Sprint(stock.Int64)
		}
		im := ""
		if img != "" {
			im = `<img class="thumb" src="/` + render.Esc(img) + `" alt="" loading="lazy">`
		}
		fmt.Fprintf(&tr, `<tr><td>%s<a href="/admin/produits/%s">%s</a><div class="row-actions"><a href="/admin/produits/%s">Modifier</a> · <a href="/produits/%s.html">Voir</a> · <a href="/admin/produits/%s?dupliquer=1" data-mx-confirm="Dupliquer ce produit ?">Dupliquer</a></div></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			im, slug, render.Esc(nom), slug, slug, slug, render.Esc(marque), pp, statutBadge(dispo), st)
	}
	a.write(w, 200, a.adminPage("Produits", "produits",
		`<p><a class="mx-btn" href="/admin/produits/nouveau">+ Ajouter un produit</a></p>`+views+`
<form class="mx-tools" method="get" action="/admin/produits"><input type="hidden" name="vue" value="`+render.Esc(vue)+`">
<input type="search" name="q" value="`+render.Esc(q)+`" placeholder="Rechercher nom, slug…">
<select name="tri"><option value="nom">Tri : nom</option><option value="prix">Tri : prix</option><option value="stock">Tri : stock</option></select>
<button class="mx-btn">OK</button></form>
<div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>Marque</th><th>Prix</th><th>Dispo</th><th>Stock</th></tr>`+tr.String()+`</table>`+render.MxPager(base+"&vue="+vue, page, pages)+`</div>`, r))
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
	isNew := slug == "nouveau" || (slug == "" && strings.HasSuffix(r.URL.Path, "/nouveau"))
	if r.URL.Query().Get("dupliquer") == "1" && !isNew {
		a.adminDupliquer(w, r, slug)
		return
	}
	var p catalog.RawProduct
	var prix, pb, pa, stock sql.NullInt64
	if !isNew {
		err := a.DB.QueryRow(`SELECT slug,nom,marque,prix,prix_boutique,prix_avant,dispo,accroche,description,image,ref_fabricant,stock,main,sub,live_top,live_sub
		FROM products WHERE slug=?`, slug).Scan(&p.Slug, &p.Nom, &p.Marque, &prix, &pb, &pa, &p.Dispo, &p.Accroche, &p.Description, &p.Image, &p.RefFabricant, &stock, &p.Main, &p.Sub, &p.LiveTop, &p.LiveSub)
		if err != nil {
			http.NotFound(w, r)
			return
		}
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
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
		if isNew {
			nslug := catalog.Slugify(nn("nom"))
			if nslug == "" {
				a.write(w, 200, a.adminPage("Nouveau produit", "produits", render.MxNotice("", "Nom requis pour créer le slug."), r))
				return
			}
			_, err := a.DB.Exec(`INSERT INTO products(slug,nom,marque,prix,prix_boutique,prix_avant,dispo,accroche,description,image,stock,main,sub,live_top,live_sub,facets_json,search_index) VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,'{}','')`,
				nslug, nn("nom"), nn("marque"), ni("prix"), ni("prix_boutique"), ni("prix_avant"), nn("dispo"), nn("accroche"), nn("description"), nn("image"), ni("stock"), nn("main"), nn("sub"), nn("live_top"), nn("live_sub"))
			if err != nil {
				a.write(w, 200, a.adminPage("Nouveau produit", "produits", render.MxNotice("", "Création impossible (slug déjà pris ?)."), r))
				return
			}
			slug = nslug
			if rel, uerr := a.saveUpload(r, slug); uerr != nil {
				a.write(w, 200, a.adminPage("Nouveau produit", "produits", render.MxNotice("", "Produit créé, mais photo refusée : "+uerr.Error()), r))
				return
			} else if rel != "" {
				_, _ = a.DB.Exec(`UPDATE products SET image=? WHERE slug=?`, rel, slug)
			}
		} else {
			_, err := a.DB.Exec(`UPDATE products SET nom=?,marque=?,prix=?,prix_boutique=?,prix_avant=?,dispo=?,accroche=?,description=?,image=?,stock=?,main=?,sub=?,live_top=?,live_sub=? WHERE slug=?`,
				nn("nom"), nn("marque"), ni("prix"), ni("prix_boutique"), ni("prix_avant"), nn("dispo"), nn("accroche"), nn("description"), nn("image"), ni("stock"), nn("main"), nn("sub"), nn("live_top"), nn("live_sub"), slug)
			if err != nil {
				http.Error(w, "enregistrement impossible", 500)
				return
			}
			if rel, uerr := a.saveUpload(r, slug); uerr != nil {
				a.write(w, 200, a.adminPage("Modifier — "+p.Nom, "produits", render.MxNotice("", "Enregistré, mais photo refusée : "+uerr.Error()), r))
				return
			} else if rel != "" {
				_, _ = a.DB.Exec(`UPDATE products SET image=? WHERE slug=?`, rel, slug)
			}
		}
		a.reload()
		db.SyncFTS(a.DB, slug)
		http.Redirect(w, r, "/admin/produits/"+slug+"?msg=Produit+enregistr%C3%A9", http.StatusSeeOther)
		return
	}
	str := func(v sql.NullInt64) string {
		if v.Valid {
			return fmt.Sprint(v.Int64)
		}
		return ""
	}
	titre := "Modifier — " + p.Nom
	if isNew {
		titre = "Nouveau produit"
	}
	dispoOpts := ""
	for _, d := range []string{"en stock", "rupture de stock", "retour"} {
		sel := ""
		if p.Dispo == d {
			sel = " selected"
		}
		dispoOpts += `<option` + sel + `>` + d + `</option>`
	}
	imgPrev := ""
	if p.Image != "" {
		imgPrev = `<p><img src="/` + render.Esc(p.Image) + `" alt="" style="max-width:220px;border:1px solid var(--mx-border);border-radius:8px"></p>`
	}
	a.write(w, 200, a.adminPage(titre, "produits",
		`<form class="mx-form" method="post" enctype="multipart/form-data"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<div class="mx-cols"><div>
<div class="mx-postbox"><h2>Contenu <button type="button" data-mx-toggle aria-label="Replier">▾</button></h2><div class="in">
<label>Nom * <input name="nom" required value="`+render.Esc(p.Nom)+`"></label>
<label>Accroche <input name="accroche" value="`+render.Esc(p.Accroche)+`"></label>
<label>Description <textarea name="description" rows="6">`+render.Esc(p.Description)+`</textarea></label>
</div></div>
<div class="mx-postbox"><h2>Prix &amp; stock <button type="button" data-mx-toggle aria-label="Replier">▾</button></h2><div class="in">
<label>Prix (vide = devis) <input name="prix" inputmode="numeric" value="`+str(prix)+`"></label>
<label>Prix boutique <input name="prix_boutique" inputmode="numeric" value="`+str(pb)+`"></label>
<label>Prix avant <input name="prix_avant" inputmode="numeric" value="`+str(pa)+`"></label>
<label>Dispo <select name="dispo">`+dispoOpts+`</select></label>
<label>Stock <input name="stock" inputmode="numeric" value="`+str(stock)+`"></label>
</div></div>
<div class="mx-postbox"><h2>Classement <button type="button" data-mx-toggle aria-label="Replier">▾</button></h2><div class="in">
<label>Marque <input name="marque" value="`+render.Esc(p.Marque)+`"></label>
<label>Catégorie <input name="main" value="`+render.Esc(p.Main)+`"></label>
<label>Sous-catégorie <input name="sub" value="`+render.Esc(p.Sub)+`"></label>
<label>Store (live_top) <input name="live_top" value="`+render.Esc(p.LiveTop)+`"></label>
<label>Sous-store (live_sub) <input name="live_sub" value="`+render.Esc(p.LiveSub)+`"></label>
</div></div>
</div><div><div class="mx-postbox mx-pub"><h2>Publication</h2><div class="in">
<label>Image <input name="image" value="`+render.Esc(p.Image)+`" placeholder="images/…"></label>
`+imgPrev+`
<label>Fichier (JPG/PNG/WebP, max 5 Mo) <input type="file" name="photo" accept=".jpg,.jpeg,.png,.webp"></label>
<div class="row"><button class="mx-btn">Enregistrer</button>`+func() string {
			if isNew {
				return ""
			}
			return `<a class="mx-btn ghost" href="/produits/` + p.Slug + `.html">Voir</a>`
		}()+`</div>
</div></div></div></div></form>`, r))
}

// adminDupliquer : copie un produit sous un slug -copie (prix/stock repris).
func (a *App) adminDupliquer(w http.ResponseWriter, r *http.Request, slug string) {
	_, err := a.DB.Exec(`INSERT INTO products(slug,nom,marque,prix,prix_boutique,prix_avant,dispo,accroche,description,image,ref_fabricant,stock,main,sub,live_top,live_sub,facets_json,search_index)
		SELECT slug||'-copie',nom||' (copie)',marque,prix,prix_boutique,prix_avant,dispo,accroche,description,image,ref_fabricant,stock,main,sub,live_top,live_sub,facets_json,search_index FROM products WHERE slug=?`, slug)
	if err != nil {
		http.Redirect(w, r, "/admin/produits?err=Duplication+impossible+(existe+d%C3%A9j%C3%A0+?)", http.StatusSeeOther)
		return
	}
	a.reload()
	db.SyncFTS(a.DB, slug+"-copie")
	http.Redirect(w, r, "/admin/produits/"+slug+"-copie?msg=Produit+dupliqu%C3%A9", http.StatusSeeOther)
}

func (a *App) AdminCommandes(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		_, _ = a.DB.Exec(`UPDATE orders SET statut=? WHERE id=?`, r.FormValue("statut"), r.FormValue("id"))
		http.Redirect(w, r, "/admin/commandes", http.StatusSeeOther)
		return
	}
	vue := r.URL.Query().Get("vue")
	if vue == "" {
		vue = "all"
	}
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const per = 30
	cond := ""
	if vue != "all" {
		cond = ` AND statut='` + strings.ReplaceAll(vue, `'`, ``) + `'`
	}
	var total int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE 1=1` + cond).Scan(&total)
	pages := (total + per - 1) / per
	if pages < 1 {
		pages = 1
	}
	if page > pages {
		page = pages
	}
	// Compteurs par statut.
	type cv struct {
		k string
		n int
	}
	var counts []cv
	crows, _ := a.DB.Query(`SELECT statut,COUNT(*) FROM orders GROUP BY statut`)
	if crows != nil {
		defer crows.Close()
		for crows.Next() {
			var k string
			var n int
			_ = crows.Scan(&k, &n)
			counts = append(counts, cv{k, n})
		}
	}
	views := []render.MxView{render.MxView{Key: "all", Href: "/admin/commandes?vue=all", Label: "Toutes", Count: total}}
	for _, c := range counts {
		views = append(views, render.MxView{Key: c.k, Href: "/admin/commandes?vue=" + c.k, Label: c.k, Count: c.n})
	}
	rows, _ := a.DB.Query(`SELECT id,statut,moyen_paiement,total,client_json,created_at FROM orders WHERE 1=1`+cond+` ORDER BY id DESC LIMIT ? OFFSET ?`, per, (page-1)*per)
	var tr strings.Builder
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var statut, moyen, client, at string
			var tot sql.NullInt64
			_ = rows.Scan(&id, &statut, &moyen, &tot, &client, &at)
			t := "devis"
			if tot.Valid {
				t = fmt.Sprint(tot.Int64)
			}
			cl := client
			if len(cl) > 60 {
				cl = cl[:60] + "…"
			}
			fmt.Fprintf(&tr, `<tr><td><a href="/admin/commandes/%d"><strong>#%d</strong></a><div class="row-actions"><a href="/admin/commandes/%d">Détail</a></div></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				id, id, id, render.Esc(at), render.Esc(moyen), t, render.Esc(cl)+` `+statutBadge(statut))
		}
	}
	a.write(w, 200, a.adminPage("Commandes", "commandes",
		render.MxViews(views, vue)+`
<div class="mx-tablewrap"><table class="mx-table"><tr><th>#</th><th>Date</th><th>Moyen</th><th>Total</th><th>Client / Statut</th></tr>`+tr.String()+`</table>`+render.MxPager("/admin/commandes?vue="+vue, page, pages)+`</div>`, r))
}
