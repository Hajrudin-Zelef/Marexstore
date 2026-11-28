package handlers

import (
	"encoding/json"
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

func testAppCommandeLivraison(t *testing.T) (*App, string) {
	t.Helper()
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO products(slug,nom,marque,prix,stock) VALUES('test-produit','Produit test','Test',10000,5)`); err != nil {
		t.Fatal(err)
	}
	// Moyens mobile + cod actifs pour les cas payeur.
	for _, q := range []string{
		`INSERT OR REPLACE INTO settings(nom,valeur) VALUES('paiement_orange_on','1')`,
		`INSERT OR REPLACE INTO settings(nom,valeur) VALUES('paiement_cod_on','1')`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
	return app, auth.CSRFToken("test-secret", "anon")
}

func posteCommande(app *App, csrf string, form url.Values, key string) *httptest.ResponseRecorder {
	form.Set("csrf", csrf)
	req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if key != "" {
		req.Header.Set("Idempotency-Key", key)
	}
	rec := httptest.NewRecorder()
	app.ApiCommande(rec, req)
	return rec
}

func baseForm() url.Values {
	return url.Values{
		"items": {"test-produitx2"},
		"moyen": {"cod"},
		"nom":   {"Awa"},
		"tel":   {"0700000000"},
	}
}

// T4 — POST /api/commande enrichi : client, livraison, numéro payeur.
// Frais calculés serveur, réponse {"ok":true,"commande":N}, idempotence inchangée.
func TestApiCommandeLivraison(t *testing.T) {
	zoneID := func(app *App, ordre int) int64 {
		var id int64
		if err := app.DB.QueryRow(`SELECT id FROM livraison_zones WHERE ordre=?`, ordre).Scan(&id); err != nil {
			t.Fatal(err)
		}
		return id
	}

	// sans nom → 400 "nom requis".
	app, csrf := testAppCommandeLivraison(t)
	f := baseForm()
	f.Del("nom")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "nom requis") {
		t.Fatalf("sans nom : code %d, body %q", rec.Code, rec.Body.String())
	}

	// sans tel → 400 "téléphone requis".
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Del("tel")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "téléphone requis") {
		t.Fatalf("sans tel : code %d, body %q", rec.Code, rec.Body.String())
	}

	// moyen orange sans numero_payeur → 400.
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("moyen", "orange")
	f.Set("livraison_id", itoa(int(zoneID(app, 1))))
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "numéro de paiement invalide") {
		t.Fatalf("orange sans payeur : code %d, body %q", rec.Code, rec.Body.String())
	}

	// moyen orange avec numero_payeur → 200, colonne persistée (normalisée).
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("moyen", "orange")
	f.Set("numero_payeur", "07 07-08 09 10")
	f.Set("livraison_id", itoa(int(zoneID(app, 1))))
	f.Set("adresse", "Plateau")
	rec := posteCommande(app, csrf, f, "")
	if rec.Code != 200 || !strings.Contains(rec.Body.String(), `"ok":true`) {
		t.Fatalf("orange avec payeur : code %d, body %q", rec.Code, rec.Body.String())
	}
	var payeur string
	_ = app.DB.QueryRow(`SELECT numero_payeur FROM orders ORDER BY id DESC LIMIT 1`).Scan(&payeur)
	if payeur != "0707080910" {
		t.Fatalf("numero_payeur persisté=%q, attendu %q (normalisé)", payeur, "0707080910")
	}

	// numero_payeur trop court → 400.
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("moyen", "mtn")
	f.Set("numero_payeur", "123")
	f.Set("livraison_id", itoa(int(zoneID(app, 1))))
	if _, err := app.DB.Exec(`INSERT OR REPLACE INTO settings(nom,valeur) VALUES('paiement_mtn_on','1')`); err != nil {
		t.Fatal(err)
	}
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "numéro de paiement invalide") {
		t.Fatalf("payeur court : code %d, body %q", rec.Code, rec.Body.String())
	}

	// cod sans numero_payeur → 200.
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("livraison_id", itoa(int(zoneID(app, 1))))
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 200 {
		t.Fatalf("cod sans payeur : code %d, body %q", rec.Code, rec.Body.String())
	}

	// Zone retrait : adresse non requise.
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("livraison_id", itoa(int(zoneID(app, 1))))
	f.Del("adresse")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 200 {
		t.Fatalf("retrait sans adresse : code %d, body %q", rec.Code, rec.Body.String())
	}

	// Zone livraison sans adresse → 400 "adresse requise".
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("livraison_id", itoa(int(zoneID(app, 2))))
	f.Del("adresse")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "adresse requise") {
		t.Fatalf("livraison sans adresse : code %d, body %q", rec.Code, rec.Body.String())
	}

	// Zone Abidjan + total-remise >= seuil → livraison_frais=0.
	app, csrf = testAppCommandeLivraison(t)
	abidjan := zoneID(app, 2)
	if _, err := app.DB.Exec(`UPDATE livraison_zones SET frais=2000, seuil_franco=10000 WHERE id=?`, abidjan); err != nil {
		t.Fatal(err)
	}
	f = baseForm() // 2 x 10000 = 20000 >= 10000 → franco.
	f.Set("livraison_id", itoa(int(abidjan)))
	f.Set("adresse", "Cocody")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 200 {
		t.Fatalf("abidjan franco : code %d, body %q", rec.Code, rec.Body.String())
	}
	var frais, total int64
	_ = app.DB.QueryRow(`SELECT livraison_frais, total FROM orders ORDER BY id DESC LIMIT 1`).Scan(&frais, &total)
	if frais != 0 || total != 20000 {
		t.Fatalf("franco : frais=%d total=%d, attendu 0/20000", frais, total)
	}

	// Zone expédition → frais ajoutés au total, persistés.
	app, csrf = testAppCommandeLivraison(t)
	exp := zoneID(app, 3)
	if _, err := app.DB.Exec(`UPDATE livraison_zones SET frais=3000, seuil_franco=0 WHERE id=?`, exp); err != nil {
		t.Fatal(err)
	}
	f = baseForm()
	f.Set("livraison_id", itoa(int(exp)))
	f.Set("adresse", "Bouaké")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 200 {
		t.Fatalf("expédition : code %d, body %q", rec.Code, rec.Body.String())
	}
	_ = app.DB.QueryRow(`SELECT livraison_frais, total FROM orders ORDER BY id DESC LIMIT 1`).Scan(&frais, &total)
	if frais != 3000 || total != 23000 {
		t.Fatalf("expédition : frais=%d total=%d, attendu 3000/23000", frais, total)
	}

	// Zone désactivée → 400 "mode de livraison indisponible".
	app, csrf = testAppCommandeLivraison(t)
	desact := zoneID(app, 3)
	if _, err := app.DB.Exec(`UPDATE livraison_zones SET actif=0 WHERE id=?`, desact); err != nil {
		t.Fatal(err)
	}
	f = baseForm()
	f.Set("livraison_id", itoa(int(desact)))
	f.Set("adresse", "Bouaké")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "mode de livraison indisponible") {
		t.Fatalf("zone désactivée : code %d, body %q", rec.Code, rec.Body.String())
	}

	// Zone inconnue → 400 "mode de livraison indisponible".
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("livraison_id", "999999")
	f.Set("adresse", "X")
	if rec := posteCommande(app, csrf, f, ""); rec.Code != 400 || !strings.Contains(rec.Body.String(), "mode de livraison indisponible") {
		t.Fatalf("zone inconnue : code %d, body %q", rec.Code, rec.Body.String())
	}

	// Réponse contient "commande":<id> + idempotence (rejeu = même id, pas de doublon).
	app, csrf = testAppCommandeLivraison(t)
	f = baseForm()
	f.Set("livraison_id", itoa(int(zoneID(app, 1))))
	rec = posteCommande(app, csrf, f, "key-liv")
	if rec.Code != 200 {
		t.Fatalf("commande : code %d, body %q", rec.Code, rec.Body.String())
	}
	var payload struct {
		OK       bool  `json:"ok"`
		Commande int64 `json:"commande"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &payload); err != nil || !payload.OK || payload.Commande == 0 {
		t.Fatalf("réponse sans commande : %q err=%v", rec.Body.String(), err)
	}
	rec2 := posteCommande(app, csrf, f, "key-liv")
	var payload2 struct {
		OK       bool  `json:"ok"`
		Commande int64 `json:"commande"`
	}
	if err := json.Unmarshal(rec2.Body.Bytes(), &payload2); err != nil || payload2.Commande != payload.Commande {
		t.Fatalf("rejeu : %q vs %q", rec.Body.String(), rec2.Body.String())
	}
	var n int
	_ = app.DB.QueryRow(`SELECT COUNT(*) FROM orders`).Scan(&n)
	if n != 1 {
		t.Fatalf("idempotence : %d commandes, attendu 1", n)
	}
}
