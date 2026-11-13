package render

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// ---------- Accueil ----------

type banner struct {
	bg, titre, sous, acc, prix, btn, lien string
}

var banners = []banner{
	{"images/banners/oukitel-rt3.jpg", "OUKITEL", "RT3 PLUS<br>16GO / 128GO", "TABLETTE PROFESSIONELLE", "200.000 CFA", "Acheter", "produits/oukitel-rt3-plus-16-go-ram-128-go.html"},
	{"images/banners/kyocera-copieurs.jpg", "DÉCOUVREZ", "NOS COPIEURS<br>KYOCERA <b>EN STOCK</b>", "A ne pas rater", "", "En savoir +", "categories/imprimantes-consommables.html"},
	{"images/banners/elitebook-850-g3.jpg", "HP ELITEBOOK 850 G3", "15\", 8Go - 256ssd", "", "185.000F", "ACHETER", "produits/hp-elitebook-elitebook-850.html"},
}

// ---------- Filtre global (accueil) ----------

// homeFilterHTML : « Trouver un produit » — un filtre complet sur l'accueil
// (marque, catégorie, type, prix, disponibilité) qui envoie vers la boutique
// déjà filtrée (boutique.html?cat=&type=&marque=&prixmin=&prixmax=&dispo=).
// Le choix du type dépend de la catégorie (app.js lit le bloc JSON
// #mediatec-types — jamais exécuté, donc hors CSP script-src).
// Le bloc et le script sont sur une seule ligne chacun (la parité les retire).
func homeFilterHTML(s *catalog.Store) string {
	type typeOpt struct {
		Slug  string `json:"slug"`
		Label string `json:"label"`
	}
	typesByMain := map[string][]typeOpt{}
	seen := map[string]map[string]bool{}
	for _, p := range s.Products {
		m := catalog.MainSlug[p.Main]
		if seen[m] == nil {
			seen[m] = map[string]bool{}
		}
		sl := catalog.Slugify(p.Sub)
		if !seen[m][sl] {
			seen[m][sl] = true
			typesByMain[m] = append(typesByMain[m], typeOpt{Slug: sl, Label: p.Sub})
		}
	}
	for _, list := range typesByMain {
		sort.Slice(list, func(i, j int) bool { return list[i].Label < list[j].Label })
	}
	jb, _ := json.Marshal(typesByMain)
	var cats strings.Builder
	cats.WriteString(`<option value="">Toutes les catégories</option>`)
	for _, m := range catalog.MainOrder {
		fmt.Fprintf(&cats, `<option value="%s">%s</option>`, esc(catalog.MainSlug[m]), esc(m))
	}
	var types strings.Builder
	types.WriteString(`<option value="">Tous les types</option>`)
	for _, m := range catalog.MainOrder {
		for _, t := range typesByMain[catalog.MainSlug[m]] {
			fmt.Fprintf(&types, `<option value="%s" data-main="%s">%s</option>`, esc(t.Slug), esc(catalog.MainSlug[m]), esc(t.Label))
		}
	}
	return `<div class="home-filter" data-home-filter><h2>Trouver un produit</h2><p class="sub">Affinez par marque, catégorie, type, prix et disponibilité — la boutique s'ouvre déjà filtrée.</p><form action="boutique.html" method="get"><label>Marque<select name="marque">` + brandOptions(s) + `</select></label><label>Catégorie<select name="cat" data-hf-cat>` + cats.String() + `</select></label><label>Type<select name="type" data-hf-type>` + types.String() + `</select></label><label>Prix min (F CFA)<input type="number" name="prixmin" min="0" placeholder="Min"></label><label>Prix max (F CFA)<input type="number" name="prixmax" min="0" placeholder="Max"></label><label>Disponibilité<select name="dispo"><option value="">Toutes</option><option value="en-stock">En stock</option><option value="rupture-de-stock">Rupture de stock</option></select></label><button class="btn" type="submit">Rechercher</button></form></div>` + "\n" + `<script type="application/json" id="mediatec-types">` + string(jb) + `</script>`
}

