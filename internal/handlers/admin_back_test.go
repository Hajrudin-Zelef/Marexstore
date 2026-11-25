package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Chaque page admin rendue avec le châssis commun doit porter un retour
// hiérarchique : sections → tableau de bord, fiches → leur liste.
func TestAdminRetourHierarchique(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,prix,dispo) VALUES('toner-retour','Toner retour',5000,'en stock')`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total) VALUES('devis','whatsapp',1000)`); err != nil {
		t.Fatal(err)
	}
	if _, err := database.Exec(`INSERT INTO users(email,pass_hash,nom) VALUES('client@test.ci','x','Client Test')`); err != nil {
		t.Fatal(err)
	}

	cas := []struct {
		path     string
		wantHref string
	}{
		{"/admin", `href="/"`},
		{"/admin/produits", `href="/admin"`},
		{"/admin/produits/nouveau", `href="/admin/produits"`},
		{"/admin/produits/toner-retour", `href="/admin/produits"`},
		{"/admin/commandes", `href="/admin"`},
		{"/admin/commandes/1", `href="/admin/commandes"`},
		{"/admin/clients", `href="/admin"`},
		{"/admin/clients/1", `href="/admin/clients"`},
		{"/admin/medias", `href="/admin"`},
		{"/admin/types", `href="/admin"`},
		{"/admin/reglages", `href="/admin"`},
		{"/admin/systeme", `href="/admin"`},
	}

	// Routeur réel : les handlers lisent chi.URLParam (slug, id).
	r := chi.NewRouter()
	r.Get("/admin", app.Admin)
	r.Get("/admin/produits", app.AdminProduits)
	r.Get("/admin/produits/nouveau", app.AdminProduitEdit)
	r.Get("/admin/produits/{slug}", app.AdminProduitEdit)
	r.Get("/admin/commandes", app.AdminCommandes)
	r.Get("/admin/commandes/{id}", app.AdminCommande)
	r.Get("/admin/clients", app.AdminClients)
	r.Get("/admin/clients/{id}", app.AdminClient)
	r.Get("/admin/medias", app.AdminMedias)
	r.Get("/admin/types", app.AdminTypes)
	r.Get("/admin/reglages", app.AdminReglages)
	r.Get("/admin/systeme", app.AdminSysteme)

	for _, c := range cas {
		req := httptest.NewRequest(http.MethodGet, c.path, nil)
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		r.ServeHTTP(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s : status %d", c.path, rec.Code)
		}
		body := rec.Body.String()
		if !strings.Contains(body, `class="mx-back"`) {
			t.Fatalf("%s : aucun bouton retour", c.path)
		}
		if !strings.Contains(body, c.wantHref) {
			t.Fatalf("%s : retour vers %s attendu", c.path, c.wantHref)
		}
	}
}

// Le retour doit rester une navigation GET sûre : pas de javascript:, pas
// d'URL pilotée par la requête (pas de redirection ouverte).
func TestAdminRetourPasDOpenRedirect(t *testing.T) {
	app, _ := testAppDesign(t)
	req := httptest.NewRequest(http.MethodGet, "/admin/produits?next="+url.QueryEscape("//evil.example"), nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminProduits(rec, req)
	body := rec.Body.String()
	if strings.Contains(body, "evil.example") {
		t.Fatal("le paramètre next influence le lien de retour (redirection ouverte)")
	}
	if strings.Contains(body, "javascript:") {
		t.Fatal("lien de retour en javascript:")
	}
}
