package render_test

import (
	"os"
	"regexp"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// storeCatalogue : les trois jeux de données, dans l'ordre d'import.
func storeCatalogue(t *testing.T) *catalog.Store {
	t.Helper()
	var list []*catalog.RawProduct
	for _, f := range []string{"produits-v3.json", "nouveaux-produits-commercial.json", "nouveaux-produits-commercial-lot2.json"} {
		l, err := catalog.LoadRaw(rel("../../catalogdata/" + f))
		if err != nil {
			t.Skip("catalogdata absent (hors dépôt local)")
		}
		list = append(list, l...)
	}
	catalog.ApplyTaxonomy(list)
	return catalog.NewStore(list)
}

// TestAccueilRecents : « Ajouté récemment » doit montrer les derniers
// produits importés (ordre de chargement = rowid), les plus récents d'abord,
// et non un ancien sitemap.
func TestAccueilRecents(t *testing.T) {
	s := storeCatalogue(t)
	r := s.Recents("")
	if len(r) != 12 {
		t.Fatalf("%d produits récents, attendu 12", len(r))
	}
	n := len(s.Products)
	if r[0] != s.Products[n-1] {
		t.Fatalf("premier récent = %q, attendu le dernier produit chargé %q", r[0].Slug, s.Products[n-1].Slug)
	}
	for i := 1; i < len(r); i++ {
		if r[i] != s.Products[n-1-i] {
			t.Fatalf("récent %d = %q, attendu %q (ordre décroissant)", i, r[i].Slug, s.Products[n-1-i].Slug)
		}
	}
}

// TestAccueilMarques : la section marques est une grille de cartes avec un
// logo SVG existant, et un lien vers la boutique filtrée sur la marque.
// Une marque sans produit est signalée (la boutique filtrée serait vide)
// sans faire échouer le test : c'est un choix éditorial.
func TestAccueilMarques(t *testing.T) {
	s := storeCatalogue(t)
	html := render.Doc("Accueil", render.IndexBody(s, s.Recents("")), "", "")
	re := regexp.MustCompile(`<a class="brand-card" href="([^"]+)"><img src="static/logos/([a-z0-9-]+)\.svg"`)
	matches := re.FindAllStringSubmatch(html, -1)
	if len(matches) != len(render.Marques) {
		t.Fatalf("%d cartes de marques, attendu %d", len(matches), len(render.Marques))
	}
	for i, m := range matches {
		href, logo := m[1], m[2]
		if !strings.HasPrefix(href, "boutique.html?marque=") {
			t.Errorf("carte %d : lien %q ne pointe pas vers la boutique filtrée", i, href)
		}
		if _, err := os.Stat("../../static/logos/" + logo + ".svg"); err != nil {
			t.Errorf("carte %q : logo static/logos/%s.svg absent", href, logo)
		}
		marque := render.Marques[i]
		found := 0
		for _, p := range s.Products {
			if strings.EqualFold(p.Marque, marque) {
				found++
			}
		}
		if found == 0 {
			t.Logf("attention : marque %q sans produit (boutique filtrée vide)", marque)
		}
	}
}
