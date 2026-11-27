package handlers

import (
	"bytes"
	"context"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// 5 photos d'un coup (plafond 4 extras) : les 4 premières passent, la 5e est
// signalée honnêtement dans le redirect au lieu d'être jetée en silence.
func TestAdminProduitEditGaleriePlafondSignale(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom) VALUES('p1','P1')`); err != nil {
		t.Fatal(err)
	}

	png := []byte("\x89PNG\r\n\x1a\n\x00\x00\x00\rIHDR\x00\x00\x00\x01\x00\x00\x00\x01\x08\x02\x00\x00\x00\x90wS\xde\x00\x00\x00\x0cIDATx\x9cc\xf8\x0f\x00\x00\x01\x01\x00\x05\x1c\xd5~\x00\x00\x00\x00IEND\xaeB`\x82")
	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("csrf", app.adminCSRF())
	_ = w.WriteField("nom", "P1")
	for i := 0; i < 5; i++ {
		fw, err := w.CreateFormFile("photos", "g.png")
		if err != nil {
			t.Fatal(err)
		}
		if _, err := fw.Write(png); err != nil {
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
		t.Fatalf("POST 5 photos : status %d, attendu 303", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(strings.ToLower(loc), "ignor") {
		t.Fatalf("5e photo jetée sans le dire : redirect %q (attendu un avertissement 'ignorée')", loc)
	}
	var js string
	if err := database.QueryRow(`SELECT images_json FROM products WHERE slug='p1'`).Scan(&js); err != nil {
		t.Fatal(err)
	}
	if strings.Count(js, "images/p1-") != 4 {
		t.Fatalf("plafond non respecté : %s", js)
	}
}
