package render_test

import (
	"crypto/sha256"
	"fmt"
	"testing"

	"path/filepath"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/db"
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
	"index":     "6708f70ed5ea98fcb3b78c7695b5f11f430f08780b8399aa86aaf67801665851",
	"boutique":  "5d65784bf653f87ab6fd73a4fa6ee4e01b16e90408617094bb180710af2772bd",
	"fiche":     "0edbf64bff08001ccd0193fe7b9a03c56ce4b1380c3ad6dc0c0bad418e95a1d1",
	"categorie": "cc340501490b8ccf8df48344a2537178991fc9788d683132182bf6c5c353b914",
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

// Phase 3.1 : avec un design publié, les textes Tx et les bascules de
// visibilité s'appliquent au rendu (en-tête + pied de page).
func TestDesign31TextesEtVisibilite(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	// Design : texte de bienvenue personnalisé, téléphone de la topbar
	// masqué, newsletter masquée, copyright personnalisé.
	doc := `{"front":{"accent":"#2563eb"},"Tx":{"bienvenue":"Bienvenue chez Mediatec","newsletter_titre":"Restez informé","copyright":"© Test 3.1"},"header":{"showPhone":false},"footer":{"showNewsletter":false}}`
	if _, err := design.Save(database, []byte(doc), "test"); err != nil {
		t.Fatal(err)
	}
	if err := design.LoadActive(database); err != nil {
		t.Fatal(err)
	}
	// Nettoyage : recharger sans design (base vide -> actif nil).
	t.Cleanup(func() {
		empty, err := db.Open(filepath.Join(t.TempDir(), "vide.db"))
		if err != nil {
			return
		}
		defer empty.Close()
		if err := db.EnsureSchema(empty, "../../db/migrations"); err != nil {
			return
		}
		_ = design.LoadActive(empty)
	})
	ancien := catalog.SiteTheme
	catalog.SiteTheme = catalog.DefaultTheme()
	defer func() { catalog.SiteTheme = ancien }()

	s := testStore(t)
	html := render.Doc("Accueil", render.IndexBodyTheme(s, nil, catalog.DefaultTheme()), "../", "")

	// Texte Tx appliqué.
	if !strings.Contains(html, "Bienvenue chez Mediatec") {
		t.Error("bienvenue personnalisé absent du rendu")
	}
	if strings.Contains(html, "Bienvenue dans notre boutique") {
		t.Error("bienvenue d'origine toujours présent")
	}
	if !strings.Contains(html, "© Test 3.1") {
		t.Error("copyright personnalisé absent du rendu")
	}
	// Téléphone de la topbar masqué (showPhone=false) ; la topbar reste.
	if !strings.Contains(html, `class="topbar"`) {
		t.Error("topbar absente (showTopbar=true par défaut)")
	}
	// Le téléphone apparaît aussi dans le pied de page — vérifier la
	// topbar seule.
	topbar := html[strings.Index(html, `class="topbar"`):]
	topbar = topbar[:strings.Index(topbar, `</div></div>`)]
	if strings.Contains(topbar, catalog.Tel) {
		t.Error("téléphone toujours présent dans la topbar malgré showPhone=false")
	}
	// Newsletter masquée (showNewsletter=false).
	if strings.Contains(html, `class="news"`) {
		t.Error("newsletter toujours rendue malgré showNewsletter=false")
	}
	if strings.Contains(html, "Restez informé") {
		t.Error("titre newsletter présent malgré showNewsletter=false")
	}
	// Les colonnes du pied de page restent (showCols=true par défaut).
	if !strings.Contains(html, "Accès Rapide") {
		t.Error("colonnes du pied de page absentes (showCols=true par défaut)")
	}
	// Les icônes de l'en-tête restent (show*=true par défaut).
	if !strings.Contains(html, "Boutique") {
		t.Error("icône Boutique absente (showBoutique=true par défaut)")
	}
}

// Phase 3.1 (visibilité) : masquer la topbar entière retire le bloc.
func TestDesign31TopbarMasquee(t *testing.T) {
	dir := t.TempDir()
	database, err := db.Open(filepath.Join(dir, "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { database.Close() })
	if err := db.EnsureSchema(database, "../../db/migrations"); err != nil {
		t.Fatal(err)
	}
	doc := `{"header":{"showTopbar":false}}`
	if _, err := design.Save(database, []byte(doc), "test"); err != nil {
		t.Fatal(err)
	}
	if err := design.LoadActive(database); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		empty, err := db.Open(filepath.Join(t.TempDir(), "vide.db"))
		if err != nil {
			return
		}
		defer empty.Close()
		if err := db.EnsureSchema(empty, "../../db/migrations"); err != nil {
			return
		}
		_ = design.LoadActive(empty)
	})
	ancien := catalog.SiteTheme
	catalog.SiteTheme = catalog.DefaultTheme()
	defer func() { catalog.SiteTheme = ancien }()

	s := testStore(t)
	html := render.Doc("Accueil", render.IndexBodyTheme(s, nil, catalog.DefaultTheme()), "../", "")
	if strings.Contains(html, `class="topbar"`) {
		t.Error("topbar toujours rendue malgré showTopbar=false")
	}
	if !strings.Contains(html, `class="main"`) {
		t.Error("en-tête principal absent (il ne dépend pas de la topbar)")
	}
}
