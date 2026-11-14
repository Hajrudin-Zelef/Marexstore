package handlers

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/design"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/go-chi/chi/v5"
)

// Studio Design (/admin/design) : personnalise le visuel du site comme le
// Customizer WordPress — aperçu live (brouillon, jamais public) puis Publier.

var designSections = []struct{ Key, Label string }{
	{"hero", "Bandeau défilant (hero)"},
	{"minis", "Mini-casiers promo"},
	{"filtre", "Bloc « Trouver un produit »"},
	{"onglets", "Onglets En vedette / Promo / Mieux noté"},
	{"produits", "Grille « Nos produits »"},
	{"vendus", "Rangée « Les plus vendus »"},
	{"banniere", "Bannière stockage"},
	{"recents", "Rangée « Ajouté récemment »"},
	{"marques", "Bandeau des marques"},
}

// reRelatif repère les attributs d'URL ; absolutizePreview préfixe les URLs
// relatives par / (la page d'aperçu est servie sous /admin/design/apercu
// mais ses liens visent la racine). Absolues (http, /, #, mailto:, data:,
// javascript:) intactes. (RE2 n'a pas de lookahead : on filtre en Go.)
var reRelatif = regexp.MustCompile(`((?:href|src|action)=")([^"]*)"`)
var reRelUrl = regexp.MustCompile(`url\(([^)]*)\)`)

// absolutizeURL rend une URL CSS absolue (racine), sauf absolues/données.
func absolutizeURL(raw string) string {
	u := strings.TrimSpace(raw)
	q := ""
	if strings.HasPrefix(u, `"`) || strings.HasPrefix(u, `'`) {
		q = u[:1]
		u = u[1:]
		if strings.HasSuffix(u, q) {
			u = u[:len(u)-1]
		}
	}
	if u == "" || strings.HasPrefix(u, "http://") || strings.HasPrefix(u, "https://") ||
		strings.HasPrefix(u, "/") || strings.HasPrefix(u, "data:") {
		return raw
	}
	return `"/` + u + `"`
}

func absolutizePreview(html string) string {
	html = reRelatif.ReplaceAllStringFunc(html, func(m string) string {
		i := strings.Index(m, `"`)
		head, val := m[:i+1], m[i+1:len(m)-1]
		if val == "" || strings.HasPrefix(val, "http://") || strings.HasPrefix(val, "https://") ||
			strings.HasPrefix(val, "/") || strings.HasPrefix(val, "#") ||
			strings.HasPrefix(val, "mailto:") || strings.HasPrefix(val, "data:") ||
			strings.HasPrefix(val, "javascript:") {
			return m
		}
		return head + "/" + val + `"`
	})
	return reRelUrl.ReplaceAllStringFunc(html, func(m string) string {
		inner := m[len("url(") : len(m)-1]
		return "url(" + absolutizeURL(inner) + ")"
	})
}

// hasThemeQuery : la requête porte-t-elle un brouillon de thème
// (paramètres de l'iframe d'aperçu en navigation GET) ?
func hasThemeQuery(r *http.Request) bool {
	q := r.URL.Query()
	for _, k := range []string{"couleur", "accent", "mode", "police", "annonce_on", "annonce_texte", "sections", "ordre", "css"} {
		if _, ok := q[k]; ok {
			return true
		}
	}
	return false
}

// themeFromForm : thème brouillon depuis le formulaire (aperçu + publication).
func themeFromForm(r *http.Request) catalog.Theme {
	t := catalog.DefaultTheme()
	couleur := r.FormValue("couleur")
	switch couleur {
	case "gazonnier", "liquid", "perso":
		t.Couleur = couleur
	default:
		t.Couleur = "auto"
	}
	t.Accent = strings.TrimSpace(r.FormValue("accent"))
	if t.Accent != "" && !strings.HasPrefix(t.Accent, "#") {
		t.Accent = "#" + t.Accent
	}
	switch r.FormValue("mode") {
	case "clair", "sombre":
		t.Mode = r.FormValue("mode")
	default:
		t.Mode = "auto"
	}
	if r.FormValue("police") == "systeme" {
		t.Police = "systeme"
	}
	t.AnnonceOn = r.FormValue("annonce_on") == "1"
	t.AnnonceTexte = strings.TrimSpace(r.FormValue("annonce_texte"))
	// Ordre souhaité (texte) filtré sur les sections cochées.
	ordre := strings.Split(strings.TrimSpace(r.FormValue("ordre")), ",")
	cochées := map[string]bool{}
	for _, s := range r.Form["sections"] {
		cochées[strings.TrimSpace(s)] = true
	}
	var secs []string
	for _, s := range ordre {
		if s = strings.TrimSpace(s); s != "" && cochées[s] {
			secs = append(secs, s)
		}
	}
	for _, d := range designSections {
		if cochées[d.Key] && !containsStr(secs, d.Key) {
			secs = append(secs, d.Key)
		}
	}
	if len(secs) > 0 {
		t.Sections = secs
	}
	t.CSS = strings.TrimSpace(r.FormValue("css"))
	return t
}

