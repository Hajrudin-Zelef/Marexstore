package handlers

import (
	"net/http"
	"net/url"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/go-chi/chi/v5"
)

// queryState : état initial anti-flash d'une page catégorie depuis l'URL.
// ?type=<slug> surligne la tuile et affiche la liste, ?all=1 affiche la liste.
func queryState(q url.Values) (active string, showList bool) {
	if t := strings.TrimSpace(q.Get("type")); t != "" {
		return t, true
	}
	if a := strings.TrimSpace(q.Get("all")); a != "" && a != "all" {
		return "", true
	}
	return "", false
}

// Les handlers passent le token via closure : on le substitue ici.
func (a *App) renderDoc(w http.ResponseWriter, r *http.Request, title, bodyHTML, prefix string) {
	// Canonical absolu = la page de référence (sans query) : le copieur est déclassé.
	path := r.URL.EscapedPath()
	if path == "/index.html" {
		path = "/"
	}
	extra := `<link rel="canonical" href="` + strings.TrimSuffix(a.Cfg.BaseURL, "/") + path + `">` + render.ThemeHead(catalog.SiteTheme)
	s := render.DocTheme(title, bodyHTML, prefix, extra, catalog.SiteTheme.HtmlAttrs())
	s = strings.ReplaceAll(s, "{{CSRF}}", auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r)))
	a.write(w, 200, s)
}

func (a *App) Index(w http.ResponseWriter, r *http.Request) {
	a.renderDoc(w, r, "Accueil", render.IndexBody(a.Store, a.Recents), "")
}

func (a *App) Boutique(w http.ResponseWriter, r *http.Request) {
	a.renderDoc(w, r, "Store", render.BoutiqueBodyQuery(a.Store, r.URL.Query()), "")
}

func (a *App) Categorie(w http.ResponseWriter, r *http.Request) {
	f := strings.TrimSuffix(chi.URLParam(r, "fichier"), ".html")
	q := r.URL.Query()
	active, showList := queryState(q)
	for _, m := range catalog.MainOrder {
		if catalog.MainSlug[m] == f {
			a.renderDoc(w, r, m, render.CategoryPageFiltered(a.Store, m, active, showList), "../")
			return
		}
	}
	// Tops live du menu « Toutes les catégories » : même page, types d'abord.
	for _, t := range catalog.LiveCats {
		if t.Slug == f {
			a.renderDoc(w, r, t.Label, render.CategoryPageLiveFiltered(a.Store, t, active, showList), "../")
			return
		}
	}
	http.NotFound(w, r)
}

// CategorieType : page dédiée d'un type (/categories/<top>/<type>.html).
// Profondeur 2 → préfixe "../../" pour le <head> (assets) comme pour le corps.
func (a *App) CategorieType(w http.ResponseWriter, r *http.Request) {
	top := chi.URLParam(r, "top")
	typ := strings.TrimSuffix(chi.URLParam(r, "type"), ".html")
	if title, body, ok := render.CategoryTypePage(a.Store, top, typ); ok {
		a.renderDoc(w, r, title, body, "../../")
		return
	}
	http.NotFound(w, r)
}

func (a *App) Produit(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimSuffix(chi.URLParam(r, "slug"), ".html")
	p, ok := a.Store.BySlug[slug]
	if !ok {
		http.NotFound(w, r)
		return
	}
	body := render.Header(a.Store, "../", "") + render.FicheBody(a.Store, p) + "\n" + render.Footer(a.Store, "../")
	a.renderDoc(w, r, p.Nom, body, "../")
}

func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

func (a *App) FacetsJS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/javascript; charset=utf-8")
	w.Header().Set("Cache-Control", "public, max-age=3600")
	w.Write([]byte(a.Store.FacetsJS()))
}
