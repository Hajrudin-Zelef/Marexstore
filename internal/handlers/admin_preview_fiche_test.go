package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
)

// Lot A — la fiche produit suit les textes du design (brouillon et publié).
// Sans surcharge : rendu inchangé (parité v3.25).
func TestAdminDesignPreviewFicheSuitLesTextes(t *testing.T) {
	app, _ := testAppDesign(t)
	app.Store = testStorePourPreview(t)
	// produit avec un prix affiché : sinon pas de bouton « Ajouter au panier »
	var slug string
	for _, prod := range app.Store.Products {
		if prod.PrixAffiche() != nil {
			slug = prod.Slug
			break
		}
	}
	if slug == "" {
		t.Skip("aucun produit avec prix")
	}

	q := url.Values{}
	q.Set("page", "fiche")
	q.Set("slug", slug)
	q.Set("t_specs", "SPECIFICATIONS TEST")
	q.Set("t_ajouter_panier", "AJOUTER TEST")
	q.Set("t_fiche_ref", "REF TEST")
	q.Set("t_tab_avis", "AVIS TEST")
	req := httptest.NewRequest(http.MethodGet, "/admin/design/apercu?"+q.Encode(), nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignPreview(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("aperçu fiche : %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"SPECIFICATIONS TEST", "AJOUTER TEST", "REF TEST", "AVIS TEST"} {
		if !strings.Contains(body, want) {
			t.Fatalf("texte %q absent de la fiche d'aperçu", want)
		}
	}
}

// Sans surcharge, la fiche garde ses libellés actuels, à l'octet près.
func TestAdminDesignPreviewFicheDefautsInchanges(t *testing.T) {
	app, _ := testAppDesign(t)
	app.Store = testStorePourPreview(t)
	slug := app.Store.Products[0].Slug
	q := url.Values{}
	q.Set("page", "fiche")
	q.Set("slug", slug)
	req := httptest.NewRequest(http.MethodGet, "/admin/design/apercu?"+q.Encode(), nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignPreview(rec, req)
	body := rec.Body.String()
	for _, want := range []string{"Caractéristiques", "Demander un devis", "Produits similaires", "Consommables", "Avis (0)", "Réf."} {
		if !strings.Contains(body, want) {
			t.Fatalf("libellé par défaut %q perdu", want)
		}
	}
}
