package handlers

import (
	"database/sql"
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
	"github.com/go-chi/chi/v5"
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
		`/admin/studio/js/view.js?v=2`,
		`/admin/studio/js/preview.js?v=1`,
		`/admin/studio/js/controller.js?v=2`,
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

// Phase 2 — API design JSON : publier un design, le recharger, le versionner,
// revenir en arrière. Le design de test porte un libellé distinctif
// (« Boutique Phase 2 ») pour vérifier l'aller-retour complet.
const designPhase2V1 = `{"front":{"accent":"#2E5136","font":"Inter","showAnnonce":true,"customCSS":""},"rows":[],"shop":{},"fiche":{},"panier":{},"Tx":{"bandeau":"Boutique Phase 2"},"txtPages":{},"menu":{},"seo":{},"reglages":{},"header":{},"footer":{},"acc":{}}`
const designPhase2V2 = `{"front":{"accent":"#84CC16","font":"Inter","showAnnonce":true,"customCSS":""},"rows":[],"shop":{},"fiche":{},"panier":{},"Tx":{"bandeau":"Boutique Phase 2 v2"},"txtPages":{},"menu":{},"seo":{},"reglages":{},"header":{},"footer":{},"acc":{}}`

func testAppDesign(t *testing.T) (*App, *sql.DB) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{AdminUser: "admin", AdminPassHash: "$2a$12$gTuewnqj.MYtMVtxRHYGteVd2uABtwzcjufgmWaqFpC51vtusLA3K", SessionSecret: "test-secret"}}
	// Sections de référence : 9, comme en prod.
	app.themeToSettings(catalog.Theme{Couleur: "gazonnier", Mode: "auto", Police: "inter", Sections: append([]string{}, catalog.DefaultSections...)})
	return app, database
}

func TestAdminDesignAPIPublieEtCharge(t *testing.T) {
	app, database := testAppDesign(t)
	// Sans design publié : GET renvoie design:null.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/design", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin/api/design (vide) : status %d, attendu 200", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), `"design":null`) {
		t.Fatalf("GET vide : attendu design:null, got %s", rec.Body.String())
	}
	// Publier v1 (CSRF en en-tête X-CSRF, corps JSON brut).
	req = httptest.NewRequest(http.MethodPost, "/admin/api/design", strings.NewReader(designPhase2V1))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF", app.adminCSRF())
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /admin/api/design : status %d, attendu 200 (corps : %s)", rec.Code, rec.Body.String())
	}
	var pub map[string]any
	if err := json.NewDecoder(rec.Body).Decode(&pub); err != nil {
		t.Fatal(err)
	}
	if pub["version"] != float64(1) || pub["theme_version"] != float64(2) {
		t.Fatalf("publication : version=%v theme_version=%v (attendu 1 et 2)", pub["version"], pub["theme_version"])
	}
	// Le thème extrait est publié : couleur gazonnier, bandeau du design.
	var couleur, bandeau, secs string
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='theme_couleur'`).Scan(&couleur); err != nil {
		t.Fatal(err)
	}
	if couleur != "gazonnier" {
		t.Fatalf("theme_couleur après publication : %q, attendu gazonnier", couleur)
	}
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='theme_annonce_texte'`).Scan(&bandeau); err != nil {
		t.Fatal(err)
	}
	if bandeau != "Boutique Phase 2" {
		t.Fatalf("theme_annonce_texte après publication : %q", bandeau)
	}
	// Les 9 sections du serveur sont préservées (pas écrasées par le design).
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='theme_sections'`).Scan(&secs); err != nil {
		t.Fatal(err)
	}
	if secs != strings.Join(catalog.DefaultSections, ",") {
		t.Fatalf("sections après publication : %q", secs)
	}
	// GET renvoie le design publié, avec son libellé.
	req = httptest.NewRequest(http.MethodGet, "/admin/api/design", nil)
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "Boutique Phase 2") {
		t.Fatalf("GET après publication : status %d, corps sans le libellé", rec.Code)
	}
}

func TestAdminDesignAPIRefuse(t *testing.T) {
	app, _ := testAppDesign(t)
	// Sans auth : 401.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/design", nil)
	rec := httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("GET sans auth : status %d, attendu 401", rec.Code)
	}
	// Sans CSRF : 403.
	req = httptest.NewRequest(http.MethodPost, "/admin/api/design", strings.NewReader(designPhase2V1))
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("POST sans CSRF : status %d, attendu 403", rec.Code)
	}
	// Design avec <script> : 400.
	req = httptest.NewRequest(http.MethodPost, "/admin/api/design", strings.NewReader(`{"front":{},"Tx":{"bandeau":"<script>x</script>"}}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF", app.adminCSRF())
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST avec script : status %d, attendu 400", rec.Code)
	}
	// Clé inconnue : 400.
	req = httptest.NewRequest(http.MethodPost, "/admin/api/design", strings.NewReader(`{"front":{},"products":[]}`))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF", app.adminCSRF())
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("POST avec clé inconnue : status %d, attendu 400", rec.Code)
	}
}

