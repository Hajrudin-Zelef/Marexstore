package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func getLivraison(t *testing.T, app *App, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminLivraison(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s : %d", path, rec.Code)
	}
	return rec.Body.String()
}

func postLivraison(t *testing.T, app *App, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	form.Set("csrf", app.adminCSRF())
	req := httptest.NewRequest(http.MethodPost, "/admin/livraison", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminLivraison(rec, req)
	return rec
}

// Onglets : ?vue=zones (défaut, liste + ajout) vs ?vue=generaux (globaux).
// Chaque onglet n'affiche que son contenu ; les POST reviennent sur leur onglet.
func TestAdminLivraisonOnglets(t *testing.T) {
	app, _ := testAppDesign(t)

	// Onglet zones par défaut : liste visible, globals absents.
	body := getLivraison(t, app, "/admin/livraison")
	for _, want := range []string{`vue=zones`, "Retrait boutique", `value="maj"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("onglet zones : %q manquant", want)
		}
	}
	if strings.Contains(body, `name="seuil_abidjan"`) {
		t.Fatal("onglet zones : formulaire globals présent (fuite)")
	}

	// Onglet généraux : globals visibles, lignes zones absentes.
	body = getLivraison(t, app, "/admin/livraison?vue=generaux")
	for _, want := range []string{`vue=generaux`, `name="seuil_abidjan"`} {
		if !strings.Contains(body, want) {
			t.Fatalf("onglet generaux : %q manquant", want)
		}
	}
	if strings.Contains(body, `value="maj"`) {
		t.Fatal("onglet generaux : lignes zones présentes (fuite)")
	}

	// POST zone → revient sur l'onglet zones.
	rec := postLivraison(t, app, url.Values{
		"action": {"ajouter"}, "nom": {"Zone onglets"}, "type": {"livraison"},
		"frais": {"0"}, "seuil_franco": {"0"}, "delai": {""}, "ordre": {"0"},
	})
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "vue=zones") {
		t.Fatalf("redirect ajout : %q devrait revenir sur vue=zones", loc)
	}

	// POST globals → revient sur l'onglet généraux.
	rec = postLivraison(t, app, url.Values{
		"action": {"globals"}, "seuil_abidjan": {"80000"}, "retrait_adresse": {"Plateau"},
	})
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "vue=generaux") {
		t.Fatalf("redirect globals : %q devrait revenir sur vue=generaux", loc)
	}
}

// Page Livraison : liste les 3 zones seed, ajout/maj/toggle/suppression,
// réglages globaux persistés sans écraser les autres settings.
func TestAdminLivraisonCRUD(t *testing.T) {
	app, database := testAppDesign(t)

	// GET : les 3 zones seed sont listées.
	body := getLivraison(t, app, "/admin/livraison")
	for _, want := range []string{"Retrait boutique", "Livraison Abidjan", "Expédition intérieur"} {
		if !strings.Contains(body, want) {
			t.Fatalf("GET /admin/livraison : %q manquant", want)
		}
	}

	// POST ajout zone → 303.
	rec := postLivraison(t, app, url.Values{
		"action": {"ajouter"}, "nom": {"Zone test"}, "type": {"livraison"},
		"frais": {"1500"}, "seuil_franco": {"50000"}, "delai": {"48h"}, "ordre": {"9"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST ajouter : %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); !strings.HasPrefix(loc, "/admin/livraison") {
		t.Fatalf("redirect ajout : %q", loc)
	}
	var id int64
	if err := database.QueryRow(`SELECT id FROM livraison_zones WHERE nom='Zone test'`).Scan(&id); err != nil {
		t.Fatalf("zone ajoutée introuvable : %v", err)
	}

	// POST maj montants → 303 + persisté.
	rec = postLivraison(t, app, url.Values{
		"action": {"maj"}, "id": {itoa(int(id))},
		"nom": {"Zone test"}, "type": {"livraison"},
		"frais": {"2000"}, "seuil_franco": {"50000"}, "delai": {"48h"}, "ordre": {"9"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST maj : %d", rec.Code)
	}
	var frais int64
	if err := database.QueryRow(`SELECT frais FROM livraison_zones WHERE id=?`, id).Scan(&frais); err != nil || frais != 2000 {
		t.Fatalf("maj frais=%d err=%v", frais, err)
	}

	// POST toggle actif → zone masquée (actif=0).
	rec = postLivraison(t, app, url.Values{"action": {"toggle"}, "id": {itoa(int(id))}})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST toggle : %d", rec.Code)
	}
	var actif int64
	if err := database.QueryRow(`SELECT actif FROM livraison_zones WHERE id=?`, id).Scan(&actif); err != nil || actif != 0 {
		t.Fatalf("toggle actif=%d err=%v", actif, err)
	}

	// POST suppression → 303 + ligne disparue.
	rec = postLivraison(t, app, url.Values{"action": {"supprimer"}, "id": {itoa(int(id))}})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST supprimer : %d", rec.Code)
	}
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM livraison_zones WHERE id=?`, id).Scan(&n); err != nil || n != 0 {
		t.Fatalf("zone non supprimée : n=%d err=%v", n, err)
	}

	// POST globals → persisté, autres settings intacts.
	if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES('whatsapp','2250705087507')`); err != nil {
		t.Fatal(err)
	}
	rec = postLivraison(t, app, url.Values{
		"action": {"globals"}, "seuil_abidjan": {"75000"}, "retrait_adresse": {"Cocody, Abidjan"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST globals : %d", rec.Code)
	}
	var seuil, adresse, wa string
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='livraison_seuil_abidjan'`).Scan(&seuil)
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='livraison_retrait_adresse'`).Scan(&adresse)
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='whatsapp'`).Scan(&wa)
	if seuil != "75000" || adresse != "Cocody, Abidjan" {
		t.Fatalf("globals non persistés : seuil=%q adresse=%q", seuil, adresse)
	}
	if wa != "2250705087507" {
		t.Fatalf("globals a écrasé whatsapp : %q", wa)
	}

	// Validation : type inconnu et nom vide → 400.
	rec = postLivraison(t, app, url.Values{
		"action": {"ajouter"}, "nom": {"X"}, "type": {"teleportation"},
		"frais": {"0"}, "seuil_franco": {"0"}, "delai": {""}, "ordre": {"0"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("type invalide : %d, attendu 400", rec.Code)
	}
	rec = postLivraison(t, app, url.Values{
		"action": {"ajouter"}, "nom": {""}, "type": {"livraison"},
		"frais": {"0"}, "seuil_franco": {"0"}, "delai": {""}, "ordre": {"0"},
	})
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("nom vide : %d, attendu 400", rec.Code)
	}
}
