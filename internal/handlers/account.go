package handlers

import (
	"net/http"
	"strings"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
)

// Auth : compte optionnel. Inscription/connexion sobres, messages génériques.
func (a *App) AuthPage(w http.ResponseWriter, r *http.Request) {
	_ = a.Tpl.ExecuteTemplate(w, "compte", map[string]any{
		"Title": "Compte — Marexstore",
		"CSRF":  auth.CSRFToken(a.Cfg.SessionSecret, a.sessionID(r)),
	})
}

func (a *App) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", 405)
		return
	}
	if !auth.CheckCSRF(a.Cfg.SessionSecret, a.sessionID(r), r.FormValue("csrf")) {
		http.Error(w, "Session expirée.", 403)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	pw := r.FormValue("password")
	nom := strings.TrimSpace(r.FormValue("nom"))
	if !strings.Contains(email, "@") || len(pw) < 8 || len(email) > 120 {
		http.Error(w, "Email valide + mot de passe 8 caractères minimum.", 400)
		return
	}
	h, _ := auth.HashPassword(pw)
	_, err := a.DB.Exec(`INSERT INTO users(email,pass_hash,nom) VALUES(?,?,?)`, email, h, nom)
	if err != nil {
		// Générique anti-énumération.
		http.Error(w, "Impossible de créer ce compte.", 400)
		return
	}
	var uid int64
	_ = a.DB.QueryRow(`SELECT id FROM users WHERE email=?`, email).Scan(&uid)
	tok, _ := auth.CreateSession(a.DB, uid, r, 30*24*time.Hour)
	auth.SetCookie(w, tok, 30*24*time.Hour, a.SecureCookies)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Méthode non autorisée", 405)
		return
	}
	if !auth.CheckCSRF(a.Cfg.SessionSecret, a.sessionID(r), r.FormValue("csrf")) {
		http.Error(w, "Session expirée.", 403)
		return
	}
	email := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
	var uid int64
	var hash string
	if err := a.DB.QueryRow(`SELECT id,pass_hash FROM users WHERE email=?`, email).Scan(&uid, &hash); err != nil ||
		!auth.CheckPassword(hash, r.FormValue("password")) {
		http.Error(w, "Identifiants invalides.", 401)
		return
	}
	tok, _ := auth.CreateSession(a.DB, uid, r, 30*24*time.Hour)
	auth.SetCookie(w, tok, 30*24*time.Hour, a.SecureCookies)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (a *App) Logout(w http.ResponseWriter, r *http.Request) {
	auth.ClearCookie(w)
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

// Admin minimal V1 : liste commandes + produits, protégé par ADMIN_PASS_HASH (bcrypt).
func (a *App) Admin(w http.ResponseWriter, r *http.Request) {
	user, pass, ok := r.BasicAuth()
	if !ok || user != a.Cfg.AdminUser || a.Cfg.AdminPassHash == "" ||
		!auth.CheckPassword(a.Cfg.AdminPassHash, pass) {
		w.Header().Set("WWW-Authenticate", `Basic realm="admin"`)
		http.Error(w, "Non autorisé.", 401)
		return
	}
	rows, _ := a.DB.Query(`SELECT id,statut,moyen_paiement,total,created_at FROM orders ORDER BY id DESC LIMIT 50`)
	type O struct {
		ID, Statut, Moyen string
		Total             *int64
		At                string
	}
	var orders []O
	if rows != nil {
		defer rows.Close()
		for rows.Next() {
			var o O
			var t interface{}
			_ = rows.Scan(&o.ID, &o.Statut, &o.Moyen, &t, &o.At)
			orders = append(orders, o)
		}
	}
	var nProd, nUsers int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&nProd)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM users`).Scan(&nUsers)
	_ = a.Tpl.ExecuteTemplate(w, "admin", map[string]any{
		"Title": "Admin — Marexstore", "Orders": orders, "NProd": nProd, "NUsers": nUsers,
	})
}
