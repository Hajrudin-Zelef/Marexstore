package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
	"github.com/go-chi/chi/v5"
)

func testAppCommander(t *testing.T) *App {
	t.Helper()
	top, drawer, foot, bodies, err := site.LoadStatic("../../templates/static")
	if err != nil {
		t.Fatal(err)
	}
	return &App{
		Cfg:    config.Config{SessionSecret: "test-secret"},
		Store:  catalog.NewStore(nil),
		Top:    top,
		Drawer: drawer,
		Foot:   foot,
		Bodies: bodies,
	}
}

// Page /commander.html : template + porte de connexion + 4 étapes du tunnel.
func TestCommanderPage(t *testing.T) {
	app := testAppCommander(t)
	req := httptest.NewRequest(http.MethodGet, "/commander.html", nil)
	rec := httptest.NewRecorder()
	app.StaticPage(rec, req, "Commander", "commander", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /commander.html = %d, voulu 200", rec.Code)
	}
	body := rec.Body.String()
	for _, m := range []string{
		"data-param-locked",
		// Porte fermée par défaut : le contenu exige un compte (JS : renderParamGate).
		"data-param-content hidden",
		// Hook livraison unique, partagé avec le tiroir (un seul sélecteur JS).
		"data-cart-livraison",
		"data-checkout-form",
		`name="nom"`,
		`name="tel"`,
		`name="adresse"`,
		"numero_payeur",
		"data-cart-paiement",
		"data-checkout-recap",
		"data-promo-input",
		"data-checkout-empty",
		"data-checkout-submit",
		"Commander",
		"data-checkout-ok",
	} {
		if !strings.Contains(body, m) {
			t.Errorf("commander.html : marqueur %q absent", m)
		}
	}
}

// Route /commander.html : câblée comme main.go (chi + StaticRoutes) et servie 200.
// Une coquille dans la clé statics échouerait ici, pas dans TestCommanderPage.
func TestCommanderRoute(t *testing.T) {
	tn, ok := StaticRoutes()["/commander.html"]
	if !ok {
		t.Fatal("StaticRoutes : clé /commander.html absente")
	}
	if tn != [2]string{"commander", "Commander"} {
		t.Fatalf("StaticRoutes : /commander.html = %q, voulu {commander Commander}", tn)
	}
	app := testAppCommander(t)
	r := chi.NewRouter()
	for route, pair := range StaticRoutes() {
		name, title := pair[0], pair[1]
		r.Get(route, func(w http.ResponseWriter, req *http.Request) {
			app.StaticPage(w, req, title, name, "")
		})
	}
	for _, path := range []string{"/commander.html", "/compte.html"} {
		req := httptest.NewRequest(http.MethodGet, path, nil)
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Errorf("GET %s via routeur = %d, voulu 200", path, rec.Code)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/commander.html", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	if !strings.Contains(rec.Body.String(), "data-cart-livraison") {
		t.Error("route /commander.html : hook data-cart-livraison absent du rendu")
	}
}
