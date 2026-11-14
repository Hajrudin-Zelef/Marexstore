package catalog_test

import (
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// P1 — le serveur gagne : data-theme-version n'est émis que pour un thème
// publié (non défaut) avec une version > 0. Jamais pour le thème par défaut
// (parité v3.25 intacte).
func TestThemeVersionAttr(t *testing.T) {
	// Thème publié (non défaut) avec version : l'attribut est émis.
	th := catalog.Theme{Couleur: "gazonnier", Mode: "auto", Police: "inter", Sections: append([]string{}, catalog.DefaultSections...), Version: 3}
	attrs := th.HtmlAttrs()
	if !strings.Contains(attrs, `data-theme-version="3"`) {
		t.Fatalf("version attendue dans HtmlAttrs, got %q", attrs)
	}
	if !strings.Contains(attrs, `data-color="gazonnier"`) {
		t.Fatalf("couleur attendue dans HtmlAttrs, got %q", attrs)
	}
	// Thème par défaut avec version (publié puis réinitialisé en défaut) :
	// aucune sortie (parité).
	def := catalog.DefaultTheme()
	def.Version = 5
	if got := def.HtmlAttrs(); got != "" {
		t.Fatalf("thème par défaut : HtmlAttrs doit être vide, got %q", got)
	}
	if !def.IsDefault() {
		t.Fatalf("la version ne doit pas rendre le thème non défaut")
	}
	// Thème non défaut sans version (publié avant P1) : pas d'attribut version.
	old := catalog.Theme{Couleur: "liquid", Mode: "auto", Police: "inter", Sections: append([]string{}, catalog.DefaultSections...)}
	if strings.Contains(old.HtmlAttrs(), "data-theme-version") {
		t.Fatalf("sans version, pas d'attribut data-theme-version, got %q", old.HtmlAttrs())
	}
}
