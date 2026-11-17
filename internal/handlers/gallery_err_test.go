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

// Un échec galerie/photo ne doit jamais perdre la modif texte ni laisser
// le store en mémoire périmé : on persiste, on recharge, on redirige avec
// l'avertissement en param err (au lieu du 200 actuel qui saute reload).
func TestAdminProduitEditGalerieEnErreurGardeTexteEtRecharge(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,image) VALUES('p1','Ancien','images/p1.jpg')`); err != nil {
		t.Fatal(err)
	}
	app.reload()

	var body bytes.Buffer
	w := multipart.NewWriter(&body)
	_ = w.WriteField("csrf", app.adminCSRF())
	_ = w.WriteField("nom", "Nouveau nom")
	_ = w.WriteField("image", "images/p1.jpg")
	fw, err := w.CreateFormFile("photos", "evil.txt")
	if err != nil {
		t.Fatal(err)
	}
	_, _ = fw.Write([]byte("pas une image"))
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
		t.Fatalf("POST avec galerie en erreur : status %d, attendu 303 (texte conservé + redirect)", rec.Code)
	}
	loc := rec.Header().Get("Location")
	if !strings.Contains(loc, "err=") {
		t.Fatalf("redirect sans avertissement err : %q", loc)
	}
	var nom string
	if err := database.QueryRow(`SELECT nom FROM products WHERE slug='p1'`).Scan(&nom); err != nil {
		t.Fatal(err)
	}
	if nom != "Nouveau nom" {
		t.Fatalf("nom en DB : %q, attendu Nouveau nom", nom)
	}
	if app.Store.BySlug["p1"] == nil || app.Store.BySlug["p1"].Nom != "Nouveau nom" {
		t.Fatalf("store en mémoire périmé après erreur galerie (reload sauté)")
	}
}
