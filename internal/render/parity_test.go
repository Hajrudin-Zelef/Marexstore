package render_test

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Charge le catalogue comme le serveur.

func rel(p string) string {
	_, f, _, _ := runtime.Caller(1)
	return filepath.Join(filepath.Dir(f), p)
}

func testStore(t *testing.T) *catalog.Store {
	t.Helper()
	list, err := catalog.LoadRaw(rel("../../catalogdata/produits-v3.json"))
	if err != nil {
		t.Skip("catalogdata absent (hors dépôt local)")
	}
	catalog.ApplyTaxonomy(list)
	return catalog.NewStore(list)
}

// Normalise une page pour comparaison : retire les différences voulues
// (csrf, scripts htmx/mirror ajoutés, texte compte synchronisé).
var reHtmx = regexp.MustCompile("<script[^>]*htmx[^>]*></script>\n?")
var reMirror = regexp.MustCompile("<script[^>]*mirror[^>]*></script>\n?")
var reVer = regexp.MustCompile(`assets/[a-z0-9_.-]+\.css\?v=\d+|assets/[a-z0-9_.-]+\.js\?v=\d+`)
var reFiltersToggle = regexp.MustCompile(`<button type="button" class="filters-toggle"[^>]*>.*?</button>\n?`)
var reMegaAll = regexp.MustCompile(`<a class="mega-all" href="[^"]*"`)
var reHubMain = regexp.MustCompile(`<a class="hub-main" href="[^"]*"`)
var reHubSub = regexp.MustCompile(`<li><a href="[^"]*">`)
var reMegaGroups = regexp.MustCompile(`(<div class="mega-groups">).*?(</div><a class="mega-all")`)
var reMegaCount = regexp.MustCompile(`<span class="mega-count">\(\d+\)</span>`)

func normPage(s string) string {
	s = reHtmx.ReplaceAllString(s, "")
	s = reMirror.ReplaceAllString(s, "")
	s = reVer.ReplaceAllString(s, "assets/X")
	s = reFiltersToggle.ReplaceAllString(s, "")
	// « Voir tous » du méga-menu : pointe maintenant vers la page catégorie
	// (hub des types d'abord), plus vers la boutique filtrée — vérifié par
	// son propre test, pas par la parité avec l'archive v3.25.
	s = reMegaAll.ReplaceAllString(s, `<a class="mega-all" href="MEGAALL"`)
	// Tuiles « Nos catégories » : mènent aux pages catégories (hub des types
	// d'abord), plus à la boutique filtrée — vérifié par TestBoutiqueHubLinks.
	s = reHubMain.ReplaceAllString(s, `<a class="hub-main" href="HUBMAIN"`)
	s = reHubSub.ReplaceAllString(s, `<li><a href="HUBSUB">`)
	// Méga-menu : les sous-catégories à 0 produit ne sont plus liées (plus de
	// page vide) et « Voir + » ne liste que les sous-catégories vivantes —
	// vérifié par TestFilterLinksAlive, pas par la parité v3.25.
	// Méga-menu : groupes de sous-catégories (données 171 vs 579 produits) et
	// compteurs du rail — vérifiés par TestFilterLinksAlive/TestMegaMenuVoirTous.
	s = reMegaGroups.ReplaceAllString(s, `${1}GROUPES${2}`)
	s = reMegaCount.ReplaceAllString(s, `<span class="mega-count">(N)</span>`)
	// Hub des types (pages catégories, nouveau) : la fermeture du bloc
	// [data-catalog-home] se colle à celle du catalogue.
	s = strings.ReplaceAll(s, "</div></div>\n</div></div></div>", "</div>\n</div></div></div>")
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, "csrf-token") {
			continue
		}
		// Nouveautés retirées de la parité (vérifiées par leurs propres tests) :
		// le hub des types, le bloc [data-catalog-home] des pages catégories,
		// le filtre global de l'accueil et sa carte des types.
		if strings.Contains(ln, "data-hub") || strings.Contains(ln, "data-catalog-home hidden") || strings.Contains(ln, "data-home-filter") || strings.Contains(ln, "MEDIATEC_TYPES") {
			continue
		}
		ln = strings.ReplaceAll(ln, "Votre compte est synchronisé sur tous vos appareils : connectez-vous pour retrouver votre panier et vos commandes.", "Votre compte est enregistré uniquement dans ce navigateur (site statique, sans serveur) : il n'est pas synchronisé avec un autre appareil. Pensez à noter votre mot de passe.")
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

func readOrig(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(rel("../../site-orig/" + path))
	if err != nil {
		t.Skip("site-orig absent")
	}
	return string(b)
}

func TestBoutiqueParity(t *testing.T) {
	s := testStore(t)
	got := normPage(render.Doc("Store", render.BoutiqueBody(s), "", ""))
	want := normPage(readOrig(t, "boutique.html"))
	// Le tri "stock d'abord" réordonne les cartes : même jeu de cartes,
	// reste de la page identique à l'octet près.
	assertSameCards(t, got, want)
}

