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

// Le moyen choisi au panier doit être persisté tel quel (pas rabattu sur
// "whatsapp"), y compris "cod" (paiement à la livraison, proposé par l'UI).
func TestApiCommandeMoyenChoisi(t *testing.T) {
	for _, moyen := range []string{"orange", "mtn", "wave", "moov", "cod", "whatsapp"} {
		t.Run(moyen, func(t *testing.T) {
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
			// Moyens cochés en admin (comme une vraie boutique configurée).
			for _, c := range []string{"orange", "mtn", "wave", "moov", "cod"} {
				if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES(?, '1')`, "paiement_"+c+"_on"); err != nil {
					t.Fatal(err)
				}
			}
			form := url.Values{
				"items": {"p1x1"}, "moyen": {moyen}, "nom": {"T"}, "tel": {"0700"},
				"numero_payeur": {"0701020304"},
				"csrf": {auth.CSRFToken("test-secret", "anon")},
			}
			req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			app.ApiCommande(rec, req)
			if rec.Code != 200 {
				t.Fatalf("code %d, body %q", rec.Code, rec.Body.String())
			}
			var got string
			if err := database.QueryRow(`SELECT moyen_paiement FROM orders ORDER BY id DESC LIMIT 1`).Scan(&got); err != nil {
				t.Fatal(err)
			}
			if got != moyen {
				t.Fatalf("moyen enregistré=%q, choisi=%q (rabattu ?)", got, moyen)
			}
		})
	}
}

// Un moyen désactivé côté admin après chargement du panier ne passe plus :
// repli explicite sur whatsapp (jamais de confiance aveugle au client).
func TestApiCommandeMoyenDesactiveRepli(t *testing.T) {
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
	app.ensurePromos()
	if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES('paiement_cod_on','1')`); err != nil {
		t.Fatal(err)
	}
	form := url.Values{
		"items": {"p1x1"}, "moyen": {"moov"}, "nom": {"T"}, "tel": {"0700"},
		"csrf": {auth.CSRFToken("test-secret", "anon")},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.ApiCommande(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var got string
	_ = database.QueryRow(`SELECT moyen_paiement FROM orders ORDER BY id DESC LIMIT 1`).Scan(&got)
	if got != "whatsapp" {
		t.Fatalf("moyen désactivé accepté : %q", got)
	}
}

// Un moyen inconnu ne doit jamais créer de valeur fantaisiste : repli
// explicite sur whatsapp.
func TestApiCommandeMoyenInconnuRepli(t *testing.T) {
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
	form := url.Values{
		"items": {"p1x1"}, "moyen": {"bitcoin"}, "nom": {"T"}, "tel": {"0700"},
		"csrf": {auth.CSRFToken("test-secret", "anon")},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.ApiCommande(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d", rec.Code)
	}
	var got string
	_ = database.QueryRow(`SELECT moyen_paiement FROM orders ORDER BY id DESC LIMIT 1`).Scan(&got)
	if got != "whatsapp" {
		t.Fatalf("moyen inconnu enregistré tel quel : %q", got)
	}
}
