package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Bouton + : chaque appui ajoute un volet d'upload (un fichier par volet),
// jusqu'à 4 extras (5 photos au total avec la principale). Fini l'input
// multiple unique : un seul fichier par volet, nom "photos" inchangé côté
// serveur.
func TestAdminProduitEditBoutonPlusGalerie(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,image,images_json) VALUES('p1','P1','images/p1.jpg','["images/p1-2.jpg"]')`); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/produits/p1", nil)
	req.SetBasicAuth("admin", "test1234")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "p1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	rec := httptest.NewRecorder()
	app.AdminProduitEdit(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET fiche : status %d, attendu 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"data-gal-add", `data-gal-free="3"`, `name="photos"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("formulaire galerie : %q absent", want)
		}
	}
	if strings.Contains(body, `name="photos" multiple`) {
		t.Fatal("formulaire galerie : l'input multiple unique doit disparaître (volets unitaires)")
	}
	if !strings.Contains(body, `value="images/p1-2.jpg"`) {
		t.Fatal("formulaire galerie : l'extra existant doit rester (case Supprimer)")
	}
}
