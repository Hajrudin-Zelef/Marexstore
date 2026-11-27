package render_test

import (
	"os"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Le menu avatar de l'en-tête : 7 entrées vers des pages réelles +
// déconnexion via le flux existant.
func TestTopAvatarMenu(t *testing.T) {
	raw, err := os.ReadFile("../../templates/static/_top.html")
	if err != nil {
		t.Skip("template absent")
	}
	s := string(raw)
	for _, want := range []string{
		`data-avatar-menu`,
		`compte.html#informations`, // Profil + Réglages
		`parametres.html`,           // Préférences
		`contact.html`,              // Obtenir de l'aide
		`faq.html`,                  // FAQ
		`apropos.html`,              // À propos
		`data-logout`,               // Déconnexion
		`aria-expanded`,
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("menu avatar : %q manquant dans _top.html", want)
		}
	}
}

// Même menu dans l'en-tête Go (pages dynamiques : accueil, boutique,
// fiche, catégories) — les deux rendus doivent rester synchronisés.
func TestHeaderAvatarMenu(t *testing.T) {
	s := testStore(t)
	h := render.Header(s, "../", "")
	for _, want := range []string{
		`data-avatar-menu`,
		`compte.html#informations`,
		`parametres.html`,
		`contact.html`,
		`faq.html`,
		`apropos.html`,
		`data-logout`,
		`data-avatar`,
	} {
		if !strings.Contains(h, want) {
			t.Fatalf("menu avatar : %q manquant dans Header() (render2.go désynchronisé de _top.html)", want)
		}
	}
}

// Même principe côté admin : l'avatar de la sidebar ouvre un menu
// (tableau de bord, voir le site, réglages, système, déconnexion).
func TestAdminAvatarMenu(t *testing.T) {
	h := render.MxSidebar("", "/admin", map[string]int{})
	for _, want := range []string{
		`data-admin-menu`,
		`href="/admin"`,
		`href="/"`,
		`href="/admin/reglages"`,
		`href="/admin/systeme"`,
		`data-admin-logout`,
		`aria-expanded`,
	} {
		if !strings.Contains(h, want) {
			t.Fatalf("menu avatar admin : %q manquant dans MxSidebar()", want)
		}
	}
}

// La page FAQ existe et couvre l'essentiel (commande, livraison,
// paiement + promo, suivi, compte, retours).
func TestFaqPage(t *testing.T) {
	raw, err := os.ReadFile("../../templates/static/faq.html")
	if err != nil {
		t.Fatalf("template faq.html manquant : %v", err)
	}
	s := string(raw)
	for _, want := range []string{"<details>", "promo", "livraison", "suivi-commande.html", "compte.html"} {
		if !strings.Contains(strings.ToLower(s), strings.ToLower(want)) {
			t.Fatalf("faq.html : %q manquant", want)
		}
	}
}
