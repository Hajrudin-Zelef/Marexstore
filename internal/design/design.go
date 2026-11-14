// Package design — le design complet du Studio, un seul document JSON
// versionné dans design_files (table créée en T0, ajout seul).
//
// Le document est l'état du Studio (les 13 clés de snap(), sans `files` —
// le contenu des fichiers de code, c'est T3) : front, rows, shop, fiche,
// panier, Tx, txtPages, menu, seo, reglages, header, footer, acc.
// `studioV16` vit dans front._studioV16 (imbriqué, pas en top-level).
//
// Règles de validation (défense en profondeur) :
//   - les clés prix/stock/commande/total existent dans l'état légitime du
//     Studio, mais seulement comme LIBELLÉS (strings : « En stock »,
//     « Sous-total », « Min »…) ou bascules d'affichage. On n'interdit donc
//     pas ces noms de clés — on interdit les VALEURS NUMÉRIQUES sous ces
//     noms exacts (un vrai prix ou un vrai stock serait un nombre).
//   - aucun contenu actif n'est autorisé dans les strings : <script>,
//     <iframe>, <form>, {{…}}, javascript:, attributs on*=.
//   - la liste des clés top-level est fermée : aucune clé produit/commande
//     (products, orders…) ne peut entrer dans le document.
package design

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// DocName : nom du document design dans design_files.
const DocName = "design.json"

// MaxDesignBytes : taille maximale d'un document design (quelques dizaines
// de Ko en pratique ; la limite bloque les envois abusifs).
const MaxDesignBytes = 2 * 1024 * 1024

// AllowedTopKeys : clés top-level autorisées (les 13 clés de snap(), sans files).
var AllowedTopKeys = map[string]bool{
	"front": true, "rows": true, "shop": true, "fiche": true, "panier": true,
	"Tx": true, "txtPages": true, "menu": true, "seo": true, "reglages": true,
	"header": true, "footer": true, "acc": true,
}

// Version : une version publiée du design.
type Version struct {
	Version int    `json:"version"`
	Auteur  string `json:"auteur"`
	CreeLe  string `json:"cree_le"`
	Actif   bool   `json:"actif"`
}

// forbiddenValue : motifs interdits dans les valeurs string (XSS / injection
// de gabarit). {{…}} est interdit car le moteur de rendu n'utilise aucun
// gabarit côté design — une accolade double ne peut être que malveillante.
var forbiddenValue = []string{
	"<script", "<iframe", "<form", "<object", "<embed", "{{", "}}",
	"javascript:", "vbscript:", "data:text/html",
}

// forbiddenAttr : attributs gestionnaires d'événements (onclick=, onerror=…).
var forbiddenAttr = regexp.MustCompile(`(?i)\son[a-z]+\s*=`)

// numericForbiddenKeys : noms de clés sous lesquels une valeur NUMÉRIQUE est
// refusée (un vrai prix, un vrai stock, un vrai total). Les libellés du
// Studio sous ces noms sont des strings et passent.
var numericForbiddenKeys = map[string]bool{
	"prix": true, "prix_boutique": true, "prixboutique": true, "stock": true,
	"quantite": true, "quantité": true, "total": true, "montant": true,
}

// Validate : valide un document design brut avant enregistrement.
func Validate(raw []byte) error {
	if len(raw) == 0 {
		return fmt.Errorf("design vide")
	}
	if len(raw) > MaxDesignBytes {
		return fmt.Errorf("design trop gros (%d octets, max %d)", len(raw), MaxDesignBytes)
	}
	var top map[string]json.RawMessage
	dec := json.NewDecoder(bytes.NewReader(raw))
	if err := dec.Decode(&top); err != nil {
		return fmt.Errorf("JSON invalide : %w", err)
	}
	if top == nil {
		return fmt.Errorf("le design doit être un objet JSON")
	}
	for k := range top {
		if !AllowedTopKeys[k] {
			return fmt.Errorf("clé non autorisée : %q", k)
		}
	}
	var v any
	if err := json.Unmarshal(raw, &v); err != nil {
		return fmt.Errorf("JSON invalide : %w", err)
	}
	return validateValue(v)
}

