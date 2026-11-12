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

func normPage(s string) string {
	s = reHtmx.ReplaceAllString(s, "")
	s = reMirror.ReplaceAllString(s, "")
	s = reVer.ReplaceAllString(s, "assets/X")
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, "csrf-token") {
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
	if got != want {
		diffFirst(t, got, want)
	}
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
		if got != want {
			t.Logf("catégorie %s", m)
			diffFirst(t, got, want)
		}
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
