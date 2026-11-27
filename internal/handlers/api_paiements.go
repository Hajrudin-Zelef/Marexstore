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

// PromoPublique : un code utilisable, exposé aux abonnés connectés.
type PromoPublique struct {
	Code        string `json:"code"`
	Type        string `json:"type"`
	Valeur      int64  `json:"valeur"`
	Description string `json:"description"`
}

// promosUtilisables : codes actifs et non expirés (la limite d'usage se
// vérifie à la commande, pas ici : l'affichage ne réserve rien).
func (a *App) promosUtilisables() []PromoPublique {
	a.ensurePromos()
	out := []PromoPublique{}
	rows, err := a.DB.Query(`SELECT code,type,valeur,description FROM promos WHERE actif=1 AND (expire='' OR expire>=date('now')) ORDER BY code`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var p PromoPublique
		if err := rows.Scan(&p.Code, &p.Type, &p.Valeur, &p.Description); err == nil && p.Valeur > 0 {
			out = append(out, p)
		}
	}
	return out
}

// ApiPromos : GET /api/promos — codes utilisables (abonnés connectés).
// Public comme les moyens de paiement : seuls les codes valables, sans
// compteurs ni limites internes.
func (a *App) ApiPromos(w http.ResponseWriter, r *http.Request) {
	if out := a.promosUtilisables(); out != nil {
		a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "promos": out})
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "promos": []PromoPublique{}})
}
