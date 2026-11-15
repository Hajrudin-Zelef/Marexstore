// Snapshot du catalogue réel pour le Studio (phase A2).
//
// GET /admin/api/snapshot (auth admin) expose les données réelles du site —
// navigation, méga-menu, hero, mini-casiers, marques, top ventes, pied de
// page, libellés Tx courants — pour que les éditeurs du Studio partent du
// réel au lieu de placeholders (« Titre de la diapositive », « Avantage 1 »,
// marques codées en dur). Lecture seule : aucun changement DB côté boutique.
//
// Le snapshot est aussi injecté dans la page du Studio
// (window.__STUDIO_SNAPSHOT__) ; il est reconstruit à chaque chargement de
// la page (données petites, une seule requête par ouverture du Studio) —
// pas de cache mémoire à invalider.

package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/design"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

type snapshotProduit struct {
	Nom         string `json:"nom"`
	Slug        string `json:"slug"`
	Prix        *int64 `json:"prix"`        // prix boutique (COALESCE), null si sur demande
	PrixAffiche string `json:"prixAffiche"` // « 45.000 F » ou « Prix sur demande »
	Image       string `json:"image"`
	Marque      string `json:"marque"`
	Accroche    string `json:"accroche,omitempty"` // rempli pour l'offre spéciale
}

type snapshotMegaTop struct {
	Slug    string `json:"slug"`
	Label   string `json:"label"`
	Groupes []struct {
		Titre string   `json:"titre"`
		Subs  []string `json:"subs"`
	} `json:"groupes"`
}

