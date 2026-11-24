package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// /api/paiements : ne renvoie QUE les moyens activés, avec leur numéro
// marchand. Aucune auth (le client doit les voir au panier), aucune fuite
// des moyens désactivés ni des autres réglages.
func TestApiPaiements(t *testing.T) {
	app, database := testAppDesign(t)

	set := func(k, v string) {
		if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES(?,?)
			ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`, k, v); err != nil {
			t.Fatal(err)
		}
	}
	set("paiement_orange_on", "1")
	set("paiement_orange_numero", "0705087507")
	set("paiement_wave_on", "1")
	set("paiement_wave_numero", "0102030405")
	set("paiement_mtn_on", "0")
	set("paiement_mtn_numero", "0500000000")
	set("paiement_cod_on", "1")
	set("paiement_moov_on", "0")

	req := httptest.NewRequest(http.MethodGet, "/api/paiements", nil)
	rec := httptest.NewRecorder()
	app.ApiPaiements(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("api paiements : %d", rec.Code)
	}
	var out struct {
		OK       bool `json:"ok"`
		Paiement []struct {
			Cle     string `json:"cle"`
			Libelle string `json:"libelle"`
			Numero  string `json:"numero"`
		} `json:"paiements"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatalf("JSON illisible : %v (%s)", err, rec.Body.String())
	}
	if !out.OK {
		t.Fatal("ok=false")
	}
	cles := map[string]string{}
	for _, p := range out.Paiement {
		cles[p.Cle] = p.Numero
	}
	if len(out.Paiement) != 3 {
		t.Fatalf("%d moyens renvoyés, attendu 3 (orange, wave, cod) : %v", len(out.Paiement), cles)
	}
	if cles["orange"] != "0705087507" || cles["wave"] != "0102030405" {
		t.Fatalf("numéros incorrects : %v", cles)
	}
	if _, ok := cles["mtn"]; ok {
		t.Fatal("un moyen désactivé (mtn) est exposé")
	}
	if num, ok := cles["cod"]; ok && num != "" {
		t.Fatalf("paiement à la livraison ne doit pas avoir de numéro : %q", num)
	}
}

// Aucun moyen actif : réponse vide, pas d'erreur.
func TestApiPaiementsAucunActif(t *testing.T) {
	app, _ := testAppDesign(t)
	req := httptest.NewRequest(http.MethodGet, "/api/paiements", nil)
	rec := httptest.NewRecorder()
	app.ApiPaiements(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("api paiements vide : %d", rec.Code)
	}
	var out struct {
		Paiement []any `json:"paiements"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
		t.Fatal(err)
	}
	if len(out.Paiement) != 0 {
		t.Fatalf("%d moyens alors qu'aucun n'est actif", len(out.Paiement))
	}
}
