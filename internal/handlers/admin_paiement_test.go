package handlers

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

// Commande de test pour le suivi paiement : statut de commande fixe,
// moyen/paiement_statut/total/payeur paramétrables.
func insertPaiementCmd(t *testing.T, database *sql.DB, moyen, pstatut string, total int64, payeur string) int64 {
	t.Helper()
	res, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total,client_json,numero_payeur,paiement_statut) VALUES('whatsapp_recu',?,?,?,?,?)`,
		moyen, total, `{"nom":"Awa","tel":"0700"}`, payeur, pstatut)
	if err != nil {
		t.Fatal(err)
	}
	id, _ := res.LastInsertId()
	return id
}

func paiementStatutDB(t *testing.T, database *sql.DB, id int64) string {
	t.Helper()
	var got string
	if err := database.QueryRow(`SELECT paiement_statut FROM orders WHERE id=?`, id).Scan(&got); err != nil {
		t.Fatal(err)
	}
	return got
}

// POST /admin/commandes/{id}/paiement {action} en contournant le routeur
// (contexte chi minimal, comme les autres tests admin).
func postPaiementAction(t *testing.T, app *App, id int64, action, csrf string, avecAuth bool) *httptest.ResponseRecorder {
	t.Helper()
	form := url.Values{}
	form.Set("csrf", csrf)
	form.Set("action", action)
	req := httptest.NewRequest(http.MethodPost, "/admin/commandes/"+fmt.Sprint(id)+"/paiement", strings.NewReader(form.Encode()))
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	if avecAuth {
		req.SetBasicAuth("admin", "test1234")
	}
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprint(id))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	app.AdminCommandePaiement(rec, req)
	return rec
}

func getFicheCommande(t *testing.T, app *App, id int64) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(http.MethodGet, "/admin/commandes/"+fmt.Sprint(id), nil)
	req.SetBasicAuth("admin", "test1234")
	rctx := chi.NewRouteContext()
	rctx.URLParams.Add("id", fmt.Sprint(id))
	req = req.WithContext(context.WithValue(req.Context(), chi.RouteCtxKey, rctx))
	rec := httptest.NewRecorder()
	app.AdminCommande(rec, req)
	return rec
}

// Les 4 moyens mobile en en_attente passent à reçu via marquer_recu.
func TestAdminPaiementMarquerRecuTousMobiles(t *testing.T) {
	for _, moyen := range []string{"orange", "mtn", "wave", "moov"} {
		t.Run(moyen, func(t *testing.T) {
			app, database := testAppDesign(t)
			id := insertPaiementCmd(t, database, moyen, "en_attente", 12000, "0701020304")
			rec := postPaiementAction(t, app, id, "marquer_recu", app.adminCSRF(), true)
			if rec.Code != http.StatusSeeOther {
				t.Fatalf("marquer_recu (%s) : status %d, attendu 303 (corps %q)", moyen, rec.Code, rec.Body.String())
			}
			if loc := rec.Header().Get("Location"); !strings.Contains(loc, "msg=") || strings.Contains(loc, "err") {
				t.Fatalf("marquer_recu (%s) : redirect %q, attendu un msg de succès", moyen, loc)
			}
			if got := paiementStatutDB(t, database, id); got != "recu" {
				t.Fatalf("marquer_recu (%s) : paiement_statut=%q, attendu recu", moyen, got)
			}
		})
	}
}

// À la livraison : a_encaisser passe à encaissé via marquer_encaisse.
func TestAdminPaiementMarquerEncaisse(t *testing.T) {
	app, database := testAppDesign(t)
	id := insertPaiementCmd(t, database, "cod", "a_encaisser", 8000, "")
	rec := postPaiementAction(t, app, id, "marquer_encaisse", app.adminCSRF(), true)
	if rec.Code != http.StatusSeeOther {
		t.Fatalf("marquer_encaisse : status %d, attendu 303", rec.Code)
	}
	if got := paiementStatutDB(t, database, id); got != "encaisse" {
		t.Fatalf("marquer_encaisse : paiement_statut=%q, attendu encaisse", got)
	}
}

// Machine à états stricte : seules en_attente→recu et a_encaisser→encaisse
// passent ; tout le reste = 400, commande inchangée.
func TestAdminPaiementTransitionsInterdites(t *testing.T) {
	cases := []struct {
		nom    string
		depart string
		action string
	}{
		{"encaisser_sur_attente", "en_attente", "marquer_encaisse"},
		{"recu_sur_a_encaisser", "a_encaisser", "marquer_recu"},
		{"encaisser_sur_recu", "recu", "marquer_encaisse"},
		{"recu_sur_encaisse", "encaisse", "marquer_recu"},
		{"action_inconnue", "en_attente", "marquer_perdu"},
		{"action_vide", "en_attente", ""},
	}
	for _, tc := range cases {
		t.Run(tc.nom, func(t *testing.T) {
			app, database := testAppDesign(t)
			id := insertPaiementCmd(t, database, "orange", tc.depart, 5000, "0701020304")
			rec := postPaiementAction(t, app, id, tc.action, app.adminCSRF(), true)
			if rec.Code != http.StatusBadRequest {
				t.Fatalf("%s : status %d, attendu 400", tc.nom, rec.Code)
			}
			if got := paiementStatutDB(t, database, id); got != tc.depart {
				t.Fatalf("%s : paiement_statut=%q, attendu inchangé %q", tc.nom, got, tc.depart)
			}
		})
	}
}

// Répéter l'action sur un état déjà atteint est idempotent (303, inchangé).
func TestAdminPaiementIdempotent(t *testing.T) {
	app, database := testAppDesign(t)
	idRecu := insertPaiementCmd(t, database, "wave", "en_attente", 5000, "0701020304")
	if rec := postPaiementAction(t, app, idRecu, "marquer_recu", app.adminCSRF(), true); rec.Code != http.StatusSeeOther {
		t.Fatalf("1er marquer_recu : %d", rec.Code)
	}
	if rec := postPaiementAction(t, app, idRecu, "marquer_recu", app.adminCSRF(), true); rec.Code != http.StatusSeeOther {
		t.Fatalf("2e marquer_recu : status %d, attendu 303 idempotent", rec.Code)
	}
	if got := paiementStatutDB(t, database, idRecu); got != "recu" {
		t.Fatalf("après double marquer_recu : %q, attendu recu", got)
	}
	idEnc := insertPaiementCmd(t, database, "cod", "a_encaisser", 5000, "")
	if rec := postPaiementAction(t, app, idEnc, "marquer_encaisse", app.adminCSRF(), true); rec.Code != http.StatusSeeOther {
		t.Fatalf("1er marquer_encaisse : %d", rec.Code)
	}
	if rec := postPaiementAction(t, app, idEnc, "marquer_encaisse", app.adminCSRF(), true); rec.Code != http.StatusSeeOther {
		t.Fatalf("2e marquer_encaisse : status %d, attendu 303 idempotent", rec.Code)
	}
	if got := paiementStatutDB(t, database, idEnc); got != "encaisse" {
		t.Fatalf("après double marquer_encaisse : %q, attendu encaisse", got)
	}
}

// Auth + CSRF + commande inconnue : 401 / 403 / 404 honnêtes.
func TestAdminPaiementGardefous(t *testing.T) {
	app, database := testAppDesign(t)
	id := insertPaiementCmd(t, database, "orange", "en_attente", 5000, "0701020304")
	if rec := postPaiementAction(t, app, id, "marquer_recu", app.adminCSRF(), false); rec.Code != http.StatusUnauthorized {
		t.Fatalf("sans auth : %d, attendu 401", rec.Code)
	}
	if rec := postPaiementAction(t, app, id, "marquer_recu", "mauvais-jeton", true); rec.Code != http.StatusForbidden {
		t.Fatalf("mauvais CSRF : %d, attendu 403", rec.Code)
	}
	if got := paiementStatutDB(t, database, id); got != "en_attente" {
		t.Fatalf("après refus : %q, attendu en_attente inchangé", got)
	}
	if rec := postPaiementAction(t, app, 99999, "marquer_recu", app.adminCSRF(), true); rec.Code != http.StatusNotFound {
		t.Fatalf("commande inconnue : %d, attendu 404", rec.Code)
	}
}

// Chaque transition réussie est auditée (traçabilité T1).
func TestAdminPaiementAudit(t *testing.T) {
	app, database := testAppDesign(t)
	id := insertPaiementCmd(t, database, "mtn", "en_attente", 5000, "0701020304")
	if rec := postPaiementAction(t, app, id, "marquer_recu", app.adminCSRF(), true); rec.Code != http.StatusSeeOther {
		t.Fatalf("marquer_recu : %d", rec.Code)
	}
	var n int
	if err := database.QueryRow(`SELECT COUNT(*) FROM admin_audit WHERE action='commande.paiement' AND objet LIKE ?`, "#"+fmt.Sprint(id)+"%").Scan(&n); err != nil {
		t.Fatal(err)
	}
	if n != 1 {
		t.Fatalf("audit commande.paiement : %d ligne(s), attendu 1", n)
	}
}

// Fiche : bloc « Paiement : <Statut> — <total> via <moyen> (payeur) »
// + bouton d'action suivante ; ligne legacy (défaut) lue « En attente ».
func TestAdminCommandeFichePaiement(t *testing.T) {
	app, database := testAppDesign(t)
	id := insertPaiementCmd(t, database, "orange", "en_attente", 12000, "0701020304")
	rec := getFicheCommande(t, app, id)
	if rec.Code != http.StatusOK {
		t.Fatalf("fiche : %d", rec.Code)
	}
	body := rec.Body.String()
	for _, want := range []string{"En attente", "12 000", "Orange Money", "0701020304", "marquer_recu", "/admin/commandes/" + fmt.Sprint(id) + "/paiement"} {
		if !strings.Contains(body, want) {
			t.Fatalf("fiche paiement : %q manquant", want)
		}
	}
	// Après réception : « Reçu », plus de bouton marquer_recu.
	if rec2 := postPaiementAction(t, app, id, "marquer_recu", app.adminCSRF(), true); rec2.Code != http.StatusSeeOther {
		t.Fatalf("marquer_recu : %d", rec2.Code)
	}
	body2 := getFicheCommande(t, app, id).Body.String()
	if !strings.Contains(body2, "Reçu") {
		t.Fatalf("fiche après réception : « Reçu » manquant")
	}
	if strings.Contains(body2, "marquer_recu") {
		t.Fatalf("fiche après réception : bouton marquer_recu encore présent")
	}
	// Ligne sans valeur explicite (défaut de colonne) : lue « En attente ».
	if _, err := database.Exec(`INSERT INTO orders(statut,moyen_paiement,total,client_json) VALUES('whatsapp_recu','moov',9000,'{}')`); err != nil {
		t.Fatal(err)
	}
	var legacyID int64
	if err := database.QueryRow(`SELECT id FROM orders ORDER BY id DESC LIMIT 1`).Scan(&legacyID); err != nil {
		t.Fatal(err)
	}
	body3 := getFicheCommande(t, app, legacyID).Body.String()
	if !strings.Contains(body3, "En attente") {
		t.Fatalf("fiche legacy (défaut) : « En attente » manquant")
	}
}

// Liste : pastille par ligne + filtre ?paiement=a_verifier|a_encaisser|tous.
func TestAdminCommandesFiltrePaiement(t *testing.T) {
	app, database := testAppDesign(t)
	idAttente := insertPaiementCmd(t, database, "orange", "en_attente", 12000, "0701020304")
	idEncaisser := insertPaiementCmd(t, database, "cod", "a_encaisser", 8000, "")
	idRecu := insertPaiementCmd(t, database, "wave", "recu", 5000, "0701")
	lien := func(id int64) string { return `href="/admin/commandes/` + fmt.Sprint(id) + `"` }
	get := func(qs string) string {
		t.Helper()
		req := httptest.NewRequest(http.MethodGet, "/admin/commandes"+qs, nil)
		req.SetBasicAuth("admin", "test1234")
		rec := httptest.NewRecorder()
		app.AdminCommandes(rec, req)
		if rec.Code != http.StatusOK {
			t.Fatalf("liste %s : %d", qs, rec.Code)
		}
		return rec.Body.String()
	}
	aEnc := get("?paiement=a_encaisser")
	if !strings.Contains(aEnc, lien(idEncaisser)) || !strings.Contains(aEnc, "À encaisser") {
		t.Fatalf("filtre a_encaisser : commande/ligne attendue absente")
	}
	for _, exclu := range []int64{idAttente, idRecu} {
		if strings.Contains(aEnc, lien(exclu)) {
			t.Fatalf("filtre a_encaisser : commande #%d présente à tort", exclu)
		}
	}
	aVer := get("?paiement=a_verifier")
	if !strings.Contains(aVer, lien(idAttente)) || !strings.Contains(aVer, "En attente") {
		t.Fatalf("filtre a_verifier : commande/ligne attendue absente")
	}
	for _, exclu := range []int64{idEncaisser, idRecu} {
		if strings.Contains(aVer, lien(exclu)) {
			t.Fatalf("filtre a_verifier : commande #%d présente à tort", exclu)
		}
	}
	tous := get("?paiement=tous")
	for _, id := range []int64{idAttente, idEncaisser, idRecu} {
		if !strings.Contains(tous, lien(id)) {
			t.Fatalf("filtre tous : commande #%d absente", id)
		}
	}
}
