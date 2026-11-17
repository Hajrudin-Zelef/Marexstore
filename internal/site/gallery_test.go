package site

import (
	"path/filepath"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
)

// Galerie persistée : la colonne images_json survit au rechargement
// (migration idempotente + rowToProduct) et alimente Gallery().
func TestLoadStoreLitImagesJson(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO products(slug,nom,image,images_json) VALUES('p1','P1','images/p1.jpg','["images/p1-2.jpg","images/p1-3.jpg"]')`); err != nil {
		t.Fatal(err)
	}
	st, err := LoadStore(database)
	if err != nil {
		t.Fatal(err)
	}
	p, ok := st.BySlug["p1"]
	if !ok {
		t.Fatal("produit p1 introuvable après LoadStore")
	}
	got := p.Gallery()
	if len(got) != 3 || got[0] != "images/p1.jpg" || got[2] != "images/p1-3.jpg" {
		t.Fatalf("galerie rechargée : %v, attendu [p1.jpg p1-2.jpg p1-3.jpg]", got)
	}
}
