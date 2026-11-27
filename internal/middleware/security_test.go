package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// Deux connexions du même client (même IP, ports éphémères différents)
// partagent le même seau : une rafale épuise les jetons et finit en 429.
func TestRateLimitPartageParIP(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := RateLimit(1, 2)(ok)
	var codes []int
	for i := 0; i < 5; i++ {
		req := httptest.NewRequest(http.MethodGet, "/admin", nil)
		// Même IP, port distant différent à chaque connexion.
		req.RemoteAddr = "10.0.0.7:5000" + string(rune('0'+i))
		rec := httptest.NewRecorder()
		h.ServeHTTP(rec, req)
		codes = append(codes, rec.Code)
	}
	n429 := 0
	for _, c := range codes {
		if c == http.StatusTooManyRequests {
			n429++
		}
	}
	if n429 == 0 {
		t.Fatalf("aucun 429 : chaque connexion a eu son propre seau (codes %v)", codes)
	}
}

// X-Real-IP (posé par nginx) sert de clé quand X-Forwarded-For est absent.
func TestRateLimitCleXRealIP(t *testing.T) {
	ok := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(200) })
	h := RateLimit(1000, 1)(ok)
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Real-IP", "10.9.9.9")
	req.RemoteAddr = "10.0.0.1:1111"
	rec := httptest.NewRecorder()
	h.ServeHTTP(rec, req)
	if rec.Code != 200 {
		t.Fatalf("1er appel : %d", rec.Code)
	}
	rec2 := httptest.NewRecorder()
	req2 := httptest.NewRequest(http.MethodGet, "/", nil)
	req2.Header.Set("X-Real-IP", "10.9.9.9")
	req2.RemoteAddr = "10.0.0.2:2222"
	h.ServeHTTP(rec2, req2)
	if rec2.Code != http.StatusTooManyRequests {
		t.Fatalf("2e appel même X-Real-IP : %d, attendu 429 (clé ignorée ?)", rec2.Code)
	}
}
