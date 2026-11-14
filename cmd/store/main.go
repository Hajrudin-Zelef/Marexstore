package main

import (
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/handlers"
	"github.com/Hajrudin-Zelef/Marexstore/internal/middleware"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
)

func main() {
	cfg := config.Load()
	database, err := db.Open(cfg.DBPath)
	if err != nil {
		log.Fatal(err)
	}
	if err := db.EnsureSchema(database, "db/migrations"); err != nil {
		log.Fatal("schema: ", err)
	}
	if err := db.EnsureFTS(database); err != nil {
		log.Fatal("fts: ", err)
	}
	// Réglages persistés : WhatsApp sans recompiler, maintenance immédiate.
	var waSetting string
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='whatsapp'`).Scan(&waSetting)
	if strings.TrimSpace(waSetting) != "" {
		catalog.WA = strings.TrimSpace(waSetting)
	}
	// Thème global du Studio Design (/admin/design) ; défaut = site d'origine.
	catalog.SiteTheme = site.ThemeFromDB(database)
	maintenanceOn := func() (bool, string) {
		var on, msg string
		_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='maintenance'`).Scan(&on)
		_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='maintenance_message'`).Scan(&msg)
		if msg == "" {
			msg = "Boutique en maintenance, retour très vite."
		}
		return on == "1", msg
	}
	st, err := site.LoadStore(database)
	if err != nil {
		log.Fatal("load store: ", err)
	}
	top, drawer, foot, bodies, err := site.LoadStatic("templates/static")
	if err != nil {
		log.Fatal("load templates: ", err)
	}
	app := &handlers.App{
		Cfg: cfg, DB: database, Store: st, Recents: nil,
		Top: top, Drawer: drawer, Foot: foot, Bodies: bodies,
		SecureCookies: os.Getenv("APP_ENV") == "prod",
	}
	app.Recents = app.Store.Recents("")

	r := chi.NewRouter()
	r.Use(chimw.Recoverer)
	r.Use(middleware.RequestID)
	r.Use(middleware.SecurityHeaders(os.Getenv("CSP_REPORT_ONLY") != "0"))
	r.Use(middleware.RateLimit(40, 80))
	// Anti-indexation des zones privées/API (complète robots.txt côté nginx).
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/admin") || strings.HasPrefix(r.URL.Path, "/api/") {
				w.Header().Set("X-Robots-Tag", "noindex, nofollow")
			}
			next.ServeHTTP(w, r)
		})
	})
	// Mode maintenance (réglages admin) : tout le public en 503, /admin et /health passent.
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.HasPrefix(r.URL.Path, "/admin") || r.URL.Path == "/health" || r.URL.Path == "/robots.txt" {
				next.ServeHTTP(w, r)
				return
			}
			if on, msg := maintenanceOn(); on {
				w.Header().Set("Content-Type", "text/html; charset=utf-8")
				w.WriteHeader(http.StatusServiceUnavailable)
				fmt.Fprintf(w, `<!DOCTYPE html><html lang="fr"><head><meta charset="utf-8"><meta name="viewport" content="width=device-width, initial-scale=1"><title>Maintenance — Marexstore</title></head><body style="font-family:sans-serif;text-align:center;padding:80px 20px"><h1>Marexstore</h1><p>%s</p></body></html>`, template.HTMLEscapeString(msg))
				return
			}
			next.ServeHTTP(w, r)
		})
	})

	r.Get("/health", app.Health)
	// robots.txt : pas d'admin ni d'API pour les crawlers ; crawl doux.
	r.Get("/robots.txt", func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Write([]byte("User-agent: *\nDisallow: /admin\nDisallow: /api/\nCrawl-delay: 10\n"))
	})
	r.Get("/", app.Index)
	r.Get("/index.html", app.Index)
	r.Get("/boutique.html", app.Boutique)
	r.Get("/categories/{top}/{type}.html", app.CategorieType)
	r.Get("/categories/{fichier}.html", app.Categorie)
	r.Get("/produits/{slug}.html", app.Produit)

	// Pages statiques (corps verbatim + chrome partagé).
	statics := map[string][2]string{
		"/apropos.html":                     {"apropos", "À propos de nous"},
		"/cgv.html":                         {"cgv", "Conditions générales de vente"},
		"/livraison.html":                   {"livraison", "Livraison & retours"},
		"/confidentialite.html":             {"confidentialite", "Politique de confidentialité"},
		"/mentions-legales.html":            {"mentions-legales", "Mentions légales"},
		"/contact.html":                     {"contact", "Contact"},
		"/contactez-nous.html":              {"contactez-nous", "Contactez-nous"},
		"/devis.html":                       {"devis", "Demande de devis"},
		"/suivi-commande.html":              {"suivi-commande", "Suivi de commande"},
		"/compte.html":                      {"compte", "Mon compte"},
		"/parametres.html":                  {"parametres", "Mes paramètres"},
		"/services.html":                    {"services", "Nos Logiciels Metiers"},
		"/nos-services.html":                {"nos-services", "Nos services"},
		"/nos-references.html":              {"nos-references", "Nos références"},
		"/qui-sommes-nous.html":             {"qui-sommes-nous", "Qui sommes-nous ?"},
		"/solutions-infrastructure-it.html": {"solutions-infrastructure-it", "Infrastructure IT"},
		"/solutions-collaboration.html":     {"solutions-collaboration", "Collaboration"},
		"/solutions-cloud-securite.html":    {"solutions-cloud-securite", "Cloud & Sécurité"},
	}
	for route, tn := range statics {
		name, title := tn[0], tn[1]
		r.Get(route, func(w http.ResponseWriter, req *http.Request) {
			app.StaticPage(w, req, title, name, "")
		})
	}

	r.Get("/assets/facets.js", app.FacetsJS)
	static := http.FileServer(http.Dir("static"))
	r.Handle("/static/*", http.StripPrefix("/static/", withCache(static)))
	r.Handle("/assets/*", http.StripPrefix("/assets/", withCache(http.FileServer(http.Dir("static")))))
	imgs := http.FileServer(http.Dir("./images"))
	r.Handle("/images/*", http.StripPrefix("/images/", withImageCache(imgs)))

	// Miroir serveur (mirror.js) + auth.
	r.Post("/api/panier", app.ApiPanier)
	r.Post("/api/commande", app.ApiCommande)
	r.Post("/api/devis", app.ApiDemande("devis"))
	r.Post("/api/suivi", app.ApiDemande("suivi"))
	r.Post("/api/contact", app.ApiDemande("contact"))
	r.Post("/api/infos", app.ApiDemande("infos"))
	r.Get("/api/csrf", app.ApiCSRF)
	r.Get("/api/recherche", app.ApiRecherche)
	r.Post("/auth/inscription", app.Register)
	r.Post("/auth/connexion", app.Login)
	r.Post("/auth/logout", app.Logout)
	r.Post("/auth/mot-de-passe", app.Password)

	// Admin maison.
	r.Get("/admin", app.Admin)
	r.Get("/admin/produits", app.AdminProduits)
	r.Get("/admin/produits/nouveau", app.AdminProduitEdit)
	r.Get("/admin/produits/{slug}", app.AdminProduitEdit)
	r.Post("/admin/produits/{slug}", app.AdminProduitEdit)
	r.Post("/admin/produits/nouveau", app.AdminProduitEdit)
	r.Get("/admin/commandes", app.AdminCommandes)
	r.Post("/admin/commandes", app.AdminCommandes)
	r.Get("/admin/commandes/{id}", app.AdminCommande)
	r.Post("/admin/commandes/{id}", app.AdminCommande)
	r.Get("/admin/clients", app.AdminClients)
	r.Get("/admin/clients/{id}", app.AdminClient)
	r.Get("/admin/medias", app.AdminMedias)
	r.Post("/admin/medias", app.AdminMedias)
	r.Get("/admin/types", app.AdminTypes)
	r.Post("/admin/types", app.AdminTypes)
	r.Get("/admin/reglages", app.AdminReglages)
	r.Post("/admin/reglages", app.AdminReglages)
	r.Get("/admin/design", app.AdminDesign)
	r.Post("/admin/design", app.AdminDesign)
	r.Post("/admin/design/reset", app.AdminDesign)
	r.Get("/admin/design/apercu", app.AdminDesignPreview)
	r.Post("/admin/design/apercu", app.AdminDesignPreview)

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           r,
		ReadTimeout:       10 * time.Second,
		ReadHeaderTimeout: 5 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}
	log.Printf("marexstore :%s env=%s db=%s produits=%d", cfg.Port, cfg.Env, cfg.DBPath, len(app.Store.Products))
	_ = database.Ping()
	log.Fatal(srv.ListenAndServe())
}

func withCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		next.ServeHTTP(w, r)
	})
}

func withImageCache(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "public, max-age=86400")
		next.ServeHTTP(w, r)
	})
}