func containsStr(ss []string, v string) bool {
	for _, s := range ss {
		if s == v {
			return true
		}
	}
	return false
}

// themeToSettings : publie un thème (live).
// P1 — le serveur gagne : theme_version est incrémenté à chaque Publier
// (une seule requête atomique, RETURNING). Toute la publication est en
// transaction : soit le thème + sa version sont écrits, soit rien.
func (a *App) themeToSettings(t catalog.Theme) {
	tx, err := a.DB.Begin()
	if err != nil {
		return
	}
	defer tx.Rollback()
	set := func(k, v string) {
		_, _ = tx.Exec(`INSERT INTO settings(nom,valeur) VALUES(?,?) ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`, k, v)
	}
	var nv int
	if err := tx.QueryRow(`INSERT INTO settings(nom,valeur) VALUES('theme_version','1')
		ON CONFLICT(nom) DO UPDATE SET valeur=CAST(settings.valeur AS INTEGER)+1
		RETURNING CAST(valeur AS INTEGER)`).Scan(&nv); err != nil {
		return
	}
	t.Version = nv
	set("theme_couleur", t.Couleur)
	set("theme_accent", t.Accent)
	set("theme_mode", t.Mode)
	set("theme_police", t.Police)
	if t.AnnonceOn {
		set("theme_annonce_on", "1")
	} else {
		set("theme_annonce_on", "0")
	}
	set("theme_annonce_texte", t.AnnonceTexte)
	set("theme_sections", strings.Join(t.Sections, ","))
	set("theme_css", t.CSS)
	if err := tx.Commit(); err != nil {
		return
	}
	catalog.SiteTheme = t
}

func designSelected(v, cur string) string {
	if v == cur {
		return " selected"
	}
	return ""
}

// AdminDesign : écran du Studio (prototype v16-2 découpé en MVC, web/admin/studio).
// GET : sert studio.html avec le thème courant + le jeton CSRF injectés
// (marqueur <!--STUDIO_INIT-->). POST : publie le thème (comme avant) ;
// si la requête demande du JSON (fetch du Studio), répond en JSON au lieu
// de rediriger. Le reste du design sera persisté en phase 2 (API JSON).
func (a *App) AdminDesign(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodPost && (r.FormValue("action") == "reset" || strings.HasSuffix(r.URL.Path, "/reset")) {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		for _, k := range []string{"theme_couleur", "theme_accent", "theme_mode", "theme_police", "theme_annonce_on", "theme_annonce_texte", "theme_sections", "theme_css", "theme_version"} {
			_, _ = a.DB.Exec(`DELETE FROM settings WHERE nom=?`, k)
		}
		catalog.SiteTheme = catalog.DefaultTheme()
		a.reload()
		a.auditAdmin(r, "design.reinitialiser", "theme")
		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			a.writeJSON(w, 200, map[string]any{"ok": true, "version": 0, "msg": "Design réinitialisé"})
			return
		}
		http.Redirect(w, r, "/admin/design?msg=Design+r%C3%A9initialis%C3%A9", http.StatusSeeOther)
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		a.themeToSettings(themeFromForm(r))
		a.reload()
		a.auditAdmin(r, "design.publier", fmt.Sprintf("theme v%d", catalog.SiteTheme.Version))
		if strings.Contains(r.Header.Get("Accept"), "application/json") {
			a.writeJSON(w, 200, map[string]any{"ok": true, "version": catalog.SiteTheme.Version, "msg": "Design publié"})
			return
		}
		http.Redirect(w, r, "/admin/design?msg=Design+publi%C3%A9", http.StatusSeeOther)
		return
	}
	// GET : la page Studio (fichier découpé, pas un formulaire Go).
	html, err := os.ReadFile("web/admin/studio/studio.html")
	if err != nil {
		http.Error(w, "studio indisponible", http.StatusInternalServerError)
		return
	}
	init := `<script>window.__STUDIO_THEME__=` + studioThemeJSON() + `;window.__STUDIO_CSRF__=` + strconv.Quote(a.adminCSRF()) + `;` + studioDesignInit(a.DB) + `</script>`
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte(strings.Replace(string(html), "<!--STUDIO_INIT-->", init, 1)))
}

