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

// Rendu : chaque carte a son "+" (details natif, sans JS), chaque ligne ses
// data attrs, et les déclarées vides s'affichent en bas avec compteur 0.
func TestAdminTypesRenduBoutonPlus(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('p1','P1','t1','x')`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO declared_types(top, sub) VALUES('t1','z')`); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/types", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET types : %d, attendu 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"data-types-add", `data-top="t1" data-sub="z"`, ">0<"} {
		if !strings.Contains(body, want) {
			t.Fatalf("rendu types : %q absent", want)
		}
	}
	ix, iz := strings.Index(body, `data-sub="x"`), strings.Index(body, `data-sub="z"`)
	if ix < 0 || iz < 0 || ix > iz {
		t.Fatal("la déclarée vide (z) doit s'afficher en bas, après x")
	}
	if strings.Contains(body, "<script") && !strings.Contains(body, `<script src=`) {
		t.Fatal("JS inline détecté (doctrine script-src 'self')")
	}
}
