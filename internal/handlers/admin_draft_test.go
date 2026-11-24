package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Brouillon serveur : sauvegarde explicite SANS publication.
// - refusé sans CSRF ;
// - refusé si le design est invalide (même validation que la publication) ;
// - accepté sinon, relu à l'identique, et le design publié n'est pas touché.
func TestAdminDesignBrouillon(t *testing.T) {
	app, _ := testAppDesign(t)

	post := func(body string, csrf bool) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, "/admin/api/design/draft", strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		if csrf {
			req.Header.Set("X-CSRF", app.adminCSRF())
		}
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		app.AdminDesignDraft(rec, req)
		return rec
	}

	// 1. sans CSRF → refus
	if rec := post(`{"front":{}}`, false); rec.Code != http.StatusForbidden {
		t.Fatalf("brouillon sans CSRF : %d, attendu 403", rec.Code)
	}

	// 2. design invalide (script interdit) → refus, rien en base
	if rec := post(`{"Tx":{"x":"<script>alert(1)</script>"}}`, true); rec.Code != http.StatusBadRequest {
		t.Fatalf("brouillon invalide : %d, attendu 400", rec.Code)
	}

	// 3. brouillon valide → 200 + version
	rec := post(`{"front":{"accent":"#2E5136"},"Tx":{"bienvenue":"Brouillon test"}}`, true)
	if rec.Code != http.StatusOK {
		t.Fatalf("brouillon valide : %d (%s)", rec.Code, rec.Body.String())
	}
	var out struct {
		OK      bool `json:"ok"`
		Version int  `json:"version"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || !out.OK || out.Version < 1 {
		t.Fatalf("réponse brouillon inattendue : %s", rec.Body.String())
	}

	// 4. le brouillon se relit, mais le design publié reste vide
	req := httptest.NewRequest(http.MethodGet, "/admin/api/design/draft", nil)
	req.SetBasicAuth("admin", "test1234")
	rrec := httptest.NewRecorder()
	app.AdminDesignDraft(rrec, req)
	if rrec.Code != http.StatusOK {
		t.Fatalf("lecture brouillon : %d", rrec.Code)
	}
	if !strings.Contains(rrec.Body.String(), "Brouillon test") {
		t.Fatalf("brouillon non relu : %s", rrec.Body.String())
	}

	preq := httptest.NewRequest(http.MethodGet, "/admin/api/design", nil)
	preq.SetBasicAuth("admin", "test1234")
	prec := httptest.NewRecorder()
	app.AdminDesignAPI(prec, preq)
	if strings.Contains(prec.Body.String(), "Brouillon test") {
		t.Fatal("le brouillon a fui dans le design publié")
	}
}
