package handlers

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// XSS filtre vue : le param brut ne doit jamais casser le href.
func TestAdminCommandesVueEchappeeHref(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total,client_json,paiement_statut) VALUES('paye','orange',5000,'{}','recu')`); err != nil {
		t.Fatal(err)
	}
	mal := `"><script>alert(1)</script>`
	req := httptest.NewRequest(http.MethodGet, "/admin/commandes?vue="+mal+"&paiement=tous", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminCommandes(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("liste : %d", rec.Code)
	}
	body := rec.Body.String()
	if strings.Contains(body, `"><script>`) {
		t.Fatalf("href brut non echappe : payload XSS present tel quel")
	}
	if !strings.Contains(body, "%22%3E%3Cscript%3E") {
		t.Fatalf("vue echappee attendue (url.QueryEscape) absente")
	}
}

// Fiche devis (total NULL) : la ligne paiement affiche « Total non défini ».
func TestAdminCommandeFicheDevisTotalNonDefini(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total,client_json,paiement_statut) VALUES('devis','orange',NULL,'{}','en_attente')`); err != nil {
		t.Fatal(err)
	}
	var id int64
	if err := database.QueryRow(`SELECT id FROM orders ORDER BY id DESC LIMIT 1`).Scan(&id); err != nil {
		t.Fatal(err)
	}
	rec := getFicheCommande(t, app, id)
	if rec.Code != http.StatusOK {
		t.Fatalf("fiche : %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, "Total non défini") {
		t.Fatalf("fiche devis : « Total non défini » manquant")
	}
	if strings.Contains(body, "devis via") {
		t.Fatalf("fiche devis : « devis via » encore present dans la ligne paiement")
	}
}
