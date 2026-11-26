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
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	marexdb "github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

func (a *App) writeAdmin(w http.ResponseWriter, r *http.Request, title, active, inner string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(a.adminPage(title, active, inner, r)))
}

func mxTable(headers []string, rows [][]string) string {
	return mxTableAlign(headers, rows, "")
}

// mxTableAlign : align vaut "lrr" pour aligner les colonnes numériques à droite.
func mxTableAlign(headers []string, rows [][]string, align string) string {
	cell := func(i int) string {
		if i < len(align) && align[i] == 'r' {
			return ` class="num"`
		}
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="mx-tablewrap mx-table-scroll"><table class="mx-table"><thead><tr>`)
	for i, h := range headers {
		b.WriteString(`<th` + cell(i) + `>` + render.Esc(h) + `</th>`)
	}
	b.WriteString(`</tr></thead><tbody>`)
	for _, row := range rows {
		b.WriteString(`<tr>`)
		for i, c := range row {
			b.WriteString(`<td` + cell(i) + `>` + c + `</td>`)
		}
		b.WriteString(`</tr>`)
	}
	b.WriteString(`</tbody></table></div>`)
	return b.String()
}

func mxEmpty(msg string) string {
	return `<div class="mx-empty">` + render.Esc(msg) + `</div>`
}

// ---------- Structure des modules (en-tête, onglets, sections) ----------

func mxPageHead(title, lead string) string {
	h := `<div class="mx-pagehead"><h2>` + render.Esc(title) + `</h2>`
	if lead != "" {
		h += `<p>` + render.Esc(lead) + `</p>`
	}
	return h + `</div>`
}

func mxSection(head, sub, body string) string {
	h := `<section class="mx-section"><div class="mx-section-head"><h3>` + render.Esc(head) + `</h3>`
	if sub != "" {
		h += `<p class="sub">` + render.Esc(sub) + `</p>`
	}
	return h + `</div>` + body + `</section>`
}

func mxCard(body string) string {
	return `<div class="mx-card">` + body + `</div>`
}

const mxIcoInfo = `<svg viewBox="0 0 24 24" width="24" height="24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="12" cy="12" r="9"/><path d="M12 16v-4M12 8h.01"/></svg>`

func mxEmptyState(title, text string) string {
	return `<div class="mx-empty"><div class="mx-empty-ic">` + mxIcoInfo + `</div><strong>` + render.Esc(title) + `</strong>` + render.Esc(text) + `</div>`
}

const (
	icCaisse  = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><rect x="2" y="7" width="20" height="14" rx="2"/><path d="M16 7V5a2 2 0 0 0-2-2H10a2 2 0 0 0-2 2v2"/></svg>`
	icRevenu  = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6"/></svg>`
	icCmd     = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M6 2 3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><path d="M3 6h18M16 10a4 4 0 0 1-8 0"/></svg>`
	icProd    = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="m7.5 4.27 9 5.15"/><path d="M21 8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><path d="M3.3 7 12 12l8.7-5M12 22V12"/></svg>`
	icStock   = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20 12V8H4v4M4 8l1-4h14l1 4M4 8v8a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8"/></svg>`
	icTag     = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M20.6 13.4 12 22l-9-9V3h10l7.6 7.6a2 2 0 0 1 0 2.8z"/><circle cx="7.5" cy="7.5" r="1.5"/></svg>`
	icPercent = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M19 5 5 19"/><circle cx="6.5" cy="6.5" r="2.5"/><circle cx="17.5" cy="17.5" r="2.5"/></svg>`
	icDown    = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M21 15v4a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2v-4"/><path d="m7 10 5 5 5-5"/><path d="M12 15V3"/></svg>`
	icAlert   = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2"><path d="M10.3 3.9 1.8 18a2 2 0 0 0 1.7 3h17a2 2 0 0 0 1.7-3L13.7 3.9a2 2 0 0 0-3.4 0z"/><path d="M12 9v4M12 17h.01"/></svg>`
)

// mxKPI : carte d'indicateur (icône + libellé + valeur). Les montants « F CFA »
// passent l'unité à la ligne pour ne jamais être rognés.
func mxKPI(label, value, color, icon string) string {
	main, unit := value, ""
	if strings.HasSuffix(value, " F CFA") {
		main, unit = strings.TrimSuffix(value, " F CFA"), "F CFA"
	}
	unitHTML, numCls := "", ""
	if unit != "" {
		unitHTML = `<span class="unit">` + unit + `</span>`
		numCls = " num"
	}
	return `<div class="mx-kpi" style="--mx-kpi:` + color + `">` +
		`<div class="mx-kpi-ic">` + icon + `</div>` +
		`<div class="mx-kpi-tx"><h3 class="lb">` + render.Esc(label) + `</h3>` +
		`<div class="vl big` + numCls + `">` + render.Esc(main) + unitHTML + `</div></div></div>`
}

func mxKPIs(cards ...string) string {
	return `<div class="mx-kpis">` + strings.Join(cards, "") + `</div>`
}

// Libellés lisibles pour les valeurs techniques des commandes.
var mxStatutLabels = map[string]string{
	"devis": "Devis", "whatsapp_recu": "WhatsApp reçu", "attente_momo": "Attente MoMo",
	"paye": "Payé", "prepare": "Préparé", "livre": "Livré", "annule": "Annulé",
}
var mxPaiementLabels = map[string]string{
	"whatsapp": "WhatsApp", "momo": "Mobile Money", "cash": "Espèces",
	"wave": "Wave", "virement": "Virement", "carte": "Carte bancaire",
}

func mxLabel(m map[string]string, v string) string {
	if lb, ok := m[v]; ok {
		return lb
	}
	return v
}

// mxDate : "2026-11-16 18:10:52" -> "16/11/2026 18:10" ; "2026-12-31" -> "31/12/2026".
func mxDate(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 10 && s[4] == '-' && s[7] == '-' {
		out := s[8:10] + "/" + s[5:7] + "/" + s[0:4]
		if len(s) >= 16 && s[10] == ' ' {
			out += " " + s[11:16]
		}
		return out
	}
	return s
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
	lead := fmt.Sprintf("%d marques référencées dans le catalogue.", len(brands))
	avecLogo, sansMarque := 0, 0
	for _, b := range brands {
		if logos[catalog.Slugify(b.k)] {
			avecLogo++
		}
		if strings.TrimSpace(b.k) == "" || catalog.NormName(b.k) == "sans marque" {
			sansMarque += b.n
		}
	}
	inner := mxPageHead("Marques", lead) +
		mxKPIs(
			mxKPI("Marques", strconv.Itoa(len(brands)), "#2271b1", icTag),
			mxKPI("Produits", strconv.Itoa(len(a.Store.Products)), "#00a32a", icProd),
			mxKPI("Avec logo", strconv.Itoa(avecLogo), "#9b51e0", icStock),
			mxKPI("Sans marque", strconv.Itoa(sansMarque), "#dba617", icAlert),
		) +
		mxSection("Répartition par marque", "Produits publiés par marque", mxCard(mxTableAlign([]string{"Logo", "Marque", "Produits"}, rows, "llr")))
	a.writeAdmin(w, r, "Produits", "produits/marques", inner)
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
	plusFourni := "—"
	if len(tags) > 0 {
		plusFourni = tags[0].k
	}
	inner := mxPageHead("Étiquettes", "Les étiquettes correspondent aux types d'équipement du catalogue.") +
		mxKPIs(
			mxKPI("Étiquettes", strconv.Itoa(len(tags)), "#2271b1", icTag),
			mxKPI("Produits classés", strconv.Itoa(len(a.Store.Products)), "#00a32a", icProd),
			mxKPI("Type le plus fourni", plusFourni, "#dba617", icAlert),
		) +
		mxSection("Types d'équipement", fmt.Sprintf("%d étiquettes", len(tags)), mxCard(mxTableAlign([]string{"Étiquette", "Produits"}, rows, "lr")))
	a.writeAdmin(w, r, "Produits", "produits/etiquettes", inner)
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
	totalValeurs := 0
	for _, at := range attrs {
		totalValeurs += len(vals[at.k])
	}
	inner := mxPageHead("Attributs", "Les attributs sont les caractéristiques présentes sur les fiches produits.") +
		mxKPIs(
			mxKPI("Attributs", strconv.Itoa(len(attrs)), "#2271b1", icTag),
			mxKPI("Valeurs distinctes", strconv.Itoa(totalValeurs), "#00a32a", icProd),
			mxKPI("Produits couverts", strconv.Itoa(len(a.Store.Products)), "#9b51e0", icStock),
		) +
		mxSection("Caractéristiques produit", fmt.Sprintf("%d attributs", len(attrs)), mxCard(mxTableAlign([]string{"Attribut", "Produits", "Valeurs distinctes"}, rows, "lrr")))
	a.writeAdmin(w, r, "Produits", "produits/attributs", inner)
}

// ---------- Produits > Avis ----------

func (a *App) AdminAvis(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	inner := mxPageHead("Avis", "Les avis clients seront affichés ici une fois activés.") +
		mxCard(mxEmptyState("Aucun avis pour le moment", "Les avis clients ne sont pas encore activés sur la boutique."))
	a.writeAdmin(w, r, "Produits", "produits/avis", inner)
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
	fcfa := func(v int64) string { return catalog.FmtPrix(&v) }
	var inner strings.Builder
	switch vue {
	case "produits":
		nbVendus := a.statInt(`SELECT COUNT(DISTINCT product_id) FROM order_items`)
		qteTotal := a.statInt(`SELECT COALESCE(SUM(qty),0) FROM order_items`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Produits au catalogue", strconv.Itoa(len(a.Store.Products)), "#9b51e0", icStock))
		inner.WriteString(mxKPI("Références vendues", strconv.FormatInt(nbVendus, 10), "#2271b1", icProd))
		inner.WriteString(mxKPI("Quantité totale vendue", strconv.FormatInt(qteTotal, 10), "#00a32a", icCmd))
		inner.WriteString(`</div>`)
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
			inner.WriteString(mxSection("Top produits vendus", "20 premiers par quantité", mxCard(mxTableAlign([]string{"Produit", "Quantité vendue", "Revenu"}, tr, "lrr"))))
		}
	case "revenu":
		revTotal := a.statInt(`SELECT COALESCE(SUM(total),0) FROM orders WHERE statut<>'annule'`)
		revMois := a.statInt(`SELECT COALESCE(SUM(total),0) FROM orders WHERE statut<>'annule' AND created_at >= datetime('now','start of month')`)
		revPanier := a.statInt(`SELECT COALESCE(AVG(total),0) FROM orders WHERE statut<>'annule'`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Revenu total", fcfa(revTotal), "#2271b1", icCaisse))
		inner.WriteString(mxKPI("Revenu ce mois-ci", fcfa(revMois), "#00a32a", icRevenu))
		inner.WriteString(mxKPI("Panier moyen", fcfa(revPanier), "#dba617", icCmd))
		inner.WriteString(`</div>`)
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
			inner.WriteString(mxSection("Revenu par mois", "24 derniers mois", mxCard(mxTableAlign([]string{"Mois", "Commandes", "Revenu"}, tr, "lrr"))))
		}
	case "commandes":
		cmdsTotal := a.statInt(`SELECT COUNT(*) FROM orders`)
		cmds7j := a.statInt(`SELECT COUNT(*) FROM orders WHERE created_at >= datetime('now','-7 days')`)
		cmdMontant := a.statInt(`SELECT COALESCE(SUM(total),0) FROM orders WHERE statut<>'annule'`)
		aTraiter := a.statInt(`SELECT COUNT(*) FROM orders WHERE statut NOT IN ('livre','annule')`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Commandes totales", strconv.FormatInt(cmdsTotal, 10), "#2271b1", icCmd))
		inner.WriteString(mxKPI("7 derniers jours", strconv.FormatInt(cmds7j, 10), "#00a32a", icCmd))
		inner.WriteString(mxKPI("Montant total", fcfa(cmdMontant), "#dba617", icCaisse))
		inner.WriteString(mxKPI("À traiter", strconv.FormatInt(aTraiter, 10), "#d63638", icAlert))
		inner.WriteString(`</div>`)
		rows, err := a.DB.Query(`SELECT statut, COUNT(*), COALESCE(SUM(total),0) FROM orders GROUP BY statut ORDER BY COUNT(*) DESC`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var st string
				var n, rev int64
				_ = rows.Scan(&st, &n, &rev)
				tr = append(tr, []string{render.Esc(mxLabel(mxStatutLabels, st)), strconv.FormatInt(n, 10), fcfa(rev)})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucune commande pour le moment."))
		} else {
			inner.WriteString(mxSection("Par statut", "Répartition des commandes", mxCard(mxTableAlign([]string{"Statut", "Commandes", "Montant"}, tr, "lrr"))))
		}
		prows, perr := a.DB.Query(`SELECT moyen_paiement, COUNT(*), COALESCE(SUM(total),0) FROM orders GROUP BY moyen_paiement ORDER BY COUNT(*) DESC`)
		var ptr [][]string
		if perr == nil {
			defer prows.Close()
			for prows.Next() {
				var moy string
				var n, rev int64
				_ = prows.Scan(&moy, &n, &rev)
				ptr = append(ptr, []string{render.Esc(mxLabel(mxPaiementLabels, moy)), strconv.FormatInt(n, 10), fcfa(rev)})
			}
		}
		if len(ptr) > 0 {
			inner.WriteString(mxSection("Par moyen de paiement", "Volume et montant par canal", mxCard(mxTableAlign([]string{"Moyen de paiement", "Commandes", "Montant"}, ptr, "lrr"))))
		}
		orows, oerr := a.DB.Query(`SELECT id, created_at, statut, COALESCE(total,0), moyen_paiement FROM orders ORDER BY id DESC LIMIT 8`)
		var otr [][]string
		if oerr == nil {
			defer orows.Close()
			for orows.Next() {
				var id, total int64
				var date, st, moy string
				_ = orows.Scan(&id, &date, &st, &total, &moy)
				otr = append(otr, []string{
					`<a href="/admin/commandes/` + strconv.FormatInt(id, 10) + `">#` + strconv.FormatInt(id, 10) + `</a>`,
					render.Esc(mxDate(date)), render.Esc(mxLabel(mxStatutLabels, st)), fcfa(total), render.Esc(mxLabel(mxPaiementLabels, moy)),
				})
			}
		}
		if len(otr) > 0 {
			inner.WriteString(mxSection("Dernières commandes", "8 commandes les plus récentes", mxCard(mxTableAlign([]string{"Commande", "Date", "Statut", "Montant", "Paiement"}, otr, "lllrl"))))
		}
	case "categories":
		nbCats := a.statInt(`SELECT COUNT(DISTINCT main) FROM products`)
		nbTypes := a.statInt(`SELECT COUNT(DISTINCT sub) FROM products`)
		qteVendue := a.statInt(`SELECT COALESCE(SUM(qty),0) FROM order_items`)
		revCat := a.statInt(`SELECT COALESCE(SUM(qty*prix_unitaire),0) FROM order_items`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Catégories", strconv.FormatInt(nbCats, 10), "#2271b1", icProd))
		inner.WriteString(mxKPI("Types d'équipement", strconv.FormatInt(nbTypes, 10), "#9b51e0", icTag))
		inner.WriteString(mxKPI("Quantité vendue", strconv.FormatInt(qteVendue, 10), "#00a32a", icCmd))
		inner.WriteString(mxKPI("Revenu des ventes", fcfa(revCat), "#dba617", icCaisse))
		inner.WriteString(`</div>`)
		rows, err := a.DB.Query(`SELECT p.main, COUNT(DISTINCT p.id), COUNT(DISTINCT CASE WHEN p.dispo='en stock' THEN p.id END), COALESCE(SUM(oi.qty),0), COALESCE(SUM(oi.qty*oi.prix_unitaire),0) FROM products p LEFT JOIN order_items oi ON oi.product_id=p.id GROUP BY p.main ORDER BY COUNT(DISTINCT p.id) DESC`)
		var tr [][]string
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var cat string
				var n, dispo, q, rev int64
				_ = rows.Scan(&cat, &n, &dispo, &q, &rev)
				tr = append(tr, []string{mxCatBadge(cat), strconv.FormatInt(n, 10), mxStockBar(dispo, n), mxNumCell(q), mxMoneyCell(rev, fcfa(rev))})
			}
		}
		if len(tr) == 0 {
			inner.WriteString(mxEmpty("Aucune catégorie dans le catalogue."))
		} else {
			inner.WriteString(mxSection("Catalogue et ventes par catégorie", "Produits publiés et volumes vendus", mxCard(mxTableAlign([]string{"Catégorie", "Produits", "En stock", "Quantité vendue", "Revenu"}, tr, "lrrrr"))))
		}
		srows, serr := a.DB.Query(`SELECT main, sub, COUNT(*) FROM products GROUP BY main, sub ORDER BY main, COUNT(*) DESC`)
		type typeRow struct {
			sub string
			n   int64
		}
		groups := map[string][]typeRow{}
		var order []string
		if serr == nil {
			defer srows.Close()
			for srows.Next() {
				var main, sub string
				var n int64
				_ = srows.Scan(&main, &sub, &n)
				if _, ok := groups[main]; !ok {
					order = append(order, main)
				}
				groups[main] = append(groups[main], typeRow{sub, n})
			}
		}
		if len(order) > 0 {
			var b strings.Builder
			b.WriteString(`<div class="mx-accordion">`)
			for _, main := range order {
				rws := groups[main]
				var total int64
				for _, rr := range rws {
					total += rr.n
				}
				typeLbl := "types"
				if len(rws) == 1 {
					typeLbl = "type"
				}
				b.WriteString(`<details class="mx-acc-item"><summary>` + mxCatTitle(main) +
					`<span class="mx-cat-count" style="--bc:` + mxBadgeColor(main) + `">` + strconv.FormatInt(total, 10) + `</span>` +
					`<span class="mx-acc-meta">` + strconv.Itoa(len(rws)) + ` ` + typeLbl + `</span></summary>`)
				var tr [][]string
				for _, rr := range rws {
					tr = append(tr, []string{render.Esc(rr.sub), strconv.FormatInt(rr.n, 10)})
				}
				b.WriteString(mxTableAlign([]string{"Type", "Produits"}, tr, "lr"))
				b.WriteString(`</details>`)
			}
			b.WriteString(`</div>`)
			inner.WriteString(mxSection("Répartition par type", "Sous-catégories du catalogue", mxCard(b.String())))
		}
	case "stock":
		enStock := a.statInt(`SELECT COUNT(*) FROM products WHERE dispo='en stock'`)
		rupture := a.statInt(`SELECT COUNT(*) FROM products WHERE dispo='rupture de stock'`)
		faible := a.statInt(`SELECT COUNT(*) FROM products WHERE stock IS NOT NULL AND stock<=2`)
		sansPrix := a.statInt(`SELECT COUNT(*) FROM products WHERE prix IS NULL AND prix_boutique IS NULL`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Produits en stock", strconv.FormatInt(enStock, 10), "#00a32a", icStock))
		inner.WriteString(mxKPI("En rupture", strconv.FormatInt(rupture, 10), "#d63638", icAlert))
		inner.WriteString(mxKPI("Stock ≤ 2", strconv.FormatInt(faible, 10), "#dba617", icProd))
		inner.WriteString(mxKPI("Sur devis (sans prix)", strconv.FormatInt(sansPrix, 10), "#9b51e0", icCaisse))
		inner.WriteString(`</div>`)
		rows, err := a.DB.Query(`SELECT slug,nom,dispo,stock FROM products WHERE dispo<>'en stock' OR (stock IS NOT NULL AND stock<=2) ORDER BY dispo, nom LIMIT 300`)
		type stockRow struct {
			slug, nom, dispo string
			stock            *int64
		}
		groupKey := func(dispo string, stock *int64) string {
			if dispo == "rupture de stock" {
				return "Rupture de stock"
			}
			if stock != nil && *stock <= 2 {
				return "Stock faible (≤ 2)"
			}
			return "Autres"
		}
		order := []string{"Rupture de stock", "Stock faible (≤ 2)", "Autres"}
		groups := map[string][]stockRow{}
		if err == nil {
			defer rows.Close()
			for rows.Next() {
				var slug, nom, dispo string
				var stock *int64
				_ = rows.Scan(&slug, &nom, &dispo, &stock)
				k := groupKey(dispo, stock)
				groups[k] = append(groups[k], stockRow{slug, nom, dispo, stock})
			}
		}
		total := 0
		for _, k := range order {
			total += len(groups[k])
		}
		if total == 0 {
			inner.WriteString(mxEmpty("Aucun produit en rupture ni en stock faible."))
		} else {
			var b strings.Builder
			b.WriteString(`<div class="mx-accordion">`)
			opened := false
			for _, k := range order {
				rws := groups[k]
				if len(rws) == 0 {
					continue
				}
				openAttr := ""
				if !opened {
					openAttr = " open"
					opened = true
				}
				prodLbl := "produits"
				if len(rws) == 1 {
					prodLbl = "produit"
				}
				col := mxStockGroupColor(k)
				b.WriteString(`<details class="mx-acc-item"` + openAttr + `><summary>` + mxGroupTitle(k, col) + mxGroupCount(len(rws), col) + `<span class="mx-acc-meta">` + strconv.Itoa(len(rws)) + ` ` + prodLbl + `</span></summary>`)
				var tr [][]string
				for _, rr := range rws {
					tr = append(tr, []string{`<a href="/admin/produits/` + render.Esc(rr.slug) + `">` + render.Esc(rr.nom) + `</a>`, mxDispoBadge(rr.dispo), mxStockCell(rr.stock)})
				}
				b.WriteString(`<div class="mx-paged" data-per="10">` + mxTableAlign([]string{"Produit", "Disponibilité", "Stock"}, tr, "lrl") + `</div>`)
				b.WriteString(`</details>`)
			}
			b.WriteString(`</div>`)
			inner.WriteString(mxSection("Produits à surveiller", "Ruptures et stocks faibles", mxCard(b.String())))
		}
	case "variations":
		nbOpts, nbProds := 0, 0
		var vtr [][]string
		for _, prod := range a.Store.Products {
			names := catalog.OptionNames(prod)
			if len(names) == 0 {
				continue
			}
			nbProds++
			nbOpts += len(names)
			vtr = append(vtr, []string{
				`<a href="/admin/produits/` + render.Esc(prod.Slug) + `">` + render.Esc(prod.Nom) + `</a>`,
				mxOptCount(len(names)),
				mxChips(names),
			})
		}
		sort.Slice(vtr, func(i, j int) bool { return vtr[i][0] < vtr[j][0] })
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Produits avec options", strconv.Itoa(nbProds), "#2271b1", icTag))
		inner.WriteString(mxKPI("Options proposées", strconv.Itoa(nbOpts), "#00a32a", icProd))
		inner.WriteString(`</div>`)
		inner.WriteString(`<p style="color:#646970;margin:6px 0 12px">Les options (capacités, variantes) sont affichées sur les fiches ; elles ne sont pas encore commandables séparément (prix sur demande).</p>`)
		if len(vtr) == 0 {
			inner.WriteString(mxEmpty("Aucun produit avec options."))
		} else {
			inner.WriteString(mxSection("Options par produit", "Capacités et variantes affichées sur les fiches", mxCard(mxTableAlign([]string{"Produit", "Options", "Valeurs"}, vtr, "lrl"))))
		}
	case "promos":
		a.ensurePromos()
		actifs := a.statInt(`SELECT COUNT(*) FROM promos WHERE actif=1`)
		total := a.statInt(`SELECT COUNT(*) FROM promos`)
		usages := a.statInt(`SELECT COALESCE(SUM(usage_count),0) FROM promos`)
		expires := a.statInt(`SELECT COUNT(*) FROM promos WHERE expire<>'' AND expire < date('now')`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Codes actifs", strconv.FormatInt(actifs, 10), "#00a32a", icTag))
		inner.WriteString(mxKPI("Codes au total", strconv.FormatInt(total, 10), "#2271b1", icTag))
		inner.WriteString(mxKPI("Utilisations", strconv.FormatInt(usages, 10), "#dba617", icCmd))
		inner.WriteString(mxKPI("Expirés", strconv.FormatInt(expires, 10), "#d63638", icAlert))
		inner.WriteString(`</div>`)
		prows, err := a.DB.Query(`SELECT code,type,valeur,usage_count,limite,expire,actif FROM promos ORDER BY code`)
		var ptr [][]string
		if err == nil {
			defer prows.Close()
			for prows.Next() {
				var code, typ, exp string
				var val, usage int64
				var limite *int64
				var actif int
				_ = prows.Scan(&code, &typ, &val, &usage, &limite, &exp, &actif)
				valLb := fcfa(val)
				if typ == "pourcentage" {
					valLb = strconv.FormatInt(val, 10) + " %"
				}
				lim := "∞"
				if limite != nil {
					lim = strconv.FormatInt(*limite, 10)
				}
				if exp == "" {
					exp = "—"
				} else {
					exp = mxDate(exp)
				}
				etat := "Actif"
				if actif == 0 {
					etat = "Inactif"
				} else if exp != "—" && exp < time.Now().Format("2006-01-02") {
					etat = "Expiré"
				}
				ptr = append(ptr, []string{render.Esc(code), render.Esc(valLb), strconv.FormatInt(usage, 10) + " / " + lim, render.Esc(exp), render.Esc(etat)})
			}
		}
		inner.WriteString(`<p style="color:#646970;margin:6px 0 12px">Gestion dans <a href="/admin/marketing/promos">Marketing &gt; Codes promo</a>. Les codes ne sont pas encore appliqués au panier : les utilisations restent à 0.</p>`)
		if len(ptr) == 0 {
			inner.WriteString(mxEmpty("Aucun code promo créé."))
		} else {
			inner.WriteString(mxSection("Codes promo", "État et utilisation", mxCard(mxTableAlign([]string{"Code", "Valeur", "Usage / Limite", "Expiration", "État"}, ptr, "lrlll"))))
		}
	case "taxes":
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Taux configuré", "0 %", "#2271b1", icPercent))
		inner.WriteString(mxKPI("Régime d'affichage", "TTC", "#00a32a", icCaisse))
		inner.WriteString(mxKPI("Règles de taxe", "0", "#dba617", icTag))
		inner.WriteString(`</div>`)
		inner.WriteString(mxEmpty("Aucune taxe configurée : les prix sont affichés TTC et aucune TVA n'est calculée à la commande."))
	case "telechargements":
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Produits téléchargeables", "0", "#2271b1", icDown))
		inner.WriteString(mxKPI("Téléchargements", "0", "#00a32a", icDown))
		inner.WriteString(`</div>`)
		inner.WriteString(mxEmpty("Aucun produit numérique dans le catalogue : tout est vendu en matériel ou sur devis."))
	default:
		ventes := a.statInt(`SELECT COALESCE(SUM(total),0) FROM orders`)
		ventesNettes := a.statInt(`SELECT COALESCE(SUM(total),0) FROM orders WHERE statut<>'annule'`)
		cmds := a.statInt(`SELECT COUNT(*) FROM orders`)
		vendus := a.statInt(`SELECT COALESCE(SUM(qty),0) FROM order_items`)
		inner.WriteString(`<div class="mx-kpis">`)
		inner.WriteString(mxKPI("Total des ventes", fcfa(ventes), "#2271b1", icCaisse))
		inner.WriteString(mxKPI("Ventes nettes", fcfa(ventesNettes), "#00a32a", icRevenu))
		inner.WriteString(mxKPI("Commandes", strconv.FormatInt(cmds, 10), "#d63638", icCmd))
		inner.WriteString(mxKPI("Produits vendus", strconv.FormatInt(vendus, 10), "#dba617", icProd))
		inner.WriteString(mxKPI("Produits au catalogue", strconv.Itoa(len(a.Store.Products)), "#9b51e0", icStock))
		inner.WriteString(`</div>`)
		if cmds == 0 {
			inner.WriteString(mxEmpty("Aucune donnée pour la période : aucune commande enregistrée."))
		}
	}
	labels := map[string]string{
		"": "Vue d'ensemble", "produits": "Produits", "revenu": "Revenu",
		"commandes": "Commandes", "variations": "Variations", "categories": "Catégories",
		"promos": "Codes promo", "taxes": "Taxes", "telechargements": "Téléchargements", "stock": "Stock",
	}
	leads := map[string]string{
		"":                "Chiffre d'affaires, commandes et volume de produits en un coup d'œil.",
		"produits":        "Classement des produits les plus vendus.",
		"revenu":          "Revenu total, revenu du mois et panier moyen, puis détail par mois.",
		"commandes":       "Volume, montant et suivi des commandes par statut et par paiement.",
		"variations":      "Options proposées sur les fiches produits.",
		"categories":      "Catalogue et ventes par catégorie, puis répartition par type.",
		"promos":          "Codes promo créés et leur état d'utilisation.",
		"taxes":           "Configuration fiscale appliquée à la boutique.",
		"telechargements": "Produits numériques et téléchargements.",
		"stock":           "Disponibilité et niveaux de stock à surveiller.",
	}
	label, lead := labels[vue], leads[vue]
	if label == "" {
		label, lead = "Vue d'ensemble", leads[""]
	}
	full := mxPageHead(label, lead) + inner.String()
	a.writeAdmin(w, r, "Statistiques", "stats", full)
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
	inner := mxPageHead("Vue d'ensemble", "Codes promo et opérations marketing de la boutique.") +
		`<div class="mx-kpis"><div class="mx-kpi" style="--mx-kpi:#00a32a"><div class="mx-kpi-ic">` + mxIcoInfo + `</div><div class="mx-kpi-tx"><h3 class="lb">Codes promo actifs</h3><div class="vl big">` + strconv.FormatInt(n, 10) + `</div></div></div></div>` +
		mxSection("Codes promo", "Créer, suivre et supprimer les codes", mxCard(`<div class="mx-card-pad"><a class="mx-btn" href="/admin/marketing/promos">Gérer les codes promo</a><p style="color:#646970;margin:10px 0 0">Les codes ne sont pas encore appliqués au panier.</p></div>`))
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
	inner.WriteString(mxPageHead("Codes promo", "Créer une remise fixe ou en pourcentage, avec limite d'usage et date d'expiration."))
	inner.WriteString(mxSection("Nouveau code", "Un code existant est mis à jour s'il porte le même nom", mxCard(`<form method="post" class="mx-form"><input type="hidden" name="csrf" value="`+render.Esc(csrf)+`"><input type="hidden" name="action" value="ajouter">
<label>Code <input name="code" required></label>
<label>Type <select name="type"><option value="fixe">Remise panier fixe (F CFA)</option><option value="pourcentage">Remise en pourcentage (%)</option></select></label>
<label>Valeur <input name="valeur" type="number" min="1" required></label>
<label>Limite d'usage <input name="limite" type="number" min="0" placeholder="∞"></label>
<label class="full">Description <input name="description"></label>
<label>Date d'expiration <input name="expire" type="date"></label>
<button class="mx-btn" type="submit">Ajouter le code</button></form>`)))
	if len(tr) == 0 {
		inner.WriteString(mxSection("Codes existants", "", mxCard(mxEmptyState("Aucun code promo", "Créez votre premier code avec le formulaire ci-dessus."))))
	} else {
		inner.WriteString(mxSection("Codes existants", fmt.Sprintf("%d codes", len(tr)), mxCard(mxTable([]string{"Code", "Type", "Valeur", "Description", "Usage / Limite", "Expiration", ""}, tr))))
	}
	a.writeAdmin(w, r, "Marketing", "marketing/promos", inner.String())
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
		form := `<form method="post" enctype="multipart/form-data" class="mx-form"><input type="hidden" name="csrf" value="` + render.Esc(a.adminCSRF()) + `"><label>Fichier JSON de produits <input type="file" name="fichier" accept=".json" required></label><button class="mx-btn" type="submit">Importer</button></form>`
		inner.WriteString(mxSection("Fichier JSON", "Même format que les fichiers catalogue (liste de produits)", mxCard(form+`<p style="padding:0 16px 16px;margin:0;color:#646970">Les produits existants sont mis à jour par slug ; les galeries gérées dans l'admin sont conservées.</p>`)))
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
		exp := `<div class="mx-card-pad"><a class="mx-btn" href="/admin/outils?vue=exporter&format=produits-csv">Produits (CSV)</a> <a class="mx-btn" href="/admin/outils?vue=exporter&format=commandes-csv">Commandes (CSV)</a> <a class="mx-btn" href="/admin/outils?vue=exporter&format=catalogue-json">Catalogue (JSON)</a><p style="margin:10px 0 0;color:#646970">Le CSV produits contient slug, nom, marque, catégorie, prix, disponibilité et référence. Le JSON reprend le catalogue complet.</p></div>`
		inner.WriteString(mxSection("Exports disponibles", "Téléchargements immédiats", mxCard(exp)))
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
		form := `<form method="post" class="mx-form"><input type="hidden" name="csrf" value="` + render.Esc(a.adminCSRF()) + `"><label>Ancien slug <input name="source" required></label><label>Slug cible <input name="cible" required></label><button class="mx-btn" type="submit">Ajouter la redirection 301</button></form>`
		inner.WriteString(mxSection("Ajouter une redirection", "Les deux slugs sont normalisés automatiquement", mxCard(form)))
		inner.WriteString(mxSection("Redirections actives", fmt.Sprintf("%d règles", len(tr)), mxCard(mxTable([]string{"Source", "Cible", "Origine"}, tr))))
	default:
		inner.WriteString(mxSection("Modules", "Trois outils pour administrer le catalogue", `<div class="mx-kpis">
<div class="mx-kpi"><div class="mx-kpi-tx"><h3 class="lb">Importer</h3><p style="margin:6px 0 12px;color:#646970">Ajouter ou mettre à jour des produits depuis un fichier JSON.</p><a class="mx-btn" href="/admin/outils?vue=importer">Ouvrir</a></div></div>
<div class="mx-kpi"><div class="mx-kpi-tx"><h3 class="lb">Exporter</h3><p style="margin:6px 0 12px;color:#646970">Produits et commandes en CSV, catalogue en JSON.</p><a class="mx-btn" href="/admin/outils?vue=exporter">Ouvrir</a></div></div>
<div class="mx-kpi"><div class="mx-kpi-tx"><h3 class="lb">Redirections</h3><p style="margin:6px 0 12px;color:#646970">Redirections 301 des anciennes fiches fusionnées.</p><a class="mx-btn" href="/admin/outils?vue=redirections">Ouvrir</a></div></div>
</div>`))
	}
	labels := map[string]string{"": "Outils disponibles", "importer": "Importer", "exporter": "Exporter", "redirections": "Redirections"}
	leads := map[string]string{
		"":             "Importer, exporter et gérer les redirections du catalogue.",
		"importer":     "Ajouter ou mettre à jour des produits depuis un fichier JSON.",
		"exporter":     "Télécharger les produits, les commandes ou le catalogue complet.",
		"redirections": "Rediriger les anciennes adresses vers les fiches actuelles.",
	}
	label, lead := labels[vue], leads[vue]
	if label == "" {
		label, lead = "Outils disponibles", leads[""]
	}
	full := mxPageHead(label, lead) + inner.String()
	a.writeAdmin(w, r, "Outils", "outils", full)
}

func (a *App) customRedirects() map[string]string {
	out := map[string]string{}
	var v string
	if err := a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom='redirections'`).Scan(&v); err == nil && v != "" {
		_ = json.Unmarshal([]byte(v), &out)
	}
	return out
}

var mxPalette = []string{"#2271b1", "#9b51e0", "#00a32a", "#dba617", "#d63638", "#00a3a3", "#e0519b", "#6b7280"}

func mxBadgeColor(s string) string {
	h := 0
	for _, c := range s {
		h += int(c)
	}
	return mxPalette[h%len(mxPalette)]
}

func mxCatBadge(s string) string {
	return `<span class="mx-catbadge" style="--bc:` + mxBadgeColor(s) + `">` + render.Esc(s) + `</span>`
}

// mxCatTitle : titre de catégorie en grand, coloré (même couleur que la pastille).
func mxCatTitle(s string) string {
	return mxGroupTitle(s, mxBadgeColor(s))
}

// mxGroupTitle : titre de groupe en grand avec une couleur imposée.
func mxGroupTitle(s, color string) string {
	return `<span class="mx-acc-cat" style="--bc:` + color + `">` + render.Esc(s) + `</span>`
}

// mxGroupCount : badge compteur coloré d'un groupe.
func mxGroupCount(n int, color string) string {
	return `<span class="mx-cat-count" style="--bc:` + color + `">` + strconv.Itoa(n) + `</span>`
}

func mxStockGroupColor(k string) string {
	switch k {
	case "Rupture de stock":
		return "#d63638"
	case "Stock faible (≤ 2)":
		return "#dba617"
	}
	return "#6b7280"
}

func mxStockBar(dispo, total int64) string {
	var pct int64
	if total > 0 {
		pct = dispo * 100 / total
	}
	return fmt.Sprintf(`<span class="mx-stockbar"><span class="mx-stockbar-fill" style="width:%d%%"></span></span><span class="mx-stockbar-txt">%d/%d</span>`, pct, dispo, total)
}

func mxNumCell(n int64) string {
	if n == 0 {
		return `<span class="mx-zero">0</span>`
	}
	return strconv.FormatInt(n, 10)
}

func mxMoneyCell(n int64, formatted string) string {
	if n == 0 {
		return `<span class="mx-zero">` + formatted + `</span>`
	}
	return formatted
}

func mxOptCount(n int) string {
	return `<span class="mx-opt-count">` + strconv.Itoa(n) + `</span>`
}

func mxChips(vals []string) string {
	const limit = 3
	var b strings.Builder
	b.WriteString(`<div class="mx-chips">`)
	for i, v := range vals {
		if i < limit {
			b.WriteString(`<span class="mx-chip">` + render.Esc(v) + `</span>`)
		}
	}
	if len(vals) > limit {
		b.WriteString(`<span class="mx-chips-extra">`)
		for _, v := range vals[limit:] {
			b.WriteString(`<span class="mx-chip">` + render.Esc(v) + `</span>`)
		}
		b.WriteString(`</span>`)
		b.WriteString(`<button type="button" class="mx-chip-more" onclick="this.parentElement.classList.toggle('open');this.textContent=this.parentElement.classList.contains('open')?'− réduire':'+` + strconv.Itoa(len(vals)-limit) + `'">+` + strconv.Itoa(len(vals)-limit) + `</button>`)
	}
	b.WriteString(`</div>`)
	return b.String()
}

// mxDispoBadge réutilise les couleurs de statut existantes (.mx-badge ok/warn/bad).
func mxDispoBadge(dispo string) string {
	cls := "mx-badge"
	switch dispo {
	case "en stock":
		cls += " ok"
	case "rupture de stock":
		cls += " bad"
	case "sur commande", "retour":
		cls += " warn"
	}
	return `<span class="` + cls + `">` + render.Esc(dispo) + `</span>`
}

func mxStockCell(stock *int64) string {
	if stock == nil {
		return "—"
	}
	n := *stock
	cls := ""
	if n <= 2 {
		cls = " mx-stock-low"
	}
	return `<span class="mx-stock-val` + cls + `">` + strconv.FormatInt(n, 10) + `</span>`
}
