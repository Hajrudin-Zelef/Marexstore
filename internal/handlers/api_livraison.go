package handlers

import (
	"net/http"
	"strconv"
)

// ZoneLivraison : une zone active, exposée au tunnel d'achat.
// Pas de champ interne (ordre, actif) : le client n'en a pas besoin.
type ZoneLivraison struct {
	ID          int64  `json:"id"`
	Nom         string `json:"nom"`
	Type        string `json:"type"`
	Frais       int64  `json:"frais"`
	SeuilFranco int64  `json:"seuil_franco"`
	Delai       string `json:"delai"`
}

// zonesLivraisonActives : zones actives triées par ordre (le client doit
// afficher frais et total ; le seuil franco par zone reste visible ici).
func (a *App) zonesLivraisonActives() []ZoneLivraison {
	out := []ZoneLivraison{}
	rows, err := a.DB.Query(`SELECT id,nom,type,frais,seuil_franco,delai FROM livraison_zones WHERE actif=1 ORDER BY ordre ASC, id ASC`)
	if err != nil {
		return out
	}
	defer rows.Close()
	for rows.Next() {
		var z ZoneLivraison
		if err := rows.Scan(&z.ID, &z.Nom, &z.Type, &z.Frais, &z.SeuilFranco, &z.Delai); err == nil {
			out = append(out, z)
		}
	}
	return out
}

// ApiLivraison : GET /api/livraison — zones actives + seuils pour le tunnel.
// Public comme les moyens de paiement : seules les zones actives, sans
// champ interne, plus le seuil Abidjan et l'adresse de retrait.
func (a *App) ApiLivraison(w http.ResponseWriter, r *http.Request) {
	seuil, err := strconv.ParseInt(a.livraisonGet("livraison_seuil_abidjan", "100000"), 10, 64)
	if err != nil || seuil < 0 {
		seuil = 100000
	}
	a.writeJSON(w, http.StatusOK, map[string]any{
		"ok":              true,
		"zones":           a.zonesLivraisonActives(),
		"seuil_abidjan":   seuil,
		"retrait_adresse": a.livraisonGet("livraison_retrait_adresse", "Plateau, Abidjan"),
	})
}
