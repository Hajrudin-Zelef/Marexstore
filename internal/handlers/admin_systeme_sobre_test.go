package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// /admin/systeme : page d'état sobre — aucun détail interne du backend
// (version Go, chemin de la base, tailles en octets, comptes FTS) ne doit
// fuiter dans le HTML, même derrière l'auth admin.
func TestAdminSystemeSobre(t *testing.T) {
	app, _ := testAppDesign(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/systeme", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminSysteme(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin/systeme : %d", rec.Code)
	}
	body := rec.Body.String()
	for _, secret := range []string{
		"go1.", "Base de données", ".db", "WAL", "Ko", " Mo", " o</", "o<",
		"runtime", "theme_version", "theme v",
	} {
		if strings.Contains(body, secret) {
			t.Fatalf("fuite backend sur /admin/systeme : %q présent", secret)
		}
	}
	for _, want := range []string{
		"État du backend", "Produits", "Clients", "Commandes",
		"Sauvegarder", "intégrité",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("état essentiel manquant : %q", want)
		}
	}
}
