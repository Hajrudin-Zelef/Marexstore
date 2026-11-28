package db

import (
	"path/filepath"
	"testing"
)

// Task 1 — module livraison : EnsureSchema sur DB vide doit créer la table
// livraison_zones (+ seed 3 zones), les colonnes orders et les 2 settings.
func TestLivraisonSchema(t *testing.T) {
	dir := t.TempDir()
	d, err := Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer d.Close()
	if err := EnsureSchema(d, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	// Table + seed 3 zones.
	var n int
	if err := d.QueryRow(`SELECT COUNT(*) FROM livraison_zones`).Scan(&n); err != nil {
		t.Fatalf("livraison_zones illisible : %v", err)
	}
	if n != 3 {
		t.Fatalf("zones = %d, attendu 3", n)
	}
	rows, err := d.Query(`SELECT nom, type, frais, seuil_franco, delai, ordre, actif FROM livraison_zones ORDER BY ordre`)
	if err != nil {
		t.Fatal(err)
	}
	defer rows.Close()
	type zone struct {
		nom, typ, delai string
		frais, seuil, ordre, actif int
	}
	var got []zone
	for rows.Next() {
		var z zone
		if err := rows.Scan(&z.nom, &z.typ, &z.frais, &z.seuil, &z.delai, &z.ordre, &z.actif); err != nil {
			t.Fatal(err)
		}
		got = append(got, z)
	}
	want := []zone{
		{"Retrait boutique — Plateau", "retrait", "Immédiat", 0, 0, 1, 1},
		{"Livraison Abidjan", "livraison", "24–72h", 0, 100000, 2, 1},
		{"Expédition intérieur", "expedition", "À confirmer", 0, 0, 3, 1},
	}
	if len(got) != len(want) {
		t.Fatalf("zones = %v, attendu %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Fatalf("zone %d = %+v, attendu %+v", i, got[i], want[i])
		}
	}
	// Colonnes orders (backfill sur base existante).
	ocols := map[string]bool{}
	orows, err := d.Query(`PRAGMA table_info(orders)`)
	if err != nil {
		t.Fatal(err)
	}
	defer orows.Close()
	for orows.Next() {
		var cid int
		var name, typ string
		var nn, dflt, pk any
		if err := orows.Scan(&cid, &name, &typ, &nn, &dflt, &pk); err != nil {
			t.Fatal(err)
		}
		ocols[name] = true
	}
	for _, c := range []string{"livraison_zone_id", "livraison_frais", "numero_payeur", "adresse"} {
		if !ocols[c] {
			t.Fatalf("colonne orders.%s manquante", c)
		}
	}
	// Settings avec défauts.
	for nom, val := range map[string]string{
		"livraison_seuil_abidjan":  "100000",
		"livraison_retrait_adresse": "Plateau, Abidjan",
	} {
		var gotVal string
		if err := d.QueryRow(`SELECT valeur FROM settings WHERE nom=?`, nom).Scan(&gotVal); err != nil {
			t.Fatalf("setting %s illisible : %v", nom, err)
		}
		if gotVal != val {
			t.Fatalf("setting %s = %q, attendu %q", nom, gotVal, val)
		}
	}
	// Idempotence : rejouer EnsureSchema ne duplique rien.
	if err := EnsureSchema(d, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	if err := d.QueryRow(`SELECT COUNT(*) FROM livraison_zones`).Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 3 {
		t.Fatalf("rejouage : zones = %d, attendu 3", n)
	}
}
