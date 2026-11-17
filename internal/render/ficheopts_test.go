package render_test

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

func ficheOptsStore() (*catalog.Store, *catalog.RawProduct) {
	img := "images/x.jpg"
	prix := int64(100000)
	p := &catalog.RawProduct{Slug: "x", Nom: "X", Image: img, Prix: &prix,
		Main: "M", Sub: "S", Dispo: "en stock", Accroche: "acc", Description: "desc",
		Specs:        json.RawMessage(`[{"label":"Poids","valeur":"1 kg"}]`),
		Consommables: []catalog.Consommable{{Nom: "C1", Prix: "5000"}},
		PointsForts:  []string{"robuste"}}
	autre := &catalog.RawProduct{Slug: "y", Nom: "Y", Image: "images/y.jpg", Main: "M", Sub: "S"}
	return catalog.NewStore([]*catalog.RawProduct{p, autre}), p
}

// Par défaut : byte-identique à FicheBody (parité prod, aucune régression).
func TestFicheBodyOptsDefautIdentique(t *testing.T) {
	s, p := ficheOptsStore()
	if got, want := render.FicheBodyOpts(s, p, render.DefaultFicheOpts()), render.FicheBody(s, p); got != want {
		t.Fatal("FicheBodyOpts par défaut diffère de FicheBody (parité cassée)")
	}
}

// Chaque bascule masque son bloc, le reste reste visible.
func TestFicheBodyOptsBascule(t *testing.T) {
	s, p := ficheOptsStore()
	cas := []struct {
		nom     string
		off     func(*render.FicheOpts)
		absent  string
		present string
	}{
		{"wa", func(o *render.FicheOpts) { o.Wa = false }, "btn wa", "Demander un devis"},
		{"devis", func(o *render.FicheOpts) { o.Devis = false }, "Demander un devis", "btn wa"},
		{"specs", func(o *render.FicheOpts) { o.Specs = false }, "Caractéristiques", "Consommables"},
		{"conso", func(o *render.FicheOpts) { o.Conso = false }, "Consommables", "Caractéristiques"},
		{"similar", func(o *render.FicheOpts) { o.Similar = false }, "Produits similaires", "Caractéristiques"},
		{"avis", func(o *render.FicheOpts) { o.Avis = false }, "Avis (0)", "Description"},
	}
	for _, c := range cas {
		o := render.DefaultFicheOpts()
		c.off(&o)
		got := render.FicheBodyOpts(s, p, o)
		if strings.Contains(got, c.absent) {
			t.Fatalf("%s OFF : %q encore présent", c.nom, c.absent)
		}
		if !strings.Contains(got, c.present) {
			t.Fatalf("%s OFF : %q a disparu (trop masqué)", c.nom, c.present)
		}
	}
}

// Libellés et nombre de similaires pilotés (textes du Studio).
func TestFicheBodyOptsTextesEtNombre(t *testing.T) {
	s, p := ficheOptsStore()
	o := render.DefaultFicheOpts()
	o.BtnWa, o.BtnDevis, o.SimilarNb = "Appelle-moi", "Devis vite", 1
	got := render.FicheBodyOpts(s, p, o)
	for _, want := range []string{"Appelle-moi", "Devis vite", ">Devis vite</a>", ">Appelle-moi</a>"} {
		if !strings.Contains(got, want) {
			t.Fatalf("libellé %q absent", want)
		}
	}
	if n := strings.Count(got, "data-live-sub="); n != 1 {
		t.Fatalf("similaires : %d cartes, attendu 1", n)
	}
}