func IndexBody(s *catalog.Store, recents []*catalog.RawProduct) string {
	var slides strings.Builder
	for i, b := range banners {
		on := ""
		if i == 0 {
			on = " on"
		}
		acc := ""
		if b.acc != "" {
			acc = `<p class="b-acc">` + b.acc + `</p>`
		}
		pr := ""
		if b.prix != "" {
			pr = `<div class="b-pr">` + b.prix + `</div>`
		}
		fmt.Fprintf(&slides, `<div class="s%s" style="background-image:url(%s)"><div class="banner-txt">
<h2>%s</h2><p class="b-sub">%s</p>
%s
%s
<a class="btn" href="%s">%s</a></div></div>`, on, esc(b.bg), b.titre, b.sous, acc, pr, esc(b.lien), b.btn)
	}
	var dots strings.Builder
	for i := range banners {
		on := ""
		if i == 0 {
			on = "on"
		}
		fmt.Fprintf(&dots, `<button type="button" data-dot class="%s" aria-label="Slide %d"></button>`, on, i+1)
	}
	deal := s.BySlug["kyocera-taskalfa-3011i"]
	tabCards := func(slugs []string) string {
		var out strings.Builder
		for _, slug := range slugs {
			if f, ok := s.BySlug[slug]; ok {
				out.WriteString(Card(s, f, "", "main_slug"))
			}
		}
		return out.String()
	}
	vedette := []string{"hp-all-in-one-24-pouce-i5", "ecran-hp-serie-527sf-s527sf", "casque-logitech-h111"}
	promo := []string{"kyocera-taskalfa-3011i", "lenovo-thinkpad-e14g5-21jk00dagq", "oukitel-rt3-plus-16-go-ram-128-go"}
	notes := []string{"lenovo-thinkpad-e14g5-21jk00dagq", "hp-elitebook-elitebook-850", "hp-elitebook-840-g8"}
	top8slugs := []string{"toner-650a-yellow", "toner-205a-yellow", "toner-711a-black",
		"toner-xerox-006r01463-magenta", "laptop-iles-de-france",
		"toner-304a-cyan", "ecran-hp-elite-m243", "toner-650a-cyan"}
	var top8 strings.Builder
	for _, slug := range top8slugs {
		if f, ok := s.BySlug[slug]; ok {
			im := ""
			if f.Image != "" {
				im = `<img src="` + esc(f.Image) + `" alt="">`
			}
			fmt.Fprintf(&top8, `<div class="hcard">%s
<div><h4><a href="produits/%s.html">%s</a></h4>
<div class="pr">%s</div></div></div>`, im, esc(slug), esc(f.Nom), PrixHTML(f))
		}
	}
	var grille strings.Builder
	for i, f := range s.Products {
		if i >= 8 {
			break
		}
		grille.WriteString(Card(s, f, "", "main_slug"))
	}
	var recentsHTML strings.Builder
	for _, f := range recents {
		recentsHTML.WriteString(Card(s, f, "", "main_slug"))
	}
	minicard := func(slug, titre, lien string) string {
		f, ok := s.BySlug[slug]
		if !ok {
			return ""
		}
		im := ""
		if f.Image != "" {
			im = `<img src="` + esc(f.Image) + `" alt="">`
		}
		if lien == "" {
			lien = "produits/" + esc(slug) + ".html"
		}
		return `<a class="minicard" href="` + lien + `">` + im + `<span>` + esc(titre) + `</span></a>`
	}
	minis := `<div class="minis">` +
		minicard("laptop-lenovo-t440", "Profitez de super offres sur le Lenovo T440", "") +
		minicard("hp-all-in-one-24-pouce-i5", "Bureau, All-in-One et plus encore -10%", "") +
		minicard("ecran-hp-serie-527sf-s527sf", "Écran HP Série 5 527SF -12%", "") +
		minicard("casque-logitech-h111", "Casque Logitech H111", "") + `</div>`
	body := Header(s, "", "") + `
<div class="wrap">
<div class="slide" id="slideshow">` + slides.String() + `
<button type="button" class="nav prev" data-prev aria-label="Précédent">‹</button>
<button type="button" class="nav next" data-next aria-label="Suivant">›</button>
<div class="dots">` + dots.String() + `</div></div>
` + minis + `
` + homeFilterHTML(s) + `
<div class="sec"><div class="deal">
<div class="dealcard"><div>` + ImgTag(deal, "") + `</div><div>
<div class="cat">Offre spéciale <span class="save">Save 10%</span></div><h2 style="margin:4px 0">` + esc(deal.Nom) + `</h2>
<p>` + esc(deal.Accroche) + `</p><div class="pr">` + PrixHTML(deal) + `</div>
<a class="btn" href="produits/` + esc(deal.Slug) + `.html">En savoir +</a></div></div>
<div data-tabs><div class="tabs">
<button type="button" class="on" data-tab="1">En vedette</button>
<button type="button" data-tab="2">En Promotion</button>
<button type="button" data-tab="3">Les Mieux noté</button></div>
<div data-pane="1"><div class="catalog view-grid" style="grid-template-columns:repeat(3,1fr)">` + tabCards(vedette) + `</div></div>
<div data-pane="2" hidden><div class="catalog view-grid" style="grid-template-columns:repeat(3,1fr)">` + tabCards(promo) + `</div></div>
<div data-pane="3" hidden><div class="catalog view-grid" style="grid-template-columns:repeat(3,1fr)">` + tabCards(notes) + `</div></div>
</div></div></div>
<div class="sec"><h2>Nos produits</h2><div class="sub">Une sélection du stock — voir tout le catalogue dans le <a href="boutique.html">Store</a></div>
<div class="catalog view-grid">` + grille.String() + `</div></div>
<div class="sec"><h2>Les plus vendus</h2><div class="sub">Top 8</div><div class="top8">` + top8.String() + `</div></div>
<div class="sec"><div class="banner"><div><h3>Clés USB Imation &amp; disques durs WD Elements</h3>
<p>Stockage fiable pour tous vos besoins — demandez votre devis.</p></div>
<div><a class="btn" href="boutique.html">VOIR LE PRODUIT</a>
<a class="btn dark" href="mailto:` + catalog.Email + `">CONTACT</a>
<a class="btn wa" href="https://wa.me/` + catalog.WA + `">DEMO</a></div></div></div>
<div class="sec"><h2>Ajouté récemment</h2><div class="sub">Les derniers produits du stock</div>
<div class="catalog view-grid">` + recentsHTML.String() + `</div></div>
<div class="sec"><div class="brands"><span>ACER</span><span>APPLE</span><span>ASUS</span><span>CISCO</span><span>DELL</span><span>EPSON</span><span>HP</span><span>KYOCERA</span><span>LENOVO</span></div></div>
</div>
` + Footer(s, "")
	return body
}