// buildSnapshot assemble le snapshot depuis le catalogue en mémoire (Store),
// les données de rendu exportées (nav, hero, minis, marques, top ventes,
// pied de page) et le design actif (libellés Tx).
func (a *App) buildSnapshot() map[string]any {
	// Méga-menu : tops live + groupes thématiques (taxonomie, jamais modifiée).
	mega := make([]snapshotMegaTop, 0, len(catalog.LiveCats))
	for _, t := range catalog.LiveCats {
		top := snapshotMegaTop{Slug: t.Slug, Label: t.Label}
		for _, g := range catalog.MegaGroups[t.Slug] {
			top.Groupes = append(top.Groupes, struct {
				Titre string   `json:"titre"`
				Subs  []string `json:"subs"`
			}{Titre: g.Title, Subs: append([]string{}, g.Subs...)})
		}
		mega = append(mega, top)
	}

	// Top ventes : les 8 slugs du site, résolus en produits réels.
	topVentes := make([]snapshotProduit, 0, len(render.Top8Slugs))
	if a.Store != nil {
		for _, slug := range render.Top8Slugs {
			p, ok := a.Store.BySlug[slug]
			if !ok {
				continue
			}
			topVentes = append(topVentes, snapshotProduit{
				Nom:         p.Nom,
				Slug:        p.Slug,
				Prix:        p.PrixAffiche(),
				PrixAffiche: catalog.FmtPrix(p.PrixAffiche()),
				Image:       p.Image,
				Marque:      p.Marque,
			})
		}
	}

	// Produits du sélecteur (phase A3) : l'union des top ventes, des
	// mini-casiers et des produits liés par les bannières du hero — les
	// produits que le site met déjà en avant, pour choisir sans taper.
	produits := make([]snapshotProduit, 0, len(topVentes)+len(render.Minis)+len(render.Banners))
	seen := map[string]bool{}
	addProduit := func(slug string) {
		if slug == "" || seen[slug] || a.Store == nil {
			return
		}
		p, ok := a.Store.BySlug[slug]
		if !ok {
			return
		}
		seen[slug] = true
		produits = append(produits, snapshotProduit{
			Nom:         p.Nom,
			Slug:        p.Slug,
			Prix:        p.PrixAffiche(),
			PrixAffiche: catalog.FmtPrix(p.PrixAffiche()),
			Image:       p.Image,
			Marque:      p.Marque,
		})
	}
	for _, p := range topVentes {
		addProduit(p.Slug)
	}
	for _, m := range render.Minis {
		addProduit(m.Slug)
	}
	for _, b := range render.Banners {
		// Lien produit : « produits/<slug>.html » ; les autres liens
		// (catégories) sont ignorés.
		if lien := b.Lien; len(lien) > len("produits/") && lien[:len("produits/")] == "produits/" && lien[len(lien)-len(".html"):] == ".html" {
			addProduit(lien[len("produits/") : len(lien)-len(".html")])
		}
	}

	// Pied de page : les 4 colonnes de liens + la colonne Contact (tél/e-mail).
	footer := make([]map[string]any, 0, len(render.FooterCols)+1)
	for _, c := range render.FooterCols {
		liens := make([]map[string]string, 0, len(c.Liens))
		for _, l := range c.Liens {
			liens = append(liens, map[string]string{"label": l.Label, "href": l.Href})
		}
		footer = append(footer, map[string]any{"titre": c.Titre, "liens": liens})
	}
	footer = append(footer, map[string]any{"titre": "Contact", "liens": []map[string]string{
		{"label": "Tél : " + catalog.Tel, "href": ""},
		{"label": "Tél : " + catalog.Tel2, "href": ""},
		{"label": "E-mail : " + catalog.Email, "href": ""},
		{"label": "Abidjan, Côte d'Ivoire", "href": ""},
	}})

	// Libellés Tx courants : ceux du design actif, null si rien de publié
	// (le Studio garde alors ses défauts).
	var tx map[string]string
	if a.DB != nil {
		if raw, _, err := design.Load(a.DB); err == nil && raw != nil {
			var doc struct {
				Tx map[string]string `json:"Tx"`
			}
			if json.Unmarshal(raw, &doc) == nil && len(doc.Tx) > 0 {
				tx = doc.Tx
			}
		}
	}

	// Bandeau sombre (C1) : le vrai bandeau de l'accueil.
	bandeauSombre := render.DarkBannerData

	// Offre spéciale (C1) : le vrai produit deal de l'accueil.
	var offreSpeciale any
	if a.Store != nil {
		if p, ok := a.Store.BySlug[render.DealSlug]; ok {
			offreSpeciale = snapshotProduit{
				Nom:         p.Nom,
				Slug:        p.Slug,
				Prix:        p.PrixAffiche(),
				PrixAffiche: catalog.FmtPrix(p.PrixAffiche()),
				Image:       p.Image,
				Marque:      p.Marque,
				Accroche:    p.Accroche,
			}
		}
	}

	// Familles du menu (C1) : les 11 tops live avec leurs compteurs réels
	// (le Studio n'en montrait que 7). Visible = au moins 1 produit.
	menuFams := make([]map[string]any, 0, len(catalog.LiveCats))
	counts := map[string]int{}
	if a.Store != nil {
		for _, p := range a.Store.Products {
			if top, _ := catalog.LiveMap(p); top != "" {
				counts[top]++
			}
		}
	}
	for _, t := range catalog.LiveCats {
		menuFams = append(menuFams, map[string]any{
			"label":   t.Label,
			"slug":    t.Slug,
			"count":   counts[t.Slug],
			"visible": counts[t.Slug] > 0,
		})
	}

	// Pages d'info (C1) : le vrai contenu des pages statiques, déjà chargé
	// dans a.Bodies par reload() (design_files actif d'abord, puis disque —
	// même source que le site). Best-effort : un App sans Bodies donne une
	// carte vide, pas une erreur.
	pagesInfo := map[string]string{}
	for _, nom := range []string{"apropos", "cgv", "contact", "livraison"} {
		if c, ok := a.Bodies[nom]; ok && c != "" {
			pagesInfo[nom] = c
		}
	}

	return map[string]any{
		"navLinks":      render.NavLinks,
		"megaMenu":      mega,
		"hero":          render.Banners,
		"minis":         render.Minis,
		"marques":       render.Marques,
		"topVentes":     topVentes,
		"produits":      produits,
		"bandeauSombre": bandeauSombre,
		"offreSpeciale": offreSpeciale,
		"menuFams":      menuFams,
		"pagesInfo":     pagesInfo,
		"footer":        footer,
		"tx":            tx,
	}
}

// AdminSnapshot : GET /admin/api/snapshot — le snapshot en JSON.
func (a *App) AdminSnapshot(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "snapshot": a.buildSnapshot()})
}

// studioSnapshotInit : le snapshot injecté dans la page du Studio
// (window.__STUDIO_SNAPSHOT__=…), re-marshalé pour un échappement sûr
// (comme studioDesignInit). Vide si le Store n'est pas chargé.
func studioSnapshotInit(a *App) string {
	snap := a.buildSnapshot()
	b, err := json.Marshal(snap)
	if err != nil {
		return `window.__STUDIO_SNAPSHOT__=null;`
	}
	return `window.__STUDIO_SNAPSHOT__=` + string(b) + `;`
}
