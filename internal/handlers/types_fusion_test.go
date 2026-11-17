package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// La fusion doit vraiment déplacer les produits (src "top|sub" → live_sub
// cible), recharger le store et l'annoncer. Sans champ "top" (plus de JS
// inline) : le store se déduit de la source côté serveur.
func TestAdminTypesFusionneVraiment(t *testing.T) {
	app, database := testAppDesign(t)
	for _, q := range []string{
		`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a1','A1','t1','x')`,
		`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a2','A2','t1','x')`,
		`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('b1','B1','t1','y')`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("src", "t1|x")
	form.Set("dst", "t1|y")
	req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")

	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST fusion : status %d, attendu 303", rec.Code)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "msg=") {
		t.Fatalf("POST fusion : redirect %q, attendu un msg de succès (pas err)", loc)
	}
	var nx, ny int
	if err := database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='x'`).Scan(&nx); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='y'`).Scan(&ny); err != nil {
		t.Fatal(err)
	}
	if nx != 0 || ny != 3 {
		t.Fatalf("fusion : x=%d y=%d, attendu x=0 y=3", nx, ny)
	}
	if app.Store == nil {
		t.Fatal("store non rechargé après fusion (aucun effet prod)")
	}
	if p := app.Store.BySlug["a1"]; p == nil || p.LiveSub != "y" {
		t.Fatalf("store périmé : a1.LiveSub=%q, attendu y", p.LiveSub)
	}
}

// Source inexistante : message honnête, pas de "fusionnés".
func TestAdminTypesFusionSourceInconnue(t *testing.T) {
	app, _ := testAppDesign(t)
	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("src", "t9|zzz")
	form.Set("dst", "t9|yyy")
	req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")

	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "err=") {
		t.Fatalf("fusion à vide : redirect %q, attendu err=", loc)
	}
}
