package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

func getReglages(t *testing.T, app *App, path string) string {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, path, nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminReglages(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET %s : %d", path, rec.Code)
	}
	return rec.Body.String()
}

func postReglages(t *testing.T, app *App, path string, form url.Values) *httptest.ResponseRecorder {
	t.Helper()
	form.Set("csrf", app.adminCSRF())
	req := httptest.NewRequest(http.MethodPost, path, strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminReglages(rec, req)
	return rec
}

// Chaque onglet n'affiche que sa section (plus les onglets eux-mêmes).
func TestAdminReglagesOnglets(t *testing.T) {
	app, _ := testAppDesign(t)

	paiement := getReglages(t, app, "/admin/reglages?vue=paiement")
	for _, want := range []string{`name="paiement_orange_on"`, "Boutique", "Contact", "Paiement", "Notifications"} {
		if !strings.Contains(paiement, want) {
			t.Fatalf("onglet paiement : %q manquant", want)
		}
	}
	for _, notWant := range []string{`name="maintenance"`, `name="whatsapp"`, `name="notif_numero"`} {
		if strings.Contains(paiement, notWant) {
			t.Fatalf("onglet paiement : %q devrait être dans un autre onglet", notWant)
		}
	}

	contact := getReglages(t, app, "/admin/reglages?vue=contact")
	if !strings.Contains(contact, `name="whatsapp"`) {
		t.Fatal("onglet contact : champ whatsapp manquant")
	}
	if strings.Contains(contact, `name="paiement_orange_on"`) {
		t.Fatal("onglet contact : section paiement présente (fuite)")
	}

	notif := getReglages(t, app, "/admin/reglages?vue=notifications")
	if !strings.Contains(notif, `name="notif_numero"`) {
		t.Fatal("onglet notifications : champ manquant")
	}
	// Le champ doit être DANS un formulaire (sinon le navigateur ne
	// l'envoie jamais — bug silencieux de l'ancien écran unique).
	if f := strings.Index(notif, "<form"); f < 0 || !strings.Contains(notif[f:], `name="notif_numero"`) {
		t.Fatal("onglet notifications : notif_numero hors formulaire")
	}

	defaut := getReglages(t, app, "/admin/reglages")
	if !strings.Contains(defaut, `name="maintenance"`) {
		t.Fatal("sans vue : l'onglet boutique (défaut) devrait s'afficher")
	}

	inconnu := getReglages(t, app, "/admin/reglages?vue=nimporte")
	if !strings.Contains(inconnu, `name="maintenance"`) {
		t.Fatal("vue inconnue : repli sur boutique attendu")
	}
}

// Chaque moyen a son logo (fichiers static/logos/paiement-*).
func TestAdminReglagesLogosPaiement(t *testing.T) {
	app, _ := testAppDesign(t)
	body := getReglages(t, app, "/admin/reglages?vue=paiement")
	for _, want := range []string{
		`src="/static/logos/paiement-orange.png"`,
		`src="/static/logos/paiement-mtn.jpg"`,
		`src="/static/logos/paiement-wave.jpg"`,
		`src="/static/logos/paiement-moov.png"`,
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("logo paiement manquant : %s", want)
		}
	}
}

// POST par onglet : seules les clés de l'onglet sont touchées.
func TestAdminReglagesPostParOnglet(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO settings(nom,valeur) VALUES('whatsapp','2250705087507'),('maintenance','0')`); err != nil {
		t.Fatal(err)
	}
	rec := postReglages(t, app, "/admin/reglages?vue=paiement", url.Values{
		"paiement_orange_on": {"1"}, "paiement_orange_numero": {"0700"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST paiement : %d", rec.Code)
	}
	if loc := rec.Header().Get("Location"); !strings.Contains(loc, "vue=paiement") {
		t.Fatalf("redirect : %q devrait revenir sur l'onglet paiement", loc)
	}
	var wa, maint, num string
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='whatsapp'`).Scan(&wa)
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='maintenance'`).Scan(&maint)
	_ = database.QueryRow(`SELECT valeur FROM settings WHERE nom='paiement_orange_numero'`).Scan(&num)
	if wa != "2250705087507" || maint != "0" {
		t.Fatalf("onglet paiement a écrasé les autres sections : whatsapp=%q maintenance=%q", wa, maint)
	}
	if num != "0700" {
		t.Fatalf("paiement non enregistré : %q", num)
	}
}

// Le numéro de notification persiste via son onglet.
func TestAdminReglagesNotifPersiste(t *testing.T) {
	app, database := testAppDesign(t)
	rec := postReglages(t, app, "/admin/reglages?vue=notifications", url.Values{
		"notif_numero": {"2250102030405"},
	})
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("POST notifications : %d", rec.Code)
	}
	var v string
	if err := database.QueryRow(`SELECT valeur FROM settings WHERE nom='notif_numero'`).Scan(&v); err != nil || v != "2250102030405" {
		t.Fatalf("notif_numero=%q err=%v", v, err)
	}
	body := getReglages(t, app, "/admin/reglages?vue=notifications")
	if !strings.Contains(body, `name="notif_numero" value="2250102030405"`) {
		t.Fatal("onglet notifications : valeur non réaffichée")
	}
}