// ---------- Boutique ----------

// SortedStockFirst : le stock d'abord — produits avec un prix (commandables),
// puis ceux sans prix (sur devis) ; dans chaque groupe, "en stock" d'abord,
// puis nom A-Z. C'est l'ordre par défaut de la boutique et des pages catégories
// (le tri "Défaut" du sélecteur restaure cet ordre).
func SortedStockFirst(fs []*catalog.RawProduct) []*catalog.RawProduct {
	out := append([]*catalog.RawProduct{}, fs...)
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		pa, pb := a.PrixAffiche(), b.PrixAffiche()
		if (pa == nil) != (pb == nil) {
			return pa != nil
		}
		da, db := a.Dispo == "en stock", b.Dispo == "en stock"
		if da != db {
			return da
		}
		return strings.ToLower(a.Nom) < strings.ToLower(b.Nom)
	})
	return out
}

func BoutiqueBody(s *catalog.Store) string {
	return boutiqueBodyQuery(s, url.Values{})
}

// HasFilterParam : true si l'URL porte déjà un filtre (?store=, ?sub=, ?cat=,
// ?q=, ?marque=, ?prixmin=, ?prixmax=, ?dispo=, ?type=, ?all=). Sert l'état
// initial anti-flash : hub masqué + liste affichée dès le premier rendu.
func HasFilterParam(q url.Values) bool {
	for _, k := range []string{"store", "sub", "cat", "q", "marque", "prixmin", "prixmax", "dispo", "type", "all"} {
		if v := strings.TrimSpace(q.Get(k)); v != "" && v != "all" {
			return true
		}
	}
	return false
}

