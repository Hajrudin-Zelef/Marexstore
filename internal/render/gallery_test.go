package render_test

import (
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

func ficheStore() *catalog.Store {
	return catalog.NewStore(nil)
}

// Sans photo supplémentaire : sortie inchangée (parité v3.25) —
// une seule image, aucune flèche, aucun data-gal.
func TestFicheBodySansExtraInchange(t *testing.T) {
	p := &catalog.RawProduct{Slug: "x", Nom: "X", Image: "images/x.jpg", Main: "M", Sub: "S"}
	got := render.FicheBody(ficheStore(), p)
	if strings.Contains(got, "data-gal") {
		t.Fatal("fiche 1 photo : data-gal inattendu (la sortie doit rester identique)")
	}
	if strings.Contains(got, "gal-prev") || strings.Contains(got, "gal-next") {
		t.Fatal("fiche 1 photo : flèches inattendues")
	}
	if n := strings.Count(got, "<img"); n != 1 {
		t.Fatalf("fiche 1 photo : %d <img, attendu 1", n)
	}
}

// Avec photos supplémentaires : carrousel avec flèches avant/arrière,
// miniatures cliquables et les 3 photos présentes.
func TestFicheBodyCarrousel(t *testing.T) {
	p := &catalog.RawProduct{Slug: "x", Nom: "X", Image: "images/x.jpg",
		Extras: []string{"images/x-2.jpg", "images/x-3.jpg"}, Main: "M", Sub: "S"}
	got := render.FicheBody(ficheStore(), p)
	for _, want := range []string{"data-gal", "data-gal-prev", "data-gal-next", "data-gal-main", "data-gal-thumb"} {
		if !strings.Contains(got, want) {
			t.Fatalf("carrousel : %q absent du rendu", want)
		}
	}
	for _, src := range []string{"images/x.jpg", "images/x-2.jpg", "images/x-3.jpg"} {
		if !strings.Contains(got, src) {
			t.Fatalf("carrousel : photo %q absente du rendu", src)
		}
	}
}
