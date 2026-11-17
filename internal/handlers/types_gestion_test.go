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

// Renommer : produits + déclaré déplacés en transaction, FTS + store
// rechargés. Cible déjà prise → err, rien ne bouge (rollback).
func TestAdminTypesRenomme(t *testing.T) {
	app, database := testAppDesign(t)
	for _, q := range []string{
		`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a1','A1','t1','x')`,
		`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a2','A2','t1','x')`,
		`INSERT INTO declared_types(top, sub) VALUES('t1','x')`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	v := url.Values{}
	v.Set("csrf", app.adminCSRF())
	v.Set("action", "renommer")
	v.Set("top", "t1")
	v.Set("ancien", "x")
	v.Set("nouveau", "z")
	req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(v.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	if rec.Code != http.StatusSeeOther || !strings.Contains(rec.Header().Get("Location"), "msg=") {
		t.Fatalf("renommer : %d %q, attendu 303 msg=", rec.Code, rec.Header().Get("Location"))
	}
	var nx, nz, dx, dz int
	_ = database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='x'`).Scan(&nx)
	_ = database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='z'`).Scan(&nz)
	_ = database.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top='t1' AND sub='x'`).Scan(&dx)
	_ = database.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top='t1' AND sub='z'`).Scan(&dz)
	if nx != 0 || nz != 2 || dx != 0 || dz != 1 {
		t.Fatalf("renommage partiel : produits x=%d z=%d, déclarés x=%d z=%d", nx, nz, dx, dz)
	}
	if p := app.Store.BySlug["a1"]; p == nil || p.LiveSub != "z" {
		t.Fatal("store périmé après renommage")
	}
}

func TestAdminTypesRenommeRefuseCiblePrise(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a1','A1','t1','x'),('b1','B1','t1','y')`); err != nil {
		t.Fatal(err)
	}
	v := url.Values{}
	v.Set("csrf", app.adminCSRF())
	v.Set("action", "renommer")
	v.Set("top", "t1")
	v.Set("ancien", "x")
	v.Set("nouveau", "y")
	req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(v.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "err=") {
		t.Fatalf("cible prise : %q, attendu err=", loc)
	}
	var nx int
	_ = database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='x'`).Scan(&nx)
	if nx != 1 {
		t.Fatalf("rollback raté : x=%d, attendu 1", nx)
	}
}

// Menu ⋯ par ligne : renommer / ajouter / fusionner / voir / supprimer,
// suppression et fusion avec confirmation, formulaires postables sans JS.
func TestAdminTypesRenduMenu(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('p1','P1','t1','x'),('p2','P2','t1','y')`); err != nil {
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
	for _, want := range []string{"data-types-menu", "Renommer", "Ajouter", "Fusionner", "Voir les produits", "Supprimer", "data-mx-confirm", `name="dst"`, `value="supprimer"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("menu ligne : %q absent", want)
		}
	}
}

// Fusion depuis la carte (même moteur que le legacy) : DB + store + FTS.
func TestAdminTypesFusionneDepuisCarte(t *testing.T) {
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
	v := url.Values{}
	v.Set("csrf", app.adminCSRF())
	v.Set("action", "fusionner")
	v.Set("src", "t1|x")
	v.Set("dst", "t1|y")
	req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(v.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	if rec.Code != http.StatusSeeOther || !strings.Contains(rec.Header().Get("Location"), "msg=") {
		t.Fatalf("fusion carte : %d %q, attendu 303 msg=", rec.Code, rec.Header().Get("Location"))
	}
	var ny int
	_ = database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='y'`).Scan(&ny)
	if ny != 3 {
		t.Fatalf("fusion carte : y=%d, attendu 3", ny)
	}
	if p := app.Store.BySlug["a1"]; p == nil || p.LiveSub != "y" {
		t.Fatal("store périmé après fusion carte")
	}
}