// BoutiqueBodyQuery : boutique avec état initial anti-flash. Sans filtre, rendu
// identique à BoutiqueBody (hub visible, liste cachée) ; avec filtre, hub masqué
// et liste affichée dès le premier rendu — le JS n'a plus rien à basculer.
func BoutiqueBodyQuery(s *catalog.Store, q url.Values) string {
	return boutiqueBodyQuery(s, q)
}

func boutiqueBodyQuery(s *catalog.Store, q url.Values) string {
	prefiltered := HasFilterParam(q)
	hubHidden, listHidden := "", " hidden"
	if prefiltered {
		hubHidden, listHidden = " hidden", ""
	}
	return Header(s, "", "") + `
<div class="wrap">
<div class="crumb"><a href="index.html">Accueil</a> › Store</div>
` + HubHTML(s, hubHidden) + `
<div class="home" data-catalog-home` + listHidden + `>
<div>` + FiltersPanel(s, "selection") + `</div>
<div><div class="crumb"><a href="boutique.html">← Toutes les catégories</a></div>
<h1 style="font-size:24px">Store</h1>
<p style="color:#666">` + fmt.Sprint(len(s.Products)) + ` produits · inventaire du 21 septembre 2026 + produits du store — prix et disponibilités à confirmer avant commande.</p>
` + CatalogWrap(s, SortedStockFirst(s.Products), "", "sub_slug", "", "list", nil, "") + `
</div></div></div>
` + Footer(s, "")
}

// ---------- Catégories ----------

func CategoryBody(s *catalog.Store, main string) []*catalog.RawProduct {
	var fs []*catalog.RawProduct
	for _, p := range s.Products {
		if p.Main == main {
			fs = append(fs, p)
		}
	}
	return SortedStockFirst(fs)
}

// ---------- Hub des types (pages catégories) ----------

// TypeHubHTML : la page d'une catégorie liste d'abord ses types d'équipements
// (les sous-catégories du catalogue, avec photo et compteur) ; un clic sur un
// type ouvre sa page dédiée (categories/<top>/<type>.html), bien structurée,
// comme le hub des catégories de la boutique. Un seul type par tuile, aucun
// type vide. active = slug du type courant : la tuile est surlignée.
func TypeHubHTML(fs []*catalog.RawProduct, active, prefix, topSlug string) string {
	counts := map[string]int{}
	rep := map[string]*catalog.RawProduct{}
	for _, p := range fs {
		counts[p.Sub]++
		if rep[p.Sub] == nil && p.Image != "" {
			rep[p.Sub] = p
		}
	}
	var subs []string
	for x := range counts {
		subs = append(subs, x)
	}
	sort.Slice(subs, func(i, j int) bool {
		if counts[subs[i]] != counts[subs[j]] {
			return counts[subs[i]] > counts[subs[j]]
		}
		return subs[i] < subs[j]
	})
	var cards strings.Builder
	liveOf := map[string]map[string]bool{}
	for _, p := range fs {
		if liveOf[p.Sub] == nil {
			liveOf[p.Sub] = map[string]bool{}
		}
		liveOf[p.Sub][p.LiveSub] = true
	}
	for _, sub := range subs {
		img := `<div class="noimg">Photo à venir</div>`
		if r := rep[sub]; r != nil {
			img = ImgTag(r, prefix)
		}
		prod := "produits"
		if counts[sub] == 1 {
			prod = "produit"
		}
		sl := catalog.Slugify(sub)
		cls, cur := "hub-main", ""
		// Surlignage : type exact (sous-catégorie) OU type live du menu dont
		// relèvent les produits de la tuile (ex. câbles HDMI/RJ45 pour
		// « Câbles & Adaptateurs »).
		if active != "" && (sl == active || liveOf[sub][active]) {
			cls, cur = "hub-main on", ` aria-current="true"`
		}
		fmt.Fprintf(&cards, `<div class="hub-card"><a class="%s" href="%scategories/%s/%s.html"%s><span class="hub-img">%s</span><span class="hub-txt"><span class="hub-name">%s</span><span class="hub-count">%d %s</span></span></a></div>`, cls, prefix, esc(topSlug), esc(sl), cur, img, esc(sub), counts[sub], prod)
	}
	return `<div data-hub data-typehub><h2 style="font-size:20px">Types d'équipements</h2><p style="color:#666">Choisissez un type pour voir ses produits.</p><div class="hub-grid">` + cards.String() + `</div><p style="margin-top:18px"><a class="btn" href="?all=1">Voir tous les produits (` + fmt.Sprint(len(fs)) + `) →</a></p></div>`
}

