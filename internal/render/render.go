// Package render porte les fonctions de rendu HTML de build_site_v3.py.
// Toutes les valeurs dynamiques passent par catalog.Esc (identique à
// html.escape(quote=True) de Python) afin d'obtenir un HTML identique
// à l'octet près au site statique v3.25.
package render

import (
	"fmt"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

const ThemeInit = `<script>try{var _t=JSON.parse(localStorage.getItem('mediatec-theme')||'{}');var _r=document.documentElement;if(_t.mode==='sombre')_r.setAttribute('data-theme','dark');if(_t.color==='liquid')_r.setAttribute('data-color','liquid');}catch(e){}</script>`

const CartSVG = `<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="21" r="1"/><circle cx="20" cy="21" r="1"/><path d="M1 1h4l2.68 13.39a2 2 0 0 0 2 1.61h9.72a2 2 0 0 0 2-1.61L23 6H6"/></svg>`

var CatIcons = map[string]string{
	"accessoires-informatiques":        ico(`<rect x="2" y="6" width="20" height="12" rx="2"/><path d="M6 10h.01M10 10h.01M14 10h.01M18 10h.01M7 14h10"/>`),
	"accessoires-telephones-tablettes": ico(`<rect x="7" y="2" width="10" height="20" rx="2"/><path d="M11 18h2"/>`),
	"cameras-videosurveillances":       ico(`<path d="M23 7l-7 5 7 5V7z"/><rect x="1" y="5" width="15" height="14" rx="2"/>`),
	"imprimantes":                      ico(`<path d="M6 9V2h12v7"/><path d="M6 18H4a2 2 0 0 1-2-2v-5a2 2 0 0 1 2-2h16a2 2 0 0 1 2 2v5a2 2 0 0 1-2 2h-2"/><rect x="6" y="14" width="12" height="8"/>`),
	"laptots":                          ico(`<rect x="3" y="4" width="18" height="12" rx="2"/><path d="M2 20h20"/>`),
	"materiels-reseaux":                ico(`<circle cx="18" cy="5" r="3"/><circle cx="6" cy="12" r="3"/><circle cx="18" cy="19" r="3"/><path d="M8.6 13.5l6.8 4M15.4 6.5l-6.8 4"/>`),
	"ordinateurs-de-bureau-neufs":      ico(`<rect x="2" y="3" width="20" height="14" rx="2"/><path d="M8 21h8M12 17v4"/>`),
	"casi-neuf":                        ico(`<path d="M23 4v6h-6"/><path d="M1 20v-6h6"/><path d="M3.51 9a9 9 0 0 1 14.85-3.36L23 10M1 14l4.64 4.36A9 9 0 0 0 20.49 15"/>`),
	"serveur":                          ico(`<rect x="2" y="2" width="20" height="8" rx="2"/><rect x="2" y="14" width="20" height="8" rx="2"/><path d="M6 6h.01M6 18h.01"/>`),
	"telephones-tablettes":             ico(`<rect x="4" y="2" width="16" height="20" rx="2"/><path d="M12 18h.01"/>`),
	"uncategorized":                    ico(`<path d="M16.5 9.4L7.55 4.24"/><path d="M21 16V8a2 2 0 0 0-1-1.73l-7-4a2 2 0 0 0-2 0l-7 4A2 2 0 0 0 3 8v8a2 2 0 0 0 1 1.73l7 4a2 2 0 0 0 2 0l7-4A2 2 0 0 0 21 16z"/><path d="M3.27 6.96L12 12.01l8.73-5.05"/><path d="M12 22.08V12"/>`),
}

func ico(paths string) string {
	return `<svg viewBox="0 0 24 24" width="18" height="18" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">` + paths + `</svg>`
}

func esc(s string) string { return catalog.Esc(s) }

// Esc est la version exportée (même échappement que le générateur Python).
func Esc(s string) string { return catalog.Esc(s) }

func ImgTag(p *catalog.RawProduct, prefix string) string {
	if p.Image != "" {
		return `<img src="` + esc(prefix+p.Image) + `" alt="` + esc(p.Nom) + `" loading="lazy">`
	}
	return `<div class="noimg">Photo à venir</div>`
}

func BadgeDispo(p *catalog.RawProduct) string {
	switch p.Dispo {
	case "en stock":
		return `<span class="stock ok">En stock</span>`
	case "rupture de stock":
		return `<span class="stock out">Rupture de stock</span>`
	default:
		return `<span class="stock ret">Retour</span>`
	}
}

func MegaPanel(s *catalog.Store, prefix string) string {
	var rail, cols strings.Builder
	first := true
	for _, t := range catalog.LiveCats {
		n := s.TopCounts[t.Slug]
		on := ""
		if first {
			on = "on"
		}
		fmt.Fprintf(&rail, `<button type="button" data-mega-top="%s" class="%s">`, t.Slug, on)
		fmt.Fprintf(&rail, `<span class="mega-lab"><span class="mega-ico">%s</span>%s</span> <span class="mega-count">(%d)</span></button>`, CatIcons[t.Slug], esc(t.Label), n)
		var groups strings.Builder
		for _, g := range catalog.MegaGroups[t.Slug] {
			var links strings.Builder
			for _, sslug := range g.Subs {
				fmt.Fprintf(&links, `<a href="%sboutique.html?sub=%s">%s <span>(%d)</span></a>`, prefix, sslug, esc(catalog.SubLabels[[2]string{t.Slug, sslug}]), s.SubCounts[[2]string{t.Slug, sslug}])
			}
			more := ""
			if len(g.Subs) > 1 {
				more = fmt.Sprintf(`<a class="mega-more" href="%sboutique.html?sub=%s">Voir +</a>`, prefix, strings.Join(g.Subs, ","))
			}
			fmt.Fprintf(&groups, `<div class="mega-group"><h5>%s</h5>%s%s</div>`, esc(g.Title), links.String(), more)
		}
		hidden := " hidden"
		if first {
			hidden = ""
		}
		fmt.Fprintf(&cols, `<div class="mega-cols" data-mega-cols="%s"%s><div class="mega-groups">%s</div><a class="mega-all" href="%scategories/%s.html">Voir tous les %s ›</a></div>`,
			t.Slug, hidden, groups.String(), prefix, t.Slug, esc(t.Label))
		first = false
	}
	return `<div class="mega-panel" data-mega-panel hidden><div class="in"><div class="mega-rail">` + rail.String() + `</div><div>` + cols.String() + `</div></div></div>`
}

// brandOptions : les options <option> du filtre Marque (mêmes valeurs que
// les cartes data-marque, tri par nombre décroissant puis libellé).
func brandOptions(s *catalog.Store) string {
	type kv struct {
		k string
		n int
	}
	var brands []kv
	for b, n := range s.BrandCounts {
		brands = append(brands, kv{b, n})
	}
	// tri Python : sorted(items, key=(-n, libellé brut)).
	for i := 0; i < len(brands); i++ {
		for j := i + 1; j < len(brands); j++ {
			if brands[j].n > brands[i].n || (brands[j].n == brands[i].n && brands[j].k < brands[i].k) {
				brands[i], brands[j] = brands[j], brands[i]
			}
		}
	}
	var opts strings.Builder
	opts.WriteString(`<option value="">Toutes les marques</option>`)
	for _, b := range brands {
		fmt.Fprintf(&opts, `<option value="%s">%s (%d)</option>`, esc(catalog.NormName(b.k)), esc(b.k), b.n)
	}
	return opts.String()
}

func FiltersPanel(s *catalog.Store, mode string) string {
	return `<button type="button" class="filters-toggle" data-f-toggle aria-expanded="false"><svg viewBox="0 0 24 24" width="16" height="16" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M22 3H2l8 9.46V19l4 2v-8.54L22 3z"/></svg><span>Filtres</span><span class="filters-chev" aria-hidden="true">▾</span></button>
<div class="filters" data-filters data-facets-mode="` + mode + `"><h3>FILTRER LES PRODUITS :</h3>
<label>Chercher une référence<input type="search" data-f-q placeholder="Désignation, modèle…"></label>
<label>Marque<select data-f-marque>` + brandOptions(s) + `</select></label>
<label>Prix (F CFA)<span class="f-range"><input type="range" data-f-prix-min-range min="0" max="` + fmt.Sprint(s.PrixMax) + `" step="100" value="0"><input type="range" data-f-prix-max-range min="0" max="` + fmt.Sprint(s.PrixMax) + `" step="100" value="` + fmt.Sprint(s.PrixMax) + `"></span><span class="f-prix"><input type="number" data-f-prix-min placeholder="Min" min="0"><span>–</span><input type="number" data-f-prix-max placeholder="Max" min="0"></span></label>
<div data-f-facets></div>
<button type="button" class="btn" data-f-more>+ DE FILTRES</button>
<div data-f-extra hidden>
<label>Disponibilité<select data-f-dispo><option value="">Toutes</option><option value="en-stock">En stock</option><option value="rupture-de-stock">Rupture de stock</option></select></label>
<label>État<select data-f-etat><option value="">Tous</option><option value="neuf">Neuf</option><option value="reconditionne">Reconditionné</option></select></label>
</div>
<label class="f-check"><input type="checkbox" data-f-stock> Voir uniquement les produits en stock</label>
<button type="button" class="btn ghost" data-f-reset>Réinitialiser les filtres</button>
</div>`
}