func TestIndexParity(t *testing.T) {
	s := testStore(t)
	recents := s.Recents(rel("../../catalogdata/store-sitemap.xml"))
	got := normPage(render.Doc("Accueil", render.IndexBody(s, recents), "", ""))
	want := normPage(readOrig(t, "index.html"))
	if got != want {
		diffFirst(t, got, want)
	}
}

func TestCategoriesParity(t *testing.T) {
	s := testStore(t)
	for _, m := range catalog.MainOrder {
		got := normPage(render.Doc(m, render.CategoryPage(s, m), "../", ""))
		want := normPage(readOrig(t, "categories/"+catalog.MainSlug[m]+".html"))
		t.Logf("catégorie %s", m)
		assertSameCards(t, got, want)
	}
}

func TestProduitsParity(t *testing.T) {
	s := testStore(t)
	for _, p := range s.Products {
		got := normPage(render.Doc(p.Nom, render.Header(s, "../", "")+render.FicheBody(s, p)+"\n"+render.Footer(s, "../"), "../", ""))
		want := normPage(readOrig(t, "produits/"+p.Slug+".html"))
		if got != want {
			t.Logf("produit %s", p.Slug)
			diffFirst(t, got, want)
		}
	}
}

func TestFacetsParity(t *testing.T) {
	s := testStore(t)
	raw, err := os.ReadFile(rel("../../site-orig/assets/facets.js"))
	if err != nil {
		t.Skip("site-orig absent")
	}
	py := strings.TrimPrefix(string(raw), "/*")
	// extrait l'objet JSON : de "window.MEDIATEC_FACETS = " jusqu'au dernier ";"
	i := strings.Index(py, "window.MEDIATEC_FACETS = ")
	j := strings.LastIndex(py, ";")
	var want any
	if err := json.Unmarshal([]byte(py[i+len("window.MEDIATEC_FACETS = "):j]), &want); err != nil {
		t.Fatal(err)
	}
	var got any
	if err := json.Unmarshal([]byte(s.FacetsJS()[strings.Index(s.FacetsJS(), "{"):strings.LastIndex(s.FacetsJS(), "}")+1]), &got); err != nil {
		t.Fatal(err)
	}
	wb, _ := json.Marshal(want)
	gb, _ := json.Marshal(got)
	if string(wb) != string(gb) {
		t.Fatalf("facettes différentes:\nwant=%s\ngot=%s", wb, gb)
	}
}

// extractCards retire les cartes produit (<div class="card" ...>...</div>,
// profondeur équilibrée) : renvoie la page sans cartes + slugs dans l'ordre.
// Sert aux pages dont seul l'ordre des cartes peut changer (tri stock d'abord).
func extractCards(html string) (string, []string) {
	var out strings.Builder
	var slugs []string
	reSlug := regexp.MustCompile(`data-slug="([^"]+)"`)
	i := 0
	for {
		j := strings.Index(html[i:], `<div class="card"`)
		if j < 0 {
			out.WriteString(html[i:])
			break
		}
		j += i
		out.WriteString(html[i:j])
		depth := 0
		k := j
		for k < len(html) {
			if strings.HasPrefix(html[k:], "<div") {
				depth++
				k += 4
			} else if strings.HasPrefix(html[k:], "</div>") {
				depth--
				k += 6
				if depth == 0 {
					break
				}
			} else {
				k++
			}
		}
		if m := reSlug.FindStringSubmatch(html[j:k]); m != nil {
			slugs = append(slugs, m[1])
		}
		out.WriteString("<!--card-->")
		i = k
	}
	return out.String(), slugs
}

// assertSameCards : hors-cartes strictement identique + même multijeu de cartes.
func assertSameCards(t *testing.T, got, want string) {
	t.Helper()
	gs, gslugs := extractCards(got)
	ws, wslugs := extractCards(want)
	if gs != ws {
		diffFirst(t, gs, ws)
	}
	if len(gslugs) != len(wslugs) {
		t.Fatalf("cartes : got %d, want %d", len(gslugs), len(wslugs))
	}
	counts := map[string]int{}
	for _, s := range wslugs {
		counts[s]++
	}
	for _, s := range gslugs {
		counts[s]--
		if counts[s] < 0 {
			t.Fatalf("carte inattendue ou en trop : %s", s)
		}
	}
}

func diffFirst(t *testing.T, got, want string) {
	t.Helper()
	gl, wl := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(gl) && i < len(wl); i++ {
		if gl[i] != wl[i] {
			ga, wa := gl[i], wl[i]
			j := 0
			for j < len(ga) && j < len(wa) && ga[j] == wa[j] {
				j++
			}
			lo := j - 60
			if lo < 0 {
				lo = 0
			}
			hi := j + 120
			if hi > len(ga) {
				hi = len(ga)
			}
			hi2 := j + 120
			if hi2 > len(wa) {
				hi2 = len(wa)
			}
			t.Fatalf("ligne %d diffère à %d:\n got: ...%s...\nwant: ...%s...", i+1, j, ga[lo:hi], wa[lo:hi2])
		}
	}
	t.Fatalf("longueurs différentes: got %d lignes, want %d lignes", len(gl), len(wl))
}
