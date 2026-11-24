package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// testStorePourPreview : petit store réel (le rendu d'accueil a besoin de
// produits) ; ignoré si le catalogue n'est pas présent.
func testStorePourPreview(t *testing.T) *catalog.Store {
	t.Helper()
	list, err := catalog.LoadRaw("../../catalogdata/produits-v3.json")
	if err != nil {
		t.Skip("catalogdata absent")
	}
	catalog.ApplyTaxonomy(list)
	return catalog.NewStore(list)
}

// L'aperçu live doit refléter un texte NON publié : le Studio envoie
// t_<clé>=<valeur> (uniquement les textes modifiés), le serveur applique la
// surcharge pour cette requête seulement (jamais le design publié).
func TestAdminDesignPreviewAppliqueTextesBrouillon(t *testing.T) {
	app, _ := testAppDesign(t)
	app.Store = testStorePourPreview(t)

	q := url.Values{}
	q.Set("page", "accueil")
	q.Set("couleur", "gazonnier")
	q.Set("t_bienvenue", "BIENVENUE BROUILLON TEST")
	req := httptest.NewRequest(http.MethodGet, "/admin/design/apercu?"+q.Encode(), nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignPreview(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("aperçu : status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "BIENVENUE BROUILLON TEST") {
		t.Fatal("le texte en brouillon n'apparaît pas dans l'aperçu")
	}
	if strings.Contains(body, "Bienvenue dans notre boutique") {
		t.Fatal("le texte publié reste affiché (surcharge ignorée)")
	}
}

// Sans surcharge, l'aperçu garde le texte publié/défaut : pas de régression.
func TestAdminDesignPreviewSansSurchargeInchange(t *testing.T) {
	app, _ := testAppDesign(t)
	app.Store = testStorePourPreview(t)
	q := url.Values{}
	q.Set("page", "accueil")
	req := httptest.NewRequest(http.MethodGet, "/admin/design/apercu?"+q.Encode(), nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesignPreview(rec, req)
	body := rec.Body.String()
	if !strings.Contains(body, "Bienvenue dans notre boutique") {
		t.Fatal("texte par défaut perdu sans surcharge")
	}
}
