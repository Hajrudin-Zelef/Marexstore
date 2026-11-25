package catalog

import (
	"encoding/json"
	"testing"
)

func gp(slug, nom, marque, specs string) *RawProduct {
	p := &RawProduct{Slug: slug, Nom: nom, Marque: marque}
	if specs != "" {
		p.Specs = json.RawMessage(specs)
	}
	return p
}

func TestGammeKey(t *testing.T) {
	cas := []struct {
		p    *RawProduct
		want string
	}{
		{gp("a", "Cisco Catalyst 9300 (48P, 24UX)", "Cisco", `[{"label":"Série","valeur":"Catalyst 9300"}]`), "cisco|catalyst9300"},
		{gp("b", "Cartouche d'encre HP 305XL", "HP", ""), "hp|305"},
		{gp("c", "Cartouche d'encre HP 305", "HP", ""), "hp|305"},
		{gp("d", "Bouteille d'encre Canon GI-490", "Canon", ""), "canon|gi"},
		{gp("e", "Écran Dell E1715S", "Dell", ""), "dell|e"},
		{gp("f", "Câble Ethernet RJ45 1 m", "Générique", ""), ""},
		{gp("g", "Kyocera ECOSYS MA2101cwfx", "Kyocera", ""), "kyocera|ecosys"},
	}
	for _, c := range cas {
		if got := GammeKey(c.p); got != c.want {
			t.Fatalf("%s : GammeKey=%q, attendu %q", c.p.Slug, got, c.want)
		}
	}
}

func TestGammeSiblingsEtLabel(t *testing.T) {
	a := gp("a", "Cartouche d'encre HP 305", "HP", "")
	b := gp("b", "Cartouche d'encre HP 305XL", "HP", "")
	c := gp("c", "Cartouche d'encre HP 653", "HP", "")
	s := NewStore([]*RawProduct{a, b, c})
	sibs := s.GammeSiblings(a)
	if len(sibs) != 1 || sibs[0].Slug != "b" {
		t.Fatalf("siblings de a : %v, attendu [b]", sibs)
	}
	if got := GammeChipLabel(b); got != "305XL" {
		t.Fatalf("chip label : %q, attendu 305XL", got)
	}
	if got := GammeLabel(a); got != "HP 305" {
		t.Fatalf("gamme label : %q, attendu HP 305", got)
	}
}
