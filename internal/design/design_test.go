package design

import (
	"database/sql"
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
)

// Design réaliste : libellés prix/stock (strings) et bascules d'affichage —
// c'est l'état légitime du Studio, il doit PASSER la validation.
const designOK = `{"front":{"accent":"#2E5136","font":"Inter","showAnnonce":true,"customCSS":""},"rows":[],"shop":{"showStockOnly":true},"fiche":{},"panier":{},"Tx":{"prix":"Prix","en_stock":"En stock","sous_total":"Sous-total","prix_min":"Min","prix_max":"Max","bandeau":"Livraison offerte"},"txtPages":{},"menu":{},"seo":{},"reglages":{},"header":{},"footer":{},"acc":{}}`

func TestValidateOK(t *testing.T) {
	if err := Validate([]byte(designOK)); err != nil {
		t.Fatalf("design légitime refusé : %v", err)
	}
}

func TestValidateRefuseScript(t *testing.T) {
	bad := `{"front":{},"Tx":{"bandeau":"<script>alert(1)</script>"}}`
	if err := Validate([]byte(bad)); err == nil {
		t.Fatal("script accepté, attendu un refus")
	}
}

func TestValidateRefuseGabari(t *testing.T) {
	bad := `{"front":{},"Tx":{"bandeau":"{{.Titre}}"}}`
	if err := Validate([]byte(bad)); err == nil {
		t.Fatal("{{…}} accepté, attendu un refus")
	}
}

func TestValidateRefuseCleInconnue(t *testing.T) {
	bad := `{"front":{},"products":[{"id":1,"prix":100}]}`
	if err := Validate([]byte(bad)); err == nil {
		t.Fatal("clé 'products' acceptée, attendu un refus")
	}
}

func TestValidateRefusePrixNumerique(t *testing.T) {
	bad := `{"front":{"prix":99999}}`
	if err := Validate([]byte(bad)); err == nil {
		t.Fatal("prix numérique accepté, attendu un refus")
	}
}

func TestValidateRefuseTropGros(t *testing.T) {
	bad := `{"front":{},"Tx":{"bandeau":"` + strings.Repeat("x", MaxDesignBytes) + `"}}`
	if err := Validate([]byte(bad)); err == nil {
		t.Fatal("design trop gros accepté, attendu un refus")
	}
}

func TestValidateRefuseJSONInvalide(t *testing.T) {
	if err := Validate([]byte(`{"front":`)); err == nil {
		t.Fatal("JSON invalide accepté, attendu un refus")
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	return database
}

func TestSaveLoadListActivate(t *testing.T) {
	database := openTestDB(t)
	// Rien publié : Load renvoie nil.
	if raw, v, err := Load(database); err != nil || raw != nil || v != 0 {
		t.Fatalf("Load sans design : raw=%v v=%d err=%v", raw, v, err)
	}
	// v1.
	v1, err := Save(database, []byte(designOK), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if v1 != 1 {
		t.Fatalf("v1 attendue 1, got %d", v1)
	}
	// v2 (accent différent).
	designV2 := strings.Replace(designOK, "#2E5136", "#84CC16", 1)
	v2, err := Save(database, []byte(designV2), "admin")
	if err != nil {
		t.Fatal(err)
	}
	if v2 != 2 {
		t.Fatalf("v2 attendue 2, got %d", v2)
	}
	// Load = v2 active.
	raw, v, err := Load(database)
	if err != nil || v != 2 || !strings.Contains(string(raw), "#84CC16") {
		t.Fatalf("Load après v2 : v=%d err=%v contient liquid=%v", v, err, strings.Contains(string(raw), "#84CC16"))
	}
	// Liste : 2 versions, v2 active, v1 inactive.
	vs, err := ListVersions(database)
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 2 || vs[0].Version != 2 || !vs[0].Actif || vs[1].Version != 1 || vs[1].Actif {
		t.Fatalf("versions inattendues : %+v", vs)
	}
	// Réactiver v1 : Load = v1.
	if err := Activate(database, 1); err != nil {
		t.Fatal(err)
	}
	raw, v, err = Load(database)
	if err != nil || v != 1 || !strings.Contains(string(raw), "#2E5136") {
		t.Fatalf("Load après activation v1 : v=%d err=%v", v, err)
	}
	// Version inexistante : refus.
	if err := Activate(database, 99); err == nil {
		t.Fatal("activation d'une version inexistante acceptée")
	}
}

func TestThemeFromDesign(t *testing.T) {
	current := catalog.Theme{Couleur: "liquid", Accent: "#84CC16", Mode: "auto", Police: "inter", Sections: []string{"hero", "minis", "produits", "vendus"}}
	th, err := ThemeFromDesign([]byte(designOK), current)
	if err != nil {
		t.Fatal(err)
	}
	if th.Couleur != "gazonnier" {
		t.Fatalf("couleur attendue gazonnier, got %q", th.Couleur)
	}
	if th.Accent != "#2E5136" {
		t.Fatalf("accent attendu #2E5136, got %q", th.Accent)
	}
	if th.AnnonceTexte != "Livraison offerte" {
		t.Fatalf("bandeau attendu 'Livraison offerte', got %q", th.AnnonceTexte)
	}
	// Sections du serveur PRÉSERVÉES (le design ne les contient pas).
	if len(th.Sections) != 4 || th.Sections[3] != "vendus" {
		t.Fatalf("sections non préservées : %v", th.Sections)
	}
}

func TestThemeFromDesignLiquid(t *testing.T) {
	current := catalog.Theme{Couleur: "gazonnier", Accent: "#2E5136", Mode: "auto", Police: "inter", Sections: []string{"hero"}}
	designV2 := strings.Replace(designOK, "#2E5136", "#84CC16", 1)
	th, err := ThemeFromDesign([]byte(designV2), current)
	if err != nil {
		t.Fatal(err)
	}
	if th.Couleur != "liquid" || th.Accent != "#84CC16" {
		t.Fatalf("liquid attendu, got couleur=%q accent=%q", th.Couleur, th.Accent)
	}
}
