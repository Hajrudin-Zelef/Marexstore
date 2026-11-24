package handlers

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/Hajrudin-Zelef/Marexstore/internal/design"
)

// AdminDesignDraft : sauvegarde explicite du brouillon (bouton Sauvegarder).
// GET  → dernier brouillon (ou vide) ;
// POST → valide + enregistre, SANS publier (le site public ne change pas).
func (a *App) AdminDesignDraft(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		raw, version, creeLe, ok, err := design.LoadDraft(a.DB)
		if err != nil {
			http.Error(w, "lecture impossible", http.StatusInternalServerError)
			return
		}
		if !ok {
			a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "brouillon": nil})
			return
		}
		a.writeJSON(w, http.StatusOK, map[string]any{
			"ok": true, "version": version, "cree_le": creeLe, "design": json.RawMessage(raw),
		})
	case http.MethodPost:
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		raw, err := io.ReadAll(io.LimitReader(r.Body, design.MaxDesignBytes+1))
		if err != nil {
			http.Error(w, "lecture impossible", http.StatusBadRequest)
			return
		}
		if len(raw) > design.MaxDesignBytes {
			http.Error(w, "brouillon trop volumineux", http.StatusRequestEntityTooLarge)
			return
		}
		version, err := design.SaveDraft(a.DB, raw, a.Cfg.AdminUser)
		if err != nil {
			http.Error(w, "brouillon refusé : "+err.Error(), http.StatusBadRequest)
			return
		}
		a.auditAdmin(r, "design.brouillon", "sauvegarde v"+itoa(version))
		a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "version": version})
	default:
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// itoa : petit utilitaire local (évite d'importer strconv dans ce fichier).
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [20]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
