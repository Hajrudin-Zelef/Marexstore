package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// /api/livraison : ne renvoie QUE les zones actives (triées par ordre), avec
// frais et seuil franco (le client doit afficher le total). Aucune auth (le
// client choisit au tunnel d'achat), aucune fuite des champs internes
// (ordre, actif) ni des autres réglages. Aucune zone active → zones:[].
func TestApiLivraison(t *testing.T) {
	app, database := testAppDesign(t)

	if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES('livraison_seuil_abidjan','75000')
		ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES('livraison_retrait_adresse','Cocody, Abidjan')
		ON CONFLICT(nom) DO UPDATE SET valeur=excluded.valeur`); err != nil {
		t.Fatal(err)
	}

	get := func() (int, map[string]any, string) {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/api/livraison", nil)
		rec := httptest.NewRecorder()
		app.ApiLivraison(rec, req)
		var out map[string]any
		if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil {
			t.Fatalf("JSON illisible : %v (%s)", err, rec.Body.String())
		}
		return rec.Code, out, rec.Body.String()
	}

	// 3 zones seed actives → 3 entrées triées par ordre, sans champ interne.
	code, out, raw := get()
	if code != http.StatusOK {
		t.Fatalf("api livraison : %d", code)
	}
	if out["ok"] != true {
		t.Fatalf("ok=false : %s", raw)
	}
	zones, _ := out["zones"].([]any)
	if len(zones) != 3 {
		t.Fatalf("%d zones renvoyées, attendu 3 : %s", len(zones), raw)
	}
	for i, z := range zones {
		m := z.(map[string]any)
		for _, k := range []string{"id", "nom", "type", "frais", "seuil_franco", "delai"} {
			if _, ok := m[k]; !ok {
				t.Fatalf("zone %d : champ %q manquant : %s", i, k, raw)
			}
		}
		for _, k := range []string{"ordre", "actif"} {
			if _, ok := m[k]; ok {
				t.Fatalf("zone %d : champ interne %q exposé : %s", i, k, raw)
			}
		}
	}
	noms := []string{
		zones[0].(map[string]any)["nom"].(string),
		zones[1].(map[string]any)["nom"].(string),
		zones[2].(map[string]any)["nom"].(string),
	}
	if !strings.Contains(noms[0], "Retrait") || !strings.Contains(noms[1], "Abidjan") || !strings.Contains(noms[2], "intérieur") {
		t.Fatalf("ordre inattendu : %v", noms)
	}
	if out["seuil_abidjan"] != float64(75000) {
		t.Fatalf("seuil_abidjan=%v, attendu 75000 : %s", out["seuil_abidjan"], raw)
	}
	if out["retrait_adresse"] != "Cocody, Abidjan" {
		t.Fatalf("retrait_adresse=%v : %s", out["retrait_adresse"], raw)
	}

	// Zone désactivée → absente.
	if _, err := database.Exec(`UPDATE livraison_zones SET actif=0 WHERE ordre=3`); err != nil {
		t.Fatal(err)
	}
	_, out, raw = get()
	zones, _ = out["zones"].([]any)
	if len(zones) != 2 {
		t.Fatalf("%d zones après désactivation, attendu 2 : %s", len(zones), raw)
	}

	// Aucune active → {ok:true, zones:[]}, 200.
	if _, err := database.Exec(`UPDATE livraison_zones SET actif=0`); err != nil {
		t.Fatal(err)
	}
	code, out, raw = get()
	if code != http.StatusOK {
		t.Fatalf("api livraison vide : %d", code)
	}
	if out["ok"] != true {
		t.Fatalf("ok=false quand vide : %s", raw)
	}
	zones, _ = out["zones"].([]any)
	if len(zones) != 0 {
		t.Fatalf("%d zones alors qu'aucune n'est active : %s", len(zones), raw)
	}
}
