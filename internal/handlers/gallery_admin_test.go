package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// POST fiche produit avec 2 photos galerie : images_json persisté,
// fichiers écrits images/<slug>-<k>.<ext>, sans toucher la principale.
func TestAdminProduitGalerieUpload(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,image) VALUES('p1','P1','images/p1.jpg')`); err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("csrf", app.adminCSRF())
	_ = w.WriteField("nom", "P1")
	_ = w.WriteField("image", "images/p1.jpg")
	for _, f := range [][2]string{{"photos", "a.png"}, {"photos", "b.jpg"}} {
		fw, err := w.CreateFormFile(f[0], f[1])
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write([]byte("fakeimg:" + f[1])); err != nil {
			t.Fatal(err)
		}
	}
	if err := w.Close(); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/produits/p1", &body)
	req.Header.Set("Content-Type", w.FormDataContentType())
	req.SetBasicAuth("admin", "test1234")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("slug", "p1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))

	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	_ = os.Chdir(t.TempDir())
	if err := os.MkdirAll("images", 0755); err != nil {
		t.Fatal(err)
	}

	rec := httptest.NewRecorder()
	app.AdminProduitEdit(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST galerie : status %d, attendu 303 (corps : %s)", rec.Code, rec.Body.String())
	}
	var js string
	if err := database.QueryRow(`SELECT images_json FROM products WHERE slug='p1'`).Scan(&js); err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(js, "images/p1-2.png") || !strings.Contains(js, "images/p1-3.jpg") {
		t.Fatalf("images_json : %s, attendu [p1-2.png p1-3.jpg]", js)
	}
	for _, f := range []string{"images/p1-2.png", "images/p1-3.jpg"} {
		if st, err := os.Stat(filepath.Join(".", f)); err != nil || st.Size() == 0 {
			t.Fatalf("fichier galerie absent ou vide : %s (%v)", f, err)
		}
	}
	var main string
	if err := database.QueryRow(`SELECT image FROM products WHERE slug='p1'`).Scan(&main); err != nil {
		t.Fatal(err)
	}
	if main != "images/p1.jpg" {
		t.Fatalf("photo principale modifiée : %q", main)
	}
}
