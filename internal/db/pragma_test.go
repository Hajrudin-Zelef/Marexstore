package db

import (
	"path/filepath"
	"testing"
)

func TestOpenPragmas(t *testing.T) {
	d, err := Open(filepath.Join(t.TempDir(), "t.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	var bt, fk int
	if err := d.QueryRow(`PRAGMA busy_timeout`).Scan(&bt); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`PRAGMA foreign_keys`).Scan(&fk); err != nil {
		t.Fatal(err)
	}
	if bt != 5000 {
		t.Errorf("busy_timeout = %d, attendu 5000", bt)
	}
	if fk != 1 {
		t.Errorf("foreign_keys = %d, attendu 1", fk)
	}
}
