package render_test

import (
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

func TestFicheBlocGamme(t *testing.T) {
	prix := int64(45000)
	a := &catalog.RawProduct{Slug: "a", Nom: "Cartouche d'encre HP 305", Marque: "HP", Main: "M", Sub: "S", Prix: &prix}
	b := &catalog.RawProduct{Slug: "b", Nom: "Cartouche d'encre HP 305XL", Marque: "HP", Main: "M", Sub: "S", Prix: &prix}
	seul := &catalog.RawProduct{Slug: "c", Nom: "Écran Dell E1715S", Marque: "Dell", Main: "M", Sub: "S"}
	s := catalog.NewStore([]*catalog.RawProduct{a, b, seul})
	body := render.FicheBody(s, a)
	for _, want := range []string{"Modèles de la gamme", `class="gchip on"`, `href="../produits/b.html"`, "305XL"} {
		if !strings.Contains(body, want) {
			t.Fatalf("fiche a : %q absent", want)
		}
	}
	if strings.Contains(render.FicheBody(s, seul), "Modèles de la gamme") {
		t.Fatal("fiche sans gamme : bloc présent à tort")
	}
}

func TestFicheCarrouselExtras(t *testing.T) {
	p := &catalog.RawProduct{Slug: "x", Nom: "X", Marque: "M", Main: "M", Sub: "S",
		Image: "images/x.jpg", Extras: []string{"images/x-2.jpg", "images/x-3.jpg", "images/x-4.jpg"}}
	s := catalog.NewStore([]*catalog.RawProduct{p})
	body := render.FicheBody(s, p)
	for _, want := range []string{"data-gal", "data-gal-main", "gal-thumbs", "images/x-4.jpg"} {
		if !strings.Contains(body, want) {
			t.Fatalf("carrousel : %q absent", want)
		}
	}
	p2 := &catalog.RawProduct{Slug: "y", Nom: "Y", Marque: "M", Main: "M", Sub: "S", Image: "images/y.jpg"}
	s2 := catalog.NewStore([]*catalog.RawProduct{p2})
	if strings.Contains(render.FicheBody(s2, p2), "data-gal") {
		t.Fatal("photo unique : carrousel présent à tort (parité)")
	}
}
