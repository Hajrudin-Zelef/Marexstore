package catalog

import (
	_ "embed"
	"encoding/json"
	"encoding/xml"
	"os"
	"sort"
	"strings"
)

//go:embed sitemap.xml
var embeddedSitemap []byte

// ---------- Compteurs ----------

type Store struct {
	Products    []*RawProduct
	BySlug      map[string]*RawProduct
	TopCounts   map[string]int
	SubCounts   map[[2]string]int
	BrandCounts map[string]int
	Facets      map[string][]Facet
	PrixMax     int64
}

type FacetOpt struct {
	V  string `json:"v"`
	Vs string `json:"vs"`
	N  int    `json:"n"`
}
type Facet struct {
	Label   string     `json:"label"`
	Slug    string     `json:"slug"`
	Options []FacetOpt `json:"options"`
}

func NewStore(list []*RawProduct) *Store {
	s := &Store{BySlug: map[string]*RawProduct{}, TopCounts: map[string]int{}, SubCounts: map[[2]string]int{}, BrandCounts: map[string]int{}}
	for _, p := range list {
		s.Products = append(s.Products, p)
		s.BySlug[p.Slug] = p
		s.TopCounts[p.LiveTop]++
		if p.LiveSub != "" {
			s.SubCounts[[2]string{p.LiveTop, p.LiveSub}]++
		}
		m := p.Marque
		if m == "" {
			m = "Sans marque"
		}
		s.BrandCounts[m]++
		if pr := p.PrixAffiche(); pr != nil && *pr > s.PrixMax {
			s.PrixMax = *pr
		}
	}
	// L'ordre du JSON source est conservé (grille d'accueil, similaires,
	// repli "récents") ; les listes triées passent par un tri explicite.
	s.Facets = s.buildFacets()
	return s
}

func (s *Store) buildFacets() map[string][]Facet {
	scopes := map[string][]*RawProduct{}
	for _, p := range s.Products {
		scopes["top:"+p.LiveTop] = append(scopes["top:"+p.LiveTop], p)
		if p.LiveSub != "" {
			scopes["sub:"+p.LiveSub] = append(scopes["sub:"+p.LiveSub], p)
		}
	}
	order := map[string]int{}
	for i, r := range FacetRules {
		order[r.Slug] = i
	}
	out := map[string][]Facet{}
	for key, fs := range scopes {
		counts := map[string]map[string]int{}
		cover := map[string]int{}
		for _, p := range fs {
			for slug, v := range FacetValues(p) {
				cover[slug]++
				if counts[slug] == nil {
					counts[slug] = map[string]int{}
				}
				switch t := v.(type) {
				case string:
					counts[slug][t]++
				case []string:
					for _, x := range t {
						counts[slug][x]++
					}
				}
			}
		}
		type wf struct {
			Facet
			cover int
		}
		var all []wf
		for _, r := range FacetRules {
			c := counts[r.Slug]
			if len(c) < 2 {
				continue
			}
			type kv struct {
				k string
				n int
			}
			var kvs []kv
			for k, n := range c {
				kvs = append(kvs, kv{k, n})
			}
			sort.Slice(kvs, func(i, j int) bool {
				if kvs[i].n != kvs[j].n {
					return kvs[i].n > kvs[j].n
				}
				return kvs[i].k < kvs[j].k
			})
			if len(kvs) > 15 {
				kvs = kvs[:15]
			}
			f := Facet{Label: r.Label, Slug: r.Slug}
			for _, e := range kvs {
				f.Options = append(f.Options, FacetOpt{V: e.k, Vs: Slugify(e.k), N: e.n})
			}
			all = append(all, wf{f, cover[r.Slug]})
		}
		w := all
		// tri stable : couverture desc (comme le sort stable de Python,
		// qui conserve l'ordre des règles en cas d'égalité), max 14.
		sort.SliceStable(w, func(i, j int) bool { return w[i].cover > w[j].cover })
		if len(w) > 14 {
			w = w[:14]
		}
		sort.Slice(w, func(i, j int) bool { return order[w[i].Slug] < order[w[j].Slug] })
		res := []Facet{}
		for _, x := range w {
			res = append(res, x.Facet)
		}
		out[key] = res
	}
	return out
}

// FacetsJS reproduit le contenu de facets.js (généré) depuis la DB.
func (s *Store) FacetsJS() string {
	b, _ := json.Marshal(s.Facets)
	return "/* Facettes des filtres par catégorie / sous-catégorie (généré par le serveur).\n   Valeurs extraites des caractéristiques réelles des fiches, rien d'inventé. */\nwindow.MEDIATEC_FACETS = " + string(b) + ";\n"
}

// HubPhoto reproduit hub_photo() + HUB_PHOTO_PREF.
func (s *Store) HubPhoto(top string) *RawProduct {
	prefs := map[string][]string{
		"accessoires-informatiques":   {"toner", "clavier"},
		"materiels-reseaux":           {"switch", "routeur", "cisco"},
		"ordinateurs-de-bureau-neufs": {"all-in-one", "tout-en-un"},
		"imprimantes":                 {"ecotank", "laserjet", "deskjet", "officejet"},
		"serveur":                     {"rack", "poweredge", "proliant", "thinksystem", "ucs"},
	}
	var cands []*RawProduct
	for _, p := range s.Products {
		if p.LiveTop == top && p.Image != "" {
			cands = append(cands, p)
		}
	}
	if len(cands) == 0 {
		return nil
	}
	sort.Slice(cands, func(i, j int) bool {
		return strings.ToLower(cands[i].Nom) < strings.ToLower(cands[j].Nom)
	})
	for _, kw := range prefs[top] {
		for _, p := range cands {
			if strings.Contains(strings.ToLower(p.Nom), kw) {
				return p
			}
		}
	}
	return cands[0]
}

// Recents reproduit la logique "Ajouté récemment" : les 8 fiches dont
// le store_url apparaît en premier dans store-sitemap.xml (trié desc).
func (s *Store) Recents(sitemapPath string) []*RawProduct {
	type urlentry struct {
		Loc     string `xml:"loc"`
		Lastmod string `xml:"lastmod"`
	}
	type urlset struct {
		URLs []urlentry `xml:"url"`
	}
	b := embeddedSitemap
	if sitemapPath != "" {
		if fb, err := os.ReadFile(sitemapPath); err == nil {
			b = fb
		}
	}
	if len(b) == 0 {
		if len(s.Products) > 8 {
			return s.Products[:8]
		}
		return s.Products
	}
	var us urlset
	if err := xml.Unmarshal(b, &us); err != nil {
		if len(s.Products) > 8 {
			return s.Products[:8]
		}
		return s.Products
	}
	sort.Slice(us.URLs, func(i, j int) bool { return us.URLs[i].Lastmod > us.URLs[j].Lastmod })
	byURL := map[string]*RawProduct{}
	for _, p := range s.Products {
		if p.StoreURL != "" {
			byURL[p.StoreURL] = p
		}
	}
	var out []*RawProduct
	seen := map[string]bool{}
	for _, u := range us.URLs {
		loc := strings.TrimSpace(u.Loc)
		if p, ok := byURL[loc]; ok && !seen[p.Slug] {
			seen[p.Slug] = true
			out = append(out, p)
		}
		if len(out) >= 8 {
			break
		}
	}
	return out
}
