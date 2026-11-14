package handlers

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

// A2 — le snapshot expose les données RÉELLES du site (pas les placeholders
// du prototype) : nav, méga-menu, hero, mini-casiers, marques, top ventes,
// pied de page. Lecture seule.
func TestAdminSnapshot(t *testing.T) {
	app, _ := testAppDesign(t)
	list, err := catalog.LoadRaw("../../catalogdata/produits-v3.json")
	if err != nil {
		t.Skip("catalogdata absent (hors dépôt local)")
	}
	catalog.ApplyTaxonomy(list)
	app.Store = catalog.NewStore(list)

	get := func(auth bool) (int, map[string]any) {
		req := httptest.NewRequest(http.MethodGet, "/admin/api/snapshot", nil)
		if auth {
			req.SetBasicAuth("admin", "test1234")
		}
		rec := httptest.NewRecorder()
		app.AdminSnapshot(rec, req)
		var body map[string]any
		_ = json.Unmarshal(rec.Body.Bytes(), &body)
		return rec.Code, body
	}

	if code, _ := get(false); code != http.StatusUnauthorized {
		t.Fatalf("sans auth : status %d, attendu 401", code)
	}
	code, body := get(true)
	if code != http.StatusOK {
		t.Fatalf("avec auth : status %d", code)
	}
	snap, ok := body["snapshot"].(map[string]any)
	if !ok {
		t.Fatal("snapshot absent de la réponse")
	}

	// Nav : 5 liens réels.
	nav, _ := snap["navLinks"].([]any)
	if len(nav) != 5 {
		t.Fatalf("navLinks : %d, attendu 5", len(nav))
	}
	if nav[0].(map[string]any)["label"] != "Accueil" {
		t.Fatalf("navLinks[0] : %v, attendu Accueil", nav[0])
	}

	// Hero : 3 vraies bannières (OUKITEL en premier).
	hero, _ := snap["hero"].([]any)
	if len(hero) != 3 {
		t.Fatalf("hero : %d, attendu 3", len(hero))
	}
	if hero[0].(map[string]any)["titre"] != "OUKITEL" {
		t.Fatalf("hero[0].titre : %v, attendu OUKITEL", hero[0].(map[string]any)["titre"])
	}

	// Minis : 4 vrais mini-casiers (Lenovo T440 en premier).
	minis, _ := snap["minis"].([]any)
	if len(minis) != 4 {
		t.Fatalf("minis : %d, attendu 4", len(minis))
	}
	if minis[0].(map[string]any)["slug"] != "laptop-lenovo-t440" {
		t.Fatalf("minis[0].slug : %v", minis[0].(map[string]any)["slug"])
	}

	// Marques : les 9 du site (pas les 6 du prototype).
	marques, _ := snap["marques"].([]any)
	if len(marques) != 9 {
		t.Fatalf("marques : %d, attendu 9", len(marques))
	}
	var hasHP, hasKyocera bool
	for _, m := range marques {
		if m == "HP" {
			hasHP = true
		}
		if m == "KYOCERA" {
			hasKyocera = true
		}
	}
	if !hasHP || !hasKyocera {
		t.Fatalf("marques : HP/KYOCERA absents (%v)", marques)
	}

	// Top ventes : 8 produits résolus, avec prix boutique.
	top, _ := snap["topVentes"].([]any)
	if len(top) != 8 {
		t.Fatalf("topVentes : %d, attendu 8", len(top))
	}
	if top[0].(map[string]any)["nom"] == nil || top[0].(map[string]any)["nom"] == "" {
		t.Fatal("topVentes[0] : nom absent")
	}
	if top[0].(map[string]any)["prixAffiche"] == nil || top[0].(map[string]any)["prixAffiche"] == "" {
		t.Fatal("topVentes[0] : prixAffiche absent")
	}

	// Produits du sélecteur (A3) : union top ventes + minis + bannières,
	// sans doublon, tous résolus.
	prods, _ := snap["produits"].([]any)
	if len(prods) < 8 {
		t.Fatalf("produits : %d, attendu au moins 8 (top ventes)", len(prods))
	}
	seen := map[string]bool{}
	for _, p := range prods {
		slug, _ := p.(map[string]any)["slug"].(string)
		if slug == "" {
			t.Fatal("produits : slug vide")
		}
		if seen[slug] {
			t.Fatalf("produits : doublon %q", slug)
		}
		seen[slug] = true
	}
	if !seen["laptop-lenovo-t440"] {
		t.Fatal("produits : le mini-casier Lenovo T440 absent")
	}
	if !seen["oukitel-rt3-plus-16-go-ram-128-go"] {
		t.Fatal("produits : le produit de la bannière OUKITEL absent")
	}

	// Pied de page : 5 colonnes (Accès Rapide en premier).
	footer, _ := snap["footer"].([]any)
	if len(footer) != 5 {
		t.Fatalf("footer : %d colonnes, attendu 5", len(footer))
	}
	if footer[0].(map[string]any)["titre"] != "Accès Rapide" {
		t.Fatalf("footer[0].titre : %v", footer[0].(map[string]any)["titre"])
	}

	// Méga-menu : des tops avec groupes.
	mega, _ := snap["megaMenu"].([]any)
	if len(mega) == 0 {
		t.Fatal("megaMenu vide")
	}
	if mega[0].(map[string]any)["groupes"] == nil {
		t.Fatal("megaMenu[0] : groupes absents")
	}

	// Tx : null sans design publié.
	if snap["tx"] != nil {
		t.Fatalf("tx : %v, attendu null sans design publié", snap["tx"])
	}
}

// A2 — le snapshot est injecté dans la page du Studio.
func TestAdminDesignServeStudioInjecteSnapshot(t *testing.T) {
	app, _ := testAppDesign(t)
	list, err := catalog.LoadRaw("../../catalogdata/produits-v3.json")
	if err != nil {
		t.Skip("catalogdata absent (hors dépôt local)")
	}
	catalog.ApplyTaxonomy(list)
	app.Store = catalog.NewStore(list)

	wd, _ := os.Getwd()
	defer func() { _ = os.Chdir(wd) }()
	_ = os.Chdir("../..")

	req := httptest.NewRequest(http.MethodGet, "/admin/design", nil)
	req.SetBasicAuth("admin", "test1234")
	rec := httptest.NewRecorder()
	app.AdminDesign(rec, req)
	if rec.Code != http.StatusOK {
		t.Fatalf("GET /admin/design : status %d", rec.Code)
	}
	body := rec.Body.String()
	if !strings.Contains(body, `window.__STUDIO_SNAPSHOT__=`) {
		t.Fatal("window.__STUDIO_SNAPSHOT__ absent de la page du Studio")
	}
	if !strings.Contains(body, `OUKITEL`) {
		t.Fatal("snapshot injecté : bannière OUKITEL absente")
	}
	if !strings.Contains(body, `laptop-lenovo-t440`) {
		t.Fatal("snapshot injecté : mini-casier Lenovo T440 absent")
	}
}
