package render

import (
	"regexp"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

func i64(v int64) *int64 { return &v }

// Le stock d'abord : avec prix avant sans prix, "en stock" avant les autres,
// puis nom A-Z. Le prix boutique compte comme un prix (PrixAffiche).
func TestSortedStockFirst(t *testing.T) {
	in := []*catalog.RawProduct{
		{Slug: "devis-1", Nom: "Toner sans prix", Dispo: "en stock"},
		{Slug: "rupture", Nom: "Imprimante en rupture", Prix: i64(50000), Dispo: "rupture de stock"},
		{Slug: "stock-b", Nom: "Souris Logitech", Prix: i64(12000), Dispo: "en stock"},
		{Slug: "stock-a", Nom: "Clavier HP", PrixBoutique: i64(15000), Dispo: "en stock"},
		{Slug: "devis-2", Nom: "Cable sans prix", Dispo: "rupture de stock"},
		{Slug: "retour", Nom: "Ecran retour", Prix: i64(80000), Dispo: "retour"},
	}
	got := SortedStockFirst(in)
	want := []string{"stock-a", "stock-b", "retour", "rupture", "devis-1", "devis-2"}
	if len(got) != len(want) {
		t.Fatalf("longueur %d, attendu %d", len(got), len(want))
	}
	for i, w := range want {
		if got[i].Slug != w {
			t.Fatalf("position %d : %s, attendu %s (ordre complet %v)", i, got[i].Slug, w, slugs(got))
		}
	}
	// L'entrée n'est pas modifiée.
	if in[0].Slug != "devis-1" {
		t.Fatalf("l'entrée a été modifiée")
	}
}

func slugs(ps []*catalog.RawProduct) []string {
	var out []string
	for _, p := range ps {
		out = append(out, p.Slug)
	}
	return out
}

// La boutique (171 fiches) ne change que l'ordre : tous les produits avec un
// prix d'abord, puis ceux sans prix — aucun produit perdu ou ajouté.
func TestBoutiqueStockFirst(t *testing.T) {
	list, err := catalog.LoadRaw("../../catalogdata/produits-v3.json")
	if err != nil {
		t.Skip("catalogdata absent")
	}
	catalog.ApplyTaxonomy(list)
	s := catalog.NewStore(list)
	html := BoutiqueBody(s)
	// data-prix des cartes, dans l'ordre d'affichage.
	re := regexp.MustCompile(`<div class="card"[^>]*data-prix="([^"]*)"`)
	ms := re.FindAllStringSubmatch(html, -1)
	if len(ms) != len(s.Products) {
		t.Fatalf("%d cartes, attendu %d", len(ms), len(s.Products))
	}
	seenPriceless := false
	for i, m := range ms {
		if m[1] == "" {
			seenPriceless = true
		} else if seenPriceless {
			t.Fatalf("carte avec prix en position %d après une carte sans prix", i)
		}
	}
	// Même ensemble de produits, juste réordonné.
	slugRe := regexp.MustCompile(`href="produits/([^"]+)\.html"`)
	seen := map[string]bool{}
	for _, m := range slugRe.FindAllStringSubmatch(html, -1) {
		seen[m[1]] = true
	}
	for _, p := range s.Products {
		if !seen[p.Slug] {
			t.Fatalf("produit %s absent de la boutique", p.Slug)
		}
	}
}
