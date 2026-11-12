package handlers

import (
	"database/sql"
	"encoding/json"
	"html/template"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/models"
)

type App struct {
	Cfg  config.Config
	DB   *sql.DB
	Tpl  *template.Template
	SecureCookies bool
}

func norm(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, "-", " ")
	var b strings.Builder
	for _, r := range s {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == ' ' {
			b.WriteRune(r)
		} else if r == ' ' {
			b.WriteRune(' ')
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

// Health : sonde Docker/nginx.
func (a *App) Health(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.Write([]byte(`{"ok":true}`))
}

type homeData struct {
	Title    string
	Count    int
	Cats     []catCount
	Products []models.Product
	Query    string
	CSRF     string
}

type catCount struct {
	Name string
	N    int
}

func (a *App) sessionID(r *http.Request) string {
	c, err := r.Cookie("marex_session")
	if err != nil {
		return "anon"
	}
	return c.Value
}

func (a *App) Home(w http.ResponseWriter, r *http.Request) {
	var n int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&n)
	rows, _ := a.DB.Query(`SELECT categorie, COUNT(*) FROM products GROUP BY categorie ORDER BY 2 DESC LIMIT 8`)
	var cats []catCount
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var c catCount
			_ = rows.Scan(&c.Name, &c.N)
			cats = append(cats, c)
		}
	}
	prods, _ := a.latest(12)
	_ = a.Tpl.ExecuteTemplate(w, "index", homeData{
		Title: "Accueil — Marexstore", Count: n, Cats: cats, Products: prods,
		CSRF: auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r)),
	})
}

func (a *App) latest(limit int) ([]models.Product, error) {
	rows, err := a.DB.Query(`SELECT id,slug,nom,marque,prix,devise,dispo,categorie,image,accroche,stock
		FROM products ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []models.Product
	for rows.Next() {
		var p models.Product
		var prix sql.NullInt64
		_ = rows.Scan(&p.ID, &p.Slug, &p.Nom, &p.Marque, &prix, &p.Devise, &p.Dispo, &p.Categorie, &p.Image, &p.Accroche, &p.Stock)
		if prix.Valid {
			v := prix.Int64
			p.Prix = &v
		}
		out = append(out, p)
	}
	return out, nil
}

func (a *App) Boutique(w http.ResponseWriter, r *http.Request) {
	q := strings.TrimSpace(r.URL.Query().Get("q"))
	cat := strings.TrimSpace(r.URL.Query().Get("cat"))
	page, _ := strconv.Atoi(r.URL.Query().Get("page"))
	if page < 1 {
		page = 1
	}
	const per = 24
	off := (page - 1) * per

	var args []any
	where := "1=1"
	if q != "" {
		// FTS si dispo, sinon LIKE normalisé.
		words := strings.Fields(norm(q))
		if len(words) > 0 {
			where += " AND (nom LIKE ? OR marque LIKE ? OR categorie LIKE ? OR accroche LIKE ?)"
			like := "%" + words[0] + "%"
			args = append(args, like, like, like, like)
		}
	}
	if cat != "" {
		where += " AND categorie = ?"
		args = append(args, cat)
	}
	rows, err := a.DB.Query(`SELECT id,slug,nom,marque,prix,devise,dispo,categorie,image,accroche,stock
		FROM products WHERE `+where+` ORDER BY id DESC LIMIT ? OFFSET ?`, append(args, per, off)...)
	if err != nil {
		http.Error(w, "Boutique indisponible", 500)
		return
	}
	defer rows.Close()
	var prods []models.Product
	for rows.Next() {
		var p models.Product
		var prix sql.NullInt64
		_ = rows.Scan(&p.ID, &p.Slug, &p.Nom, &p.Marque, &prix, &p.Devise, &p.Dispo, &p.Categorie, &p.Image, &p.Accroche, &p.Stock)
		if prix.Valid {
			v := prix.Int64
			p.Prix = &v
		}
		prods = append(prods, p)
	}
	// Fragment htmx : ne renvoie que les cartes.
	if r.Header.Get("HX-Request") == "true" {
		_ = a.Tpl.ExecuteTemplate(w, "cards", prods)
		return
	}
	_ = a.Tpl.ExecuteTemplate(w, "boutique", map[string]any{
		"Title": "Boutique — Marexstore", "Products": prods, "Q": q, "Cat": cat, "Page": page,
		"CSRF": auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r)),
	})
}

func (a *App) Produit(w http.ResponseWriter, r *http.Request) {
	slug := strings.TrimPrefix(r.URL.Path, "/produits/")
	slug = strings.Split(slug, "/")[0]
	if ok := len(slug) > 2 && len(slug) < 120; !ok {
		http.NotFound(w, r)
		return
	}
	var p models.Product
	var prix sql.NullInt64
	err := a.DB.QueryRow(`SELECT id,slug,nom,marque,prix,devise,dispo,categorie,sous_categorie,image,accroche,description,specs_json,stock
		FROM products WHERE slug=?`, slug).Scan(&p.ID, &p.Slug, &p.Nom, &p.Marque, &prix, &p.Devise, &p.Dispo,
		&p.Categorie, &p.SousCategorie, &p.Image, &p.Accroche, &p.Description, &p.SpecsJSON, &p.Stock)
	if err != nil {
		http.NotFound(w, r)
		return
	}
	if prix.Valid {
		v := prix.Int64
		p.Prix = &v
	}
	_ = json.Unmarshal([]byte(p.SpecsJSON), &p.Specs)
	_ = a.Tpl.ExecuteTemplate(w, "produit", map[string]any{
		"Title": p.Nom + " — Marexstore", "P": p,
		"CSRF":  auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r)),
		"WA":    "https://wa.me/" + a.Cfg.WhatsApp + "?text=" + url.QueryEscape("Bonjour, je veux : "+p.Nom),
	})
}