// categoryPageFor : page d'une catégorie (hub des types d'abord, puis les
// produits) pour une liste de produits et son titre. Sert pour les familles
// (pages /categories/<famille>.html) et pour les tops live du menu
// « Toutes les catégories » (pages /categories/<top-live>.html).
// Le hub des types reste TOUJOURS visible (pas de flash) ; showList affiche
// la liste dès le premier rendu quand un filtre est déjà posé (?type=, ?all=1)
// et activeType surligne la tuile courante. topSlug sert aux liens des tuiles.
func categoryPageFor(s *catalog.Store, fs []*catalog.RawProduct, title, topSlug, activeType string, showList bool) string {
	subs := map[string]bool{}
	for _, p := range fs {
		subs[p.Sub] = true
	}
	var subList []string
	for x := range subs {
		subList = append(subList, x)
	}
	sort.Strings(subList)
	var chips [][2]string
	for _, x := range subList {
		chips = append(chips, [2]string{x, catalog.Slugify(x)})
	}
	n := fmt.Sprintf("%d produit", len(fs))
	if len(fs) > 1 {
		n = fmt.Sprintf("%d produits", len(fs))
	}
	listHidden := " hidden"
	if showList {
		listHidden = ""
	}
	return Header(s, "../", "") + `
<div class="wrap"><div class="home">
<div>` + FiltersPanel(s, "page") + `</div>
<div><div class="crumb"><a href="../index.html">Accueil</a> › <a href="../boutique.html">Store</a> › ` + esc(title) + `</div>
<h1 style="font-size:24px">` + esc(title) + `</h1>
<p style="color:#666">` + n + ` · ` + esc(strings.Join(subList, ", ")) + `</p>
 ` + TypeHubHTML(fs, activeType, "../", topSlug) + `
<div data-catalog-home` + listHidden + `>
` + CatalogWrap(s, fs, "../", "sub_slug", "sub", "list", chips, "") + `</div>
</div></div></div>
` + Footer(s, "../")
}

func CategoryPage(s *catalog.Store, main string) string {
	return categoryPageFor(s, CategoryBody(s, main), main, catalog.MainSlug[main], "", false)
}

// CategoryPageFiltered : variante avec état initial (anti-flash) — activeType
// (?type=) surligne la tuile, showList affiche la liste dès le premier rendu.
func CategoryPageFiltered(s *catalog.Store, main, activeType string, showList bool) string {
	return categoryPageFor(s, CategoryBody(s, main), main, catalog.MainSlug[main], activeType, showList)
}

