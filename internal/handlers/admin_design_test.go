package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
)

// P1 — le serveur gagne : chaque Publier incrémente theme_version (1, 2, ...),
// ThemeFromDB le relit, et la réinitialisation l'efface (retour au défaut).
func TestThemeServeurGagneApresPublish(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database}
	publie := func(couleur string) catalog.Theme {
		th := catalog.Theme{Couleur: couleur, Mode: "auto", Police: "inter", Sections: append([]string{}, catalog.DefaultSections...)}
		app.themeToSettings(th)
		return catalog.SiteTheme
	}
	// Premier Publier : version 1.
	if got := publie("gazonnier"); got.Version != 1 {
		t.Fatalf("1er Publier : version attendue 1, got %d", got.Version)
	}
	// Deuxième Publier : version 2.
	if got := publie("liquid"); got.Version != 2 {
		t.Fatalf("2e Publier : version attendue 2, got %d", got.Version)
	}
	// ThemeFromDB relit la version et le thème.
	relu := site.ThemeFromDB(database)
	if relu.Version != 2 || relu.Couleur != "liquid" {
		t.Fatalf("ThemeFromDB : version 2 + liquid attendus, got version=%d couleur=%q", relu.Version, relu.Couleur)
	}
	// HtmlAttrs émet la version (thème non défaut).
	if got := relu.HtmlAttrs(); !strings.Contains(got, `data-theme-version="2"`) {
		t.Fatalf("HtmlAttrs doit émettre data-theme-version=\"2\", got %q", got)
	}
	// Réinitialisation : version effacée, thème par défaut, pas d'attribut.
	if _, err := database.Exec(`DELETE FROM settings WHERE nom LIKE 'theme_%'`); err != nil {
		t.Fatal(err)
	}
	catalog.SiteTheme = catalog.DefaultTheme()
	relu2 := site.ThemeFromDB(database)
	if relu2.Version != 0 || !relu2.IsDefault() || relu2.HtmlAttrs() != "" {
		t.Fatalf("après reset : version 0 + défaut + HtmlAttrs vide attendus, got version=%d défaut=%v attrs=%q", relu2.Version, relu2.IsDefault(), relu2.HtmlAttrs())
	}
}

