package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
)

// Statut paiement par défaut à la création : mobile → en_attente,
// cod → a_encaisser, sinon → en_attente.
func TestApiCommandePaiementStatut(t *testing.T) {
	cases := []struct {
		nom     string
		moyen   string
		payeur  string
		attendu string
	}{
		{"mobile+payeur", "orange", "0701020304", "en_attente"},
		{"cod", "cod", "", "a_encaisser"},
		{"whatsapp", "whatsapp", "", "en_attente"},
	}
	for _, tc := range cases {
		t.Run(tc.nom, func(t *testing.T) {
			dir := t.TempDir()
			database, err := db.Open(filepath.Join(dir, "test.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
				t.Fatal(err)
			}
			if _, err := database.Exec(`INSERT INTO products(slug,nom,prix,stock) VALUES('p1','P1',10000,5)`); err != nil {
				t.Fatal(err)
			}
			app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
			for _, c := range []string{"orange", "mtn", "wave", "moov", "cod"} {
				if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES(?, '1')`, "paiement_"+c+"_on"); err != nil {
					t.Fatal(err)
				}
			}
			form := url.Values{
				"items": {"p1x1"}, "moyen": {tc.moyen}, "nom": {"T"}, "tel": {"0700"},
				"csrf": {auth.CSRFToken("test-secret", "anon")},
			}
			if tc.payeur != "" {
				form.Set("numero_payeur", tc.payeur)
			}
			req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			app.ApiCommande(rec, req)
			if rec.Code != 200 {
				t.Fatalf("code %d, body %q", rec.Code, rec.Body.String())
			}
			var got string
			if err := database.QueryRow(`SELECT paiement_statut FROM orders ORDER BY id DESC LIMIT 1`).Scan(&got); err != nil {
				t.Fatalf("paiement_statut illisible: %v", err)
			}
			if got != tc.attendu {
				t.Fatalf("paiement_statut=%q, attendu %q (moyen %q)", got, tc.attendu, tc.moyen)
			}
		})
	}
}
