package handlers

import (
	"net/http"
	"strings"
)

// MoyenPaiement : un moyen activé dans /admin/reglages, exposé au panier.
type MoyenPaiement struct {
	Cle     string `json:"cle"`
	Libelle string `json:"libelle"`
	Numero  string `json:"numero"`
}

// moyensPaiementActifs lit les réglages et ne retient que les moyens cochés.
// Le paiement à la livraison n'a pas de numéro marchand (champ vide).
func (a *App) moyensPaiementActifs() []MoyenPaiement {
	defs := []struct{ cle, libelle string }{
		{"orange", "Orange Money"},
		{"mtn", "MTN Mobile Money"},
		{"wave", "Wave"},
		{"moov", "Moov Money"},
		{"cod", "Paiement à la livraison"},
	}
	out := []MoyenPaiement{}
	for _, d := range defs {
		var on, num string
		_ = a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom=?`, "paiement_"+d.cle+"_on").Scan(&on)
		if on != "1" {
			continue
		}
		_ = a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom=?`, "paiement_"+d.cle+"_numero").Scan(&num)
		out = append(out, MoyenPaiement{Cle: d.cle, Libelle: d.libelle, Numero: strings.TrimSpace(num)})
	}
	return out
}

// ApiPaiements : GET /api/paiements — moyens actifs pour le panier.
// Public (le client doit choisir), sans aucune donnée sensible : seulement
// les moyens cochés par l'admin et leurs numéros marchands.
func (a *App) ApiPaiements(w http.ResponseWriter, r *http.Request) {
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "paiements": a.moyensPaiementActifs()})
}
