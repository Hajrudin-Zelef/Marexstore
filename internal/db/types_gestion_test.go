package db

import (
	"path/filepath"
	"testing"
)

// Les catégories déclarées (même vides, 0 produit) doivent survivre en base :
// sans table dédiée, un type sans produit reste invisible des cartes.
func TestEnsureSchemaCreeDeclaredTypes(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := EnsureSchema(d, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if _, err := d.Exec(`INSERT INTO declared_types(top, sub) VALUES('t1','x')`); err != nil {
		t.Fatalf("insert déclaré : %v", err)
	}
	if _, err := d.Exec(`INSERT INTO declared_types(top, sub) VALUES('t1','x')`); err == nil {
		t.Fatal("doublon (t1|x) accepté, PRIMARY KEY attendue")
	}
}
