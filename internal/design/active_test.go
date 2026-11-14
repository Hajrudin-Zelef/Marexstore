package design

import (
	"testing"
)

// Design partiel : seules quelques clés sont présentes — les accesseurs
// doivent retourner les valeurs du design là où elles existent, et le défaut
// partout ailleurs, sans jamais planter.
const designPartiel = `{"front":{"accent":"#2563eb"},"Tx":{"bandeau":"Bandeau test"},"header":{"showTopbar":false},"shop":{"perPage":24}}`

// Design aux types incorrects : les accesseurs doivent retomber sur le
// défaut, jamais paniquer.
const designMauvaisTypes = `{"Tx":{"bandeau":42},"header":{"showTopbar":"oui"},"shop":{"perPage":"24"}}`

func TestLoadActiveSansDesign(t *testing.T) {
	database := openTestDB(t)
	defer func() { _ = LoadActive(database) }() // remet l'actif à nil
	if err := LoadActive(database); err != nil {
		t.Fatal(err)
	}
	if Current() != nil {
		t.Fatal("actif attendu nil sans design publié")
	}
	d := Current()
	if d.Version() != 0 {
		t.Fatalf("version attendue 0, got %d", d.Version())
	}
	if got := d.Text("bandeau", "Défaut"); got != "Défaut" {
		t.Fatalf("Text sans design : %q", got)
	}
	if got := d.Bool("header", "showTopbar", true); got != true {
		t.Fatalf("Bool sans design : %v", got)
	}
	if got := d.Str("front", "accent", "#000"); got != "#000" {
		t.Fatalf("Str sans design : %q", got)
	}
	if got := d.Int("shop", "perPage", 12); got != 12 {
		t.Fatalf("Int sans design : %d", got)
	}
	if got := d.List("header", "navLinks"); got != nil {
		t.Fatalf("List sans design : %v", got)
	}
}

func TestLoadActiveEtAccesseurs(t *testing.T) {
	database := openTestDB(t)
	defer func() { _ = LoadActive(database) }()
	if _, err := Save(database, []byte(designPartiel), "admin"); err != nil {
		t.Fatal(err)
	}
	if err := LoadActive(database); err != nil {
		t.Fatal(err)
	}
	d := Current()
	if d == nil {
		t.Fatal("actif nil après publication")
	}
	if d.Version() != 1 {
		t.Fatalf("version attendue 1, got %d", d.Version())
	}
	// Valeurs présentes dans le design.
	if got := d.Text("bandeau", "Défaut"); got != "Bandeau test" {
		t.Fatalf("Text bandeau : %q", got)
	}
	if got := d.Bool("header", "showTopbar", true); got != false {
		t.Fatalf("Bool showTopbar : %v", got)
	}
	if got := d.Str("front", "accent", "#000"); got != "#2563eb" {
		t.Fatalf("Str accent : %q", got)
	}
	if got := d.Int("shop", "perPage", 12); got != 24 {
		t.Fatalf("Int perPage : %d", got)
	}
	// Clés absentes → défaut.
	if got := d.Text("inexistant", "Défaut"); got != "Défaut" {
		t.Fatalf("Text absent : %q", got)
	}
	if got := d.Bool("header", "showPhone", true); got != true {
		t.Fatalf("Bool absent : %v", got)
	}
	if got := d.Str("footer", "newsletter", "N"); got != "N" {
		t.Fatalf("Str section absente : %q", got)
	}
	if got := d.Int("fiche", "similaires_nb", 4); got != 4 {
		t.Fatalf("Int absent : %d", got)
	}
	if !d.Has("front") || d.Has("inexistant") {
		t.Fatal("Has incohérent")
	}
}

func TestAccesseursMauvaisTypes(t *testing.T) {
	database := openTestDB(t)
	defer func() { _ = LoadActive(database) }()
	if _, err := Save(database, []byte(designMauvaisTypes), "admin"); err != nil {
		t.Fatal(err)
	}
	if err := LoadActive(database); err != nil {
		t.Fatal(err)
	}
	d := Current()
	if got := d.Text("bandeau", "Défaut"); got != "Défaut" {
		t.Fatalf("Text nombre : %q", got)
	}
	if got := d.Bool("header", "showTopbar", true); got != true {
		t.Fatalf("Bool chaîne : %v", got)
	}
	if got := d.Int("shop", "perPage", 12); got != 12 {
		t.Fatalf("Int chaîne : %d", got)
	}
}

func TestRefreshNouvelleVersion(t *testing.T) {
	database := openTestDB(t)
	defer func() { _ = LoadActive(database) }()
	if _, err := Save(database, []byte(designOK), "admin"); err != nil {
		t.Fatal(err)
	}
	if err := LoadActive(database); err != nil {
		t.Fatal(err)
	}
	if Current().Version() != 1 {
		t.Fatalf("v1 attendue, got %d", Current().Version())
	}
	if _, err := Save(database, []byte(designPartiel), "admin"); err != nil {
		t.Fatal(err)
	}
	if err := Refresh(database); err != nil {
		t.Fatal(err)
	}
	d := Current()
	if d.Version() != 2 {
		t.Fatalf("v2 attendue après Refresh, got %d", d.Version())
	}
	if got := d.Text("bandeau", "Défaut"); got != "Bandeau test" {
		t.Fatalf("bandeau v2 : %q", got)
	}
}

func TestAccesseursRecepteurNil(t *testing.T) {
	var d *Document
	if got := d.Text("bandeau", "Défaut"); got != "Défaut" {
		t.Fatalf("Text nil : %q", got)
	}
	if got := d.Bool("header", "showTopbar", true); got != true {
		t.Fatalf("Bool nil : %v", got)
	}
	if got := d.Str("front", "accent", "#000"); got != "#000" {
		t.Fatalf("Str nil : %q", got)
	}
	if got := d.Int("shop", "perPage", 12); got != 12 {
		t.Fatalf("Int nil : %d", got)
	}
	if got := d.List("header", "navLinks"); got != nil {
		t.Fatalf("List nil : %v", got)
	}
	if d.Version() != 0 || d.Has("front") {
		t.Fatal("Version/Has nil incohérents")
	}
}
