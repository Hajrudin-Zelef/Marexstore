package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"

	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
)

// C2 — l'endpoint /admin/api/fichiers expose les VRAIS fichiers du site
// (pas les extraits inventés du prototype) : chrome + pages + app.js /
// style.css / theme.js. Le faux theme.css n'existe pas.
func TestAdminFichiers(t *testing.T) {
	app, _ := testAppDesign(t)
	// Pages statiques : charger les vraies pages du dépôt dans app.Bodies
	// (comme reload() le fait au démarrage du serveur).
	if top, drawer, foot, bodies, _, err := site.LoadStaticDB(app.DB, "../../templates/static"); err == nil {
		app.Top, app.Drawer, app.Foot, app.Bodies = top, drawer, foot, bodies
	}

	get := func(auth bool) (int, map[string]any) {
		req := httptest.NewRequest(http.MethodGet, "/admin/api/fichiers", nil)
		if auth {
			req.SetBasicAuth("admin", "test1234")
		}
		rec := httptest.NewRecorder()
		app.AdminFichiers(rec, req)
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return rec.Code, body
	}

	if code, _ := get(false); code != http.StatusUnauthorized {
		t.Fatalf("sans auth : status %d, attendu 401", code)
	}
	code, body := get(true)
	if code != http.StatusOK {
		t.Fatalf("avec auth : status %d, attendu 200", code)
	}
	list, _ := body["fichiers"].([]any)
	if len(list) < 20 {
		t.Fatalf("fichiers : %d, attendu au moins 20 (21 templates + 3 static)", len(list))
	}
	byNom := map[string]map[string]any{}
	for _, f := range list {
		m := f.(map[string]any)
		byNom[m["nom"].(string)] = m
	}

	// Le faux theme.css n'existe pas ; le vrai thème est theme.js + style.css.
	if _, ok := byNom["static/theme.css"]; ok {
		t.Fatal("theme.css ne devrait pas exister (c'est le prototype)")
	}
	theme, ok := byNom["static/theme.js"]
	if !ok {
		t.Fatal("static/theme.js absent")
	}
	if theme["lignes"].(float64) < 10 {
		t.Fatalf("theme.js : %v lignes, attendu le vrai fichier", theme["lignes"])
	}

	// app.js : le vrai fichier (1 202 lignes), pas l'extrait de 5 lignes.
	appjs, ok := byNom["static/app.js"]
	if !ok {
		t.Fatal("static/app.js absent")
	}
	if appjs["lignes"].(float64) < 1000 {
		t.Fatalf("app.js : %v lignes, attendu ~1202", appjs["lignes"])
	}
	if !strings.Contains(appjs["contenu"].(string), "localStorage") && !strings.Contains(appjs["contenu"].(string), "theme") {
		t.Fatal("app.js : contenu ne ressemble pas au vrai fichier")
	}

	// _top.html : le vrai (22 lignes, vraie topbar), pas l'extrait de 7 lignes.
	top, ok := byNom["templates/static/_top.html"]
	if !ok {
		t.Fatal("templates/static/_top.html absent")
	}
	if top["lignes"].(float64) < 15 {
		t.Fatalf("_top.html : %v lignes, attendu ~22", top["lignes"])
	}
	if !strings.Contains(top["contenu"].(string), "Bienvenue") {
		t.Fatal("_top.html : la vraie topbar (« Bienvenue ») absente")
	}

	// apropos.html : le vrai (62 lignes), pas le résumé de 3 lignes.
	apropos, ok := byNom["templates/static/apropos.html"]
	if !ok {
		t.Fatal("templates/static/apropos.html absent")
	}
	if apropos["lignes"].(float64) < 40 {
		t.Fatalf("apropos.html : %v lignes, attendu ~62", apropos["lignes"])
	}
}

