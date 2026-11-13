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

// TestLiveTopPages : chaque top live du menu « Toutes les catégories » a sa
// page (/categories/<slug>.html) avec le hub de SES types d'abord, puis la
// liste cachée — pas la boutique filtrée. (3 tops live n'ont aucun produit
// dans le catalogue actuel : leur page est vide, c'est normal.)
func TestLiveTopPages(t *testing.T) {
	s := testStoreFull(t)
	for _, lt := range catalog.LiveCats {
		fs := render.CategoryBodyLive(s, lt.Slug)
		page := render.CategoryPageLive(s, lt)
		if !strings.Contains(page, `<h1 style="font-size:24px">`+catalog.Esc(lt.Label)+`</h1>`) {
			t.Errorf("%s : titre absent", lt.Label)
		}
		if len(fs) == 0 {
			t.Logf("%s : aucun produit (page vide)", lt.Label)
			continue
		}
		if !strings.Contains(page, `data-hub`) {
			t.Errorf("%s : hub des types absent", lt.Label)
		}
		if !strings.Contains(page, `data-catalog-home hidden`) {
			t.Errorf("%s : liste des produits pas cachée derrière le hub", lt.Label)
		}
		// Le hub liste les types de CE top live, avec le bon compteur.
		counts := map[string]int{}
		for _, p := range fs {
			counts[p.Sub]++
		}
		hub := render.TypeHubHTML(fs)
		if got := strings.Count(hub, `class="hub-card"`); got != len(counts) {
			t.Errorf("%s : %d tuiles, %d types attendus", lt.Label, got, len(counts))
		}
		for sub, n := range counts {
			if !strings.Contains(hub, `href="?type=`+catalog.Slugify(sub)+`"`) {
				t.Errorf("%s : lien ?type= absent pour %q", lt.Label, sub)
			}
			want := fmt.Sprintf(`%d produit`, n)
			if n > 1 {
				want = fmt.Sprintf(`%d produits`, n)
			}
			if !strings.Contains(hub, want) {
				t.Errorf("%s : compteur %q absent pour %q", lt.Label, want, sub)
			}
		}
	}
	// Le cas de Zelef : Matériels réseaux doit montrer ses 5 types.
	fs := render.CategoryBodyLive(s, "materiels-reseaux")
	hub := render.TypeHubHTML(fs)
	for _, sub := range []string{"Réseau & câbles", "Cisco & réseau", "Téléphonie IP", "Réseau Legrand", "Adaptateurs"} {
		if !strings.Contains(hub, catalog.Esc(sub)) {
			t.Errorf("Matériels réseaux : type %q absent du hub", sub)
		}
	}
}

// TestMegaMenuVoirTous : « Voir tous les X » du méga-menu mène à la page
// catégorie (hub des types d'abord), jamais à la boutique filtrée.
func TestMegaMenuVoirTous(t *testing.T) {
	s := testStoreFull(t)
	panel := render.MegaPanel(s, "")
	re := regexp.MustCompile(`<a class="mega-all" href="([^"]*)">`)
	matches := re.FindAllStringSubmatch(panel, -1)
	if len(matches) != len(catalog.LiveCats) {
		t.Fatalf("méga-menu : %d liens « Voir tous », %d tops live attendus", len(matches), len(catalog.LiveCats))
	}
	seen := map[string]bool{}
	for _, m := range matches {
		href := m[1]
		seen[href] = true
		if strings.Contains(href, "boutique.html?store=") {
			t.Errorf("« Voir tous » : href=%q pointe encore vers la boutique filtrée", href)
		}
	}
	for _, lt := range catalog.LiveCats {
		want := "categories/" + lt.Slug + ".html"
		if !seen[want] {
			t.Errorf("top %q : pas de lien « Voir tous » vers %q", lt.Label, want)
		}
	}
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
