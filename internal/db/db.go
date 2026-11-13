package db

import (
	"database/sql"
	"fmt"
	"os"
	"strings"

	_ "modernc.org/sqlite"
)

func Open(path string) (*sql.DB, error) {
	d, err := sql.Open("sqlite", path+"?cache=shared")
	if err != nil {
		return nil, err
	}
	d.SetMaxOpenConns(1)
	if _, err := d.Exec(`PRAGMA journal_mode=WAL; PRAGMA foreign_keys=ON;`); err != nil {
		return nil, err
	}
	return d, nil
}

// EnsureFTS remplit l'index de recherche FTS5 s'il est vide
// (auto-réparation après déploiement : l'import le remplit aussi).
func EnsureFTS(d *sql.DB) error {
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM products_fts`).Scan(&n); err != nil {
		return err
	}
	if n > 0 {
		return nil
	}
	_, err := d.Exec(`INSERT INTO products_fts(slug, nom, marque, categorie, sous_categorie, accroche, description, search_index)
		SELECT slug, nom, marque, categorie, sous_categorie, accroche, description, search_index FROM products`)
	return err
}

// SyncFTS resynchronise la ligne FTS d'un produit (après édition admin).
func SyncFTS(d *sql.DB, slug string) {
	_, _ = d.Exec(`DELETE FROM products_fts WHERE rowid IN (SELECT rowid FROM products_fts WHERE slug=?)`, slug)
	_, _ = d.Exec(`INSERT INTO products_fts(slug, nom, marque, categorie, sous_categorie, accroche, description, search_index)
		SELECT slug, nom, marque, categorie, sous_categorie, accroche, description, search_index FROM products WHERE slug=?`, slug)
}

// Migrate applique un fichier SQL brut (usage : création initiale).
func Migrate(d *sql.DB, sqlPath string) error {
	b, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	_, err = d.Exec(string(b))
	return err
}

// EnsureSchema rend le schéma idempotent : 001 est rejouable sans risque
// (tout est IF NOT EXISTS), puis les colonnes manquantes sont ajoutées.
func EnsureSchema(d *sql.DB, migDir string) error {
	if err := Migrate(d, migDir+"/001_init.sql"); err != nil {
		return fmt.Errorf("001: %w", err)
	}
	cols := map[string]bool{}
	rows, err := d.Query(`PRAGMA table_info(products)`)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var cid int
		var n, t string
		var nn, dflt, pk any
		if err := rows.Scan(&cid, &n, &t, &nn, &dflt, &pk); err != nil {
			return err
		}
		cols[n] = true
	}
	want := map[string]string{
		"prix_boutique": "INTEGER NULL", "prix_avant": "INTEGER NULL",
		"pointsforts_json":  "TEXT NOT NULL DEFAULT '[]'",
		"consommables_json": "TEXT NOT NULL DEFAULT '[]'",
		"store_url":         "TEXT NOT NULL DEFAULT ''",
		"ref":               "TEXT NOT NULL DEFAULT ''",
		"ref_fabricant":     "TEXT NOT NULL DEFAULT ''",
		"main":              "TEXT NOT NULL DEFAULT ''",
		"main_slug":         "TEXT NOT NULL DEFAULT ''",
		"sub":               "TEXT NOT NULL DEFAULT ''",
		"sub_slug":          "TEXT NOT NULL DEFAULT ''",
		"live_top":          "TEXT NOT NULL DEFAULT ''",
		"live_sub":          "TEXT NOT NULL DEFAULT ''",
		"facets_json":       "TEXT NOT NULL DEFAULT '{}'",
	}
	for col, typ := range want {
		if !cols[col] {
			if _, err := d.Exec(fmt.Sprintf("ALTER TABLE products ADD COLUMN %s %s", col, typ)); err != nil {
				if !strings.Contains(err.Error(), "duplicate column") {
					return fmt.Errorf("col %s: %w", col, err)
				}
			}
		}
	}
	return nil
}
