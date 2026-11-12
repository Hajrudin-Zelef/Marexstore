package db

import (
	"database/sql"
	"os"

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

// Migrate applique le fichier SQL embarqué (simple, V1).
func Migrate(d *sql.DB, sqlPath string) error {
	b, err := os.ReadFile(sqlPath)
	if err != nil {
		return err
	}
	_, err = d.Exec(string(b))
	return err
}
