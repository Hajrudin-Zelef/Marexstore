package main

import (
	"net/http"
	"strings"
)

// adminBodyMax : plafond du corps des POST admin (API design/fichiers,
// fiche client). 4 Mo > MaxFichierBytes (2 Mo) + enveloppe JSON.
const adminBodyMax = 4 << 20

func bodyLimit(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		p := r.URL.Path
		if r.Method == http.MethodPost && (strings.HasPrefix(p, "/admin/api/") ||
			strings.HasPrefix(p, "/admin/design") || strings.HasPrefix(p, "/admin/clients/")) {
			r.Body = http.MaxBytesReader(w, r.Body, adminBodyMax)
		}
		next.ServeHTTP(w, r)
	})
}