// CategoryBodyLive : produits d'un top live (catégorie du menu
// « Toutes les catégories »).
func CategoryBodyLive(s *catalog.Store, liveTop string) []*catalog.RawProduct {
	var fs []*catalog.RawProduct
	for _, p := range s.Products {
		if p.LiveTop == liveTop {
			fs = append(fs, p)
		}
	}
	return SortedStockFirst(fs)
}

// CategoryPageLive : page d'un top live (catégorie du menu) : mêmes types
// d'équipements d'abord, puis les produits.
func CategoryPageLive(s *catalog.Store, t catalog.LiveTop) string {
	return categoryPageFor(s, CategoryBodyLive(s, t.Slug), t.Label, t.Slug, "", false)
}

// CategoryPageLiveFiltered : variante avec état initial (anti-flash).
func CategoryPageLiveFiltered(s *catalog.Store, t catalog.LiveTop, activeType string, showList bool) string {
	return categoryPageFor(s, CategoryBodyLive(s, t.Slug), t.Label, t.Slug, activeType, showList)
}

// typeScope : produits d'un top (famille ou live) + libellé du top.
func typeScope(s *catalog.Store, topSlug string) (fs []*catalog.RawProduct, topLabel string, ok bool) {
	for _, m := range catalog.MainOrder {
		if catalog.MainSlug[m] == topSlug {
			return CategoryBody(s, m), m, true
		}
	}
	for _, t := range catalog.LiveCats {
		if t.Slug == topSlug {
			return CategoryBodyLive(s, t.Slug), t.Label, true
		}
	}
	return nil, "", false
}

// CategoryTypePage : page dédiée d'un type d'équipement
// (/categories/<top>/<type>.html) — fil d'Ariane complet, titre, hub des types
// (tuile active surlignée, toujours visible) puis la liste des produits.
// Le type matche la sous-catégorie (data-sub) OU la sous-catégorie live du menu
// (data-live-sub), comme le filtre ?type=. Retourne ok=false si le couple
// top/type ne désigne aucun produit (pas de page vide).
func CategoryTypePage(s *catalog.Store, topSlug, typeSlug string) (title, body string, ok bool) {
	topFs, topLabel, found := typeScope(s, topSlug)
	if !found {
		return "", "", false
	}
	var fs []*catalog.RawProduct
	label := ""
	for _, p := range topFs {
		subMatch := catalog.Slugify(p.Sub) == typeSlug
		liveMatch := p.LiveSub == typeSlug
		if !subMatch && !liveMatch {
			continue
		}
		fs = append(fs, p)
		if label == "" {
			if subMatch {
				label = p.Sub
			} else if l := catalog.SubLabels[[2]string{topSlug, typeSlug}]; l != "" {
				label = l
			} else {
				label = p.Sub
			}
		}
	}
	if len(fs) == 0 || label == "" {
		return "", "", false
	}
	n := fmt.Sprintf("%d produit", len(fs))
	if len(fs) > 1 {
		n = fmt.Sprintf("%d produits", len(fs))
	}
	body = Header(s, "../../", "") + `
<div class="wrap"><div class="home">
<div>` + FiltersPanel(s, "page") + `</div>
<div><div class="crumb"><a href="../../index.html">Accueil</a> › <a href="../../boutique.html">Store</a> › <a href="../` + esc(topSlug) + `.html">` + esc(topLabel) + `</a> › ` + esc(label) + `</div>
<h1 style="font-size:24px">` + esc(label) + `</h1>
<p style="color:#666">` + n + ` · ` + esc(topLabel) + `</p>
 ` + "<p class=\"back\"><a href=\"../" + esc(topSlug) + ".html\">← Tous les types de " + esc(topLabel) + "</a></p>" + `
<div data-catalog-home data-locked>
` + CatalogWrap(s, fs, "../../", "sub_slug", "sub", "list", nil, "") + `</div>
</div></div></div>
` + Footer(s, "../../")
	return label + " — " + topLabel, body, true
}