// studioThemeJSON : thème courant du site, au format attendu par le Studio
// (front.accent, front.font, showAnnonce, bandeau, customCSS, couleur).
func studioThemeJSON() string {
	t := catalog.SiteTheme
	accent := "#2E5136"
	switch t.Couleur {
	case "liquid":
		accent = "#84cc16"
	case "perso":
		if strings.HasPrefix(t.Accent, "#") {
			accent = t.Accent
		} else if t.Accent != "" {
			accent = "#" + t.Accent
		}
	}
	b, _ := json.Marshal(map[string]any{
		"couleur":      t.Couleur,
		"accent":       accent,
		"font":         t.Police,
		"showAnnonce":  t.AnnonceOn,
		"annonceTexte": t.AnnonceTexte,
		"css":          t.CSS,
		"version":      t.Version,
		// Sections d'accueil actuelles : le Studio ne les gère pas en phase 1,
		// il les renvoie telles quelles à la publication (ne jamais les
		// déduire des lignes de l'éditeur visuel, qui n'en a que 3).
		"sections": t.Sections,
	})
	return string(b)
}

// studioDesignInit : design actif injecté dans la page Studio
// (window.__STUDIO_DESIGN__ + __STUDIO_DESIGN_VERSION__). Le JSON est
// re-marshalé par Go (qui échappe < > & en \u003c…) : aucune valeur ne peut
// fermer le bloc <script>. Sans design publié : null + version 0 (le Studio
// démarre alors sur son défaut local + le thème injecté).
func studioDesignInit(db *sql.DB) string {
	raw, version, err := design.Load(db)
	if err != nil || raw == nil {
		return `window.__STUDIO_DESIGN__=null;window.__STUDIO_DESIGN_VERSION__=0;`
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return `window.__STUDIO_DESIGN__=null;window.__STUDIO_DESIGN_VERSION__=0;`
	}
	b, err := json.Marshal(v)
	if err != nil {
		return `window.__STUDIO_DESIGN__=null;window.__STUDIO_DESIGN_VERSION__=0;`
	}
	return `window.__STUDIO_DESIGN__=` + string(b) + `;window.__STUDIO_DESIGN_VERSION__=` + strconv.Itoa(version) + `;`
}

// AdminDesignAPI : le design complet du Studio en JSON (phase 2).
// GET : renvoie le design actif (ou design:null si rien n'est publié).
// POST : publie un nouveau design (CSRF en en-tête X-CSRF, corps JSON brut) :
// validation serveur, nouvelle version dans design_files, puis extraction de
// la partie thème (front) vers settings — la boutique continue d'utiliser le
// thème publié, sans changer son rendu.
func (a *App) AdminDesignAPI(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodGet {
		raw, version, err := design.Load(a.DB)
		if err != nil {
			http.Error(w, "lecture impossible", http.StatusInternalServerError)
			return
		}
		if raw == nil {
			a.writeJSON(w, 200, map[string]any{"ok": true, "version": 0, "design": nil})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`{"ok":true,"version":` + strconv.Itoa(version) + `,"design":` + string(raw) + `}`))
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, design.MaxDesignBytes+1))
		if err != nil {
			http.Error(w, "lecture impossible", http.StatusBadRequest)
			return
		}
		version, err := design.Save(a.DB, raw, a.Cfg.AdminUser)
		if err != nil {
			http.Error(w, "design refusé : "+err.Error(), http.StatusBadRequest)
			return
		}
		// Extraire le thème et le publier (theme_version +1, sections préservées).
		if th, err := design.ThemeFromDesign(raw, catalog.SiteTheme); err == nil {
			a.themeToSettings(th)
			a.reload()
		}
		a.auditAdmin(r, "design.publier", fmt.Sprintf("design v%d (theme v%d)", version, catalog.SiteTheme.Version))
		a.writeJSON(w, 200, map[string]any{"ok": true, "version": version, "theme_version": catalog.SiteTheme.Version, "msg": "Design publié"})
		return
	}
	http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
}

