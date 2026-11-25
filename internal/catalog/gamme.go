package catalog

import (
	"encoding/json"
	"regexp"
	"sort"
	"strings"
	"unicode"
)

// Gamme : regroupement « modèles de la gamme » de la fiche produit.
// Clé = marque + ligne. La ligne vient, par priorité, de la spec Série,
// de la spec Gamme, des mots de ligne du nom (Catalyst, ECOSYS…), ou à
// défaut de la racine du modèle (305 pour 305/305XL, GI pour GI-490/790).
// Aucun regroupement inventé : seuls les produits réels du catalogue.

var typeWords = map[string]bool{
	"ecran": true, "commutateur": true, "routeur": true, "scanner": true,
	"imprimante": true, "matricielle": true, "cartouche": true,
	"bouteille": true, "encre": true, "toner": true, "point": true,
	"acces": true, "webcam": true, "terminal": true, "controleur": true,
	"lecteur": true, "camera": true, "enregistreur": true, "pare": true,
	"feu": true, "kit": true, "clavier": true, "souris": true,
	"client": true, "leger": true, "appliance": true, "securite": true,
	"disque": true, "dur": true, "nas": true, "ssd": true, "chaine": true,
	"banque": true, "pack": true, "cadre": true, "prise": true,
	"adaptateur": true, "housse": true, "carte": true, "memoire": true,
	"parasurtenseur": true, "telephone": true, "batteries": true,
	"externe": true, "modulaire": true, "saillie": true, "protection": true,
	"pour": true, "avec": true, "wifi": true, "wi": true, "fi": true,
	"pro": true, "max": true, "mini": true, "plus": true, "ultra": true,
	"noir": true, "jaune": true, "magenta": true, "cyan": true,
	"couleur": true, "blanc": true, "gris": true, "rouge": true,
	"bleu": true, "vert": true,
}

func normTok(s string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(s) {
		switch r {
		case 'à', 'á', 'â', 'ä':
			r = 'a'
		case 'é', 'è', 'ê', 'ë':
			r = 'e'
		case 'î', 'ï':
			r = 'i'
		case 'ô', 'ö':
			r = 'o'
		case 'ù', 'û', 'ü':
			r = 'u'
		case 'ç':
			r = 'c'
		case 'ÿ':
			r = 'y'
		}
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func nomTokens(nom string) []string {
	return strings.FieldsFunc(nom, func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r)
	})
}

func specVal(p *RawProduct, label string) string {
	var specs []struct {
		Label  string `json:"label"`
		Valeur string `json:"valeur"`
	}
	if len(p.Specs) == 0 || json.Unmarshal(p.Specs, &specs) != nil {
		return ""
	}
	for _, s := range specs {
		if strings.TrimSpace(s.Label) == label {
			return strings.TrimSpace(s.Valeur)
		}
	}
	return ""
}

var reModelRoot = regexp.MustCompile(`^([A-Za-z]+)?([0-9]+)?`)

// gammeLigne : ligne de gamme du produit ("" si indéterminable).
func gammeLigne(p *RawProduct) string {
	if v := specVal(p, "Série"); v != "" {
		return v
	}
	if v := specVal(p, "Gamme"); v != "" {
		return v
	}
	marqueToks := map[string]bool{}
	for _, t := range nomTokens(p.Marque) {
		marqueToks[normTok(t)] = true
	}
	var ligne []string
	var modelTok string
	for _, t := range nomTokens(p.Nom) {
		nt := normTok(t)
		if nt == "" || marqueToks[nt] || typeWords[nt] {
			continue
		}
		if strings.IndexFunc(t, unicode.IsDigit) >= 0 {
			if modelTok == "" {
				modelTok = t
			}
			continue
		}
		if len([]rune(nt)) >= 2 {
			ligne = append(ligne, t)
		}
	}
	if len(ligne) > 0 {
		return strings.Join(ligne, " ")
	}
	if modelTok != "" {
		m := reModelRoot.FindStringSubmatch(modelTok)
		if m[1] != "" {
			return strings.ToUpper(m[1])
		}
		return m[2]
	}
	return ""
}

// GammeKey : clé de regroupement ("" = pas de gamme affichable).
func GammeKey(p *RawProduct) string {
	m := normTok(p.Marque)
	if m == "" || p.Marque == "Générique" {
		return ""
	}
	l := normTok(gammeLigne(p))
	if l == "" {
		return ""
	}
	return m + "|" + l
}

// GammeLabel : libellé affiché (marque + ligne).
func GammeLabel(p *RawProduct) string {
	l := gammeLigne(p)
	if l == "" {
		return p.Marque
	}
	// La spec Série contient souvent déjà la marque (« HPE Aruba CX 6200F »).
	if strings.HasPrefix(normTok(l), normTok(p.Marque)) || strings.Contains(normTok(l), normTok(p.Marque)) {
		return l
	}
	return strings.TrimSpace(p.Marque + " " + l)
}

// GammeSiblings : autres produits de la même gamme, triés par nom.
func (s *Store) GammeSiblings(p *RawProduct) []*RawProduct {
	key := GammeKey(p)
	if key == "" {
		return nil
	}
	var out []*RawProduct
	for _, g := range s.Products {
		if g.Slug != p.Slug && GammeKey(g) == key {
			out = append(out, g)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Nom < out[j].Nom })
	return out
}

// GammeChipLabel : nom court du modèle (sans type ni marque).
func GammeChipLabel(p *RawProduct) string {
	marqueToks := map[string]bool{}
	for _, t := range nomTokens(p.Marque) {
		marqueToks[normTok(t)] = true
	}
	var keep []string
	started := false
	for _, t := range nomTokens(p.Nom) {
		nt := normTok(t)
		if !started && (marqueToks[nt] || typeWords[nt] || len([]rune(nt)) < 2) {
			continue
		}
		started = true
		keep = append(keep, t)
	}
	if len(keep) == 0 {
		return p.Nom
	}
	return strings.Join(keep, " ")
}
