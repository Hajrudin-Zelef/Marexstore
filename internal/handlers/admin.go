package handlers

import (
	"database/sql"
	"fmt"
	"net/http"
	"net/url"
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
	// Produits par catégorie (donut) : libellés lisibles, tops à 0 exclus.
	var catSegs []dashSeg
	if crows, err := a.DB.Query(`SELECT live_top, COUNT(*) FROM products GROUP BY live_top ORDER BY COUNT(*) DESC`); err == nil && crows != nil {
		defer crows.Close()
		i := 0
		for crows.Next() {
			var top string
			var n int
			_ = crows.Scan(&top, &n)
			if n <= 0 {
				continue
			}
			lb, ok := topLabels[top]
			if !ok {
				lb = top
			}
			catSegs = append(catSegs, dashSeg{Label: lb, Count: n, Color: dashColor(i)})
			i++
		}
	}
	// Commandes par statut (donut).
	stLabel := map[string]string{
		"devis": "Devis", "whatsapp_recu": "WhatsApp reçu", "attente_momo": "Attente MoMo",
		"paye": "Payé", "prepare": "Préparé", "livre": "Livré", "annule": "Annulé",
	}
	var stSegs []dashSeg
	if srows, err := a.DB.Query(`SELECT statut, COUNT(*) FROM orders GROUP BY statut ORDER BY COUNT(*) DESC`); err == nil && srows != nil {
		defer srows.Close()
		i := 0
		for srows.Next() {
			var st string
			var n int
			_ = srows.Scan(&st, &n)
			if n <= 0 {
				continue
			}
			lb, ok := stLabel[st]
			if !ok {
				lb = st
			}
			stSegs = append(stSegs, dashSeg{Label: lb, Count: n, Color: dashColor(i)})
			i++
		}
	}
	// Top marques (barres) : marque vide = « Sans marque ».
	var brandSegs []dashSeg
	maxBrand := 0
	if brows, err := a.DB.Query(`SELECT COALESCE(NULLIF(TRIM(marque),''),'Sans marque'), COUNT(*) FROM products GROUP BY 1 ORDER BY COUNT(*) DESC LIMIT 8`); err == nil && brows != nil {
		defer brows.Close()
		i := 0
		for brows.Next() {
			var m string
			var n int
			_ = brows.Scan(&m, &n)
			if n <= 0 {
				continue
			}
			if n > maxBrand {
				maxBrand = n
			}
			brandSegs = append(brandSegs, dashSeg{Label: m, Count: n, Color: dashColor(i)})
			i++
		}
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
	// Donut statuts : état vide élégant quand aucune commande.
	stDonut := donutSVG(stSegs, 170, fmt.Sprint(nOrders), "commandes")
	if nOrders == 0 {
		stDonut = `<div class="mx-empty"><div class="mx-empty-big">0</div><div class="mx-muted">Aucune commande pour le moment.<br>Les commandes apparaîtront ici.</div></div>`
	}
	zeroNote := ""
	zeroTops := 0
	for _, t := range []string{"cameras-videosurveillances", "accessoires-telephones-tablettes", "uncategorized"} {
		found := false
		for _, s := range catSegs {
			if s.Label == topLabels[t] {
				found = true
				break
			}
		}
		if !found {
			zeroTops++
		}
	}
	if zeroTops > 0 {
		zeroNote = `<div class="mx-note">` + fmt.Sprint(zeroTops) + ` catégorie(s) sans produit (Caméras &amp; vidéosurveillance, Accessoires téléphones, Non classé) — exclue(s) du graphique.</div>`
	}
	body := `<div class="mx-grid kpi">` +
		kpiCard(icCash, "#2e5136", "Chiffre encaissé", caStr, `F CFA (payé/préparé/livré)`) +
		kpiCard(icBag, "#2271b1", "Commandes", fmt.Sprint(nOrders), `<a href="/admin/commandes">Tout voir →</a>`) +
		kpiCard(icBox, "#2e5136", "Produits", fmt.Sprint(nProd), `<a href="/admin/produits">Gérer →</a>`) +
		kpiCard(icWarn, "#d63638", "Ruptures", fmt.Sprint(rupt), `<a href="/admin/produits?vue=rupture">Voir →</a>`) +
		kpiCard(icUser, "#7c3aed", "Comptes", fmt.Sprint(nUsers), `<a href="/admin/clients">Voir →</a>`) +
		`</div>` +
		`<div class="mx-grid charts">` +
		`<div class="mx-card mx-chart"><h2>Produits par catégorie</h2>` + donutSVG(catSegs, 180, fmt.Sprint(nProd), "produits") + zeroNote + `</div>` +
		`<div class="mx-card mx-chart"><h2>Commandes par statut</h2>` + stDonut + `</div>` +
		`</div>` +
		`<div class="mx-card mx-chart"><h2>Top marques</h2>` + barRows(brandSegs, maxBrand) + `</div>` +
		`<div class="mx-grid two">` +
		`<div class="mx-card"><h2>Dernières commandes</h2><div class="mx-tablewrap"><table class="mx-table"><tr><th>#</th><th>Date</th><th>Moyen</th><th>Total</th><th>Statut</th></tr>` + last.String() + `</table></div></div>` +
		`<div class="mx-card"><h2>À réassortir</h2><div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>État</th></tr>` + low.String() + `</table></div></div>` +
		`</div>`
	a.write(w, 200, a.adminPage("Tableau de bord", "", body, r))
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
	cat := r.URL.Query().Get("cat")
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const per = 50
	cond := ""
	// Recherche multi-mots : chaque mot doit apparaître (nom ou slug),
	// sinon « toner 650a » ne trouve rien alors que « Toner HP 650A » existe.
	var args []any
	words := strings.Fields(q)
	if len(words) == 0 {
		cond = ""
	} else {
		var ands []string
		for _, w := range words {
			ands = append(ands, `(nom LIKE ? OR slug LIKE ?)`)
			args = append(args, "%"+w+"%", "%"+w+"%")
		}
		cond = " AND " + strings.Join(ands, " AND ")
	}
	switch vue {
	case "stock":
		cond = ` AND dispo='en stock'`
	case "rupture":
		cond = ` AND (dispo='rupture de stock' OR (stock IS NOT NULL AND stock<=0))`
	case "sansprix":
		cond = ` AND prix IS NULL AND prix_boutique IS NULL`
	}
	if cat != "" && cat != "all" {
		cond += ` AND live_top=?`
		args = append(args, cat)
	}
	order := "nom"
	switch tri {
	case "prix":
		order = "COALESCE(prix_boutique,prix,999999999)"
	case "stock":
		order = "stock"
	}
	var total int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products WHERE 1=1`+cond, args...).Scan(&total)
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
	// Options du filtre catégorie (live_top + compteur).
	var catOpts strings.Builder
	catOpts.WriteString(`<option value="all">Toutes catégories</option>`)
	if crows, err := a.DB.Query(`SELECT live_top, COUNT(*) FROM products WHERE live_top IS NOT NULL AND live_top<>'' GROUP BY live_top ORDER BY COUNT(*) DESC`); err == nil && crows != nil {
		defer crows.Close()
		for crows.Next() {
			var top string
			var n int
			_ = crows.Scan(&top, &n)
			lb, ok := topLabels[top]
			if !ok {
				lb = top
			}
			sel := ""
			if top == cat {
				sel = " selected"
			}
			fmt.Fprintf(&catOpts, `<option value="%s"%s>%s (%d)</option>`, render.Esc(top), sel, render.Esc(lb), n)
		}
	}
	// Bandeau mini-stats (compteurs globaux, liens vers les vues).
	stats := `<div class="mx-stats">` +
		miniStat("/admin/produits?vue=all", "Produits", fmt.Sprint(cAll), "#2e5136") +
		miniStat("/admin/produits?vue=stock", "En stock", fmt.Sprint(cStock), "#2271b1") +
		miniStat("/admin/produits?vue=rupture", "Rupture", fmt.Sprint(cRupt), "#d63638") +
		miniStat("/admin/produits?vue=sansprix", "Sans prix", fmt.Sprint(cSP), "#dba617") +
		`</div>`
	base := "/admin/produits?q=" + url.QueryEscape(q) + "&tri=" + tri + "&cat=" + cat
	views := render.MxViews([]render.MxView{
		render.MxView{Key: "all", Href: base + "&vue=all", Label: "Tous", Count: cAll},
		render.MxView{Key: "stock", Href: base + "&vue=stock", Label: "En stock", Count: cStock},
		render.MxView{Key: "rupture", Href: base + "&vue=rupture", Label: "Rupture", Count: cRupt},
		render.MxView{Key: "sansprix", Href: base + "&vue=sansprix", Label: "Sans prix", Count: cSP},
	}, vue)
	selTri := func(v string) string {
		if tri == v {
			return " selected"
		}
		return ""
	}
	rows, err := a.DB.Query(`SELECT slug,nom,marque,prix,prix_boutique,dispo,stock,image,live_top FROM products WHERE 1=1`+cond+` ORDER BY `+order+` LIMIT ? OFFSET ?`, append(args, per, (page-1)*per)...)
	if err != nil {
		http.Error(w, "indisponible", 500)
		return
	}
	defer rows.Close()
	var tr, cards, minis strings.Builder
	for rows.Next() {
		var slug, nom, marque, dispo, img, top string
		var prix, pb, stock sql.NullInt64
		_ = rows.Scan(&slug, &nom, &marque, &prix, &pb, &dispo, &stock, &img, &top)
		pp := `<span class="mx-muted">devis</span>`
		ppTxt := "devis"
		if pb.Valid {
			pp = `<span class="mx-price">` + fmtMoney(pb.Int64) + ` F CFA</span>`
			ppTxt = fmtMoney(pb.Int64) + ` F CFA`
		} else if prix.Valid {
			pp = `<span class="mx-price">` + fmtMoney(prix.Int64) + ` F CFA</span>`
			ppTxt = fmtMoney(prix.Int64) + ` F CFA`
		}
		// Stock : nombre suivi ; sinon pastille explicite (jamais « — » vide).
		st := ""
		switch {
		case stock.Valid && stock.Int64 <= 0:
			st = `<span class="mx-badge bad" title="Stock épuisé">0</span>`
		case stock.Valid && stock.Int64 <= 3:
			st = `<span class="mx-badge warn" title="Stock faible">` + fmt.Sprint(stock.Int64) + `</span>`
		case stock.Valid:
			st = fmt.Sprint(stock.Int64)
		case !pb.Valid && !prix.Valid:
			st = `<span class="mx-badge info" title="Prix sur devis : pas de stock suivi">Devis</span>`
		default:
			st = `<span class="mx-badge" title="Sur commande : stock non suivi">Cde</span>`
		}
		catLb, ok := topLabels[top]
		if !ok || top == "" {
			catLb = "Non classé"
		}
		im := ""
		if img != "" {
			im = `<img class="thumb" src="/` + render.Esc(img) + `" alt="" loading="lazy">`
		}
		fmt.Fprintf(&tr, `<tr><td><div class="mx-prod">%s<div><a class="mx-prod-nm" href="/admin/produits/%s">%s</a><div class="mx-prod-cat">%s</div><div class="row-actions"><a href="/admin/produits/%s">Modifier</a> · <a href="/produits/%s.html">Voir</a> · <a href="/admin/produits/%s?dupliquer=1" data-mx-confirm="Dupliquer ce produit ?">Dupliquer</a></div></div></div></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
			im, slug, render.Esc(nom), render.Esc(catLb), slug, slug, slug, render.Esc(marque), pp, statutBadge(dispo), st)
		fmt.Fprintf(&cards, `<div class="mx-gcard"><a href="/admin/produits/%s">%s</a><div class="in"><a class="mx-prod-nm" href="/admin/produits/%s">%s</a><div class="mx-prod-cat">%s · %s</div><div class="mx-gmeta">%s %s %s</div><div class="row-actions"><a href="/admin/produits/%s">Modifier</a> · <a href="/produits/%s.html">Voir</a> · <a href="/admin/produits/%s?dupliquer=1" data-mx-confirm="Dupliquer ce produit ?">Dupliquer</a></div></div></div>`,
			slug, im, slug, render.Esc(nom), render.Esc(catLb), render.Esc(marque), pp, statutBadge(dispo), st, slug, slug, slug)
		fmt.Fprintf(&minis, `<a class="mx-mini" href="/admin/produits/%s" title="%s — %s">%s<span>%s</span></a>`,
			slug, render.Esc(nom), render.Esc(ppTxt), im, render.Esc(nom))
	}
	nbTxt := fmt.Sprintf(`<span class="mx-count">%d produit%s</span>`, total, func() string {
		if total > 1 {
			return "s"
		}
		return ""
	}())
	a.write(w, 200, a.adminPage("Produits", "produits",
		`<p><a class="mx-btn" href="/admin/produits/nouveau">+ Ajouter un produit</a></p>`+stats+views+`
<form class="mx-tools" method="get" action="/admin/produits"><input type="hidden" name="vue" value="`+render.Esc(vue)+`">
<input type="search" name="q" value="`+render.Esc(q)+`" placeholder="Rechercher nom, slug…">
<select name="cat">`+catOpts.String()+`</select>
<select name="tri"><option value="nom"`+selTri("nom")+`>Tri : nom</option><option value="prix"`+selTri("prix")+`>Tri : prix</option><option value="stock"`+selTri("stock")+`>Tri : stock</option></select>
<button class="mx-btn">OK</button>`+nbTxt+`
<span class="mx-aff" role="group" aria-label="Affichage"><button type="button" data-aff="list" class="on" title="Liste">☰ Liste</button><button type="button" data-aff="grid" title="Grille">▦ Grille</button><button type="button" data-aff="minis" title="Miniatures">▦ Minis</button></span></form>
<div data-mx-aff="list"><div class="mx-tablewrap"><table class="mx-table"><tr><th>Produit</th><th>Marque</th><th>Prix</th><th>Dispo</th><th>Stock</th></tr>`+tr.String()+`</table></div>
<div class="mx-gridcards">`+cards.String()+`</div>
<div class="mx-minis">`+minis.String()+`</div>`+render.MxPager(base+"&vue="+vue, page, pages)+`</div>`, r))
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
	// Aperçu live (carte produit, comme le badge employé) : admin.js le
	// met à jour en direct pendant la saisie.
	pvNom := p.Nom
	if pvNom == "" {
		pvNom = "Nouveau produit"
	}
	pvPrix := `devis`
	pvPrixCls := "pv-prix devis"
	if pb.Valid {
		pvPrix = fmtMoney(pb.Int64) + ` F CFA`
		pvPrixCls = "pv-prix"
	} else if prix.Valid {
		pvPrix = fmtMoney(prix.Int64) + ` F CFA`
		pvPrixCls = "pv-prix"
	}
	pvStock := "—"
	if stock.Valid {
		pvStock = fmt.Sprint(stock.Int64)
	}
	pvDispo := p.Dispo
	if pvDispo == "" {
		pvDispo = "en stock"
	}
	pvImg := `<div id="pv-img-ph" class="pv-img-ph">Aucune photo</div><img id="pv-img" src="" alt="" style="display:none">`
	if p.Image != "" {
		pvImg = `<div id="pv-img-ph" class="pv-img-ph" style="display:none">Aucune photo</div><img id="pv-img" src="/` + render.Esc(p.Image) + `" alt="">`
	}
	pvBadgeStyle := ""
	if p.Marque == "" {
		pvBadgeStyle = ` style="display:none"`
	}
	voirBtn := ""
	if !isNew {
		voirBtn = `<a class="mx-btn ghost" href="/produits/` + p.Slug + `.html">Voir</a>`
	}
	a.write(w, 200, a.adminPage(titre, "produits",
		`<form class="mx-form" id="mx-edit-form" method="post" enctype="multipart/form-data"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<div class="mx-edit">
<div class="mx-preview">
<div class="pv-top">
<div class="pv-avatar">`+pvImg+`</div>
<div class="pv-nom" id="pv-nom">`+render.Esc(pvNom)+`</div>
<div class="pv-badge" id="pv-marque"`+pvBadgeStyle+`>`+render.Esc(p.Marque)+`</div>
</div>
<div class="pv-body">
<div class="pv-info">
<div class="pv-row"><span>Prix</span><strong class="`+pvPrixCls+`" id="pv-prix">`+pvPrix+`</strong></div>
<div class="pv-row"><span>Catégorie</span><strong id="pv-cat">`+render.Esc(p.Main)+`</strong></div>
<div class="pv-row"><span>Stock</span><strong id="pv-stock">`+pvStock+`</strong></div>
<div class="pv-row"><span>Statut</span><span id="pv-dispo">`+statutBadge(pvDispo)+`</span></div>
</div>
</div>
<div class="pv-photo">
<label>Image <input name="image" id="f-image" value="`+render.Esc(p.Image)+`" placeholder="images/…"></label>
<label>Fichier (JPG/PNG/WebP, max 5 Mo) <input type="file" name="photo" accept=".jpg,.jpeg,.png,.webp"></label>
</div>
</div>
<div class="mx-fcol">
<div class="mx-fsec"><h3>Contenu</h3><div class="mx-fgrid">
<label class="span2">Nom * <input name="nom" id="f-nom" required value="`+render.Esc(p.Nom)+`"></label>
<label class="span2">Accroche <input name="accroche" value="`+render.Esc(p.Accroche)+`"></label>
<label class="span2">Description <textarea name="description" rows="5">`+render.Esc(p.Description)+`</textarea></label>
</div></div>
<div class="mx-fsec"><h3>Prix &amp; stock</h3><div class="mx-fgrid">
<label>Prix (vide = devis) <input name="prix" id="f-prix" inputmode="numeric" value="`+str(prix)+`"></label>
<label>Prix boutique <input name="prix_boutique" id="f-pb" inputmode="numeric" value="`+str(pb)+`"></label>
<label>Prix avant <input name="prix_avant" inputmode="numeric" value="`+str(pa)+`"></label>
<label>Dispo <select name="dispo" id="f-dispo">`+dispoOpts+`</select></label>
<label>Stock <input name="stock" id="f-stock" inputmode="numeric" value="`+str(stock)+`"></label>
</div></div>
<div class="mx-fsec"><h3>Classement</h3><div class="mx-fgrid">
<label>Marque <input name="marque" id="f-marque" value="`+render.Esc(p.Marque)+`"></label>
<label>Catégorie <input name="main" id="f-main" value="`+render.Esc(p.Main)+`"></label>
<label>Sous-catégorie <input name="sub" value="`+render.Esc(p.Sub)+`"></label>
<label>Store (live_top) <input name="live_top" value="`+render.Esc(p.LiveTop)+`"></label>
<label>Sous-store (live_sub) <input name="live_sub" value="`+render.Esc(p.LiveSub)+`"></label>
</div></div>
<div class="mx-form-actions"><button class="mx-btn primary">Enregistrer</button>`+voirBtn+`</div>
</div>
</div></form>`, r))
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
	// Mini-stats : CA encaissé (payée/préparée/livrée), en cours, livrées.
	var ca, enCours, livrees int64
	_ = a.DB.QueryRow(`SELECT COALESCE(SUM(total),0) FROM orders WHERE statut IN ('paye','prepare','livre')`).Scan(&ca)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE statut NOT IN ('livre','annule')`).Scan(&enCours)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM orders WHERE statut='livre'`).Scan(&livrees)
	rows, _ := a.DB.Query(`SELECT id,statut,moyen_paiement,total,client_json,created_at FROM orders WHERE 1=1`+cond+` ORDER BY id DESC LIMIT ? OFFSET ?`, per, (page-1)*per)
	var tr strings.Builder
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var id int64
			var statut, moyen, client, at string
			var tot sql.NullInt64
			_ = rows.Scan(&id, &statut, &moyen, &tot, &client, &at)
			t := `<span class="mx-muted">devis</span>`
			if tot.Valid {
				t = `<span class="mx-price">` + fmtMoney(tot.Int64) + ` F CFA</span>`
			}
			cl := clientLabel(client)
			fmt.Fprintf(&tr, `<tr><td><a href="/admin/commandes/%d"><strong>#%d</strong></a><div class="row-actions"><a href="/admin/commandes/%d">Détail</a></div></td><td>%s</td><td>%s</td><td>%s</td><td>%s</td><td>%s</td></tr>`,
				id, id, id, render.Esc(at), render.Esc(moyen), t, render.Esc(cl), statutBadge(statut))
		}
	}
	a.write(w, 200, a.adminPage("Commandes", "commandes",
		`<div class="mx-stats">`+
			miniStat("/admin/commandes?vue=all", "Commandes", fmt.Sprint(total), "#2e5136")+
			miniStat("/admin/commandes?vue=paye", "CA encaissé", fmtMoney(ca)+" F CFA", "#2271b1")+
			miniStat("/admin/commandes?vue=all", "En cours", fmt.Sprint(enCours), "#dba617")+
			miniStat("/admin/commandes?vue=livre", "Livrées", fmt.Sprint(livrees), "#6fa82e")+
			`</div>`+
			render.MxViews(views, vue)+`
<div class="mx-tablewrap"><table class="mx-table"><tr><th>#</th><th>Date</th><th>Moyen</th><th>Total</th><th>Client</th><th>Statut</th></tr>`+tr.String()+`</table></div>`+render.MxPager("/admin/commandes?vue="+vue, page, pages), r))
}