// AdminDesignVersions : liste les versions publiées (la plus récente d'abord).
func (a *App) AdminDesignVersions(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	versions, err := design.ListVersions(a.DB)
	if err != nil {
		http.Error(w, "lecture impossible", http.StatusInternalServerError)
		return
	}
	if versions == nil {
		versions = []design.Version{}
	}
	a.writeJSON(w, 200, map[string]any{"ok": true, "versions": versions})
}

// AdminDesignActivate : réactive une ancienne version (retour en arrière en
// 1 clic). Le thème extrait de cette version est republié (theme_version +1).
func (a *App) AdminDesignActivate(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	if !a.checkAdminCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	n, err := strconv.Atoi(chi.URLParam(r, "n"))
	if err != nil || n < 1 {
		http.Error(w, "version invalide", http.StatusBadRequest)
		return
	}
	if err := design.Activate(a.DB, n); err != nil {
		http.Error(w, "activation impossible : "+err.Error(), http.StatusBadRequest)
		return
	}
	// Republier le thème de la version réactivée (la boutique suit).
	if raw, _, err := design.Load(a.DB); err == nil && raw != nil {
		if th, err := design.ThemeFromDesign(raw, catalog.SiteTheme); err == nil {
			a.themeToSettings(th)
			a.reload()
		}
	}
	a.auditAdmin(r, "design.activer", fmt.Sprintf("design v%d (theme v%d)", n, catalog.SiteTheme.Version))
	a.writeJSON(w, 200, map[string]any{"ok": true, "version": n, "theme_version": catalog.SiteTheme.Version, "msg": "Version réactivée"})
}

// AdminStudioFile : fichiers du Studio (CSS/JS découpés, web/admin/studio).
// Protégé par adminAuth comme la page ; le chemin est nettoyé et borné au
// dossier du Studio (pas de remontée ../). Type MIME selon l'extension.
func (a *App) AdminStudioFile(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	rel := strings.TrimPrefix(r.URL.Path, "/admin/studio/")
	rel = path.Clean("/" + rel)[1:]
	if rel == "" || rel == "." || strings.Contains(rel, "..") {
		http.NotFound(w, r)
		return
	}
	full := filepath.Join("web", "admin", "studio", rel)
	data, err := os.ReadFile(full)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	switch filepath.Ext(rel) {
	case ".css":
		w.Header().Set("Content-Type", "text/css; charset=utf-8")
	case ".js":
		w.Header().Set("Content-Type", "text/javascript; charset=utf-8")
	case ".html":
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
	default:
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(data)
}

// AdminDesignPreview : rendu de l'accueil (brouillon POST ou thème live en GET).
func (a *App) AdminDesignPreview(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	th := catalog.SiteTheme
	// Ne SURTOUT pas appeler r.ParseForm() ici : il fige r.Form à vide
	// et casse ensuite r.FormValue sur les corps multipart.
	if r.Method == http.MethodPost || hasThemeQuery(r) {
		th = themeFromForm(r)
	}
	body := render.IndexBodyTheme(a.Store, a.Recents, th)
	// <base href="/"> : la page est servie sous /admin/design/apercu mais ses
	// URLs relatives (assets/, images/, produits/) visent la racine du site.
	extra := render.ThemeHead(th) + `<script src="/static/theme.js?v=2" defer></script><script src="/static/app.js?v=339" defer></script><script src="/static/htmx.min.js?v=1912" defer></script><script src="/static/mirror.js?v=1" defer></script>`
	s := render.DocTheme("Accueil — aperçu", body, "", extra, th.HtmlAttrs())
	a.write(w, 200, absolutizePreview(s))
}
