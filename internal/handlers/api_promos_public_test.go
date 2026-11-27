package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/config"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
)

// GET /api/promos (public, sans auth) : uniquement les codes utilisables
// (actifs, non expirés). Ni les inactifs, ni les expirés, ni usage_count.
func TestApiPromosPublic(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database, Cfg: config.Config{SessionSecret: "test-secret"}}
	app.ensurePromos()
	for _, q := range []string{
		`INSERT INTO promos(code,type,valeur,description,actif) VALUES('ACTIF','fixe',1000,'Moins 1000',1)`,
		`INSERT INTO promos(code,type,valeur,expire,actif) VALUES('VIEUX','fixe',500,'2000-01-01',1)`,
		`INSERT INTO promos(code,type,valeur,actif) VALUES('OFF','fixe',500,0)`,
		`INSERT INTO promos(code,type,valeur,expire,limite,actif) VALUES('FUTUR','pourcentage',20,'2099-12-31',10,1)`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatal(err)
		}
	}
	req := httptest.NewRequest(http.MethodGet, "/api/promos", nil)
	rec := httptest.NewRecorder()
	app.ApiPromos(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /api/promos : %d, body %q", rec.Code, rec.Body.String())
	}
	var out struct {
		Ok     bool `json:"ok"`
		Promos []struct {
			Code   string `json:"code"`
			Type   string `json:"type"`
			Valeur int64  `json:"valeur"`
		} `json:"promos"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || !out.Ok {
		t.Fatalf("JSON invalide : %v %q", err, rec.Body.String())
	}
	got := map[string]bool{}
	for _, p := range out.Promos {
		got[p.Code] = true
	}
	for _, want := range []string{"ACTIF", "FUTUR"} {
		if !got[want] {
			t.Fatalf("promo utilisable absente : %s (reçu %v)", want, got)
		}
	}
	for _, notWant := range []string{"VIEUX", "OFF"} {
		if got[notWant] {
			t.Fatalf("promo inutilisable exposée : %s", notWant)
		}
	}
	body := rec.Body.String()
	if strings.Contains(body, "usage_count") || strings.Contains(body, "limite") {
		t.Fatal("compteurs internes exposés au public")
	}
}
