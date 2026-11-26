package render

// Chrome dédié /admin (style WordPress : sidebar + topbar + wrap).
// Classes préfixées mx- (static/admin.css) : zéro impact sur la boutique.

import (
	"fmt"
	"strings"
)

type MxSub struct {
	Href, Label string
}

type MxItem struct {
	Slug, Label, Icon string
	Count             int
	ShowCount         bool
	Subs              []MxSub
}

// MxSidebar : menu latéral avec compteurs et sous-menus (style WP).
func MxSidebar(active, current string, counts map[string]int) string {
	items := []MxItem{
		{Slug: "", Label: "Tableau de bord", Icon: "dashboard"},
		{Slug: "produits", Label: "Produits", Icon: "archive", Count: counts["produits"], ShowCount: true, Subs: []MxSub{
			{"/admin/produits", "Tous les produits"},
			{"/admin/produits/nouveau", "Ajouter un nouveau produit"},
			{"/admin/produits/marques", "Marques"},
			{"/admin/types", "Catégories"},
			{"/admin/produits/etiquettes", "Étiquettes"},
			{"/admin/produits/attributs", "Attributs"},
			{"/admin/produits/avis", "Avis"},
		}},
		{Slug: "stats", Label: "Statistiques", Icon: "activity", Subs: []MxSub{
			{"/admin/stats", "Vue d'ensemble"},
			{"/admin/stats?vue=produits", "Produits"},
			{"/admin/stats?vue=revenu", "Revenu"},
			{"/admin/stats?vue=commandes", "Commandes"},
			{"/admin/stats?vue=variations", "Variations"},
			{"/admin/stats?vue=categories", "Catégories"},
			{"/admin/stats?vue=promos", "Codes promo"},
			{"/admin/stats?vue=taxes", "Taxes"},
			{"/admin/stats?vue=telechargements", "Téléchargements"},
			{"/admin/stats?vue=stock", "Stock"},
		}},
		{Slug: "marketing", Label: "Marketing", Icon: "tag", Subs: []MxSub{
			{"/admin/marketing", "Vue d'ensemble"},
			{"/admin/marketing/promos", "Codes promo"},
		}},
		{Slug: "design", Label: "Design", Icon: "palette"},
		{Slug: "commandes", Label: "Commandes", Icon: "cart", Count: counts["commandes"], ShowCount: true},
		{Slug: "clients", Label: "Clients", Icon: "users", Count: counts["clients"], ShowCount: true},
		{Slug: "medias", Label: "Médias", Icon: "image"},
		{Slug: "outils", Label: "Outils", Icon: "cog", Subs: []MxSub{
			{"/admin/outils", "Outils disponibles"},
			{"/admin/outils?vue=importer", "Importer"},
			{"/admin/outils?vue=exporter", "Exporter"},
			{"/admin/outils?vue=redirections", "Redirections"},
		}},
		{Slug: "reglages", Label: "Réglages", Icon: "cog"},
		{Slug: "systeme", Label: "Système", Icon: "activity"},
	}
	var b strings.Builder
	b.WriteString(`<aside class="mx-side"><div class="mx-brand"><img src="/images/logo-mediatec.png" alt=""><div>Marexstore<small>Administration</small></div><button class="mx-side-x" type="button" data-side-close aria-label="Fermer le menu">✕</button></div><nav class="mx-nav"><div class="sep">Administration</div>`)
	for _, it := range items {
		href := "/admin"
		if it.Slug != "" {
			href = "/admin/" + it.Slug
		}
		open := active == it.Slug || (it.Slug != "" && strings.HasPrefix(active, it.Slug+"/"))
		cls := ""
		if open {
			cls = ` class="on"`
		}
		cnt := ""
		if it.ShowCount {
			cnt = fmt.Sprintf(`<span class="cnt">%d</span>`, it.Count)
		}
		fmt.Fprintf(&b, `<a href="%s"%s>%s<span>%s</span>%s</a>`, href, cls, MxSVG(it.Icon), it.Label, cnt)
		if open && len(it.Subs) > 0 {
			b.WriteString(`<div class="mx-subnav">`)
			for _, s := range it.Subs {
				cls := ""
				if s.Href == current {
					cls = ` class="on"`
				}
				fmt.Fprintf(&b, `<a href="%s"%s>%s</a>`, s.Href, cls, Esc(s.Label))
			}
			b.WriteString(`</div>`)
		}
	}
	b.WriteString(`<div class="sep">Boutique</div><a href="/"><svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round" aria-hidden="true"><path d="M18 13v6a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2V8a2 2 0 0 1 2-2h6"/><polyline points="15 3 21 3 21 9"/><line x1="10" y1="14" x2="21" y2="3"/></svg><span>Voir le site</span></a></nav><div class="mx-user"><span class="mx-user-av">A</span><span class="mx-user-tx"><strong>admin</strong><small>Administrateur</small></span><a class="mx-user-out" href="/" title="Retour boutique">↗</a></div></aside>`)
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
		"palette":   `<circle cx="12" cy="12" r="10"/><circle cx="8.5" cy="10" r="1.4" fill="currentColor" stroke="none"/><circle cx="12" cy="7.5" r="1.4" fill="currentColor" stroke="none"/><circle cx="15.5" cy="10" r="1.4" fill="currentColor" stroke="none"/><path d="M12 22c5 0 8-3 8-8h-3a4 4 0 0 1-4-4H9a5 5 0 0 0-5 5c0 4 3 7 8 7z"/>`,
		"activity":  `<polyline points="22 12 18 12 15 21 9 3 6 12 2 12"/>`,
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
// MxBackLink : destination du retour hiérarchique d'une page admin.
type MxBackLink struct {
	Href  string
	Label string
}

// MxBack déduit le parent logique d'une page admin depuis son chemin :
// fiche → sa liste, section → tableau de bord, tableau de bord → site.
// Le parent vient du chemin, jamais d'un paramètre de requête : aucune
// redirection ouverte possible.
func MxBack(path string) MxBackLink {
	p := strings.TrimSuffix(path, "/")
	if !strings.HasPrefix(p, "/admin") {
		return MxBackLink{}
	}
	const dash = "/admin"
	if p == dash {
		return MxBackLink{Href: "/", Label: "le site"}
	}
	switch {
	case p == dash+"/produits":
		return MxBackLink{Href: dash, Label: "le tableau de bord"}
	case strings.HasPrefix(p, dash+"/produits/"):
		return MxBackLink{Href: dash + "/produits", Label: "la liste des produits"}
	case p == dash+"/commandes":
		return MxBackLink{Href: dash, Label: "le tableau de bord"}
	case strings.HasPrefix(p, dash+"/commandes/"):
		return MxBackLink{Href: dash + "/commandes", Label: "la liste des commandes"}
	case p == dash+"/clients":
		return MxBackLink{Href: dash, Label: "le tableau de bord"}
	case strings.HasPrefix(p, dash+"/clients/"):
		return MxBackLink{Href: dash + "/clients", Label: "la liste des clients"}
	}
	// Sections simples (médias, types, réglages, système, design…) et tout
	// chemin inconnu sous /admin : on remonte au tableau de bord.
	return MxBackLink{Href: dash, Label: "le tableau de bord"}
}

// MxBackHTML rend le bouton retour ; chaîne vide pour un chemin hors admin.
func MxBackHTML(path string) string {
	b := MxBack(path)
	if b.Href == "" {
		return ""
	}
	return `<a class="mx-back" href="` + b.Href + `" title="Retour à ` + Esc(b.Label) +
		`" aria-label="Retour à ` + Esc(b.Label) + `">← <span class="lb">Retour</span></a>`
}

func MxDoc(title, active, current string, counts map[string]int, back, body string) string {
	return `<!DOCTYPE html><html lang="fr"><head><meta charset="utf-8">
<meta name="viewport" content="width=device-width, initial-scale=1">
<title>` + Esc(title) + ` — Marexstore Admin</title>
<link href="https://fonts.googleapis.com/css2?family=Inter:wght@400;600;700;800&display=swap" rel="stylesheet">
<link rel="stylesheet" href="/static/admin.css?v=15">
<script src="/static/theme.js?v=2"></script>
</head><body class="mx"><div class="mx-shell">` +
		MxSidebar(active, current, counts) +
		`<div class="mx-scrim" data-side-close></div>` +
		`<div class="mx-main"><div class="mx-top"><button class="mx-burger" type="button" onclick="document.body.classList.toggle('nav-open')" aria-label="Menu">☰</button>` + back + `<h1>` + Esc(title) + `</h1><span class="sp"></span><a class="mx-btn ghost" href="/">Voir le site</a></div><div class="mx-body">` +
		body + `</div></div></div><script src="/static/admin.js?v=9"></script></body></html>`
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
