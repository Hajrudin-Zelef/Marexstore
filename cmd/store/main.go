package main

import (
	"database/sql"
	"html/template"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/handlers"
	"github.com/Hajrudin-Zelef/Marexstore/internal/middleware"
)

func main() {
	cfg := config.Load()
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	if _, err := os.Stat(cfg.DBPath); err == nil {
		// Migrations idempotentes (IF NOT EXISTS).
		_ = db.Migrate(database, "db/migrations/001_init.sql")
	}
	tpl := template.Must(template.ParseGlob("templates/pages/*.html"))
	// Partials inclus dans les pages via {{define}} ; pas de parse séparé nécessaire.
	app := &handlers.App{
		Cfg: cfg, DB: database, Tpl: tpl,
		SecureCookies: os.Getenv("APP_ENV") == "prod",
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /health", app.Health)
	mux.HandleFunc("GET /", app.Home)
	mux.HandleFunc("GET /boutique", app.Boutique)
	mux.HandleFunc("GET /produits/", app.Produit)
	mux.HandleFunc("POST /panier/ajouter", app.PanierAjouter)
	mux.HandleFunc("GET /commande", app.Commande)
	mux.HandleFunc("POST /commande", app.Commande)
	mux.HandleFunc("GET /compte", app.AuthPage)
	mux.HandleFunc("POST /auth/inscription", app.Register)
	mux.HandleFunc("POST /auth/connexion", app.Login)
	mux.HandleFunc("POST /auth/logout", app.Logout)
	mux.HandleFunc("GET /admin", app.Admin)

	// Statique : cache long + immutable (les fichiers sont versionnés par ?v=).
	static := http.FileServer(http.Dir("static"))
	mux.Handle("GET /static/", http.StripPrefix("/static/", withCache(static)))
	// Images importées de l'archive (transitoire avant thumbs/).
	mux.Handle("GET /images/", http.StripPrefix("/", http.FileServer(http.Dir(".")) ))

	reportOnly := os.Getenv("CSP_REPORT_ONLY") != "0"
	var h http.Handler = mux
	h = middleware.RequestID(h)
	h = middleware.SecurityHeaders(reportOnly)(h)
	// Rate-limit doux global : 40 req/s burst 80 (n'ennuie pas l'acheteur normal).
	h = middleware.RateLimit(40, 80)(h)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           h,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("marexstore :%s env=%s db=%s", cfg.Port, cfg.Env, cfg.DBPath)
	_ = database.Ping()
	log.Fatal(srv.ListenAndServe())
}

func withCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

var _ = sql.ErrNoRows
