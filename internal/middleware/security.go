package middleware

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"net/http"
	"sync"
	"time"
)

// RequestID ajoute X-Request-ID pour corréler logs.
func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		b := make([]byte, 8)
		_, _ = rand.Read(b)
		id := hex.EncodeToString(b)
		w.Header().Set("X-Request-ID", id)
		next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), "reqid", id)))
	})
}

// SecurityHeaders : durcissement sans casser htmx/images/fonts.
// CSP en Report-Only par défaut (passer en enforce via CSP_REPORT_ONLY=0 quand prêt).
//   - script-src 'self' strict : aucun script inline dans le HTML (theme.js externe,
//     JSON inerte application/json, handlers via addEventListener).
//   - style-src 'unsafe-inline' : les attributs style="" sont partout (rendu Go,
//     htmx, app.js) — les hasher est ingérable et CSS ≠ exécution de script.
//   - form-action 'self' : aucun <form action="mailto:"> (newsletter via JS).
func SecurityHeaders(reportOnly bool) func(http.Handler) http.Handler {
	csp := "default-src 'self'; img-src 'self' data:; script-src 'self'; " +
		"style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; " +
		"connect-src 'self'; frame-ancestors 'none'; base-uri 'self'; form-action 'self'"
	cspOpen := "default-src 'self'; img-src 'self' data:; script-src 'self'; " +
		"style-src 'self' 'unsafe-inline' fonts.googleapis.com; font-src 'self' fonts.gstatic.com; " +
		"connect-src 'self'; frame-ancestors 'self'; base-uri 'self'; form-action 'self'"
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			h := w.Header()
			useCSP := csp
			xfo := "DENY"
			if r.URL.Path == "/admin/design/apercu" {
				// Aperçu du Studio Design : iframé par /admin/design (même
				// origine, page protégée par auth admin). SAMEORIGIN uniquement.
				useCSP = cspOpen
				xfo = "SAMEORIGIN"
			}
			if reportOnly {
				h.Set("Content-Security-Policy-Report-Only", useCSP)
			} else {
				h.Set("Content-Security-Policy", useCSP)
			}
			h.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
			h.Set("X-Frame-Options", xfo)
			h.Set("X-Content-Type-Options", "nosniff")
			h.Set("Referrer-Policy", "strict-origin-when-cross-origin")
			h.Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")
			h.Set("Cross-Origin-Opener-Policy", "same-origin")
			next.ServeHTTP(w, r)
		})
	}
}

// RateLimit : token bucket simplifié par IP, doux (429 + Retry-After).
// Suffisant V1 ; passer à un redis/compteur distribué si multi-instances.
func RateLimit(rps int, burst int) func(http.Handler) http.Handler {
	type bucket struct {
		tokens float64
		last   time.Time
	}
	var mu sync.Mutex
	buckets := map[string]*bucket{}
	// Purge des IP inactives (> 10 min) pour éviter la croissance mémoire.
	go func() {
		for range time.Tick(5 * time.Minute) {
			mu.Lock()
			for ip, b := range buckets {
				if time.Since(b.last) > 10*time.Minute {
					delete(buckets, ip)
				}
			}
			mu.Unlock()
		}
	}()
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ip := r.Header.Get("X-Forwarded-For")
			if ip == "" {
				ip = r.RemoteAddr
			}
			mu.Lock()
			b, ok := buckets[ip]
			if !ok {
				b = &bucket{tokens: float64(burst), last: time.Now()}
				buckets[ip] = b
			}
			now := time.Now()
			elapsed := now.Sub(b.last).Seconds()
			b.tokens += elapsed * float64(rps)
			if b.tokens > float64(burst) {
				b.tokens = float64(burst)
			}
			b.last = now
			if b.tokens < 1 {
				mu.Unlock()
				w.Header().Set("Retry-After", "5")
				http.Error(w, "Trop de requêtes, réessayez dans quelques secondes.", http.StatusTooManyRequests)
				return
			}
			b.tokens--
			mu.Unlock()
			next.ServeHTTP(w, r)
		})
	}
}
