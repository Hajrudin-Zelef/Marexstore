package handlers

import (
	"fmt"
	"math"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Dashboard premium : graphiques SVG générés côté serveur (zéro JS,
// zéro CDN, zéro dépendance). Le donut utilise la technique
// stroke-dasharray : chaque segment est un cercle dont on ne montre
// qu'un arc, décalé pour s'enchaîner au précédent.

type dashSeg struct {
	Label string
	Count int
	Color string
}

// Palette premium (distincte, lisible sur fond blanc).
var dashPalette = []string{
	"#2e5136", "#2271b1", "#d63638", "#dba617", "#7c3aed",
	"#0e9f8a", "#e67e22", "#8e44ad", "#16a085", "#c0392b",
	"#2980b9", "#95a5a6",
}

func dashColor(i int) string { return dashPalette[i%len(dashPalette)] }

// donutSVG rend un donut (arcs) + sa légende. size = diamètre en px.
// Le total est affiché au centre. Les segments à 0 sont ignorés.
func donutSVG(segs []dashSeg, size int, centerTop, centerSub string) string {
	total := 0
	for _, s := range segs {
		total += s.Count
	}
	if total <= 0 || len(segs) == 0 {
		return `<div class="mx-empty"><div class="mx-empty-big">0</div><div class="mx-muted">` + render.Esc(centerSub) + `</div></div>`
	}
	r := 48.0
	c := 2 * math.Pi * r
	var arcs strings.Builder
	off := 0.0
	for _, s := range segs {
		if s.Count <= 0 {
			continue
		}
		frac := float64(s.Count) / float64(total)
		length := frac * c
		// Petit espace entre segments (1.5 unité) pour la lisibilité.
		gap := 1.5
		if length-gap < 0.5 {
			gap = 0
		}
		fmt.Fprintf(&arcs, `<circle cx="60" cy="60" r="48" fill="none" stroke="%s" stroke-width="18" stroke-dasharray="%.2f %.2f" stroke-dashoffset="%.2f" transform="rotate(-90 60 60)"/>`,
			s.Color, length-gap, c-length+gap, -off)
		off += length
	}
	var leg strings.Builder
	for _, s := range segs {
		if s.Count <= 0 {
			continue
		}
		pct := float64(s.Count) * 100 / float64(total)
		fmt.Fprintf(&leg, `<li><span class="dot" style="background:%s"></span><span class="lb">%s</span><span class="vl">%d</span><span class="pc">%.0f%%</span></li>`,
			s.Color, render.Esc(s.Label), s.Count, pct)
	}
	return `<div class="mx-donut"><svg width="` + fmt.Sprint(size) + `" height="` + fmt.Sprint(size) + `" viewBox="0 0 120 120" role="img" aria-label="Graphique"><circle cx="60" cy="60" r="48" fill="none" stroke="#eef0f2" stroke-width="18"/>` +
		arcs.String() +
		`<text x="60" y="57" text-anchor="middle" class="mx-donut-total">` + render.Esc(centerTop) + `</text>` +
		`<text x="60" y="74" text-anchor="middle" class="mx-donut-sub">` + render.Esc(centerSub) + `</text></svg>` +
		`<ul class="mx-legend">` + leg.String() + `</ul></div>`
}

// barRows rend des barres horizontales (label + barre + compteur).
// max est la valeur de référence (100%).
func barRows(rows []dashSeg, max int) string {
	if max <= 0 {
		max = 1
	}
	var b strings.Builder
	for _, s := range rows {
		if s.Count <= 0 {
			continue
		}
		w := float64(s.Count) * 100 / float64(max)
		fmt.Fprintf(&b, `<div class="mx-bar-row"><span class="lb">%s</span><span class="tr"><span class="fl" style="width:%.1f%%;background:%s"></span></span><span class="vl">%d</span></div>`,
			render.Esc(s.Label), w, s.Color, s.Count)
	}
	return b.String()
}

// kpiCard : carte KPI premium (icône, accent couleur, gros chiffre, sous-ligne).
func kpiCard(icon, accent, label, big, sub string) string {
	return `<div class="mx-card mx-kpi" style="--mx-kpi:` + accent + `"><div class="mx-kpi-ic">` + icon + `</div><div class="mx-kpi-tx"><h3>` + label + `</h3><div class="big">` + big + `</div><div class="sub">` + sub + `</div></div></div>`
}

// Icônes SVG (trait, 24x24) — style sobre, assorti à la sidebar.
const (
	icCash = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><rect x="2" y="6" width="20" height="12" rx="2"/><circle cx="12" cy="12" r="3"/><path d="M6 12h.01M18 12h.01"/></svg>`
	icBag  = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M6 7h12l1.5 13.5a1 1 0 0 1-1 1.1H5.5a1 1 0 0 1-1-1.1L6 7z"/><path d="M9 10V6a3 3 0 0 1 6 0v4"/></svg>`
	icBox  = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M21 8l-9-5-9 5v8l9 5 9-5V8z"/><path d="M3.5 8.5L12 13l8.5-4.5"/><path d="M12 13v8"/></svg>`
	icWarn = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><path d="M12 3.5L2.5 20h19L12 3.5z"/><path d="M12 10v4.2"/><path d="M12 17.4h.01"/></svg>`
	icUser = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.8" stroke-linecap="round" stroke-linejoin="round"><circle cx="9" cy="8" r="3.4"/><path d="M2.8 20c.8-3.1 3.3-4.8 6.2-4.8s5.4 1.7 6.2 4.8"/><circle cx="17" cy="9" r="2.4"/><path d="M16.2 15.4c2.4.4 4.2 1.8 4.9 4.1"/></svg>`
)