// C4 — POST /admin/api/fichiers : enregistre une nouvelle version, visible
// en GET, avec historique + rollback. Un gabarit prend effet via reload().
// Le test se place à la racine du dépôt (reload() lit "templates/static").
func TestAdminFichiersEnregistrer(t *testing.T) {
	app, _ := testAppDesign(t)
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	if top, drawer, foot, bodies, _, err := site.LoadStaticDB(app.DB, "templates/static"); err == nil {
		app.Top, app.Drawer, app.Foot, app.Bodies = top, drawer, foot, bodies
	}
	post := func(nom, contenu string, csrf bool) (int, map[string]any) {
		payload := `{"nom":` + strconv.Quote(nom) + `,"contenu":` + strconv.Quote(contenu) + `}`
		req := httptest.NewRequest(http.MethodPost, "/admin/api/fichiers", strings.NewReader(payload))
		req.SetBasicAuth("admin", "test1234")
		if csrf {
			req.Header.Set("X-CSRF", app.adminCSRF())
		}
		rec := httptest.NewRecorder()
		app.AdminFichiers(rec, req)
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return rec.Code, body
	}
	if code, _ := post("templates/static/apropos.html", "x", false); code != http.StatusForbidden {
		t.Fatalf("sans CSRF : %d, attendu 403", code)
	}
	// Gabarit : v1 puis v2, GET reflète v2, reload a mis à jour app.Bodies.
	if code, body := post("templates/static/apropos.html", "<h1>Test C4 v1</h1>", true); code != http.StatusOK || body["version"].(float64) != 1 {
		t.Fatalf("POST apropos v1 : code=%d body=%v", code, body)
	}
	if code, body := post("templates/static/apropos.html", "<h1>Test C4 v2</h1>", true); code != http.StatusOK || body["version"].(float64) != 2 {
		t.Fatalf("POST apropos v2 : code=%d body=%v", code, body)
	}
	if !strings.Contains(app.Bodies["apropos"], "Test C4 v2") {
		t.Fatal("reload() n'a pas appliqué le gabarit enregistré (app.Bodies)")
	}
	// Fichier static : enregistré en base, GET le reflète.
	if code, body := post("static/app.js", "/* Test C4 */\nconsole.log(1);", true); code != http.StatusOK || body["version"].(float64) != 1 {
		t.Fatalf("POST app.js : code=%d body=%v", code, body)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/api/fichiers", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminFichiers(rec, req)
	var list map[string]any
	_ = json.Unmarshal(rec.Body.Bytes(), &list)
	for _, f := range list["fichiers"].([]any) {
		m := f.(map[string]any)
		if m["nom"] == "static/app.js" && !strings.Contains(m["contenu"].(string), "Test C4") {
			t.Fatal("GET /admin/api/fichiers ne reflète pas app.js enregistré")
		}
	}
	// Versions + rollback (via un vrai routeur chi pour le paramètre {n}).
	r := chi.NewRouter()
	r.Get("/admin/api/fichiers/versions", app.AdminFichiersVersions)
	r.Post("/admin/api/fichiers/versions/{n}/activer", app.AdminFichiersActiver)
	getV := func(nom string) []any {
		req := httptest.NewRequest(http.MethodGet, "/admin/api/fichiers/versions?nom="+url.QueryEscape(nom), nil)
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("versions %s : code %d", nom, rec.Code)
		}
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return body["versions"].([]any)
	}
	vs := getV("templates/static/apropos.html")
	if len(vs) != 2 {
		t.Fatalf("versions apropos : %d, attendu 2", len(vs))
	}
	act := func(nom string, n int) int {
		req := httptest.NewRequest(http.MethodPost, "/admin/api/fichiers/versions/"+strconv.Itoa(n)+"/activer?nom="+url.QueryEscape(nom), nil)
		req.SetBasicAuth("admin", "test1234")
		req.Header.Set("X-CSRF", app.adminCSRF())
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		return rec.Code
	}
	if code := act("templates/static/apropos.html", 1); code != http.StatusOK {
		t.Fatalf("activer v1 : code %d", code)
	}
	if !strings.Contains(app.Bodies["apropos"], "Test C4 v1") {
		t.Fatal("rollback : app.Bodies n'est pas revenu à v1")
	}
	if code := act("templates/static/apropos.html", 99); code != http.StatusBadRequest {
		t.Fatalf("activer v99 : code %d, attendu 400", code)
	}
}

