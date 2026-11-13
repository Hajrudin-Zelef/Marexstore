package render_test

import (
	"fmt"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Charge le catalogue complet (171 + lot commercial), comme le serveur.
func testStoreFull(t *testing.T) *catalog.Store {
	t.Helper()
	list, err := catalog.LoadRaw(rel("../../catalogdata/produits-v3.json"))
	if err != nil {
		t.Skip("catalogdata absent (hors dépôt local)")
	}
	lot, err := catalog.LoadRaw(rel("../../catalogdata/nouveaux-produits-commercial.json"))
	if err != nil {
		t.Skip("lot commercial absent")
	}
	list = append(list, lot...)
	catalog.ApplyTaxonomy(list)
	return catalog.NewStore(list)
}

// TestTypeHubComplete : chaque page catégorie liste tous ses types, avec le
// bon compteur, une photo (ou « Photo à venir ») et le lien ?type=.
func TestTypeHubComplete(t *testing.T) {
	s := testStoreFull(t)
	for _, m := range catalog.MainOrder {
		fs := render.CategoryBody(s, m)
		top := catalog.MainSlug[m]
		hub := render.TypeHubHTML(fs, "", "../", top)
		counts := map[string]int{}
		for _, p := range fs {
			counts[p.Sub]++
		}
		if got := strings.Count(hub, `class="hub-card"`); got != len(counts) {
			t.Errorf("%s : %d tuiles, %d types attendus", m, got, len(counts))
		}
		for sub, n := range counts {
			sl := catalog.Slugify(sub)
			if !strings.Contains(hub, `href="../categories/`+top+`/`+sl+`.html"`) {
				t.Errorf("%s : lien page type absent pour %q", m, sub)
			}
			if !strings.Contains(hub, `>`+catalog.Esc(sub)+`<`) {
				t.Errorf("%s : libellé %q absent du hub", m, sub)
			}
			want := fmt.Sprintf(`%d produit`, n)
			if n > 1 {
				want = fmt.Sprintf(`%d produits`, n)
			}
			if !strings.Contains(hub, want) {
				t.Errorf("%s : compteur %q absent pour %q", m, want, sub)
			}
			// La tuile du type : une photo ou le repli « Photo à venir ».
			i := strings.Index(hub, `href="../categories/`+top+`/`+sl+`.html"`)
			if i < 0 {
				continue
			}
			tile := hub[i:]
			if len(tile) > 300 {
				tile = tile[:300]
			}
			if !strings.Contains(tile, "<img") && !strings.Contains(tile, "noimg") {
				t.Errorf("%s : tuile %q sans photo ni repli", m, sub)
			}
		}
		if !strings.Contains(hub, fmt.Sprintf(`Voir tous les produits (%d)`, len(fs))) {
			t.Errorf("%s : lien « Voir tous les produits (%d) » absent", m, len(fs))
		}
	}
}

// TestHomeFilter : le filtre global de l'accueil propose marque, catégorie,
// type (tous les types du catalogue) et disponibilité, et la carte des types.
func TestHomeFilter(t *testing.T) {
	s := testStoreFull(t)
	recents := s.Recents(rel("../../catalogdata/store-sitemap.xml"))
	body := render.IndexBody(s, recents)
	if !strings.Contains(body, `data-home-filter`) {
		t.Fatal("filtre global absent de l'accueil")
	}
	if !strings.Contains(body, `action="boutique.html"`) {
		t.Error("le filtre doit envoyer vers boutique.html")
	}
	// Marques : les mêmes valeurs que le filtre latéral (NormName).
	if !strings.Contains(body, `<option value="hp">HP (`) {
		t.Error("marque HP absente du filtre")
	}
	if !strings.Contains(body, `<option value="cisco">Cisco (`) {
		t.Error("marque Cisco absente du filtre")
	}
	// Catégories : les 7 mains.
	for _, m := range catalog.MainOrder {
		if !strings.Contains(body, `<option value="`+catalog.MainSlug[m]+`">`+catalog.Esc(m)+`</option>`) {
			t.Errorf("catégorie %q absente du filtre", m)
		}
	}
	// Types : tous les types du catalogue.
	seen := map[string]bool{}
	for _, p := range s.Products {
		seen[p.Sub] = true
	}
	for sub := range seen {
		if !strings.Contains(body, `value="`+catalog.Slugify(sub)+`"`) {
			t.Errorf("type %q absent du filtre", sub)
		}
	}
	if !strings.Contains(body, `id="mediatec-types"`) {
		t.Error("carte des types (#mediatec-types JSON inerte) absente")
	}
	for _, m := range catalog.MainOrder {
		if !strings.Contains(body, `"`+catalog.MainSlug[m]+`":`) {
			t.Errorf("catégorie %q absente de MEDIATEC_TYPES", m)
		}
	}
	t.Logf("filtre : %d types", len(seen))
}

// TestLiveTopPages : chaque top live du menu « Toutes les catégories » a sa
// page (/categories/<slug>.html) avec le hub de SES types d'abord, puis la
// liste cachée — pas la boutique filtrée. (3 tops live n'ont aucun produit
// dans le catalogue actuel : leur page est vide, c'est normal.)
func TestLiveTopPages(t *testing.T) {
	s := testStoreFull(t)
	for _, lt := range catalog.LiveCats {
		fs := render.CategoryBodyLive(s, lt.Slug)
		page := render.CategoryPageLive(s, lt)
		if !strings.Contains(page, `<h1 style="font-size:24px">`+catalog.Esc(lt.Label)+`</h1>`) {
			t.Errorf("%s : titre absent", lt.Label)
		}
		if len(fs) == 0 {
			t.Logf("%s : aucun produit (page vide)", lt.Label)
			continue
		}
		if !strings.Contains(page, `data-hub`) {
			t.Errorf("%s : hub des types absent", lt.Label)
		}
		if !strings.Contains(page, `data-catalog-home hidden`) {
			t.Errorf("%s : liste des produits pas cachée derrière le hub", lt.Label)
		}
		// Le hub liste les types de CE top live, avec le bon compteur.
		counts := map[string]int{}
		for _, p := range fs {
			counts[p.Sub]++
		}
		hub := render.TypeHubHTML(fs, "", "../", lt.Slug)
		if got := strings.Count(hub, `class="hub-card"`); got != len(counts) {
			t.Errorf("%s : %d tuiles, %d types attendus", lt.Label, got, len(counts))
		}
		for sub, n := range counts {
			if !strings.Contains(hub, `href="../categories/`+lt.Slug+`/`+catalog.Slugify(sub)+`.html"`) {
				t.Errorf("%s : lien page type absent pour %q", lt.Label, sub)
			}
			want := fmt.Sprintf(`%d produit`, n)
			if n > 1 {
				want = fmt.Sprintf(`%d produits`, n)
			}
			if !strings.Contains(hub, want) {
				t.Errorf("%s : compteur %q absent pour %q", lt.Label, want, sub)
			}
		}
	}
	// Le cas de Zelef : Matériels réseaux doit montrer ses 5 types.
	fs := render.CategoryBodyLive(s, "materiels-reseaux")
	hub := render.TypeHubHTML(fs, "", "../", "materiels-reseaux")
	for _, sub := range []string{"Réseau & câbles", "Cisco & réseau", "Téléphonie IP", "Réseau Legrand", "Adaptateurs"} {
		if !strings.Contains(hub, catalog.Esc(sub)) {
			t.Errorf("Matériels réseaux : type %q absent du hub", sub)
		}
	}
}

// TestBoutiqueHubLinks : les tuiles du hub boutique (« Nos catégories »)
// mènent à la page catégorie (hub des types d'abord), jamais à la boutique
// filtrée — sinon le hub s'affiche 1 s puis disparaît (syncHub le masque
// quand un filtre est posé). Les sous-catégories mènent à la page catégorie
// filtrée sur le type (?type=).
func TestBoutiqueHubLinks(t *testing.T) {
	s := testStoreFull(t)
	hub := render.HubHTML(s, "")
	reMain := regexp.MustCompile(`<a class="hub-main" href="([^"]*)">`)
	mains := reMain.FindAllStringSubmatch(hub, -1)
	if len(mains) == 0 {
		t.Fatalf("hub boutique : aucune tuile « hub-main »")
	}
	for _, m := range mains {
		href := m[1]
		if strings.Contains(href, "boutique.html?store=") {
			t.Errorf("tuile hub : href=%q pointe encore vers la boutique filtrée", href)
		}
		if !strings.HasPrefix(href, "categories/") || !strings.HasSuffix(href, ".html") {
			t.Errorf("tuile hub : href=%q attendu categories/<slug>.html", href)
		}
	}
	reSub := regexp.MustCompile(`<li><a href="([^"]*)">`)
	subs := reSub.FindAllStringSubmatch(hub, -1)
	if len(subs) == 0 {
		t.Fatalf("hub boutique : aucun lien de sous-catégorie")
	}
	reType := regexp.MustCompile(`^categories/([a-z0-9-]+)/([a-z0-9-]+)\.html$`)
	for _, m := range subs {
		href := m[1]
		if strings.Contains(href, "boutique.html") || strings.Contains(href, "?type=") {
			t.Errorf("sous-catégorie hub : href=%q doit mener à la page type dédiée", href)
		}
		mm := reType.FindStringSubmatch(href)
		if mm == nil {
			t.Errorf("sous-catégorie hub : href=%q attendu categories/<top>/<type>.html", href)
			continue
		}
		if _, _, ok := render.CategoryTypePage(s, mm[1], mm[2]); !ok {
			t.Errorf("sous-catégorie hub : page type %q vide ou introuvable", href)
		}
	}
}

// TestMegaMenuVoirTous : « Voir tous les X » du méga-menu mène à la page
// catégorie (hub des types d'abord), jamais à la boutique filtrée.
func TestMegaMenuVoirTous(t *testing.T) {
	s := testStoreFull(t)
	panel := render.MegaPanel(s, "")
	re := regexp.MustCompile(`<a class="mega-all" href="([^"]*)">`)
	matches := re.FindAllStringSubmatch(panel, -1)
	if len(matches) != len(catalog.LiveCats) {
		t.Fatalf("méga-menu : %d liens « Voir tous », %d tops live attendus", len(matches), len(catalog.LiveCats))
	}
	seen := map[string]bool{}
	for _, m := range matches {
		href := m[1]
		seen[href] = true
		if strings.Contains(href, "boutique.html?store=") {
			t.Errorf("« Voir tous » : href=%q pointe encore vers la boutique filtrée", href)
		}
	}
	for _, lt := range catalog.LiveCats {
		want := "categories/" + lt.Slug + ".html"
		if !seen[want] {
			t.Errorf("top %q : pas de lien « Voir tous » vers %q", lt.Label, want)
		}
	}
}

// TestBoutiqueDataSub : les cartes de la boutique portent data-sub = la
// sous-catégorie (le filtre ?type= s'applique partout, pas seulement sur les
// pages catégories), jamais le slug d'une catégorie principale.
func TestBoutiqueDataSub(t *testing.T) {
	s := testStoreFull(t)
	body := render.BoutiqueBody(s)
	mainSlugs := map[string]bool{}
	for _, m := range catalog.MainOrder {
		mainSlugs[catalog.MainSlug[m]] = true
	}
	subSlugs := map[string]bool{}
	for _, p := range s.Products {
		subSlugs[catalog.Slugify(p.Sub)] = true
	}
	re := regexp.MustCompile(`data-sub="([^"]+)"`)
	vals := re.FindAllStringSubmatch(body, -1)
	if len(vals) != len(s.Products) {
		t.Fatalf("boutique : %d data-sub, %d produits attendus", len(vals), len(s.Products))
	}
	for _, v := range vals {
		if mainSlugs[v[1]] {
			t.Errorf("data-sub = %q, un slug de catégorie principale (ancien comportement)", v[1])
		}
		if !subSlugs[v[1]] {
			t.Errorf("data-sub = %q, pas un type connu", v[1])
		}
	}
}

// TestFilterLinksAlive : chaque lien ?type= (hub boutique, hub des types) matche
// au moins une carte — via data-sub OU data-live-sub, comme le filtre JS — et
// chaque lien ?sub= du méga-menu matche au moins une carte (data-live-sub).
// Un lien qui ne matche rien = page vide = « la page disparaît ».
func TestFilterLinksAlive(t *testing.T) {
	s := testStoreFull(t)
	matchType := func(v string) bool {
		for _, p := range s.Products {
			if catalog.Slugify(p.Sub) == v || p.LiveSub == v {
				return true
			}
		}
		return false
	}
	matchLiveSub := func(v string) bool {
		for _, p := range s.Products {
			if p.LiveSub == v {
				return true
			}
		}
		return false
	}
	// 1. Hub boutique : tuiles ?type= des sous-listes (+ tuiles catégories).
	hub := render.HubHTML(s, "")
	for _, m := range regexp.MustCompile(`href="categories/[^"]+\.html\?type=([^"]+)"`).FindAllStringSubmatch(hub, -1) {
		if !matchType(m[1]) {
			t.Errorf("hub boutique : ?type=%q ne matche aucune carte", m[1])
		}
	}
	// 2. Hub des types de chaque page catégorie (familles + tops live) :
	// chaque tuile ouvre une page type non vide.
	reTypePage := regexp.MustCompile(`categories/([a-z0-9-]+)/([a-z0-9-]+)\.html`)
	checkHub := func(hub, where, top string) {
		found := false
		for _, mt := range reTypePage.FindAllStringSubmatch(hub, -1) {
			found = true
			if _, _, ok := render.CategoryTypePage(s, mt[1], mt[2]); !ok {
				t.Errorf("types %s : page %q vide ou introuvable", where, mt[0])
			}
		}
		if !found {
			t.Errorf("types %s (top %s) : aucune tuile vers une page type", where, top)
		}
	}
	for _, m := range catalog.MainOrder {
		if len(render.CategoryBody(s, m)) == 0 {
			continue
		}
		checkHub(render.TypeHubHTML(render.CategoryBody(s, m), "", "../", catalog.MainSlug[m]), m, catalog.MainSlug[m])
	}
	for _, lt := range catalog.LiveCats {
		fs := render.CategoryBodyLive(s, lt.Slug)
		if len(fs) == 0 {
			continue
		}
		checkHub(render.TypeHubHTML(fs, "", "../", lt.Slug), lt.Label, lt.Slug)
	}
	// 3. Méga-menu : sous-liens simples -> pages types dédiées non vides ;
	// « Voir + » multi -> boutique ?sub= (match data-live-sub).
	for _, m := range reTypePage.FindAllStringSubmatch(render.MegaPanel(s, ""), -1) {
		if _, _, ok := render.CategoryTypePage(s, m[1], m[2]); !ok {
			t.Errorf("méga-menu : page type %q vide ou introuvable", m[0])
		}
	}
	for _, m := range regexp.MustCompile(`boutique\.html\?sub=([a-z0-9-,]+)"`).FindAllStringSubmatch(render.MegaPanel(s, ""), -1) {
		for _, v := range strings.Split(m[1], ",") {
			if !matchLiveSub(v) {
				t.Errorf("méga-menu « Voir + » : ?sub=%q ne matche aucune carte", v)
			}
		}
	}
}

// TestInitialStateFromQuery : état initial anti-flash rendu par le serveur.
// Boutique filtrée (?sub=) : hub masqué + liste affichée dès le premier rendu.
// Page catégorie ?type= : hub des types TOUJOURS visible, tuile active marquée,
// liste affichée. Sans query : hub visible, liste cachée (inchangé).
func TestInitialStateFromQuery(t *testing.T) {
	s := testStoreFull(t)
	must := func(cond bool, msg string) {
		t.Helper()
		if !cond {
			t.Error(msg)
		}
	}
	// Boutique par défaut : comme avant.
	def := render.BoutiqueBody(s)
	must(strings.Contains(def, `<div data-hub>`), "boutique : hub visible par défaut")
	must(strings.Contains(def, `<div class="home" data-catalog-home hidden>`), "boutique : liste cachée par défaut")
	// Boutique filtrée : pas de bascule JS à attendre.
	f := render.BoutiqueBodyQuery(s, testQuery("sub=souris"))
	must(strings.Contains(f, `<div data-hub hidden>`), "boutique ?sub= : hub masqué au premier rendu")
	must(!strings.Contains(f, `<div class="home" data-catalog-home hidden>`), "boutique ?sub= : liste affichée au premier rendu")
	must(strings.Contains(f, `<div class="home" data-catalog-home>`), "boutique ?sub= : bloc liste présent")
	// Catégorie par défaut : hub visible, liste cachée.
	cat := render.CategoryPage(s, "Imprimantes")
	must(strings.Contains(cat, `<div data-hub data-typehub>`), "catégorie : hub des types marqué data-typehub")
	must(strings.Contains(cat, `<div data-catalog-home hidden>`), "catégorie : liste cachée par défaut")
	must(!strings.Contains(cat, `hub-main on`), "catégorie : aucune tuile active par défaut")
	// Catégorie ?type= : hub TOUJOURS visible, tuile active, liste affichée.
	ct := render.CategoryPageFiltered(s, "Réseaux", "reseau-cables", true)
	must(strings.Contains(ct, `<div data-hub data-typehub>`), "catégorie ?type= : hub visible (jamais hidden)")
	must(!strings.Contains(ct, `<div data-hub data-typehub hidden>`), "catégorie ?type= : hub non masqué")
	must(strings.Contains(ct, `class="hub-main on" href="../categories/reseaux/reseau-cables.html" aria-current="true"`), "catégorie ?type= : tuile active marquée")
	must(strings.Contains(ct, `<div data-catalog-home>`), "catégorie ?type= : liste affichée")
	must(!strings.Contains(ct, `<div data-catalog-home hidden>`), "catégorie ?type= : liste non cachée")
}

func testQuery(raw string) url.Values {
	v, _ := url.ParseQuery(raw)
	return v
}

// TestNoInlineExecutableScript : zéro <script> exécutable inline dans le HTML
// rendu (CSP script-src 'self' strict, console sans warning). Seul autorisé :
// <script type="application/json"> (données inertes, jamais exécutées).
func TestNoInlineExecutableScript(t *testing.T) {
	s := testStoreFull(t)
	pages := map[string]string{
		"head":      render.Head("T", "", ""),
		"index":     render.Doc("T", render.IndexBody(s, s.Recents("")), "", ""),
		"boutique":  render.Doc("T", render.BoutiqueBody(s), "", ""),
		"categorie": render.Doc("T", render.CategoryPage(s, "Réseaux"), "../", ""),
		"fiche":     render.Doc("T", render.Header(s, "../", "")+render.FicheBody(s, s.Products[0])+render.Footer(s, "../"), "../", ""),
	}
	re := regexp.MustCompile(`<script[^>]*>`)
	for name, html := range pages {
		for _, tag := range re.FindAllString(html, -1) {
			if strings.Contains(tag, "src=") {
				continue
			}
			if strings.Contains(tag, `type="application/json"`) {
				continue
			}
			t.Errorf("%s : script inline exécutable : %s", name, tag)
		}
		if strings.Contains(html, "onsubmit=") || strings.Contains(html, "onclick=") {
			t.Errorf("%s : gestionnaire inline (onclick/onsubmit)", name)
		}
		if strings.Contains(html, `action="mailto:`) {
			t.Errorf("%s : <form action=\"mailto:\"> (Mixed Content + form-action)", name)
		}
	}
	// JSON inerte présent sur l'accueil et lisible.
	if !strings.Contains(pages["index"], `id="mediatec-types"`) {
		t.Errorf("accueil : bloc #mediatec-types absent")
	}
}

// TestTypePages : chaque tuile de type (hub boutique, hubs des pages catégories,
// méga-menu) ouvre une page dédiée bien structurée : fil d'Ariane complet,
// titre, hub des types TOUJOURS visible avec tuile(s) active(s), liste affichée.
func TestTypePages(t *testing.T) {
	s := testStoreFull(t)
	seen := map[string]bool{}
	check := func(top, typ, where string) {
		t.Helper()
		key := top + "/" + typ
		if seen[key] {
			return
		}
		seen[key] = true
		title, body, ok := render.CategoryTypePage(s, top, typ)
		if !ok {
			t.Errorf("page type %s (%s) : introuvable ou vide", key, where)
			return
		}
		if title == "" {
			t.Errorf("page type %s : titre vide", key)
		}
		must := func(cond bool, msg string) {
			t.Helper()
			if !cond {
				t.Errorf("page type %s : %s", key, msg)
			}
		}
		must(!strings.Contains(body, `data-typehub`), "le hub des types ne doit plus s'afficher sur la page type")
		must(strings.Contains(body, `Tous les types de`), "lien de retour vers les types absent")
		must(strings.Contains(body, `<div data-catalog-home data-locked>`), "liste des produits absente ou non verrouillée")
		must(!strings.Contains(body, `<div data-catalog-home hidden`), "liste des produits cachée")
		must(strings.Contains(body, `class="card"`), "aucune carte produit")
		must(strings.Contains(body, `<a href="../../boutique.html">Store</a>`), "fil d'Ariane incomplet")
		must(strings.Contains(body, `<a href="../`+top+`.html">`), "fil d'Ariane sans lien catégorie")
	}
	reType := regexp.MustCompile(`categories/([a-z0-9-]+)/([a-z0-9-]+)\.html`)
	for _, m := range reType.FindAllStringSubmatch(render.HubHTML(s, ""), -1) {
		check(m[1], m[2], "hub boutique")
	}
	for _, m := range catalog.MainOrder {
		fs := render.CategoryBody(s, m)
		if len(fs) == 0 {
			continue
		}
		for _, mt := range reType.FindAllStringSubmatch(render.TypeHubHTML(fs, "", "../", catalog.MainSlug[m]), -1) {
			check(catalog.MainSlug[m], mt[2], "hub "+m)
		}
	}
	for _, lt := range catalog.LiveCats {
		fs := render.CategoryBodyLive(s, lt.Slug)
		if len(fs) == 0 {
			continue
		}
		for _, mt := range reType.FindAllStringSubmatch(render.TypeHubHTML(fs, "", "../", lt.Slug), -1) {
			check(lt.Slug, mt[2], "hub "+lt.Label)
		}
	}
	for _, m := range reType.FindAllStringSubmatch(render.MegaPanel(s, ""), -1) {
		check(m[1], m[2], "méga-menu")
	}
	t.Logf("pages types vérifiées : %d", len(seen))
}
