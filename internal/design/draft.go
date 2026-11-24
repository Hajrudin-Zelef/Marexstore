package design

import (
	"database/sql"
	"errors"
)

// Brouillon de travail : document NON publié, sauvegardé explicitement par
// l'admin (« Sauvegarder »). Il vit dans design_files sous un autre nom, donc
// il ne remplace jamais DocName : le site public continue de rendre la
// version publiée tant que l'admin n'a pas cliqué « Publier ».
const DraftName = "design.draft.json"

// SaveDraft : valide (mêmes règles que la publication) puis enregistre le
// brouillon en nouvelle version. Aucun effet sur le rendu public.
func SaveDraft(db *sql.DB, raw []byte, auteur string) (int, error) {
	if err := Validate(raw); err != nil {
		return 0, err
	}
	tx, err := db.Begin()
	if err != nil {
		return 0, err
	}
	defer tx.Rollback()
	var next int
	if err := tx.QueryRow(`SELECT COALESCE(MAX(version),0)+1 FROM design_files WHERE nom=?`, DraftName).Scan(&next); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`UPDATE design_files SET actif=0 WHERE nom=? AND actif=1`, DraftName); err != nil {
		return 0, err
	}
	if _, err := tx.Exec(`INSERT INTO design_files(nom, contenu, version, auteur, actif) VALUES(?,?,?,?,1)`,
		DraftName, string(raw), next, auteur); err != nil {
		return 0, err
	}
	if err := tx.Commit(); err != nil {
		return 0, err
	}
	return next, nil
}

// LoadDraft : dernier brouillon actif. ok=false si aucun brouillon.
func LoadDraft(db *sql.DB) (raw []byte, version int, creeLe string, ok bool, err error) {
	var contenu string
	err = db.QueryRow(`SELECT contenu, version, cree_le FROM design_files
		WHERE nom=? AND actif=1 ORDER BY version DESC LIMIT 1`, DraftName).Scan(&contenu, &version, &creeLe)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, 0, "", false, nil
	}
	if err != nil {
		return nil, 0, "", false, err
	}
	return []byte(contenu), version, creeLe, true, nil
}
