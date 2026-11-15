package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