func TestAdminDesignVersionsEtActiver(t *testing.T) {
	app, database := testAppDesign(t)
	publie := func(design string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodPost, "/admin/api/design", strings.NewReader(design))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF", app.adminCSRF())
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		app.AdminDesignAPI(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("POST /admin/api/design : status %d", rec.Code)
		}
	}
	publie(designPhase2V1)
	publie(designPhase2V2)
	// Liste : 2 versions, v2 active.
	req := httptest.NewRequest(http.MethodGet, "/admin/api/design/versions", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignVersions(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET versions : status %d", rec.Code)
	}
	var list struct {
		Versions []struct {
			Version int  `json:"version"`
			Actif   bool `json:"actif"`
		} `json:"versions"`
	}
	if err := json.NewDecoder(rec.Body).Decode(&list); err != nil {
		t.Fatal(err)
	}
	if len(list.Versions) != 2 || list.Versions[0].Version != 2 || !list.Versions[0].Actif || list.Versions[1].Version != 1 || list.Versions[1].Actif {
		t.Fatalf("versions inattendues : %+v", list.Versions)
	}
	// Réactiver v1 (via le routeur chi, pour le paramètre {n}).
	r := chi.NewRouter()
	r.Post("/admin/api/design/versions/{n}/activer", app.AdminDesignActivate)
	req = httptest.NewRequest(http.MethodPost, "/admin/api/design/versions/1/activer", nil)
	req.Header.Set("X-CSRF", app.adminCSRF())
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST activer v1 : status %d, corps %s", rec.Code, rec.Body.String())
	}
	// Le design actif est redevenu v1 (libellé v1, accent gazonnier).
	req = httptest.NewRequest(http.MethodGet, "/admin/api/design", nil)
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if !strings.Contains(rec.Body.String(), "Boutique Phase 2") || strings.Contains(rec.Body.String(), "v2") {
		t.Fatalf("GET après activation v1 : corps inattendu %s", rec.Body.String())
	}
	var couleur string
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='theme_couleur'`).Scan(&couleur); err != nil {
		t.Fatal(err)
	}
	if couleur != "gazonnier" {
		t.Fatalf("theme_couleur après activation v1 : %q, attendu gazonnier", couleur)
	}
	// Version inexistante : 400.
	req = httptest.NewRequest(http.MethodPost, "/admin/api/design/versions/99/activer", nil)
	req.Header.Set("X-CSRF", app.adminCSRF())
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("activer v99 : status %d, attendu 400", rec.Code)
	}
}

func TestAdminDesignServeStudioInjecteDesign(t *testing.T) {
	app, _ := testAppDesign(t)
	// Publier un design d'abord.
	req := httptest.NewRequest(http.MethodPost, "/admin/api/design", strings.NewReader(designPhase2V1))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-CSRF", app.adminCSRF())
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignAPI(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("POST /admin/api/design : status %d", rec.Code)
	}
	// GET /admin/design : le design actif est injecté dans la page.
	wd, _ := os.Getwd()
	if err := os.Chdir("../.."); err != nil {
		t.Fatal(err)
	}
	defer func() { _ = os.Chdir(wd) }()
	req = httptest.NewRequest(http.MethodGet, "/admin/design", nil)
	req.SetBasicAuth("admin", "test1234")
	rec = httptest.NewRecorder()
	app.AdminDesign(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin/design : status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `window.__STUDIO_DESIGN__`) {
		t.Fatal("design non injecté dans la page Studio")
	}
	if !strings.Contains(body, "Boutique Phase 2") {
		t.Fatal("libellé du design absent de la page injectée")
	}
	if !strings.Contains(body, `window.__STUDIO_DESIGN_VERSION__=1`) {
		t.Fatal("version du design non injectée (attendue 1)")
	}
}
