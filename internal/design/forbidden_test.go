package design

import "testing"

func TestForbiddenContournements(t *testing.T) {
	ok := `{"front":{},"Tx":{"bandeau":"Bonjour"}}`
	if err := Validate([]byte(ok)); err != nil {
		t.Fatalf("contrôle positif refusé : %v", err)
	}
	bad := map[string]string{
		"svg onload sans espace":     `<svg/onload=alert(1)>`,
		"attribut après guillemet":   `<a href=\"x\"onmouseover=alert(1)>`,
		"fermeture style":            `</style><b>`,
		"fermeture style majuscules": `</STYLE><b>`,
		"import css":                 `@import url(x.css);`,
	}
	for name, v := range bad {
		doc := `{"front":{},"Tx":{"bandeau":"` + v + `"}}`
		if err := Validate([]byte(doc)); err == nil {
			t.Errorf("%s : accepté à tort", name)
		}
	}
}
