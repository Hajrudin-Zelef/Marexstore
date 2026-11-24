package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Réglages > Notifications : numéro destinataire mémorisé, lien wa.me
// prérempli avec le rapport de stock (rupture + stock faible).
func TestAdminReglagesNotificationsStock(t *testing.T) {
	app, database := testAppDesign(t)
	for _, q := range []string{
		`INSERT INTO products(slug,nom,dispo,stock) VALUES('p-rupt','Toner RUPTURE TEST','rupture de stock',0)`,
		`INSERT INTO products(slug,nom,dispo,stock) VALUES('p-faible','Cable STOCK FAIBLE','en stock',2)`,
	} {
		if _, err := database.Exec(q); err != nil {
			t.Fatal(err)
		}
	}

	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("maintenance", "0")
	form.Set("whatsapp", "2250705087507")
	form.Set("notif_numero", "2250102030405")
	req := httptest.NewRequest(http.MethodPost, "/admin/reglages", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminReglages(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("enregistrement : %d", rec.Code)
	}

	req2 := httptest.NewRequest(http.MethodGet, "/admin/reglages", nil)
	req2.SetBasicAuth("admin", "test1234")
	rec2 := httptest.NewRecorder()
	app.AdminReglages(rec2, req2)
	body := rec2.Body.String()

	for _, want := range []string{
		"Notifications",
		`name="notif_numero" value="2250102030405"`,
		"https://wa.me/2250102030405?text=",
		"Rupture", "stock%20faible",
		"Toner%20RUPTURE%20TEST",
		"Cable%20STOCK%20FAIBLE",
	} {
		if !strings.Contains(body, want) && !strings.Contains(strings.ToLower(body), strings.ToLower(want)) {
			t.Fatalf("notifications : %q absent", want)
		}
	}
}

// Le numéro est saisi par l'admin : il ne doit jamais pouvoir casser l'attribut
// href (guillemet, chevron) — échappement obligatoire.
func TestAdminReglagesNotificationsEchappement(t *testing.T) {
	app, _ := testAppDesign(t)
	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("maintenance", "0")
	form.Set("notif_numero", `225"onmouseover="alert(1)`)
	req := httptest.NewRequest(http.MethodPost, "/admin/reglages", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminReglages(rec, req)

	req2 := httptest.NewRequest(http.MethodGet, "/admin/reglages", nil)
	req2.SetBasicAuth("admin", "test1234")
	rec2 := httptest.NewRecorder()
	app.AdminReglages(rec2, req2)
	// La tentative doit être neutralisée partout : soit échappée en &quot;,
	// soit percent-encodée — jamais un attribut exécutable dans le href.
	body := rec2.Body.String()
	if strings.Contains(body, `href="https://wa.me/225"onmouseover`) ||
		strings.Contains(body, `"onmouseover="alert(1)`) {
		t.Fatal("le numéro casse l'attribut href (non échappé)")
	}
}