// Supprimer : vide OK (msg + ligne partie), non vide refusé (err + intact).
func TestAdminTypesSupprimeVideEtRefuseNonVide(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a1','A1','t1','x')`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO declared_types(top, sub) VALUES('t1','z')`); err != nil {
		t.Fatal(err)
	}
	post := func(top, sub string) string {
		v := url.Values{}
		v.Set("csrf", app.adminCSRF())
		v.Set("action", "supprimer")
		v.Set("top", top)
		v.Set("sub", sub)
		req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		app.AdminTypes(rec, req)
		return rec.Header().Get("Location")
	}
	if loc := post("t1", "x"); !strings.Contains(loc, "err=") {
		t.Fatalf("supprimer non vide : %q, attendu err=", loc)
	}
	if loc := post("t1", "z"); !strings.Contains(loc, "msg=") {
		t.Fatalf("supprimer vide : %q, attendu msg=", loc)
	}
	var n int
	_ = database.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top='t1' AND sub='z'`).Scan(&n)
	if n != 0 {
		t.Fatal("type vide toujours déclaré après suppression")
	}
	var nx int
	_ = database.QueryRow(`SELECT COUNT(*) FROM products WHERE live_top='t1' AND live_sub='x'`).Scan(&nx)
	if nx != 1 {
		t.Fatal("type non vide touché par la suppression")
	}
}

// Nouveau store (niveau 1) + audit complet : chaque action laisse une
// trace (creer ×2, renommer, fusionner, supprimer) ; fusion et suppression
// exigent une confirmation côté page.
func TestAdminTypesCreeStoreEtAudit(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,live_top,live_sub) VALUES('a1','A1','t1','x')`); err != nil {
		t.Fatal(err)
	}
	post := func(v url.Values) string {
		v.Set("csrf", app.adminCSRF())
		req := httptest.NewRequest(http.MethodPost, "/admin/types", strings.NewReader(v.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		app.AdminTypes(rec, req)
		if rec.Code != http.StatusSeeOther {
			t.Fatalf("POST types : %d, attendu 303", rec.Code)
		}
		return rec.Header().Get("Location")
	}
	act := func(name string, kv ...string) url.Values {
		v := url.Values{}
		v.Set("action", name)
		for i := 0; i+1 < len(kv); i += 2 {
			v.Set(kv[i], kv[i+1])
		}
		return v
	}
	if loc := post(act("creer", "top", "t1", "sub", "z")); !strings.Contains(loc, "msg=") {
		t.Fatalf("creer : %q", loc)
	}
	if loc := post(act("creer_top", "top", "nouveau")); !strings.Contains(loc, "msg=") {
		t.Fatalf("creer_top : %q", loc)
	}
	if loc := post(act("renommer", "top", "t1", "ancien", "z", "nouveau", "w")); !strings.Contains(loc, "msg=") {
		t.Fatalf("renommer : %q", loc)
	}
	if loc := post(act("fusionner", "src", "t1|x", "dst", "t1|y")); !strings.Contains(loc, "msg=") {
		t.Fatalf("fusionner : %q", loc)
	}
	if loc := post(act("supprimer", "top", "t1", "sub", "w")); !strings.Contains(loc, "msg=") {
		t.Fatalf("supprimer : %q", loc)
	}
	var nAudit int
	_ = database.QueryRow(`SELECT COUNT(*) FROM admin_audit WHERE action LIKE 'types.%'`).Scan(&nAudit)
	if nAudit != 5 {
		t.Fatalf("audit : %d traces types.*, attendu 5", nAudit)
	}
	var nCard int
	_ = database.QueryRow(`SELECT COUNT(*) FROM declared_types WHERE top='nouveau'`).Scan(&nCard)
	if nCard != 1 {
		t.Fatal("store déclaré absent de declared_types")
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/types", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminTypes(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, `data-top="nouveau"`) {
		t.Fatal("carte du nouveau store absente du rendu")
	}
	for _, want := range []string{"Fusionner", "Supprimer", "data-mx-confirm"} {
		if c := strings.Count(body, want); c < 1 {
			t.Fatalf("confirmation %q absente (%d)", want, c)
		}
	}
}
