package handlers

import (
	"path/filepath"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
)

// P1 — le serveur gagne : chaque Publier incrémente theme_version (1, 2, ...),
// ThemeFromDB le relit, et la réinitialisation l'efface (retour au défaut).
func TestThemeServeurGagneApresPublish(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer database.Close()
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	app := &App{DB: database}
	publie := func(couleur string) catalog.Theme {
		th := catalog.Theme{Couleur: couleur, Mode: "auto", Police: "inter", Sections: append([]string{}, catalog.DefaultSections...)}
		app.themeToSettings(th)
		return catalog.SiteTheme
	}
	// Premier Publier : version 1.
	if got := publie("gazonnier"); got.Version != 1 {
		t.Fatalf("1er Publier : version attendue 1, got %d", got.Version)
	}
	// Deuxième Publier : version 2.
	if got := publie("liquid"); got.Version != 2 {
		t.Fatalf("2e Publier : version attendue 2, got %d", got.Version)
	}
	// ThemeFromDB relit la version et le thème.
	relu := site.ThemeFromDB(database)
	if relu.Version != 2 || relu.Couleur != "liquid" {
		t.Fatalf("ThemeFromDB : version 2 + liquid attendus, got version=%d couleur=%q", relu.Version, relu.Couleur)
	}
	// HtmlAttrs émet la version (thème non défaut).
	if got := relu.HtmlAttrs(); !strings.Contains(got, `data-theme-version="2"`) {
		t.Fatalf("HtmlAttrs doit émettre data-theme-version=\"2\", got %q", got)
	}
	// Réinitialisation : version effacée, thème par défaut, pas d'attribut.
	if _, err := database.Exec(`DELETE FROM settings WHERE nom LIKE 'theme_%'`); err != nil {
		t.Fatal(err)
	}
	catalog.SiteTheme = catalog.DefaultTheme()
	relu2 := site.ThemeFromDB(database)
	if relu2.Version != 0 || !relu2.IsDefault() || relu2.HtmlAttrs() != "" {
		t.Fatalf("après reset : version 0 + défaut + HtmlAttrs vide attendus, got version=%d défaut=%v attrs=%q", relu2.Version, relu2.IsDefault(), relu2.HtmlAttrs())
	}
}