func validateValue(v any) error {
	switch t := v.(type) {
	case map[string]any:
		for k, val := range t {
			if err := validateValue(val); err != nil {
				return err
			}
			// Valeur numérique sous un nom de donnée métier → refus.
			if _, isNum := val.(float64); isNum && numericForbiddenKeys[strings.ToLower(k)] {
				return fmt.Errorf("champ numérique interdit : %q (prix/stock/total)", k)
			}
		}
	case []any:
		for _, val := range t {
			if err := validateValue(val); err != nil {
				return err
			}
		}
	case string:
		low := strings.ToLower(t)
		for _, f := range forbiddenValue {
			if strings.Contains(low, f) {
				return fmt.Errorf("contenu interdit dans une valeur : %q", f)
			}
		}
		if forbiddenAttr.MatchString(t) {
			return fmt.Errorf("attribut gestionnaire d'événement interdit (on*=)")
		}
	}
	return nil
}

// Load : charge la version active du design. Retourne (nil, 0, nil) si
// aucune version n'est publiée (le Studio démarre alors sur son défaut).
func Load(db *sql.DB) (raw []byte, version int, err error) {
	var contenu string
	err = db.QueryRow(`SELECT contenu, version FROM design_files WHERE nom=? AND actif=1 ORDER BY version DESC LIMIT 1`, DocName).Scan(&contenu, &version)
	if err == sql.ErrNoRows {
		return nil, 0, nil
	}
	if err != nil {
		return nil, 0, err
	}
	return []byte(contenu), version, nil
}

// Save : valide puis enregistre le design en nouvelle version (transaction :
// l'ancienne version active est désactivée, la nouvelle devient active).
// Retourne le numéro de la nouvelle version.
func Save(db *sql.DB, raw []byte, auteur string) (int, error) {
	if err := Validate(raw); err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var next int
	if err := tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM design_files WHERE nom=?`, DocName).Scan(&next); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=0 WHERE nom=? AND actif=1`, DocName); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`INSERT INTO design_files(nom, contenu, version, auteur, actif) VALUES(?,?,?,?,1)`, DocName, string(raw), next, auteur); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return next, nil
}

// ListVersions : liste les versions, la plus récente d'abord.
func ListVersions(db *sql.DB) ([]Version, error) {
	rows, err := db.Query(`SELECT version, auteur, cree_le, actif FROM design_files WHERE nom=? ORDER BY version DESC`, DocName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Version
	for rows.Next() {
		var v Version
		var actif int
		if err := rows.Scan(&v.Version, &v.Auteur, &v.CreeLe, &actif); err != nil {
			return nil, err
		}
		v.Actif = actif == 1
		out = append(out, v)
	}
	return out, rows.Err()
}

// Activate : réactive une ancienne version (retour en arrière en 1 clic).
// Le contenu de la version devient le design actif ; l'historique est conservé.
func Activate(db *sql.DB, version int) error {
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM design_files WHERE nom=? AND version=?`, DocName, version).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("version %d introuvable", version)
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=0 WHERE nom=?`, DocName); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=1 WHERE nom=? AND version=?`, DocName, version); err != nil {
		return err
	}
	return tx.Commit()
}

// designTheme : partie thème d'un document design (front + Tx.bandeau).
type designTheme struct {
	Front struct {
		Accent      string `json:"accent"`
		Font        string `json:"font"`
		ShowAnnonce bool   `json:"showAnnonce"`
		CustomCSS   string `json:"customCSS"`
	} `json:"front"`
	Tx struct {
		Bandeau string `json:"bandeau"`
	} `json:"Tx"`
}

// ThemeFromDesign : extrait le thème (partie `front` + bandeau) d'un document
// design, en partant du thème courant — les sections d'accueil du serveur
// sont PRÉSERVÉES (le design ne les contient pas : l'éditeur visuel n'a que
// 3 lignes, le site en a 9 — voir phase 1).
func ThemeFromDesign(raw []byte, current catalog.Theme) (catalog.Theme, error) {
	var d designTheme
	if err := json.Unmarshal(raw, &d); err != nil {
		return current, fmt.Errorf("design illisible : %w", err)
	}
	t := current
	accent := strings.ToUpper(strings.TrimSpace(d.Front.Accent))
	switch accent {
	case "#2E5136":
		t.Couleur = "gazonnier"
	case "#84CC16":
		t.Couleur = "liquid"
	case "":
		// Pas d'accent dans le design : garder la couleur courante.
	default:
		t.Couleur = "perso"
	}
	if accent != "" {
		t.Accent = accent
	}
	t.Mode = "auto"
	if d.Front.Font == "Inter" {
		t.Police = "inter"
	} else if d.Front.Font != "" {
		t.Police = "systeme"
	}
	t.AnnonceOn = d.Front.ShowAnnonce
	if d.Tx.Bandeau != "" {
		t.AnnonceTexte = d.Tx.Bandeau
	}
	t.CSS = d.Front.CustomCSS
	return t, nil
}
