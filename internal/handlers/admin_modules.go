package handlers

// Modules admin façon WooCommerce : sous-modules Produits (Marques,
// Étiquettes, Attributs, Avis), Statistiques, Marketing (codes promo)
// et Outils (importer / exporter / redirections).

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	marexdb "github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

func (a *App) writeAdmin(w http.ResponseWriter, r *http.Request, title, active, inner string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(a.adminPage(title, active, inner, r)))
}

func mxTable(headers []string, rows [][]string) string {
	var b strings.Builder
	b.WriteString(`<table class="mx-table"><thead><tr>`)
	for _, h := range headers {
		b.WriteString(`<th>` + render.Esc(h) + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		b.WriteString(`<tr>`)
		for _, c := range row {
			b.WriteString(`<td>` + c + `</td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table>`)
	return b.String()
}

func mxEmpty(msg string) string {
	return `<div class="mx-empty">` + render.Esc(msg) + `</div>`
}

// ---------- Produits > Marques ----------

func (a *App) AdminMarques(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	type kv struct {
		k string
		n int
	}
	var brands []kv
	for b, n := range a.Store.BrandCounts {
		brands = append(brands, kv{b, n})
	}
	sort.Slice(brands, func(i, j int) bool {
		if brands[i].n != brands[j].n {
			return brands[i].n > brands[j].n
		}
		return brands[i].k < brands[j].k
	})
	logos := map[string]bool{"acer": true, "apple": true, "asus": true, "cisco": true, "dell": true, "epson": true, "hp": true, "kyocera": true, "lenovo": true}
	var rows [][]string
	for _, b := range brands {
		slug := catalog.Slugify(b.k)
		logo := "—"
		if logos[slug] {
			logo = `<img src="/static/logos/` + slug + `.svg" alt="" style="height:20px;vertical-align:middle">`
		}
		rows = append(rows, []string{
			logo,
			`<a href="/boutique.html?marque=` + render.Esc(catalog.NormName(b.k)) + `">` + render.Esc(b.k) + `</a>`,
			strconv.Itoa(b.n),
		})
	}
	inner := `<p>` + strconv.Itoa(len(brands)) + ` marques dans le catalogue.</p>` + mxTable([]string{"Logo", "Marque", "Produits"}, rows)
	a.writeAdmin(w, r, "Marques", "produits/marques", inner)
}

// ---------- Produits > Étiquettes ----------

func (a *App) AdminEtiquettes(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	counts := map[string]int{}
	for _, p := range a.Store.Products {
		if p.Sub != "" {
			counts[p.Sub]++
		}
	}
	type kv struct {
		k string
		n int
	}
	var tags []kv
	for k, n := range counts {
		tags = append(tags, kv{k, n})
	}
	sort.Slice(tags, func(i, j int) bool {
		if tags[i].n != tags[j].n {
			return tags[i].n > tags[j].n
		}
		return tags[i].k < tags[j].k
	})
	var rows [][]string
	for _, t := range tags {
		rows = append(rows, []string{
			`<a href="/boutique.html?type=` + render.Esc(catalog.Slugify(t.k)) + `">` + render.Esc(t.k) + `</a>`,
			strconv.Itoa(t.n),
		})
	}
	inner := `<p>Les étiquettes correspondent aux types d'équipement du catalogue.</p>` + mxTable([]string{"Étiquette", "Produits"}, rows)
	a.writeAdmin(w, r, "Étiquettes", "produits/etiquettes", inner)
}

// ---------- Produits > Attributs ----------

func (a *App) AdminAttributs(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	prods := map[string]int{}
	vals := map[string]map[string]bool{}
	for _, p := range a.Store.Products {
		for _, s := range p.SpecList() {
			if s.Label == "" {
				continue
			}
			prods[s.Label]++
			if vals[s.Label] == nil {
				vals[s.Label] = map[string]bool{}
			}
			vals[s.Label][s.Valeur] = true
		}
	}
	type kv struct {
		k string
		n int
	}
	var attrs []kv
	for k, n := range prods {
		attrs = append(attrs, kv{k, n})
	}
	sort.Slice(attrs, func(i, j int) bool {
		if attrs[i].n != attrs[j].n {
			return attrs[i].n > attrs[j].n
		}
		return attrs[i].k < attrs[j].k
	})
	var rows [][]string
	for _, at := range attrs {
		rows = append(rows, []string{render.Esc(at.k), strconv.Itoa(at.n), strconv.Itoa(len(vals[at.k]))})
	}
	inner := `<p>Les attributs sont les caractéristiques présentes sur les fiches produits.</p>` + mxTable([]string{"Attribut", "Produits", "Valeurs distinctes"}, rows)
	a.writeAdmin(w, r, "Attributs", "produits/attributs", inner)
}

// ---------- Produits > Avis ----------

func (a *App) AdminAvis(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	inner := mxEmpty("Aucun avis pour le moment. Les avis clients ne sont pas encore activés sur la boutique.")
	a.writeAdmin(w, r, "Avis", "produits/avis", inner)
}

// ---------- Statistiques ----------

func (a *App) statInt(q string, args ...any) int64 {
	var v int64
	_ = a.DB.QueryRow(q, args...).Scan(&v)
	return v
}

func (a *App) AdminStats(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	vue := r.URL.Query().Get("vue")
	kpi := func(lb string, vl string) string {
		return `<div class="mx-kpi"><div class="lb">` + render.Esc(lb) + `</div><div class="vl">` + render.Esc(vl) + `</div></div>`
	}
	fcfa := func(v int64) string { return catalog.FmtPrix(&v) }
	var inner strings.Builder
	switch vue {
	case "produits":
		rows, err := a.DB.Query(`SELECT p.nom, SUM(oi.qty), COALESCE(SUM(oi.qty*oi.prix_unitaire),0) FROM order_items oi JOIN products p ON p.id=oi.product_id GROUP BY oi.product_id ORDER BY SUM(oi.qty) DESC LIMIT 20`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var nom string
				var q, rev int64
				_ = rows.Scan(&nom, &q, &rev)
				tr = append(tr, []string{render.Esc(nom), strconv.FormatInt(q, 10), fcfa(rev)})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucune vente pour le moment."))
		} else {
			inner.WriteString(mxTable([]string{"Produit", "Quantité vendue", "Revenu"}, tr))
		}
	case "revenu":
		rows, err := a.DB.Query(`SELECT strftime('%Y-%m',created_at), COUNT(*), COALESCE(SUM(total),0) FROM orders GROUP BY 1 ORDER BY 1 DESC LIMIT 24`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var m string
				var n, rev int64
				_ = rows.Scan(&m, &n, &rev)
				tr = append(tr, []string{render.Esc(m), strconv.FormatInt(n, 10), fcfa(rev)})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucune commande pour le moment."))
		} else {
			inner.WriteString(mxTable([]string{"Mois", "Commandes", "Revenu"}, tr))
		}
	case "commandes":
		rows, err := a.DB.Query(`SELECT statut, COUNT(*), COALESCE(SUM(total),0) FROM orders GROUP BY statut ORDER BY COUNT(*) DESC`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var st string
				var n, rev int64
				_ = rows.Scan(&st, &n, &rev)
				tr = append(tr, []string{render.Esc(st), strconv.FormatInt(n, 10), fcfa(rev)})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucune commande pour le moment."))
		} else {
			inner.WriteString(mxTable([]string{"Statut", "Commandes", "Montant"}, tr))
		}
	case "categories":
		rows, err := a.DB.Query(`SELECT p.main, SUM(oi.qty), COALESCE(SUM(oi.qty*oi.prix_unitaire),0) FROM order_items oi JOIN products p ON p.id=oi.product_id GROUP BY p.main ORDER BY SUM(oi.qty) DESC`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var cat string
				var q, rev int64
				_ = rows.Scan(&cat, &q, &rev)
				tr = append(tr, []string{render.Esc(cat), strconv.FormatInt(q, 10), fcfa(rev)})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucune vente pour le moment."))
		} else {
			inner.WriteString(mxTable([]string{"Catégorie", "Quantité vendue", "Revenu"}, tr))
		}
	case "stock":
		rows, err := a.DB.Query(`SELECT slug,nom,dispo,stock FROM products WHERE dispo<>'en stock' OR (stock IS NOT NULL AND stock<=2) ORDER BY dispo, nom LIMIT 100`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var slug, nom, dispo string
				var stock *int64
				_ = rows.Scan(&slug, &nom, &dispo, &stock)
				st := "—"
				if stock != nil {
					st = strconv.FormatInt(*stock, 10)
				}
				tr = append(tr, []string{`<a href="/admin/produits/` + render.Esc(slug) + `">` + render.Esc(nom) + `</a>`, render.Esc(dispo), st})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucun produit en rupture ni en stock faible."))
		} else {
			inner.WriteString(mxTable([]string{"Produit", "Disponibilité", "Stock"}, tr))
		}
	case "variations":
		inner.WriteString(mxEmpty("Aucune variation vendue : les options des fiches sont informatives (prix sur demande)."))
	case "promos":
		inner.WriteString(mxEmpty("Voir le module Marketing > Codes promo pour la gestion des codes."))
	case "taxes":
		inner.WriteString(mxEmpty("Aucune taxe configurée : les prix sont affichés TTC."))
	case "telechargements":
		inner.WriteString(mxEmpty("Aucun produit téléchargeable dans le catalogue."))
	default:
		ventes := a.statInt(`SELECT COALESCE(SUM(total),0) FROM orders`)
		cmds := a.statInt(`SELECT COUNT(*) FROM orders`)
		vendus := a.statInt(`SELECT COALESCE(SUM(qty),0) FROM order_items`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(kpi("Total des ventes", fcfa(ventes)))
		inner.WriteString(kpi("Ventes nettes", fcfa(ventes)))
		inner.WriteString(kpi("Commandes", strconv.FormatInt(cmds, 10)))
		inner.WriteString(kpi("Produits vendus", strconv.FormatInt(vendus, 10)))
		inner.WriteString(kpi("Produits au catalogue", strconv.Itoa(len(a.Store.Products))))
		inner.WriteString(`</div>`)
		if cmds == 0 {
			inner.WriteString(mxEmpty("Aucune donnée pour la période : aucune commande enregistrée."))
		}
	}
	a.writeAdmin(w, r, "Statistiques", "stats", inner.String())
}

// ---------- Marketing ----------

func (a *App) ensurePromos() {
	_, _ = a.DB.Exec(`CREATE TABLE IF NOT EXISTS promos(code TEXT PRIMARY KEY, type TEXT NOT NULL DEFAULT 'fixe', valeur INTEGER NOT NULL DEFAULT 0, description TEXT NOT NULL DEFAULT '', usage_count INTEGER NOT NULL DEFAULT 0, limite INTEGER NULL, expire TEXT NOT NULL DEFAULT '', actif INTEGER NOT NULL DEFAULT 1)`)
}

func (a *App) AdminMarketing(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	a.ensurePromos()
	n := a.statInt(`SELECT COUNT(*) FROM promos WHERE actif=1`)
	inner := `<div class="mx-kpis"><div class="mx-kpi"><div class="lb">Codes promo actifs</div><div class="vl">` + strconv.FormatInt(n, 10) + `</div></div></div>` +
		`<p><a class="mx-btn" href="/admin/marketing/promos">Gérer les codes promo</a></p>`
	a.writeAdmin(w, r, "Marketing", "marketing", inner)
}

func (a *App) AdminPromos(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	a.ensurePromos()
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		switch r.FormValue("action") {
		case "ajouter":
			code := strings.ToUpper(strings.TrimSpace(r.FormValue("code")))
			typ := r.FormValue("type")
			if typ != "pourcentage" {
				typ = "fixe"
			}
			val, _ := strconv.ParseInt(r.FormValue("valeur"), 10, 64)
			var limite any
			if lv, err := strconv.ParseInt(r.FormValue("limite"), 10, 64); err == nil && lv > 0 {
				limite = lv
			}
			if code != "" && val > 0 {
				_, _ = a.DB.Exec(`INSERT INTO promos(code,type,valeur,description,limite,expire) VALUES(?,?,?,?,?,?) ON CONFLICT(code) DO UPDATE SET type=excluded.type, valeur=excluded.valeur, description=excluded.description, limite=excluded.limite, expire=excluded.expire, actif=1`,
					code, typ, val, r.FormValue("description"), limite, r.FormValue("expire"))
				a.auditAdmin(r, "promos.ajouter", code)
			}
		case "supprimer":
			_, _ = a.DB.Exec(`DELETE FROM promos WHERE code=?`, r.FormValue("code"))
			a.auditAdmin(r, "promos.supprimer", r.FormValue("code"))
		}
		http.Redirect(w, r, "/admin/marketing/promos?msg=Code+promo+enregistr%C3%A9", http.StatusSeeOther)
		return
	}
	rows, err := a.DB.Query(`SELECT code,type,valeur,description,usage_count,limite,expire FROM promos ORDER BY code`)
	var tr [][]string
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var code, typ, desc, exp string
			var val, usage int64
			var limite *int64
			_ = rows.Scan(&code, &typ, &val, &desc, &usage, &limite, &exp)
			typLb := "Remise panier fixe"
			valLb := catalog.FmtPrix(&val)
			if typ == "pourcentage" {
				typLb = "Remise en pourcentage"
				valLb = strconv.FormatInt(val, 10) + " %"
			}
			lim := "∞"
			if limite != nil {
				lim = strconv.FormatInt(*limite, 10)
			}
			if exp == "" {
				exp = "—"
			}
			tr = append(tr, []string{
				render.Esc(code), render.Esc(typLb), render.Esc(valLb), render.Esc(desc),
				strconv.FormatInt(usage, 10) + " / " + lim, render.Esc(exp),
				`<form method="post" style="display:inline"><input type="hidden" name="csrf" value="` + render.Esc(a.adminCSRF()) + `"><input type="hidden" name="action" value="supprimer"><input type="hidden" name="code" value="` + render.Esc(code) + `"><button class="mx-btn ghost" type="submit">Supprimer</button></form>`,
			})
		}
	}
	csrf := a.adminCSRF()
	var inner strings.Builder
	inner.WriteString(`<form method="post" class="mx-form"><input type="hidden" name="csrf" value="` + render.Esc(csrf) + `"><input type="hidden" name="action" value="ajouter">
<label>Code <input name="code" required></label>
<label>Type <select name="type"><option value="fixe">Remise panier fixe (F CFA)</option><option value="pourcentage">Remise en pourcentage (%)</option></select></label>
<label>Valeur <input name="valeur" type="number" min="1" required></label>
<label>Description <input name="description"></label>
<label>Limite d'usage <input name="limite" type="number" min="0" placeholder="∞"></label>
<label>Date d'expiration <input name="expire" type="date"></label>
<button class="mx-btn" type="submit">Ajouter le code</button></form>`)
	if len(tr) == 0 {
		inner.WriteString(mxEmpty("Aucun code promo."))
	} else {
		inner.WriteString(mxTable([]string{"Code", "Type de code promo", "Valeur du code promo", "Description", "Usage / Limite", "Date d'expiration", ""}, tr))
	}
	a.writeAdmin(w, r, "Codes promo", "marketing/promos", inner.String())
}

// ---------- Outils ----------

func (a *App) AdminOutils(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	vue := r.URL.Query().Get("vue")
	var inner strings.Builder
	switch vue {
	case "importer":
		if r.Method == http.MethodPost {
			if !a.checkAdminCSRF(r) {
				http.Error(w, "refusé", http.StatusForbidden)
				return
			}
			f, _, err := r.FormFile("fichier")
			if err != nil {
				http.Redirect(w, r, "/admin/outils?vue=importer&err=Fichier+manquant", http.StatusSeeOther)
				return
			}
			defer f.Close()
			var prods []*catalog.RawProduct
			if err := json.NewDecoder(f).Decode(&prods); err != nil {
				http.Redirect(w, r, "/admin/outils?vue=importer&err=JSON+illisible", http.StatusSeeOther)
				return
			}
			catalog.ApplyTaxonomy(prods)
			n := 0
			for _, p := range prods {
				if p.Slug == "" || p.Nom == "" {
					continue
				}
				specs, _ := json.Marshal(p.SpecList())
				pf, _ := json.Marshal(p.PointsForts)
				cons, _ := json.Marshal(p.Consommables)
				opts, _ := json.Marshal(p.Options)
				ext := p.Extras
				if ext == nil {
					ext = []string{}
				}
				extras, _ := json.Marshal(ext)
				var prix any
				if p.Prix != nil {
					prix = *p.Prix
				}
				devise := p.Devise
				if devise == "" {
					devise = "F CFA"
				}
				dispo := p.Dispo
				if dispo == "" {
					dispo = "en stock"
				}
				_, err := a.DB.Exec(`INSERT INTO products(slug,nom,marque,prix,devise,dispo,categorie,accroche,description,image,specs_json,pointsforts_json,consommables_json,options_json,store_url,ref,ref_fabricant,main,main_slug,sub,sub_slug,live_top,live_sub,search_index,images_json)
					VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
					ON CONFLICT(slug) DO UPDATE SET nom=excluded.nom, marque=excluded.marque, prix=excluded.prix, devise=excluded.devise, dispo=excluded.dispo, categorie=excluded.categorie, accroche=excluded.accroche, description=excluded.description, image=excluded.image, specs_json=excluded.specs_json, pointsforts_json=excluded.pointsforts_json, consommables_json=excluded.consommables_json, options_json=excluded.options_json, store_url=excluded.store_url, ref=excluded.ref, ref_fabricant=excluded.ref_fabricant, main=excluded.main, main_slug=excluded.main_slug, sub=excluded.sub, sub_slug=excluded.sub_slug, live_top=excluded.live_top, live_sub=excluded.live_sub, search_index=excluded.search_index,
					images_json=CASE WHEN excluded.images_json<>'[]' THEN excluded.images_json ELSE products.images_json END`,
					p.Slug, p.Nom, p.Marque, prix, devise, dispo, p.Categorie, p.Accroche, p.Description, p.Image, string(specs), string(pf), string(cons), string(opts), p.StoreURL, p.Ref, p.RefFabricant, p.Main, p.MainSlug, p.Sub, p.SubSlug, p.LiveTop, p.LiveSub, catalog.SearchIndex(p), string(extras))
				if err == nil {
					n++
					marexdb.SyncFTS(a.DB, p.Slug)
				}
			}
			a.reload()
			a.auditAdmin(r, "outils.importer", fmt.Sprintf("%d produits", n))
			http.Redirect(w, r, "/admin/outils?vue=importer&msg="+strconv.Itoa(n)+"+produits+import%C3%A9s", http.StatusSeeOther)
			return
		}
		inner.WriteString(`<form method="post" enctype="multipart/form-data" class="mx-form"><input type="hidden" name="csrf" value="` + render.Esc(a.adminCSRF()) + `"><label>Fichier JSON de produits <input type="file" name="fichier" accept=".json" required></label><button class="mx-btn" type="submit">Importer</button></form><p>Même format que les fichiers catalogue (liste de produits). Les produits existants sont mis à jour par slug.</p>`)
	case "exporter":
		switch r.URL.Query().Get("format") {
		case "produits-csv":
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			w.Header().Set("Content-Disposition", "attachment; filename=produits.csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"slug", "nom", "marque", "categorie", "prix", "disponibilite", "reference"})
			for _, p := range a.Store.Products {
				prix := ""
				if p.PrixAffiche() != nil {
					prix = strconv.FormatInt(*p.PrixAffiche(), 10)
				}
				_ = cw.Write([]string{p.Slug, p.Nom, p.Marque, p.Categorie, prix, p.Dispo, p.RefFabricant})
			}
			cw.Flush()
			return
		case "commandes-csv":
			w.Header().Set("Content-Type", "text/csv; charset=utf-8")
			w.Header().Set("Content-Disposition", "attachment; filename=commandes.csv")
			cw := csv.NewWriter(w)
			_ = cw.Write([]string{"id", "date", "statut", "total", "paiement"})
			rows, err := a.DB.Query(`SELECT id,created_at,statut,COALESCE(total,0),moyen_paiement FROM orders ORDER BY id`)
			if err == nil {
				defer rows.Close()
				for rows.Next() {
					var id, total int64
					var date, st, pay string
					_ = rows.Scan(&id, &date, &st, &total, &pay)
					_ = cw.Write([]string{strconv.FormatInt(id, 10), date, st, strconv.FormatInt(total, 10), pay})
				}
			}
			cw.Flush()
			return
		case "catalogue-json":
			w.Header().Set("Content-Type", "application/json")
			w.Header().Set("Content-Disposition", "attachment; filename=catalogue.json")
			_ = json.NewEncoder(w).Encode(a.Store.Products)
			return
		}
		inner.WriteString(`<p><a class="mx-btn" href="/admin/outils?vue=exporter&format=produits-csv">Produits (CSV)</a> <a class="mx-btn" href="/admin/outils?vue=exporter&format=commandes-csv">Commandes (CSV)</a> <a class="mx-btn" href="/admin/outils?vue=exporter&format=catalogue-json">Catalogue (JSON)</a></p>`)
	case "redirections":
		if r.Method == http.MethodPost {
			if !a.checkAdminCSRF(r) {
				http.Error(w, "refusé", http.StatusForbidden)
				return
			}
			src := catalog.Slugify(r.FormValue("source"))
			dst := catalog.Slugify(r.FormValue("cible"))
			if src != "" && dst != "" && src != dst {
				custom := a.customRedirects()
				custom[src] = dst
				b, _ := json.Marshal(custom)
				_, _ = a.DB.Exec(`INSERT INTO settings(nom,valeur) VALUES('redirections',?) ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`, string(b))
				a.auditAdmin(r, "outils.redirection", src+" -> "+dst)
			}
			http.Redirect(w, r, "/admin/outils?vue=redirections&msg=Redirection+enregistr%C3%A9e", http.StatusSeeOther)
			return
		}
		var tr [][]string
		for src, dst := range slugRedirects {
			tr = append(tr, []string{render.Esc(src), render.Esc(dst), "Fusion de fiches (intégrée)"})
		}
		for src, dst := range a.customRedirects() {
			tr = append(tr, []string{render.Esc(src), render.Esc(dst), "Personnalisée"})
		}
		sort.Slice(tr, func(i, j int) bool { return tr[i][0] < tr[j][0] })
		inner.WriteString(`<form method="post" class="mx-form"><input type="hidden" name="csrf" value="` + render.Esc(a.adminCSRF()) + `"><label>Ancien slug <input name="source" required></label><label>Slug cible <input name="cible" required></label><button class="mx-btn" type="submit">Ajouter la redirection 301</button></form>`)
		inner.WriteString(mxTable([]string{"Source", "Cible", "Origine"}, tr))
	default:
		inner.WriteString(`<div class="mx-kpis">
<div class="mx-kpi"><div class="lb">Importer</div><p>Ajouter ou mettre à jour des produits depuis un fichier JSON.</p><a class="mx-btn" href="/admin/outils?vue=importer">Ouvrir</a></div>
<div class="mx-kpi"><div class="lb">Exporter</div><p>Produits et commandes en CSV, catalogue en JSON.</p><a class="mx-btn" href="/admin/outils?vue=exporter">Ouvrir</a></div>
<div class="mx-kpi"><div class="lb">Redirections</div><p>Redirections 301 des anciennes fiches fusionnées.</p><a class="mx-btn" href="/admin/outils?vue=redirections">Ouvrir</a></div>
</div>`)
	}
	a.writeAdmin(w, r, "Outils", "outils", inner.String())
}

func (a *App) customRedirects() map[string]string {
	out := map[string]string{}
	var v string
	if err := a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom='redirections'`).Scan(&v); err == nil && v != "" {
		_ = json.Unmarshal([]byte(v), &out)
	}
	return out
}
