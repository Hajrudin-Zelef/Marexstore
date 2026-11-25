package render

import "testing"

// Le retour admin suit la hiérarchie réelle :
// fiche → sa liste → tableau de bord → site public.
func TestMxBackHierarchie(t *testing.T) {
	cas := []struct {
		path  string
		href  string
		label string
	}{
		// niveau 0 : le tableau de bord remonte au site public
		{"/admin", "/", "le site"},
		{"/admin/", "/", "le site"},

		// sections : retour au tableau de bord
		{"/admin/produits", "/admin", "le tableau de bord"},
		{"/admin/commandes", "/admin", "le tableau de bord"},
		{"/admin/clients", "/admin", "le tableau de bord"},
		{"/admin/medias", "/admin", "le tableau de bord"},
		{"/admin/types", "/admin", "le tableau de bord"},
		{"/admin/reglages", "/admin", "le tableau de bord"},
		{"/admin/systeme", "/admin", "le tableau de bord"},

		// fiches : retour à leur liste, pas au tableau de bord
		{"/admin/produits/nouveau", "/admin/produits", "la liste des produits"},
		{"/admin/produits/toner-hp-26a", "/admin/produits", "la liste des produits"},
		{"/admin/commandes/12", "/admin/commandes", "la liste des commandes"},
		{"/admin/clients/7", "/admin/clients", "la liste des clients"},
	}

	for _, c := range cas {
		got := MxBack(c.path)
		if got.Href != c.href || got.Label != c.label {
			t.Errorf("%s → {%q, %q}, attendu {%q, %q}", c.path, got.Href, got.Label, c.href, c.label)
		}
	}
}

// Le lien doit être rendu avec un libellé accessible et un href échappé.
func TestMxBackHTML(t *testing.T) {
	html := MxBackHTML("/admin/commandes/3")
	if html == "" {
		t.Fatal("aucun lien de retour rendu")
	}
	for _, want := range []string{
		`class="mx-back"`,
		`href="/admin/commandes"`,
		`<span class="lb">Retour</span>`,
		`title="Retour à la liste des commandes"`,
		`aria-label="Retour à la liste des commandes"`,
	} {
		if !contains(html, want) {
			t.Errorf("lien de retour : %q absent dans %s", want, html)
		}
	}
	// Un chemin inconnu ne doit jamais produire de lien cassé.
	if h := MxBackHTML(""); h != "" {
		t.Errorf("chemin vide : lien inattendu %q", h)
	}
}

func contains(s, sub string) bool {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
