package catalog

import "fmt"

// Moteur de thème global (Studio Design /admin/design, esprit Customizer WP).
// Le thème par défaut reproduit le site actuel au pixel près : aucune sortie
// HTML/CSS n'est émise tant que l'admin ne personnalise rien (parité v3.25).

// DefaultSections : ordre exact des blocs de la page d'accueil actuelle.
var DefaultSections = []string{"hero", "minis", "filtre", "onglets", "produits", "vendus", "banniere", "recents", "marques"}

type Theme struct {
	Couleur      string   // "auto" (défaut = comportement actuel), "gazonnier", "liquid", "perso"
	Accent       string   // hexa utilisé si Couleur == "perso"
	Mode         string   // "auto" (défaut), "clair", "sombre"
	Police       string   // "inter" (défaut), "systeme"
	AnnonceOn    bool     // bandeau haut de page
	AnnonceTexte string   // texte du bandeau
	Sections     []string // blocs d'accueil activés, dans l'ordre (défaut = DefaultSections)
	CSS          string   // CSS personnalisé (filtré : pas de <script>)
}

func DefaultTheme() Theme {
	return Theme{Couleur: "auto", Mode: "auto", Police: "inter", Sections: append([]string{}, DefaultSections...)}
}

// IsDefault : vrai = aucune personnalisation = sortie identique au site d'origine.
func (t Theme) IsDefault() bool {
	if t.Couleur != "auto" || t.Mode != "auto" || t.Police != "inter" {
		return false
	}
	if t.AnnonceOn || t.AnnonceTexte != "" || t.CSS != "" {
		return false
	}
	if len(t.Sections) != len(DefaultSections) {
		return false
	}
	for i := range t.Sections {
		if t.Sections[i] != DefaultSections[i] {
			return false
		}
	}
	return true
}

// HtmlAttrs : attributs <html> du thème global (vide = défaut, parité intacte).
func (t Theme) HtmlAttrs() string {
	out := ""
	if t.Couleur == "gazonnier" || t.Couleur == "liquid" || t.Couleur == "perso" {
		out += ` data-color="` + t.Couleur + `"`
	}
	if t.Mode == "sombre" {
		out += ` data-theme="dark"`
	} else if t.Mode == "clair" {
		out += ` data-theme="light"`
	}
	return out
}

// SiteTheme : thème global actif (chargé depuis settings au démarrage,
// modifiable dans /admin/design). Défaut = site d’origine.
var SiteTheme = DefaultTheme()

// shade assombrit un hexa #rrggbb (facteur 0..1), repli = couleur telle quelle.
func Shade(hex string, f float64) string {
	var r, g, b int
	if len(hex) == 7 && hex[0] == '#' {
		fmt.Sscanf(hex[1:3], "%x", &r)
		fmt.Sscanf(hex[3:5], "%x", &g)
		fmt.Sscanf(hex[5:], "%x", &b)
		return fmt.Sprintf("#%02x%02x%02x", int(float64(r)*f), int(float64(g)*f), int(float64(b)*f))
	}
	return hex
}
