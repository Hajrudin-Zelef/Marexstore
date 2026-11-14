package handlers

import (
	"fmt"
	"net/http"
	"regexp"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
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

// AdminDesign : écran du studio (formulaire + iframe d'aperçu live).
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
		http.Redirect(w, r, "/admin/design?msg=Design+publi%C3%A9", http.StatusSeeOther)
		return
	}
	t := catalog.SiteTheme
	var secs strings.Builder
	on := map[string]bool{}
	for _, s := range t.Sections {
		on[s] = true
	}
	for _, d := range designSections {
		chk := ""
		if on[d.Key] {
			chk = " checked"
		}
		fmt.Fprintf(&secs, `<label class="mx-check"><input type="checkbox" name="sections" value="%s"%s> %s</label>`, d.Key, chk, d.Label)
	}
	a.write(w, 200, a.adminPage("Design du site", "design",
		`<form class="mx-form" id="mx-design-form" method="post" action="/admin/design"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`">
<div class="mx-cols"><div>
<div class="mx-postbox"><h2>Couleurs &amp; ambiance</h2><div class="in">
<label>Palette
<select name="couleur">
<option value="auto"`+designSelected("auto", t.Couleur)+`>Auto (défaut boutique)</option>
<option value="gazonnier"`+designSelected("gazonnier", t.Couleur)+`>Gazonnier (vert Mediatec)</option>
<option value="liquid"`+designSelected("liquid", t.Couleur)+`>Liquid (lime)</option>
<option value="perso"`+designSelected("perso", t.Couleur)+`>Personnalisée (code ci-dessous)</option>
</select></label>
<label>Couleur personnalisée (hexa, ex. 8a2be2) <input name="accent" value="`+render.Esc(strings.TrimPrefix(t.Accent, "#"))+`" placeholder="8a2be2"></label>
<label>Mode <select name="mode">
<option value="auto"`+designSelected("auto", t.Mode)+`>Auto (défaut)</option>
<option value="clair"`+designSelected("clair", t.Mode)+`>Clair</option>
<option value="sombre"`+designSelected("sombre", t.Mode)+`>Sombre</option>
</select></label>
<label>Police <select name="police">
<option value="inter"`+designSelected("inter", t.Police)+`>Inter (défaut)</option>
<option value="systeme"`+designSelected("systeme", t.Police)+`>Système</option>
</select></label>
</div></div>
<div class="mx-postbox"><h2>Page d'accueil</h2><div class="in">
<label class="mx-check"><input type="checkbox" name="annonce_on" value="1"`+func() string {
			if t.AnnonceOn {
				return " checked"
			}
			return ""
		}()+`> Bande annonce en haut du site</label>
<label>Texte du bandeau <input name="annonce_texte" value="`+render.Esc(t.AnnonceTexte)+`" placeholder="Livraison offerte dès 100 000 F CFA"></label>
<div class="mx-hint">Sections affichées (décocher pour masquer) :</div>
`+secs.String()+`
<label>Ordre des sections (séparées par des virgules) <input name="ordre" value="`+render.Esc(strings.Join(t.Sections, ","))+`"></label>
<div class="mx-hint">Ordre possible : hero, minis, filtre, onglets, produits, vendus, banniere, recents, marques.</div>
</div></div>
<div class="mx-postbox"><h2>CSS personnalisé</h2><div class="in">
<label>CSS (sans balise &lt;script&gt;, appliqué après le thème) <textarea name="css" rows="4" placeholder=".mx-annonce{font-size:15px}">`+render.Esc(t.CSS)+`</textarea></label>
<div class="mx-form-actions">
<button class="mx-btn" formaction="/admin/design/apercu" formtarget="mx-preview">Actualiser l'aperçu</button>
<button class="mx-btn primary">Publier</button>
<button class="mx-btn danger" formaction="/admin/design/reset" formmethod="post" data-mx-confirm="Réinitialiser tout le design ?">Réinitialiser</button>
</div></div></div>
</form></div>
<div><div class="mx-postbox mx-pub"><h2>Aperçu live</h2><div class="in">
<iframe name="mx-preview" src="/admin/design/apercu" title="Aperçu du site" style="width:100%;height:640px;border:1px solid var(--mx-border);border-radius:8px;background:#fff"></iframe>
<p class="mx-hint">Le brouillon ne s'affiche que ici. « Publier » le met en ligne.</p>
</div></div></div></div>`, r))
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
