package catalog

import (
	"encoding/json"
	"testing"
)

// Galerie fiche produit : photo principale + extras, plafond 5 photos.
func TestGaleriePlafondCinqPhotos(t *testing.T) {
	p := &RawProduct{Slug: "x", Nom: "X", Image: "images/x.jpg",
		Extras: []string{"images/x-2.jpg", "images/x-3.jpg", "images/x-4.jpg", "images/x-5.jpg", "images/x-6.jpg", "images/x-7.jpg"}}
	got := p.Gallery()
	if len(got) != 5 {
		t.Fatalf("galerie : %d photos, attendu 5 (1 principale + 4 extras)", len(got))
	}
	if got[0] != "images/x.jpg" {
		t.Fatalf("galerie[0] : %q, attendu la photo principale", got[0])
	}
	if got[4] != "images/x-5.jpg" {
		t.Fatalf("galerie[4] : %q, attendu images/x-5.jpg", got[4])
	}
}

func TestGalerieSansExtra(t *testing.T) {
	p := &RawProduct{Slug: "x", Nom: "X", Image: "images/x.jpg"}
	got := p.Gallery()
	if len(got) != 1 || got[0] != "images/x.jpg" {
		t.Fatalf("galerie sans extra : %v, attendu [images/x.jpg]", got)
	}
}

func TestGalerieSansPhoto(t *testing.T) {
	p := &RawProduct{Slug: "x", Nom: "X"}
	if got := p.Gallery(); len(got) != 0 {
		t.Fatalf("galerie sans photo : %v, attendu vide", got)
	}
}

// Le champ "images" des JSON sources est mort (fichiers souvent absents) :
// il doit rester ignoré pour ne pas casser les fiches ni la parité v3.25.
func TestGalerieIgnoreImagesJSONSources(t *testing.T) {
	var p RawProduct
	if err := json.Unmarshal([]byte(`{"slug":"x","nom":"X","image":"images/x.jpg","images":["images/x-2.jpg"]}`), &p); err != nil {
		t.Fatal(err)
	}
	got := p.Gallery()
	if len(got) != 1 || got[0] != "images/x.jpg" {
		t.Fatalf("galerie : %v, attendu [images/x.jpg] (images JSON ignorées)", got)
	}
}
