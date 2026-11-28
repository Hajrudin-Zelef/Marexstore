package handlers

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// La fiche commande affiche le code promo et la remise quand la commande
// en porte une (sinon rien ne change à l'affichage).
func TestAdminCommandeAffichePromo(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total,client_json,promo_code,remise) VALUES('devis','moov',17000,'{}','REMISE3000',3000)`); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/commandes/1", nil)
	req.SetBasicAuth("admin", "test1234")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	app.AdminCommande(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("fiche : %d", rec.Code)
	}
	for _, want := range []string{"REMISE3000", "3 000", "Moov Money"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("promo absente de la fiche : %q manquant", want)
		}
	}
}

// La fiche commande affiche la livraison : zone, frais, numéro payeur,
// adresse (sinon rien ne change à l'affichage).
func TestAdminCommandeLivraison(t *testing.T) {
	app, database := testAppDesign(t)
	var zoneID int64
	var zoneNom string
	if err := database.QueryRow(`SELECT id, nom FROM livraison_zones WHERE ordre=2`).Scan(&zoneID, &zoneNom); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`UPDATE livraison_zones SET frais=2000 WHERE id=?`, zoneID); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total,client_json,livraison_zone_id,livraison_frais,numero_payeur,adresse) VALUES('devis','orange',12000,'{}',?,?,?,?)`, zoneID, 2000, "0701020304", "Cocody Riviera"); err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodGet, "/admin/commandes/1", nil)
	req.SetBasicAuth("admin", "test1234")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", "1")
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	app.AdminCommande(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("fiche : %d", rec.Code)
	}
	for _, want := range []string{zoneNom, "2 000", "0701020304", "Cocody Riviera"} {
		if !strings.Contains(rec.Body.String(), want) {
			t.Fatalf("livraison absente de la fiche : %q manquant", want)
		}
	}
}
