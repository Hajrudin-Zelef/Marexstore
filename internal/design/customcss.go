package design

import (
	"fmt"
	"regexp"
)

// forbiddenCSS : le CSS personnalisé ne charge aucune ressource distante
// (@import, url(http…), url(//…)) et n'utilise ni expression() ni binding.
var forbiddenCSS = regexp.MustCompile(`(?i)@import|expression\s*\(|behavior\s*:|-moz-binding|url\s*\(\s*["']?\s*(https?:)?//`)

// validateCustomCSS : contrôle front.customCSS dans un document design décodé.
func validateCustomCSS(v any) error {
	top, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	front, ok := top["front"].(map[string]any)
	if !ok {
		return nil
	}
	css, ok := front["customCSS"].(string)
	if !ok || css == "" {
		return nil
	}
	if forbiddenCSS.MatchString(css) {
		return fmt.Errorf("CSS personnalisé : @import, expression() et ressources distantes (url(http…)) interdits")
	}
	return nil
}
