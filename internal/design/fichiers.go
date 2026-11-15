// Fichiers du site versionnés en base (mode Code du Studio, phase C4 / T3).
//
// Contrairement au document design (design.json, validé comme JSON), un
// fichier est du TEXTE BRUT (HTML / CSS / JS) : aucune validation de contenu
// — le code du site contient légitimement <script> et {{...}} (gabarits).
// La sécurité, c'est la liste fermée des noms + l'auth admin + le
// kill-switch DISALLOW_FILE_EDIT, pas le contenu.
//
// Stockage : table design_files (nom, contenu, version, auteur, cree_le,
// actif), append-only, une version active par nom — comme le design, donc
// chaque enregistrement est réversible (rollback = activer une ancienne
// version). Le nom en base :
//   - templates : le nom de fichier NU (ex. "_top.html", "apropos.html") —
//     c'est exactement ce que LoadStaticDB interroge déjà (T0) ; un fichier
//     enregistré prend effet au prochain reload(), sans redémarrage.
//   - static/* : le chemin complet ("static/app.js", ...) — aucun conflit
//     avec les noms de gabarits, et la requête "pages en base" de
//     LoadStaticDB (nom LIKE '%.html') ne les ramasse pas.
//
// Pourquoi en base et pas sur disque : le conteneur est read_only et
// ./static est monté ro — écrire sur disque échouerait, et serait perdu au
// prochain rebuild de l'image. La base, elle, est persistée (/data) et
// sauvegardée chaque nuit par tools/backup.sh.

package design

import (
	"database/sql"
	"fmt"
	"strings"
)

// MaxFichierBytes : taille maximale d'un fichier enregistré (2 Mo —
// app.js fait ~65 Ko, style.css ~75 Ko ; la limite est large).
const MaxFichierBytes = 2 * 1024 * 1024

// Fichiers statiques modifiables (les 3 seuls ; le reste de static/ est
// servi tel quel et ne passe pas par le Studio).
var fichiersStatiques = map[string]bool{
	"static/app.js":    true,
	"static/style.css": true,
	"static/theme.js":  true,
}

// FichierNom valide un nom affiché par le Studio ("templates/static/_top.html",
// "static/app.js") et retourne le nom à utiliser en base + true.
// Refuse tout le reste : traversée de chemin, chemin absolu, autre dossier,
// autre extension. Le nom de gabarit doit être un nom de fichier simple
// (sans "/" ni "..") se terminant par ".html".
func FichierNom(nom string) (string, bool) {
	if strings.HasPrefix(nom, "templates/static/") {
		base := strings.TrimPrefix(nom, "templates/static/")
		if len(base) <= len(".html") || strings.Contains(base, "/") || strings.Contains(base, "\\") ||
			strings.Contains(base, "..") || !strings.HasSuffix(base, ".html") {
			return "", false
		}
		return base, true
	}
	if fichiersStatiques[nom] {
		return nom, true
	}
	return "", false
}

// SaveFichier enregistre une nouvelle version d'un fichier (transaction :
// numéro = MAX(version)+1 pour ce nom, ancienne version désactivée).
// Le nom affiché est validé par FichierNom ; le contenu n'est pas validé
// (texte brut). Retourne le numéro de la nouvelle version.
func SaveFichier(db *sql.DB, nom, contenu, auteur string) (int, error) {
	dbNom, ok := FichierNom(nom)
	if !ok {
		return 0, fmt.Errorf("fichier non autorisé : %q", nom)
	}
	if len(contenu) > MaxFichierBytes {
		return 0, fmt.Errorf("fichier trop gros (%d octets, max %d)", len(contenu), MaxFichierBytes)
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var next int
	if err := tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM design_files WHERE nom=?`, dbNom).Scan(&next); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=0 WHERE nom=? AND actif=1`, dbNom); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`INSERT INTO design_files(nom, contenu, version, auteur, actif) VALUES(?,?,?,?,1)`, dbNom, contenu, next, auteur); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return next, nil
}

// LoadFichier charge la version active d'un fichier. Retourne
// (contenu, version, true, nil) si une version active existe,
// ("", 0, false, nil) sinon.
func LoadFichier(db *sql.DB, nom string) (string, int, bool, error) {
	dbNom, ok := FichierNom(nom)
	if !ok {
		return "", 0, false, fmt.Errorf("fichier non autorisé : %q", nom)
	}
	var contenu string
	var version int
	err := db.QueryRow(`SELECT contenu, version FROM design_files WHERE nom=? AND actif=1 ORDER BY version DESC LIMIT 1`, dbNom).Scan(&contenu, &version)
	if err == sql.ErrNoRows {
		return "", 0, false, nil
	}
	if err != nil {
		return "", 0, false, err
	}
	return contenu, version, true, nil
}

// ListFichierVersions liste les versions d'un fichier, la plus récente
// d'abord (tableau vide, jamais nil).
func ListFichierVersions(db *sql.DB, nom string) ([]Version, error) {
	dbNom, ok := FichierNom(nom)
	if !ok {
		return nil, fmt.Errorf("fichier non autorisé : %q", nom)
	}
	rows, err := db.Query(`SELECT version, auteur, cree_le, actif FROM design_files WHERE nom=? ORDER BY version DESC`, dbNom)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Version{}
	for rows.Next() {
		var v Version
		var actif int
		if err := rows.Scan(&v.Version, &v.Auteur, &v.CreeLe, &actif); err != nil {
			return nil, err
		}
		v.Actif = actif == 1
		out = append(out, v)
	}
	return out, nil
}

// ActivateFichier active une version d'un fichier (rollback 1 clic).
// Erreur si la version n'existe pas.
func ActivateFichier(db *sql.DB, nom string, version int) error {
	dbNom, ok := FichierNom(nom)
	if !ok {
		return fmt.Errorf("fichier non autorisé : %q", nom)
	}
	tx, err := db.Begin()
	if err != nil {
		return err
	}
	defer tx.Rollback()
	var exists int
	if err := tx.QueryRow(`SELECT COUNT(*) FROM design_files WHERE nom=? AND version=?`, dbNom, version).Scan(&exists); err != nil {
		return err
	}
	if exists == 0 {
		return fmt.Errorf("version %d introuvable pour %q", version, nom)
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=0 WHERE nom=?`, dbNom); err != nil {
		return err
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=1 WHERE nom=? AND version=?`, dbNom, version); err != nil {
		return err
	}
	return tx.Commit()
}
