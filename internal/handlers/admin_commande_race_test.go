package handlers

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Deux admins valident LA MÊME transition au même instant depuis "devis"
// (ou un double-clic) : une seule doit gagner. Sans compare-and-swap en
// base, les deux UPDATE inconditionnels passent et la commande porte 2
// événements identiques (voire un stock décrémenté deux fois à "paye").
// Note : on utilise le même `vers` des deux côtés car deux transitions
// divergentes resteraient toutes deux légitimes en séquence (ex. annule
// est valide depuis devis ET depuis whatsapp_recu) — ce test pinne le
// doublon, pas l'arbitrage entre deux choix différents.
func TestAdminCommandeTransitionsConcurrentesUnSeulGagnant(t *testing.T) {
	app, database := testAppDesign(t)
	if _, err := database.Exec(`INSERT INTO products(slug,nom,prix,stock) VALUES('p1','P1',1000,50)`); err != nil {
		t.Fatal(err)
	}
	poste := func(oid int64, vers string, start <-chan struct{}, wg *sync.WaitGroup, code *int) {
		defer wg.Done()
		<-start
		form := url.Values{"csrf": {app.adminCSRF()}, "statut": {vers}}
		req := httptest.NewRequest(http.MethodPost, fmt.Sprintf("/admin/commandes/%d", oid), strings.NewReader(form.Encode()))
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
		req.SetBasicAuth("admin", "test1234")
		rctx := chi.NewRouteContext()
		rctx.URLParams.Add("id", fmt.Sprint(oid))
		req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
		rec := httptest.NewRecorder()
		app.AdminCommande(rec, req)
		*code = rec.Code
	}

	const paires = 10
	for i := 0; i < paires; i++ {
		res, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total) VALUES('devis','whatsapp',1000)`)
		if err != nil {
			t.Fatal(err)
		}
		oid, _ := res.LastInsertId()
		start := make(chan struct{})
		var wg sync.WaitGroup
		var c1, c2 int
		wg.Add(2)
		go poste(oid, "whatsapp_recu", start, &wg, &c1)
		go poste(oid, "whatsapp_recu", start, &wg, &c2)
		close(start)
		wg.Wait()
		if c1 != http.StatusSeeOther || c2 != http.StatusSeeOther {
			t.Fatalf("paire %d : codes %d/%d, attendu 303/303", i, c1, c2)
		}
	}
	var nEvents int
	if err := database.QueryRow(`SELECT COUNT(*) FROM order_events`).Scan(&nEvents); err != nil {
		t.Fatal(err)
	}
	if nEvents != paires {
		t.Fatalf("course non arbitrée : %d événements pour %d paires (attendu 1 gagnant par paire)", nEvents, paires)
	}
	var nWa int
	_ = database.QueryRow(`SELECT COUNT(*) FROM orders WHERE statut='whatsapp_recu'`).Scan(&nWa)
	if nWa != paires {
		t.Fatalf("états finaux incohérents : %d whatsapp_recu pour %d paires", nWa, paires)
	}
}