// Phase 1 — intégration du Studio v16-2 : GET /admin/design sert la page
// découpée (studio.html) avec le thème courant + le CSRF injectés ; POST
// avec Accept: application/json répond en JSON (fetch du Studio) ; les
// fichiers découpés sont servis sous /admin/studio/ (auth admin requise).
func TestAdminDesignServeStudio(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	// Thème publié au préalable : le Studio doit démarrer avec lui.
	app := &App{DB: database, Cfg: config.Config{AdminUser: "admin", AdminPassHash: "$2a$12$gTuewnqj.MYtMVtxRHYGteVd2uABtwzcjufgmWaqFpC51vtusLA3K", SessionSecret: "test-secret"}}
	app.themeToSettings(catalog.Theme{Couleur: "liquid", Mode: "auto", Police: "inter", Sections: append([]string{}, catalog.DefaultSections...)})

	// Le handler lit web/admin/studio/studio.html depuis la racine du dépôt.
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	req := httptest.NewRequest(http.MethodGet, "/admin/design", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesign(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin/design : status %d, attendu 200", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{
		`window.__STUDIO_THEME__`,
		`"couleur":"liquid"`,     // thème courant injecté
		`window.__STUDIO_CSRF__`, // jeton CSRF injecté
		`/admin/studio/studio.css?v=1`,
		`/admin/studio/js/model.js?v=1`,
		`/admin/studio/js/view.js?v=1`,
		`/admin/studio/js/preview.js?v=1`,
		`/admin/studio/js/controller.js?v=1`,
		`id="tab-front"`,
		`id="tab-boutique"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET /admin/design : %q absent du HTML servi", want)
		}
	}
	// Sans auth : 401.
	req2 := httptest.NewRequest(http.MethodGet, "/admin/design", nil)
	rec2 := httptest.NewRecorder()
	app.AdminDesign(rec2, req2)
	if rec2.Code != http.StatusUnauthorized {
		t.Fatalf("GET /admin/design sans auth : status %d, attendu 401", rec2.Code)
	}
}

func TestAdminDesignPublieJSON(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{AdminUser: "admin", AdminPassHash: "$2a$12$gTuewnqj.MYtMVtxRHYGteVd2uABtwzcjufgmWaqFpC51vtusLA3K", SessionSecret: "test-secret"}}

	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("couleur", "gazonnier")
	form.Set("accent", "")
	form.Set("mode", "auto")
	form.Set("police", "inter")
	form.Set("annonce_on", "1")
	form.Set("annonce_texte", "Livraison offerte")
	for _, s := range catalog.DefaultSections {
		form.Add("sections", s)
	}
	req := httptest.NewRequest(http.MethodPost, "/admin/design", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesign(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /admin/design JSON : status %d, attendu 200", rec.Code)
	}
	var got struct {
		OK      bool `json:"ok"`
		Version int  `json:"version"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&got); err != nil {
		t.Fatalf("POST /admin/design JSON : réponse illisible : %v", err)
	}
	if !got.OK || got.Version != 1 {
		t.Fatalf("POST /admin/design JSON : ok=%v version=%d, attendus ok=true version=1", got.OK, got.Version)
	}
	// Le thème est bien persisté (la boutique le relira).
	var couleur string
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='theme_couleur'`).Scan(&couleur); err != nil || couleur != "gazonnier" {
		t.Fatalf("thème non persisté : couleur=%q err=%v", couleur, err)
	}
}

func TestAdminStudioFile(t *testing.T) {
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()

	app := &App{Cfg: config.Config{AdminUser: "admin", AdminPassHash: "$2a$12$gTuewnqj.MYtMVtxRHYGteVd2uABtwzcjufgmWaqFpC51vtusLA3K", SessionSecret: "test-secret"}}

	req := httptest.NewRequest(http.MethodGet, "/admin/studio/js/model.js", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminStudioFile(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin/studio/js/model.js : status %d, attendu 200", rec.Code)
	}
	if ct := rec.Header().Get("Content-Type"); !strings.Contains(ct, "javascript") {
		t.Fatalf("model.js : Content-Type %q, attendu javascript", ct)
	}
	if !strings.Contains(rec.Body.String(), "DEFAULT_FRONT") {
		t.Fatal("model.js : contenu attendu absent")
	}
	// Traversée de dossier bloquée.
	req2 := httptest.NewRequest(http.MethodGet, "/admin/studio/../../internal/handlers/admin_design.go", nil)
	req2.SetBasicAuth("admin", "test1234")
	rec2 := httptest.NewRecorder()
	app.AdminStudioFile(rec2, req2)
	if rec2.Code != http.StatusNotFound {
		t.Fatalf("traversée ../ : status %d, attendu 404", rec2.Code)
	}
}

// Phase 1 — la publication du Studio ne doit JAMAIS réduire la liste des
// sections d'accueil : le Studio renvoie les sections du serveur telles
// quelles (les lignes de l'éditeur visuel n'en sont que 3, le site en a 9).
func TestAdminDesignPubliePreserveSections(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{AdminUser: "admin", AdminPassHash: "$2a$12$gTuewnqj.MYtMVtxRHYGteVd2uABtwzcjufgmWaqFpC51vtusLA3K", SessionSecret: "test-secret"}}
	// Sections personnalisées (7 sur 9) publiées au préalable.
	custom := []string{"hero", "minis", "produits", "vendus", "banniere", "recents", "marques"}
	app.themeToSettings(catalog.Theme{Couleur: "liquid", Mode: "auto", Police: "inter", Sections: custom})

	// Le Studio renvoie les sections du serveur (comme le fait themeFromFront).
	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("couleur", "gazonnier")
	form.Set("accent", "")
	form.Set("mode", "auto")
	form.Set("police", "inter")
	for _, s := range custom {
		form.Add("sections", s)
	}
	form.Set("ordre", strings.Join(custom, ","))
	req := httptest.NewRequest(http.MethodPost, "/admin/design", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Accept", "application/json")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesign(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /admin/design : status %d, attendu 200", rec.Code)
	}
	var secs string
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='theme_sections'`).Scan(&secs); err != nil {
		t.Fatal(err)
	}
	if secs != strings.Join(custom, ",") {
		t.Fatalf("sections après Publier : %q, attendues %q", secs, strings.Join(custom, ","))
	}
}
