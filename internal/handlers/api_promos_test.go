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

// Un code promo valide (fixe) réduit le total, est persisté sur la commande
// et incrémente usage_count — au lieu d'être silencieusement ignoré.
func TestApiCommandeAppliquePromoFixe(t *testing.T) {
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
	if _, err := database.Exec(`INSERT INTO promos(code,type,valeur,actif) VALUES('REMISE3000','fixe',3000,1)`); err != nil {
		t.Fatal(err)
	}
	form := url.Values{
		"items": {"p1x2"}, "moyen": {"whatsapp"}, "nom": {"T"}, "tel": {"0700"},
		"csrf": {auth.CSRFToken("test-secret", "anon")}, "promo": {"REMISE3000"},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.ApiCommande(rec, req)
	if rec.Code != 200 {
		t.Fatalf("commande avec promo : %d, body %q", rec.Code, rec.Body.String())
	}
	var total, remise int64
	var code string
	if err := database.QueryRow(`SELECT total,remise,promo_code FROM orders ORDER BY id DESC LIMIT 1`).Scan(&total, &remise, &code); err != nil {
		t.Fatalf("colonnes promo manquantes ou illisibles : %v", err)
	}
	if total != 17000 || remise != 3000 || code != "REMISE3000" {
		t.Fatalf("promo non appliquée : total=%d remise=%d code=%q (attendu 17000/3000/REMISE3000)", total, remise, code)
	}
	var usages int64
	_ = database.QueryRow(`SELECT usage_count FROM promos WHERE code='REMISE3000'`).Scan(&usages)
	if usages != 1 {
		t.Fatalf("usage_count=%d, attendu 1", usages)
	}
}

// Table des cas limites : un code invalide n'est jamais une erreur, il est
// ignoré (total plein, usage_count inchangé) ; la remise ne rend jamais le
// total négatif.
func TestApiCommandePromoCasLimites(t *testing.T) {
	cas := []struct {
		nom         string
		setup       string // INSERT promo
		promo       string
		totalAvant  int64  // prix unitaire x1
		totalApres  int64
		remise      int64
		codeEnBase  string
		usagesApres int64
	}{
		{"pourcentage", `INSERT INTO promos(code,type,valeur,actif) VALUES('MOINS20','pourcentage',20,1)`, "moins20", 20000, 16000, 4000, "MOINS20", 1},
		{"inconnu", ``, "NOPE", 20000, 20000, 0, "", -1},
		{"expire", `INSERT INTO promos(code,type,valeur,expire,actif) VALUES('VIEUX','fixe',1000,'2000-01-01',1)`, "VIEUX", 20000, 20000, 0, "", 0},
		{"limite", `INSERT INTO promos(code,type,valeur,usage_count,limite,actif) VALUES('UNIQUE','fixe',1000,1,1,1)`, "UNIQUE", 20000, 20000, 0, "", 1},
		{"inactif", `INSERT INTO promos(code,type,valeur,actif) VALUES('OFF','fixe',1000,0)`, "OFF", 20000, 20000, 0, "", 0},
		{"plafond", `INSERT INTO promos(code,type,valeur,actif) VALUES('GENEREUX','fixe',999999,1)`, "GENEREUX", 20000, 0, 20000, "GENEREUX", 1},
	}
	for _, c := range cas {
		t.Run(c.nom, func(t *testing.T) {
			dir := t.TempDir()
			database, err := db.Open(filepath.Join(dir, "test.db"))
			if err != nil {
				t.Fatal(err)
			}
			defer database.Close()
			if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
				t.Fatal(err)
			}
			if _, err := database.Exec(`INSERT INTO products(slug,nom,prix,stock) VALUES('p1','P1',20000,5)`); err != nil {
				t.Fatal(err)
			}
			app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
			app.ensurePromos()
			if c.setup != "" {
				if _, err := database.Exec(c.setup); err != nil {
					t.Fatal(err)
				}
			}
			form := url.Values{
				"items": {"p1x1"}, "moyen": {"whatsapp"}, "nom": {"T"}, "tel": {"0700"},
				"csrf": {auth.CSRFToken("test-secret", "anon")}, "promo": {c.promo},
			}
			req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
			req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
			rec := httptest.NewRecorder()
			app.ApiCommande(rec, req)
			if rec.Code != 200 {
				t.Fatalf("code %d, body %q", rec.Code, rec.Body.String())
			}
			var total, remise int64
			var code string
			if err := database.QueryRow(`SELECT total,remise,promo_code FROM orders ORDER BY id DESC LIMIT 1`).Scan(&total, &remise, &code); err != nil {
				t.Fatal(err)
			}
			if total != c.totalApres || remise != c.remise || code != c.codeEnBase {
				t.Fatalf("total=%d remise=%d code=%q, attendu %d/%d/%q", total, remise, code, c.totalApres, c.remise, c.codeEnBase)
			}
			if c.usagesApres >= 0 {
				var u int64
				_ = database.QueryRow(`SELECT usage_count FROM promos WHERE code=?`, strings.ToUpper(c.promo)).Scan(&u)
				if u != c.usagesApres {
					t.Fatalf("usage_count=%d, attendu %d", u, c.usagesApres)
				}
			}
		})
	}
}

// Un slug contenant un "x" (ex. lux-pad-ux) ne doit pas casser le découpage
// slugxqté : c'est le DERNIER "x" qui sépare la quantité.
func TestApiCommandeSlugAvecX(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO products(slug,nom,prix,stock) VALUES('lux-pad-ux','Lux Pad',5000,5)`); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
	form := url.Values{
		"items": {"lux-pad-uxx3"}, "moyen": {"whatsapp"}, "nom": {"T"}, "tel": {"0700"},
		"csrf": {auth.CSRFToken("test-secret", "anon")},
	}
	req := httptest.NewRequest(http.MethodPost, "/api/commande", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	app.ApiCommande(rec, req)
	if rec.Code != 200 {
		t.Fatalf("code %d, body %q", rec.Code, rec.Body.String())
	}
	var total, qty int64
	if err := database.QueryRow(`SELECT total FROM orders ORDER BY id DESC LIMIT 1`).Scan(&total); err != nil {
		t.Fatal(err)
	}
	if err := database.QueryRow(`SELECT qty FROM order_items ORDER BY order_id DESC LIMIT 1`).Scan(&qty); err != nil {
		t.Fatal(err)
	}
	if total != 15000 || qty != 3 {
		t.Fatalf("slug à 'x' mal découpé : total=%d qty=%d, attendu 15000/3", total, qty)
	}
}
