package render

import (
	"fmt"
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

func SortedByNom(fs []*catalog.RawProduct) []*catalog.RawProduct {
	out := append([]*catalog.RawProduct{}, fs...)
	sort.Slice(out, func(i, j int) bool {
		return strings.ToLower(out[i].Nom) < strings.ToLower(out[j].Nom)
	})
	return out
}

func BoutiqueBody(s *catalog.Store) string {
	return Header(s, "", "") + `
<div class="wrap">
<div class="crumb"><a href="index.html">Accueil</a> › Store</div>
` + HubHTML(s) + `
<div class="home" data-catalog-home hidden>
<div>` + FiltersPanel(s, "selection") + `</div>
<div><div class="crumb"><a href="boutique.html">← Toutes les catégories</a></div>
<h1 style="font-size:24px">Store</h1>
<p style="color:#666">` + fmt.Sprint(len(s.Products)) + ` produits · inventaire du 21 septembre 2026 + produits du store — prix et disponibilités à confirmer avant commande.</p>
` + CatalogWrap(s, SortedByNom(s.Products), "", "main_slug", "", "list", nil, "") + `
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
	return SortedByNom(fs)
}

func CategoryPage(s *catalog.Store, main string) string {
	fs := CategoryBody(s, main)
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
	return Header(s, "../", "") + `
<div class="wrap"><div class="home">
<div>` + FiltersPanel(s, "page") + `</div>
<div><div class="crumb"><a href="../index.html">Accueil</a> › <a href="../boutique.html">Store</a> › ` + esc(main) + `</div>
<h1 style="font-size:24px">` + esc(main) + `</h1>
<p style="color:#666">` + n + ` · ` + esc(strings.Join(subList, ", ")) + `</p>
` + CatalogWrap(s, fs, "../", "sub_slug", "sub", "list", chips, "") + `
</div></div></div>
` + Footer(s, "../")
}
