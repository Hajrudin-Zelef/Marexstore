package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Créer une catégorie (même vide) : persistée en declared_types, store
// rechargé, message honnête. Vide / "|" / doublon → err, rien en DB.
func TestAdminTypesCreeCategorie(t *testing.T) {
	app, database := testAppDesign(t)
	post := func(v url.Values) (int, string) {
		req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		app.AdminTypes(rec, req)
		return rec.Code, rec.Header().Get("Location")
	}
	newForm := func(top, sub string) url.Values {
		v := url.Values{}
		v.Set("csrf", app.adminCSRF())
		v.Set("action", "creer")
		v.Set("top", top)
		v.Set("sub", sub)
		return v
	}

	code, loc := post(newForm("t1", "z"))
	if code != http.StatusSeeOther || !strings.Contains(loc, "msg=") {
		t.Fatalf("créer : %d %q, attendu 303 msg=", code, loc)
	}
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top='t1' AND sub='z'`).Scan(&n); err != nil || n != 1 {
		t.Fatalf("déclarée absente en DB (n=%d, err=%v)", n, err)
	}
	if app.Store == nil {
		t.Fatal("store non rechargé après création")
	}

	for _, bad := range [][2]string{{"t1", ""}, {"t1", "a|b"}, {"t1", "  "}} {
		code, loc := post(newForm(bad[0], bad[1]))
		if code != http.StatusSeeOther || !strings.Contains(loc, "err=") {
			t.Fatalf("créer %q : %d %q, attendu 303 err=", bad[1], code, loc)
		}
	}
	if _, err := database.Exec(`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('p1','P1','t1','y')`); err != nil {
		t.Fatal(err)
	}
	if code, loc := post(newForm("t1", "y")); !strings.Contains(loc, "err=") {
		t.Fatalf("doublon : %d %q, attendu err=", code, loc)
	}
}
