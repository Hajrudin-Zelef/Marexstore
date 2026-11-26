package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Non-régression : la création d'un produit via POST /admin/produits/nouveau
// doit insérer la ligne (303 vers la fiche), pas répondre 200
// « Création impossible ». L'édition (UPDATE) et la duplication (-copie)
// sont vérifiées dans la foulée : elles partagent le même formulaire.
func TestAdminProduitCreerEditerDupliquer(t *testing.T) {
	app, database := testAppDesign(t)

	post := func(target string, form url.Values) *httptest.ResponseRecorder {
		form.Set("csrf", app.adminCSRF())
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		r := chi.NewRouter()
		r.Post("/admin/produits/nouveau", app.AdminProduitEdit)
		r.Post("/admin/produits/{slug}", app.AdminProduitEdit)
		r.Get("/admin/produits/{slug}", app.AdminProduitEdit)
		r.ServeHTTP(rec, req)
		return rec
	}

	// 1. Création.
	rec := post("/admin/produits/nouveau", url.Values{
		"nom": {"Produit Test Creation"}, "marque": {"ZZ"}, "prix": {"1000"},
		"dispo": {"en stock"}, "stock": {"5"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("création : status %d, attendu 303 (body: %.200s)", rec.Code, rec.Body.String())
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "/admin/produits/produit-test-creation") {
		t.Fatalf("création : Location %q inattendue", loc)
	}
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM products WHERE slug='produit-test-creation'`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("création : ligne absente en base (n=%d, err=%v)", n, err)
	}

	// 2. Fiche : le produit créé s'affiche.
	req := httptest.NewRequest(http.MethodGet, "/admin/produits/produit-test-creation", nil)
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	r := chi.NewRouter()
	r.Get("/admin/produits/{slug}", app.AdminProduitEdit)
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Produit Test Creation") {
		t.Fatalf("fiche : status %d", rec.Code)
	}

	// 3. Édition.
	rec = post("/admin/produits/produit-test-creation", url.Values{
		"nom": {"Produit Test Creation"}, "marque": {"ZZ2"}, "prix": {"2000"}, "dispo": {"en stock"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("édition : status %d, attendu 303", rec.Code)
	}
	var marque string
	if err := database.QueryRow(`SELECT marque FROM products WHERE slug='produit-test-creation'`).Scan(&marque); err != nil || marque != "ZZ2" {
		t.Fatalf("édition : marque=%q, err=%v", marque, err)
	}

	// 4. Duplication.
	req = httptest.NewRequest(http.MethodGet, "/admin/produits/produit-test-creation?dupliquer=1", nil)
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	rd := chi.NewRouter()
	rd.Get("/admin/produits/{slug}", app.AdminProduitEdit)
	rd.ServeHTTP(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("duplication : status %d, attendu 303", rec.Code)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM products WHERE slug='produit-test-creation-copie'`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("duplication : copie absente (n=%d, err=%v)", n, err)
	}
}
