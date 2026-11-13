package render_test

import (
	"fmt"
	"regexp"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Charge le catalogue complet (171 + lot commercial), comme le serveur.
func testStoreFull(t *testing.T) *catalog.Store {
	t.Helper()
	list, err := catalog.LoadRaw(rel("../../catalogdata/produits-v3.json"))
	if err != nil {
		t.Skip("catalogdata absent (hors dépôt local)")
	}
	lot, err := catalog.LoadRaw(rel("../../catalogdata/nouveaux-produits-commercial.json"))
	if err != nil {
		t.Skip("lot commercial absent")
	}
	list = append(list, lot...)
	catalog.ApplyTaxonomy(list)
	return catalog.NewStore(list)
}

// TestTypeHubComplete : chaque page catégorie liste tous ses types, avec le
// bon compteur, une photo (ou « Photo à venir ») et le lien ?type=.
func TestTypeHubComplete(t *testing.T) {
	s := testStoreFull(t)
	for _, m := range catalog.MainOrder {
		fs := render.CategoryBody(s, m)
		hub := render.TypeHubHTML(fs)
		counts := map[string]int{}
		for _, p := range fs {
			counts[p.Sub]++
		}
		if got := strings.Count(hub, `class="hub-card"`); got != len(counts) {
			t.Errorf("%s : %d tuiles, %d types attendus", m, got, len(counts))
		}
		for sub, n := range counts {
			sl := catalog.Slugify(sub)
			if !strings.Contains(hub, `href="?type=`+sl+`"`) {
				t.Errorf("%s : lien ?type= absent pour %q", m, sub)
			}
			if !strings.Contains(hub, `>`+catalog.Esc(sub)+`<`) {
				t.Errorf("%s : libellé %q absent du hub", m, sub)
			}
			want := fmt.Sprintf(`%d produit`, n)
			if n > 1 {
				want = fmt.Sprintf(`%d produits`, n)
			}
			if !strings.Contains(hub, want) {
				t.Errorf("%s : compteur %q absent pour %q", m, want, sub)
			}
			// La tuile du type : une photo ou le repli « Photo à venir ».
			i := strings.Index(hub, `href="?type=`+sl+`"`)
			if i < 0 {
				continue
			}
			tile := hub[i:]
			if len(tile) > 300 {
				tile = tile[:300]
			}
			if !strings.Contains(tile, "<img") && !strings.Contains(tile, "noimg") {
				t.Errorf("%s : tuile %q sans photo ni repli", m, sub)
			}
		}
		if !strings.Contains(hub, fmt.Sprintf(`Voir tous les produits (%d)`, len(fs))) {
			t.Errorf("%s : lien « Voir tous les produits (%d) » absent", m, len(fs))
		}
	}
}

// TestHomeFilter : le filtre global de l'accueil propose marque, catégorie,
// type (tous les types du catalogue) et disponibilité, et la carte des types.
func TestHomeFilter(t *testing.T) {
	s := testStoreFull(t)
	recents := s.Recents(rel("../../catalogdata/store-sitemap.xml"))
	body := render.IndexBody(s, recents)
	if !strings.Contains(body, `data-home-filter`) {
		t.Fatal("filtre global absent de l'accueil")
	}
	if !strings.Contains(body, `action="boutique.html"`) {
		t.Error("le filtre doit envoyer vers boutique.html")
	}
	// Marques : les mêmes valeurs que le filtre latéral (NormName).
	if !strings.Contains(body, `<option value="hp">HP (`) {
		t.Error("marque HP absente du filtre")
	}
	if !strings.Contains(body, `<option value="cisco">Cisco (`) {
		t.Error("marque Cisco absente du filtre")
	}
	// Catégories : les 7 mains.
	for _, m := range catalog.MainOrder {
		if !strings.Contains(body, `<option value="`+catalog.MainSlug[m]+`">`+catalog.Esc(m)+`</option>`) {
			t.Errorf("catégorie %q absente du filtre", m)
		}
	}
	// Types : tous les types du catalogue.
	seen := map[string]bool{}
	for _, p := range s.Products {
		seen[p.Sub] = true
	}
	for sub := range seen {
		if !strings.Contains(body, `value="`+catalog.Slugify(sub)+`"`) {
			t.Errorf("type %q absent du filtre", sub)
		}
	}
	if !strings.Contains(body, `window.MEDIATEC_TYPES=`) {
		t.Error("carte des types (MEDIATEC_TYPES) absente")
	}
	for _, m := range catalog.MainOrder {
		if !strings.Contains(body, `"`+catalog.MainSlug[m]+`":`) {
			t.Errorf("catégorie %q absente de MEDIATEC_TYPES", m)
		}
	}
	t.Logf("filtre : %d types", len(seen))
}

// TestBoutiqueDataSub : les cartes de la boutique portent data-sub = la
// sous-catégorie (le filtre ?type= s'applique partout, pas seulement sur les
// pages catégories), jamais le slug d'une catégorie principale.
func TestBoutiqueDataSub(t *testing.T) {
	s := testStoreFull(t)
	body := render.BoutiqueBody(s)
	mainSlugs := map[string]bool{}
	for _, m := range catalog.MainOrder {
		mainSlugs[catalog.MainSlug[m]] = true
	}
	subSlugs := map[string]bool{}
	for _, p := range s.Products {
		subSlugs[catalog.Slugify(p.Sub)] = true
	}
	re := regexp.MustCompile(`data-sub="([^"]+)"`)
	vals := re.FindAllStringSubmatch(body, -1)
	if len(vals) != len(s.Products) {
		t.Fatalf("boutique : %d data-sub, %d produits attendus", len(vals), len(s.Products))
	}
	for _, v := range vals {
		if mainSlugs[v[1]] {
			t.Errorf("data-sub = %q, un slug de catégorie principale (ancien comportement)", v[1])
		}
		if !subSlugs[v[1]] {
			t.Errorf("data-sub = %q, pas un type connu", v[1])
		}
	}
}
