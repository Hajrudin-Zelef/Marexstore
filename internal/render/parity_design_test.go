package render_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/design"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Parité phase 3 (socle 3.0) : sans design publié, le HTML rendu doit être
// identique à aujourd'hui, octet pour octet (après normalisation des
// versions d'assets, comme la parité v3.25). Les empreintes ci-dessous sont
// celles du code actuel ; toute sous-phase de la phase 3 qui modifierait le
// rendu sans design les fera échouer — c'est le filet de sécurité.
//
// Pour régénérer les empreintes après un changement VOLONTAIRE du rendu :
// go test ./internal/render/ -run TestPariteSansDesign -v (les valeurs
// calculées sont journalisées), puis les recopier ici.
var empreintesSansDesign = map[string]string{
	"index":     "32847560d95c166d3db4b6e4a31e162a37d17eb05fcc25563c3dfbaf76ad8a3b",
	"boutique":  "f495fc5a4f3fc42be743a5e69e16f59c5e38b9c424464803f79ff3fd7be597e3",
	"fiche":     "daa0547cfe2f3cf5005bedb5d5bd5a6fe0e16d04d4f9281dfce47f631a725c6c",
	"categorie": "4b9b44a78db1466a95c5aed6e5b910c3f35b84d1bf02ea70baaff64bbcdaef4b",
}

func hashPage(s string) string {
	return fmt.Sprintf("%x", sha256.Sum256([]byte(s)))
}

func TestPariteSansDesign(t *testing.T) {
	if design.Current() != nil {
		t.Fatal("un design est chargé : la parité sans design exige Current()==nil")
	}
	// Thème par défaut (le global peut avoir été touché par un autre test).
	ancien := catalog.SiteTheme
	catalog.SiteTheme = catalog.DefaultTheme()
	defer func() { catalog.SiteTheme = ancien }()

	s := testStore(t)
	pages := map[string]string{
		"index":     render.Doc("Accueil", render.IndexBodyTheme(s, nil, catalog.DefaultTheme()), "../", ""),
		"boutique":  render.Doc("Boutique", render.BoutiqueBodyQuery(s, nil), "../", ""),
		"categorie": render.Doc("Catégorie", render.CategoryPage(s, "materiels-reseaux"), "../", ""),
	}
	// Une fiche : le premier produit du catalogue.
	var fiche string
	for _, p := range s.Products {
		fiche = render.Doc(p.Nom, render.Header(s, "../", "")+render.FicheBody(s, p)+"\n"+render.Footer(s, "../"), "../", "")
		break
	}
	pages["fiche"] = fiche

	for nom, html := range pages {
		h := hashPage(normPage(html))
		t.Logf("empreinte %s : %s", nom, h)
		if want := empreintesSansDesign[nom]; want == "" {
			t.Errorf("empreinte %s non renseignée (calculée : %s)", nom, h)
			continue
		} else if h != want {
			t.Errorf("parité %s rompue : got %s, want %s (rendu modifié sans design ?)", nom, h, want)
		}
	}
}
