package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// Double verrou de publication : mot de passe admin → jeton court → publication.
// Sans jeton valide, aucune publication (même avec Basic Auth + CSRF).
func TestPublicationExigeJetonConfirme(t *testing.T) {
	app, _ := testAppDesign(t)

	post := func(url string, body string, extra map[string]string) *httptest.ResponseRecorder {
		req := httptest.NewRequest(http.MethodPost, url, strings.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-CSRF", app.adminCSRF())
		for k, v := range extra {
			req.Header.Set(k, v)
		}
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		switch url {
		case "/admin/api/publish/confirm":
			app.AdminPublishConfirm(rec, req)
		default:
			app.AdminDesignAPI(rec, req)
		}
		return rec
	}

	design := `{"front":{"accent":"#2E5136"},"Tx":{"bienvenue":"Test"}}`

	// 1. sans jeton → refus
	rec := post("/admin/api/design", design, nil)
	if rec.Code != http.StatusForbidden {
		t.Fatalf("publication sans jeton : %d, attendu 403", rec.Code)
	}

	// 2. mauvais mot de passe → pas de jeton
	rec = post("/admin/api/publish/confirm", `{"password":"faux"}`, nil)
	if rec.Code != http.StatusUnauthorized {
		t.Fatalf("mauvais mot de passe : %d, attendu 401", rec.Code)
	}

	// 3. bon mot de passe → jeton
	rec = post("/admin/api/publish/confirm", `{"password":"test1234"}`, nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("confirmation : %d, attendu 200 (%s)", rec.Code, rec.Body.String())
	}
	var out struct {
		Token string `json:"token"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &out); err != nil || out.Token == "" {
		t.Fatalf("jeton absent : %s", rec.Body.String())
	}

	// 4. jeton valide → publication OK
	rec = post("/admin/api/design", design, map[string]string{"X-Publish-Confirm": out.Token})
	if rec.Code != http.StatusOK {
		t.Fatalf("publication avec jeton : %d (%s)", rec.Code, rec.Body.String())
	}

	// 5. jeton bidon → refus
	rec = post("/admin/api/design", design, map[string]string{"X-Publish-Confirm": "bidon"})
	if rec.Code != http.StatusForbidden {
		t.Fatalf("jeton bidon : %d, attendu 403", rec.Code)
	}
}
