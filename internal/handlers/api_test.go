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

// T1 — Idempotency-Key : rejouer la même commande ne crée pas de doublon.
// Commande + lignes + événement de création en une seule transaction.
func TestApiCommandeIdempotente(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO products(slug,nom,marque,prix,stock) VALUES('test-produit','Produit test','Test',10000,5)`); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
	csrf := auth.CSRFToken("test-secret", "anon")
	poste := func(key string) *httptest.ResponseRecorder {
		form := url.Values{
			"items": {"test-produitx2"},
			"moyen": {"whatsapp"},
			"nom":   {"Test"},
			"tel":   {"0700000000"},
			"csrf":  {csrf},
		}
		req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		if key != "" {
			req.Header.Set("Idempotency-Key", key)
		}
		rec := httptest.NewRecorder()
		app.ApiCommande(rec, req)
		return rec
	}
	// 1re requête (clé A) : crée la commande.
	if rec := poste("key-A"); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("1re commande : code %d, body %q", rec.Code, rec.Body.String())
	}
	// Rejeu (même clé A) : OK mais pas de doublon.
	if rec := poste("key-A"); rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("rejeu clé A : code %d, body %q", rec.Code, rec.Body.String())
	}
	// Autre clé (B) : 2e commande.
	if rec := poste("key-B"); rec.Code != 200 {
		t.Fatalf("clé B : code %d", rec.Code)
	}
	var nOrders, nItems, nEvents, nKeys int
	_ = database.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&nOrders)
	_ = database.QueryRow(`SELECT COUNT(*) FROM order_items`).Scan(&nItems)
	_ = database.QueryRow(`SELECT COUNT(*) FROM order_events`).Scan(&nEvents)
	_ = database.QueryRow(`SELECT COUNT(*) FROM idempotency_keys`).Scan(&nKeys)
	if nOrders != 2 || nItems != 2 || nEvents != 2 || nKeys != 2 {
		t.Fatalf("attendu 2 commandes / 2 lignes / 2 événements / 2 clés, got %d / %d / %d / %d", nOrders, nItems, nEvents, nKeys)
	}
	var total int64
	_ = database.QueryRow(`SELECT total FROM orders ORDER BY id LIMIT 1`).Scan(&total)
	if total != 20000 {
		t.Fatalf("total attendu 20000 (2 x 10000, calculé serveur), got %d", total)
	}
	var vers, par string
	_ = database.QueryRow(`SELECT vers_statut, par FROM order_events ORDER BY id LIMIT 1`).Scan(&vers, &par)
	if vers != "whatsapp_recu" || par != "site" {
		t.Fatalf("événement de création attendu (whatsapp_recu/site), got %q/%q", vers, par)
	}
}

// Le client paie le prix affiché : si prix_boutique existe, le total et le
// prix unitaire utilisent le prix boutique, pas le prix stock.
func TestApiCommandeTotalisePrixBoutique(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO products(slug,nom,marque,prix,prix_boutique,stock) VALUES('toner-test','Toner Test','HP',145000,110000,5)`); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
	csrf := auth.CSRFToken("test-secret", "anon")
	form := url.Values{
		"items": {"toner-testx1"},
		"moyen": {"whatsapp"},
		"nom":   {"Test"},
		"tel":   {"0700000000"},
		"csrf":  {csrf},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Set("Idempotency-Key", "key-boutique")
	rec := httptest.NewRecorder()
	app.ApiCommande(rec, req)
	if rec.Code != 200 {
		t.Fatalf("commande : code %d, body %q", rec.Code, rec.Body.String())
	}
	var total, pu int64
	_ = database.QueryRow(`SELECT total FROM orders`).Scan(&total)
	_ = database.QueryRow(`SELECT prix_unitaire FROM order_items`).Scan(&pu)
	if total != 110000 || pu != 110000 {
		t.Fatalf("total=%d prix_unitaire=%d, attendu 110000/110000 (prix boutique)", total, pu)
	}
}
