package render

// Chrome dédié /admin (style WordPress : sidebar + topbar + wrap).
// Classes préfixées mx- (static/admin.css) : zéro impact sur la boutique.

import (
	"fmt"
	"strings"
)

type MxItem struct {
	Slug, Label, Icon string
	Count             int
	ShowCount         bool
}

// MxSidebar : menu latéral avec compteurs.
func MxSidebar(active string, counts map[string]int) string {
	items := []MxItem{
		{"", "Tableau de bord", "dashboard", 0, false},
		{"produits", "Produits", "archive", counts["produits"], true},
		{"types", "Types", "tag", 0, false},
		{"commandes", "Commandes", "cart", counts["commandes"], true},
		{"clients", "Clients", "users", counts["clients"], true},
		{"medias", "Médias", "image", 0, false},
		{"reglages", "Réglages", "cog", 0, false},
	}
	var b strings.Builder
	b.WriteString(`<aside class="mx-side"><div class="mx-brand"><img src="/images/logo-mediatec.png" alt=""><div>Marexstore<small>Administration</small></div></div><nav class="mx-nav">`)
	for _, it := range items {
		href := "/admin"
		if it.Slug != "" {
			href = "/admin/" + it.Slug
		}
		cls := ""
		if active == it.Slug {
			cls = ` class="on"`
		}
		cnt := ""
		if it.ShowCount {
			cnt = fmt.Sprintf(`<span class="cnt">%d</span>`, it.Count)
		}
		fmt.Fprintf(&b, `<a href="%s"%s>%s<span>%s</span>%s</a>`, href, cls, MxSVG(it.Icon), it.Label, cnt)
	}
	b.WriteString(`<div class="sep">Boutique</div><a href="/"><span>Voir le site</span></a></nav><div class="mx-side-foot">Connecté : admin<br><a href="/">Retour boutique</a></div></aside>`)
	return b.String()
}

// MxSVG : pictos inline (stroke currentColor, style dashicons).
func MxSVG(name string) string {
	p := map[string]string{
		"dashboard": `<rect x="3" y="3" width="7" height="9" rx="1"/><rect x="14" y="3" width="7" height="5" rx="1"/><rect x="14" y="12" width="7" height="9" rx="1"/><rect x="3" y="16" width="7" height="5" rx="1"/>`,
		"archive":   `<rect x="2" y="3" width="20" height="5" rx="1"/><path d="M4 8v12a1 1 0 0 0 1 1h14a1 1 0 0 0 1-1V8M10 12h4"/>`,
		"tag":       `<path d="M20 12l-8 8-9-9V4h7z"/><circle cx="7.5" cy="7.5" r="1.5"/>`,
		"cart":      `<path d="M6 2L3 6v14a2 2 0 0 0 2 2h14a2 2 0 0 0 2-2V6l-3-4z"/><path d="M3 6h18M16 10a4 4 0 0 1-8 0"/>`,
		"users":     `<path d="M20 21v-2a4 4 0 0 0-4-4H8a4 4 0 0 0-4 4v2"/><circle cx="12" cy="7" r="4"/>`,
		"image":     `<rect x="3" y="3" width="18" height="18" rx="2"/><circle cx="8.5" cy="8.5" r="1.5"/><path d="M21 15l-5-5L5 21"/>`,
		"cog":       `<circle cx="12" cy="12" r="3"/><path d="M19.4 15a1.65 1.65 0 0 0 .33 1.82l.06.06a2 2 0 1 1-2.83 2.83l-.06-.06a1.65 1.65 0 0 0-1.82-.33 1.65 1.65 0 0 0-1 1.51V21a2 2 0 1 1-4 0v-.09A1.65 1.65 0 0 0 9 19.4a1.65 1.65 0 0 0-1.82.33l-.06.06a2 2 0 1 1-2.83-2.83l.06-.06a1.65 1.65 0 0 0 .33-1.82 1.65 1.65 0 0 0-1.51-1H3a2 2 0 1 1 0-4h.09A1.65 1.65 0 0 0 4.6 9a1.65 1.65 0 0 0-.33-1.82l-.06-.06a2 2 0 1 1 2.83-2.83l.06.06a1.65 1.65 0 0 0 1.82.33H9a1.65 1.65 0 0 0 1-1.51V3a2 2 0 1 1 4 0v.09a1.65 1.65 0 0 0 1 1.51 1.65 1.65 0 0 0 1.82-.33l.06-.06a2 2 0 1 1 2.83 2.83l-.06.06a1.65 1.65 0 0 0-.33 1.82V9a1.65 1.65 0 0 0 1.51 1H21a2 2 0 1 1 0 4h-.09a1.65 1.65 0 0 0-1.51 1z"/>`,
	}
	return `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true">` + p[name] + `</svg>`
}

// MxNotice : bandeau succès/erreur (?msg= / ?err=).
func MxNotice(okMsg, errMsg string) string {
	if okMsg != "" {
		return `<div class="mx-notice">` + Esc(okMsg) + `</div>`
	}
	if errMsg != "" {
		return `<div class="mx-notice err">` + Esc(errMsg) + `</div>`
	}
	return ""
}

// MxDoc : document admin complet (sans chrome boutique).
func MxDoc(title, active string, counts map[string]int, body string) string {
	return `<!DOCTYPE html><html lang="fr"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + Esc(title) + ` — Marexstore Admin</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
<link rel="stylesheet" href="/static/admin.css?v=4">
<script src="/static/theme.js?v=1"></script>
</head><body class="mx"><div class="mx-shell">` +
		MxSidebar(active, counts) +
		`<div class="mx-main"><div class="mx-top"><button class="mx-burger" type="button" onclick="document.body.classList.toggle('nav-open')" aria-label="Menu">☰</button><h1>` + Esc(title) + `</h1><span class="sp"></span><a class="mx-btn ghost" href="/">Voir le site</a></div><div class="mx-body">` +
		body + `</div></div></div><script src="/static/admin.js?v=3"></script></body></html>`
}

// MxView : un onglet de vue (Tous/En stock/...) avec compteur.
type MxView struct {
	Key   string
	Href  string
	Label string
	Count int
}

// MxViews : onglets de vues avec compteurs.
func MxViews(views []MxView, cur string) string {
	var b strings.Builder
	b.WriteString(`<ul class="mx-views">`)
	for i, v := range views {
		if i > 0 {
			b.WriteString(`<li aria-hidden="true">|</li>`)
		}
		cls := ""
		if v.Key == cur {
			cls = ` class="on"`
		}
		fmt.Fprintf(&b, `<li><a href="%s"%s>%s <span class="c">(%d)</span></a></li>`, v.Href, cls, Esc(v.Label), v.Count)
	}
	b.WriteString(`</ul>`)
	return b.String()
}

// MxPager : pagination simple.
func MxPager(base string, page, pages int) string {
	if pages < 2 {
		return ""
	}
	var b strings.Builder
	b.WriteString(`<div class="mx-pager">`)
	sep := "?"
	if strings.Contains(base, "?") {
		sep = "&"
	}
	prev, next := page-1, page+1
	if prev < 1 {
		prev = 1
	}
	if next > pages {
		next = pages
	}
	fmt.Fprintf(&b, `<a href="%s%spage=%d">‹</a><span class="cur">%d / %d</span><a href="%s%spage=%d">›</a>`, base, sep, prev, page, pages, base, sep, next)
	return b.String() + `</div>`
}
