package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Réglages paiement : l'admin enregistre les moyens activés + numéros
// marchands, et les retrouve cochés/remplis au rechargement de la page.
func TestAdminReglagesPaiement(t *testing.T) {
	app, _ := testAppDesign(t)

	form := url.Values{}
	form.Set("csrf", app.adminCSRF())
	form.Set("paiement_orange_on", "1")
	form.Set("paiement_orange_numero", "0705087507")
	form.Set("paiement_mtn_on", "1")
	form.Set("paiement_mtn_numero", "0501234567")
	form.Set("paiement_cod_on", "1")
	// wave / moov volontairement décochés

	req := httptest.NewRequest(http.MethodPost, "/admin/reglages?vue=paiement", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminReglages(rec, req)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("enregistrement réglages : %d", rec.Code)
	}

	// Relecture : GET affiche les valeurs enregistrées.
	req2 := httptest.NewRequest(http.MethodGet, "/admin/reglages?vue=paiement", nil)
	req2.SetBasicAuth("admin", "test1234")
	rec2 := httptest.NewRecorder()
	app.AdminReglages(rec2, req2)
	body := rec2.Body.String()

	for _, want := range []string{
		"paiement_orange_numero\" value=\"0705087507\"",
		"paiement_mtn_numero\" value=\"0501234567\"",
		"Orange Money", "MTN Mobile Money", "Wave", "Moov Money", "Paiement à la livraison",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("réglages : %q absent du rendu", want)
		}
	}

	// Les cases : orange/mtn/cod cochées, wave/moov non.
	coche := func(name string) bool {
		i := strings.Index(body, `name="`+name+`_on"`)
		if i < 0 {
			t.Fatalf("case %s absente", name)
		}
		fin := i + 200
		if fin > len(body) {
			fin = len(body)
		}
		return strings.Contains(body[i:fin], "checked")
	}
	for _, n := range []string{"paiement_orange", "paiement_mtn", "paiement_cod"} {
		if !coche(n) {
			t.Fatalf("%s devrait être cochée", n)
		}
	}
	for _, n := range []string{"paiement_wave", "paiement_moov"} {
		if coche(n) {
			t.Fatalf("%s ne devrait pas être cochée", n)
		}
	}
}
