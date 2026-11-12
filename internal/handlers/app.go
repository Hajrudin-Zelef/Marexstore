package handlers

import (
	"database/sql"
	"net/http"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

type App struct {
	Cfg           config.Config
	DB            *sql.DB
	Store         *catalog.Store
	Recents       []*catalog.RawProduct
	Top           string
	Drawer        string
	Foot          string
	Bodies        map[string]string
	SecureCookies bool
}

func (a *App) sessionID(r *http.Request) string {
	c, err := r.Cookie("marex_session")
	if err != nil || c.Value == "" {
		return "anon"
	}
	return c.Value
}

func (a *App) csrf(r *http.Request) string {
	return auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r))
}

func (a *App) write(w http.ResponseWriter, code int, html string) {
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.WriteHeader(code)
	w.Write([]byte(html))
}

// page statique : head + top + corps + drawer + widgets + pied.
func (a *App) StaticPage(w http.ResponseWriter, r *http.Request, title, name, prefix string) {
	body, ok := a.Bodies[name]
	if !ok {
		http.NotFound(w, r)
		return
	}
	csrf := a.csrf(r)
	fill := func(s string) string {
		s = strings.ReplaceAll(s, "{{.Prefix}}", prefix)
		return strings.ReplaceAll(s, "{{.CSRF}}", csrf)
	}
	// NB : le pied statique se termine déjà par </body></html> (verbatim d'origine).
	out := `<!DOCTYPE html><html lang="fr"><head>` + render.Head(title, prefix, "") + `</head><body>` +
		fill(a.Top) + "\n" + fill(a.Drawer) + fill(body) +
		render.WidgetsHTML(a.Store, prefix) +
		strings.ReplaceAll(strings.ReplaceAll(a.Foot, "{{.Widgets}}", ""), "{{.Prefix}}", prefix)
	a.write(w, 200, out)
}