// C4 — refus : 401 sans auth, 403 sans CSRF, 400 nom hors liste / JSON
// invalide, 405 méthode non autorisée.
func TestAdminFichiersRefus(t *testing.T) {
	app, _ := testAppDesign(t)
	post := func(method, payload string, auth, csrf bool) int {
		req := httptest.NewRequest(method, "/admin/api/fichiers", strings.NewReader(payload))
		if auth {
			req.SetBasicAuth("admin", "test1234")
		}
		if csrf {
			req.Header.Set("X-CSRF", app.adminCSRF())
		}
		rec := httptest.NewRecorder()
		app.AdminFichiers(rec, req)
		return rec.Code
	}
	if code := post(http.MethodPost, `{"nom":"templates/static/apropos.html","contenu":"x"}`, false, true); code != http.StatusUnauthorized {
		t.Fatalf("POST sans auth : %d, attendu 401", code)
	}
	if code := post(http.MethodPost, `{"nom":"templates/static/apropos.html","contenu":"x"}`, true, false); code != http.StatusForbidden {
		t.Fatalf("POST sans CSRF : %d, attendu 403", code)
	}
	if code := post(http.MethodPost, `{"nom":"../main.go","contenu":"x"}`, true, true); code != http.StatusBadRequest {
		t.Fatalf("POST ../main.go : %d, attendu 400", code)
	}
	if code := post(http.MethodPost, `{"nom":"static/theme.css","contenu":"x"}`, true, true); code != http.StatusBadRequest {
		t.Fatalf("POST theme.css : %d, attendu 400", code)
	}
	if code := post(http.MethodPost, `{json invalide`, true, true); code != http.StatusBadRequest {
		t.Fatalf("POST JSON invalide : %d, attendu 400", code)
	}
	if code := post(http.MethodPut, `{"nom":"static/app.js","contenu":"x"}`, true, true); code != http.StatusMethodNotAllowed {
		t.Fatalf("PUT : %d, attendu 405", code)
	}
}

// C4 — StaticFile : version base servie avec no-cache, disque avec cache
// long. Le test se place à la racine du dépôt (FileServer relatif).
func TestStaticFile(t *testing.T) {
	app, _ := testAppDesign(t)
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	get := func(path string) (int, string, string) {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		app.StaticFile().ServeHTTP(rec, req)
		return rec.Code, rec.Body.String(), rec.Header().Get("Cache-Control")
	}
	// Disque : theme.js (aucune version base) → cache long, vrai contenu.
	if code, body, cc := get("/theme.js"); code != http.StatusOK || !strings.Contains(cc, "max-age=31536000") || !strings.Contains(body, "theme") {
		t.Fatalf("theme.js disque : code=%d cc=%q (contenu %d car.)", code, cc, len(body))
	}
	// Base : enregistrer app.js puis le demander → contenu base, no-cache.
	payload := `{"nom":"static/app.js","contenu":"/* C4 test */\nvar x=1;"}`
	req := httptest.NewRequest(http.MethodPost, "/admin/api/fichiers", strings.NewReader(payload))
	req.SetBasicAuth("admin", "test1234")
	req.Header.Set("X-CSRF", app.adminCSRF())
	rec := httptest.NewRecorder()
	app.AdminFichiers(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST app.js : code %d", rec.Code)
	}
	if code, body, cc := get("/app.js"); code != http.StatusOK || body != "/* C4 test */\nvar x=1;" || cc != "no-cache" {
		t.Fatalf("app.js base : code=%d cc=%q body=%q", code, cc, body)
	}
}
