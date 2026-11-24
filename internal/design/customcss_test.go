package design

import "testing"

func TestCustomCSSInterdits(t *testing.T) {
	ok := `{"front":{"customCSS":"body{background:url(/images/x.png)}"}}`
	if err := Validate([]byte(ok)); err != nil {
		t.Fatalf("contrôle positif refusé : %v", err)
	}
	bad := map[string]string{
		"url http":      `body{background:url(https://evil.example/x.png)}`,
		"url protocole": `body{background:url(//evil.example/x.png)}`,
		"url espaces":   `body{background:url( \"https://evil.example/x.png\")}`,
		"expression":    `a{width:expression(alert(1))}`,
		"binding":       `a{-moz-binding:url(x)}`,
	}
	for name, v := range bad {
		doc := `{"front":{"customCSS":"` + v + `"}}`
		if err := Validate([]byte(doc)); err == nil {
			t.Errorf("%s : accepté à tort", name)
		}
	}
}
