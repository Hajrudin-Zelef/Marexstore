package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
)

// Page /commander.html : template + porte de connexion + 4 étapes du tunnel.
func TestCommanderPage(t *testing.T) {
	top, drawer, foot, bodies, err := site.LoadStatic("../../templates/static")
	if err != nil {
		t.Fatal(err)
	}
	app := &App{
		Cfg:    config.Config{SessionSecret: "test-secret"},
		Store:  catalog.NewStore(nil),
		Top:    top,
		Drawer: drawer,
		Foot:   foot,
		Bodies: bodies,
	}
	req := httptest.NewRequest(http.MethodGet, "/commander.html", nil)
	rec := httptest.NewRecorder()
	app.StaticPage(rec, req, "Commander", "commander", "")
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /commander.html = %d, voulu 200", rec.Code)
	}
	body := rec.Body.String()
	for _, m := range []string{
		"data-param-locked",
		"data-param-content",
		"data-checkout-livraison",
		"data-checkout-form",
		`name="nom"`,
		`name="tel"`,
		`name="adresse"`,
		"numero_payeur",
		"data-cart-paiement",
		"data-checkout-recap",
		"Commander",
		"data-checkout-ok",
	} {
		if !strings.Contains(body, m) {
			t.Errorf("commander.html : marqueur %q absent", m)
		}
	}
}
