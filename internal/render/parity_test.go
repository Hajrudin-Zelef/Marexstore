package render_test

import (
	"encoding/json"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
)

// Charge le catalogue comme le serveur.

func rel(p string) string {
	_, f, _, _ := runtime.Caller(1)
	return filepath.Join(filepath.Dir(f), p)
}

func testStore(t *testing.T) *catalog.Store {
	t.Helper()
	list, err := catalog.LoadRaw(rel("../../catalogdata/produits-v3.json"))
	if err != nil {
		t.Skip("catalogdata absent (hors dépôt local)")
	}
	catalog.ApplyTaxonomy(list)
	return catalog.NewStore(list)
}

// Normalise une page pour comparaison : retire les différences voulues
// (csrf, scripts htmx/mirror ajoutés, texte compte synchronisé).
var reHtmx = regexp.MustCompile("<script[^>]*htmx[^>]*></script>\n?")
var reMirror = regexp.MustCompile("<script[^>]*mirror[^>]*></script>\n?")
var reVer = regexp.MustCompile(`assets/[a-z0-9_.-]+\.css\?v=\d+|assets/[a-z0-9_.-]+\.js\?v=\d+`)
var reFiltersToggle = regexp.MustCompile(`<button type="button" class="filters-toggle"[^>]*>.*?</button>\n?`)
var reHubImg = regexp.MustCompile(`(<span class="hub-img"><img src=")[^"]*(" alt=")[^"]*(")`)
var reMegaAll = regexp.MustCompile(`<a class="mega-all" href="[^"]*"`)
var reHubMain = regexp.MustCompile(`<a class="hub-main" href="[^"]*"`)
var reHubSub = regexp.MustCompile(`<li><a href="[^"]*">`)
var reMegaGroups = regexp.MustCompile(`(<div class="mega-groups">).*?(</div><a class="mega-all")`)
var reMegaCount = regexp.MustCompile(`<span class="mega-count">\(\d+\)</span>`)
var reThemeJs = regexp.MustCompile(`<script src="[^"]*theme\.js[^"]*"></script>\n?`)
var reHrowPrev = regexp.MustCompile(`<div class="hrow-wrap"><button type="button" class="hrow-btn prev" data-hrow="-1" aria-label="Précédent">‹</button>`)
var reHrowNext = regexp.MustCompile(`<button type="button" class="hrow-btn next" data-hrow="1" aria-label="Suivant">›</button></div>`)
var reNewsletter = regexp.MustCompile(`<form[^>]*><input type="email"[^>]*><button type="[a-z]*"[^>]*>S'inscrire</button></form>`)
var rePromoInput = regexp.MustCompile(`<label class="drawer-promo"[^>]*><input data-promo-input[^>]*></label>\n?`)
var reAvatarMenu = regexp.MustCompile(`<div class="avatar-menu"[^>]*>.*?</div>\n?`)
var reViewBtn = regexp.MustCompile(`<button type="button" class="[^"]*" data-view="([a-z]+)">`)
var reGamme = regexp.MustCompile(`(?s)<div class="gamme">.*?(<div style="margin-top:12px)`)
var reGalCarousel = regexp.MustCompile(`(?s)<div class="gal" data-gal>.*?<img src="([^"]*)"[^>]*data-gal-main>.*?</div></div>`)
var reGalSimple = regexp.MustCompile(`(?s)<div class="gal"><img src="([^"]*)"[^>]*></div>`)

func normPage(s string) string {
	s = reHtmx.ReplaceAllString(s, "")
	s = reMirror.ReplaceAllString(s, "")
	// theme.js : script externe ajouté après v3.25 (zéro script inline).
	s = reThemeJs.ReplaceAllString(s, "")
	// Rangées défilantes de l'accueil (ajout volontaire) : retire wrappers et
	// flèches pour comparer le contenu des cartes à la v3.25.
	s = reHrowPrev.ReplaceAllString(s, "")
	s = reHrowNext.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, `catalog view-grid hrow"`, `catalog view-grid"`)
	s = strings.ReplaceAll(s, `top8 hrow"`, `top8"`)
	// Vue grille/liste : défaut volontairement passé à grille (choix produit),
	// la parité v3.25 (liste) ne porte que sur le contenu.
	s = strings.ReplaceAll(s, `catalog view-list"`, `catalog view-VIEW"`)
	s = strings.ReplaceAll(s, `catalog view-grid"`, `catalog view-VIEW"`)
	s = strings.ReplaceAll(s, `catalog view-thumbs"`, `catalog view-VIEW"`)
	s = reViewBtn.ReplaceAllString(s, `<button data-view="$1">`)
	// Newsletter : <form action="mailto:"> remplacé par JS (même UX, sans Mixed Content).
	s = reNewsletter.ReplaceAllString(s, `<form>NEWSLETTER</form>`)
	// Champ code promo du panier (ajout volontaire) : absent de la v3.25,
	// retiré avant comparaison — vérifié par E2E Playwright (input présent,
	// code repris dans le lien WhatsApp, sans erreur JS).
	s = rePromoInput.ReplaceAllString(s, "")
	// Menu avatar (ajout volontaire) : même traitement — vérifié par
	// TestTopAvatarMenu/TestHeaderAvatarMenu + E2E Playwright (ouverture,
	// navigation, déconnexion).
	s = reAvatarMenu.ReplaceAllString(s, "")
	s = strings.ReplaceAll(s, ` data-auth-open data-avatar aria-haspopup="true" aria-expanded="false"`, ` data-auth-open`)
	s = reVer.ReplaceAllString(s, "assets/X")
	s = reFiltersToggle.ReplaceAllString(s, "")
	// « Voir tous » du méga-menu : pointe maintenant vers la page catégorie
	// (hub des types d'abord), plus vers la boutique filtrée — vérifié par
	// son propre test, pas par la parité avec l'archive v3.25.
	s = reMegaAll.ReplaceAllString(s, `<a class="mega-all" href="MEGAALL"`)
	// Tuiles « Nos catégories » : mènent aux pages catégories (hub des types
	// d'abord), plus à la boutique filtrée — vérifié par TestBoutiqueHubLinks.
	s = reHubMain.ReplaceAllString(s, `<a class="hub-main" href="HUBMAIN"`)
	s = reHubSub.ReplaceAllString(s, `<li><a href="HUBSUB">`)
	// Méga-menu : les sous-catégories à 0 produit ne sont plus liées (plus de
	// page vide) et « Voir + » ne liste que les sous-catégories vivantes —
	// vérifié par TestFilterLinksAlive, pas par la parité v3.25.
	// Méga-menu : groupes de sous-catégories (données 171 vs 579 produits) et
	// compteurs du rail — vérifiés par TestFilterLinksAlive/TestMegaMenuVoirTous.
	s = reMegaGroups.ReplaceAllString(s, `${1}GROUPES${2}`)
	s = reMegaCount.ReplaceAllString(s, `<span class="mega-count">(N)</span>`)
	// Hub des types (pages catégories, nouveau) : la fermeture du bloc
	// [data-catalog-home] se colle à celle du catalogue.
	s = strings.ReplaceAll(s, "</div></div>\n</div></div></div>", "</div>\n</div></div></div>")
	// Bloc « Modèles de la gamme » (ajout volontaire 2026-11-25) : absent de
	// la v3.25, retiré avant comparaison — vérifié par TestFicheBlocGamme.
	s = reGamme.ReplaceAllString(s, "$1")
	// Carrousel (ajout volontaire) : la v3.25 n'affichait qu'une photo. On
	// compare la photo principale, sans les flèches ni les miniatures —
	// la structure du carrousel est couverte par gallery_test.go.
	var out []string
	for _, ln := range strings.Split(s, "\n") {
		if strings.Contains(ln, "csrf-token") {
			continue
		}
		// Scripts inline du v3.25 (thème + flag js) : externalisés en theme.js.
		t := strings.TrimSpace(ln)
		if t == `<script>try{var _t=JSON.parse(localStorage.getItem('mediatec-theme')||'{}');var _r=document.documentElement;if(_t.mode==='sombre')_r.setAttribute('data-theme','dark');if(_t.color==='liquid')_r.setAttribute('data-color','liquid');}catch(e){}</script>` ||
			t == `<script>document.documentElement.classList.add('js')</script>` {
			continue
		}
		// Nouveautés retirées de la parité (vérifiées par leurs propres tests) :
		// le hub des types, le bloc [data-catalog-home] des pages catégories,
		// le filtre global de l'accueil et sa carte des types.
		if strings.Contains(ln, "data-hub") || strings.Contains(ln, "data-catalog-home hidden") || strings.Contains(ln, "data-home-filter") || strings.Contains(ln, "MEDIATEC_TYPES") || strings.Contains(ln, "mediatec-types") {
			continue
		}
		ln = strings.ReplaceAll(ln, "Votre compte est synchronisé sur tous vos appareils : connectez-vous pour retrouver votre panier et vos commandes.", "Votre compte est enregistré uniquement dans ce navigateur (site statique, sans serveur) : il n'est pas synchronisé avec un autre appareil. Pensez à noter votre mot de passe.")
		out = append(out, ln)
	}
	return strings.Join(out, "\n")
}

// normPageV325 : normalisation du test de parité v3.25 uniquement. Elle ne
// doit PAS entrer dans normPage, qui sert aussi aux empreintes de
// TestPariteSansDesign (toute modification y change les empreintes).
func normPageV325(s string) string {
	s = normPage(s)
	// Galerie : la v3.25 n'avait qu'une photo, le lot photos ajoute des
	// extras. Contenu couvert par gallery_test.go / gamme_test.go.
	s = reGalCarousel.ReplaceAllString(s, `<div class="gal">@@</div>`)
	s = reGalSimple.ReplaceAllString(s, `<div class="gal">@@</div>`)
	s = strings.ReplaceAll(s, `<div class="gal"><div class="noimg">Photo à venir</div></div>`, `<div class="gal">@@</div>`)
	// Retrait du bloc gamme : rétablir le saut de ligne consommé.
	s = strings.ReplaceAll(s, `</a><div style="margin-top:12px`, "</a>\n<div style=\"margin-top:12px")
	return s
}

func readOrig(t *testing.T, path string) string {
	t.Helper()
	b, err := os.ReadFile(rel("../../site-orig/" + path))
	if err != nil {
		t.Skip("site-orig absent")
	}
	return string(b)
}

// renommages2026 : corrections de noms produits du 2026-11-25 (contenu
// éditorial voulu, cf. plan/RAPPORT-CORRECTION-NOMS.md). La référence
// site-orig reste la v3.25 : on y applique ces remplacements avant de
// comparer, pour que la parité continue de porter à l'octet près sur la
// structure, les prix, les descriptions et les attributs.
var renommages2026 = []struct{ ancien, nouveau string }{
	{"Toner 655A YELLOW", "Toner HP 655A Jaune (CF452A)"},
	{"Toner 659A NOIR", "Toner HP 659A Noir (W2010A)"},
	{"Toner 78A NOIR", "Toner HP 78A Noir (CE278A)"},
	{"Toner 79A NOIR", "Toner HP 79A Noir (CF279A)"},
	{"Toner 80A NOIR", "Toner HP 80A Noir (CF280A)"},
	{"Toner 827A NOIR", "Toner HP 827A Noir (CF300A)"},
	{"Toner 85A NOIR", "Toner HP 85A Noir (CE285A)"},
	{"Toner 90A NOIR", "Toner HP 90A Noir (CE390A)"},
	{"Toner CEXV32 NOIR", "Toner Canon C-EXV32 Noir"},
	{"Toner CEXV42 NOIR", "Toner Canon C-EXV42 Noir"},
	{"Toner CEXV58 MAGENTA", "Toner Canon C-EXV58 Magenta"},
	{"Toner CEXV63 NOIR", "Toner Canon C-EXV63 Noir"},
	{"Toner TK-410 TONER KIT", "Toner Kyocera TK-410 Noir"},
	{"Toner TK-450 TONER KIT", "Toner Kyocera TK-450 Noir"},
	// Lot 2 (même date) : type court ajouté devant marque + modèle.
	{"Cadre saillie 40 mm Mosaic pour 2 modules", "Cadre saillie Legrand Mosaic 40 mm pour 2 modules"},
	{"Dell E1715S", "Écran Dell E1715S"},
	{"Dell E2720H", "Écran Dell E2720H"},
	{"HP P19b G4", "Écran HP P19b G4"},
	{"HP CS500", "Kit clavier et souris HP CS500"},
	{"HP CS700", "Kit clavier et souris HP CS700"},
	{"Kodak Alaris i3500", "Scanner Kodak Alaris i3500"},
	{"LG 27UP85NP-W", "Écran LG 27UP85NP-W"},
	{"Cisco Meraki MX105", "Appliance de sécurité Cisco Meraki MX105"},
	{"Cisco Meraki MX250", "Appliance de sécurité Cisco Meraki MX250"},
	{"NComputing L300", "Client léger NComputing L300"},
	{"Prise RJ45 Cat. 6 STP blindée Mosaic 1 module", "Prise RJ45 Legrand Cat. 6 STP blindée Mosaic 1 module"},
	{"Fujitsu fi-7460", "Scanner Fujitsu fi-7460"},
	{"Epson LQ-2190", "Imprimante matricielle Epson LQ-2190"},
	{"Logitech C270 HD", "Webcam Logitech C270 HD"},
}

// qescTest reprend l'encodage des liens rendus (qesc dans render2.go).
func qescTest(s string) string {
	q := url.QueryEscape(s)
	q = strings.ReplaceAll(q, "+", "%20")
	return strings.ReplaceAll(q, "%2F", "/")
}

// applyRenommages ne touche QUE les positions qui portent le nom affiché :
// titre, attributs alt/data-nom, texte de lien/titre, et liens percent-encodés.
// Un remplacement global réécrirait les descriptions (« Le <nom> est … »).
func applyRenommages(s string) string {
	for _, r := range renommages2026 {
		remplacements := [][2]string{
			{"<title>" + r.ancien, "<title>" + r.nouveau},
			{`alt="` + r.ancien + `"`, `alt="` + r.nouveau + `"`},
			{`data-nom="` + r.ancien + `"`, `data-nom="` + r.nouveau + `"`},
			// Attributs normalisés (minuscules sans accents) : ancrés au début
			// de la valeur, sinon on toucherait le nom dans la description.
			{`data-name="` + catalog.NormName(r.ancien), `data-name="` + catalog.NormName(r.nouveau)},
			{`data-search="` + catalog.NormName(r.ancien), `data-search="` + catalog.NormName(r.nouveau)},
			{">" + r.ancien + "<", ">" + r.nouveau + "<"},
			// Fil d'Ariane : le nom est le dernier segment, avant </div>.
			{"› " + r.ancien, "› " + r.nouveau},
			// Liens devis / WhatsApp : même encodage que qesc() dans render2.go.
			{qescTest(r.ancien), qescTest(r.nouveau)},
		}
		for _, m := range remplacements {
			s = strings.ReplaceAll(s, m[0], m[1])
		}
	}
	return s
}

// readOrigNorm : référence site-orig normalisée, avec les renommages
// volontaires appliqués (le rendu courant, lui, porte déjà les nouveaux noms).
func readOrigNorm(t *testing.T, path string) string {
	t.Helper()
	return normPageV325(applyRenommages(readOrig(t, path)))
}

func TestBoutiqueParity(t *testing.T) {
	s := testStore(t)
	got := normPageV325(render.Doc("Store", render.BoutiqueBody(s), "", ""))
	want := readOrigNorm(t, "boutique.html")
	// Le tri "stock d'abord" réordonne les cartes : même jeu de cartes,
	// reste de la page identique à l'octet près.
	assertSameCards(t, got, want)
}

func TestIndexParity(t *testing.T) {
	s := testStore(t)
	recents := s.Recents(rel("../../catalogdata/store-sitemap.xml"))
	got := dropHomeLiveBlocks(normPageV325(render.Doc("Accueil", render.IndexBody(s, recents), "", "")))
	want := dropHomeLiveBlocks(readOrigNorm(t, "index.html"))
	if got != want {
		diffFirst(t, got, want)
	}
}

// dropHomeLiveBlocks retire les sections d'accueil dont le contenu est vivant
// (« Ajouté récemment » = derniers produits importés) ou refondu (cartes de
// marques avec logos), qui ne peuvent plus être comparées à la v3.25 : les
// produits changent à chaque import, et les marques sont passées de simples
// libellés à des cartes cliquables. Ces deux blocs sont couverts par
// TestAccueilRecents / TestAccueilMarques et par l'empreinte sans design.
func dropHomeLiveBlocks(s string) string {
	starts := []string{
		`<div class="sec"><h2>Ajouté récemment</h2>`,
		`<div class="sec"><h2>Nos marques</h2>`,
		`<div class="sec"><div class="brands">`,
	}
	for {
		idx := -1
		for _, st := range starts {
			if i := strings.Index(s, st); i >= 0 && (idx < 0 || i < idx) {
				idx = i
			}
		}
		if idx < 0 {
			break
		}
		// Fin du bloc : prochaine section de premier niveau (les blocs sont
		// concaténés avec un saut de ligne).
		end := len(s)
		for _, st := range []string{"\n<div class=\"sec\">", "\n<div class=\"widgets\">", "\n<div class=\"banner\""} {
			if i := strings.Index(s[idx:], st); i >= 0 && idx+i < end {
				end = idx + i
			}
		}
		s = s[:idx] + s[end:]
	}
	return s
}

func TestCategoriesParity(t *testing.T) {
	s := testStore(t)
	for _, m := range catalog.MainOrder {
		got := normPageV325(render.Doc(m, render.CategoryPage(s, m), "../", ""))
		want := readOrigNorm(t, "categories/"+catalog.MainSlug[m]+".html")
		t.Logf("catégorie %s", m)
		assertSameCards(t, got, want)
	}
}

func TestProduitsParity(t *testing.T) {
	s := testStore(t)
	for _, p := range s.Products {
		got := normPageV325(render.Doc(p.Nom, render.Header(s, "../", "")+render.FicheBody(s, p)+"\n"+render.Footer(s, "../"), "../", ""))
		want := readOrigNorm(t, "produits/"+p.Slug+".html")
		if got != want {
			t.Logf("produit %s", p.Slug)
			diffFirst(t, got, want)
		}
	}
}

func TestFacetsParity(t *testing.T) {
	s := testStore(t)
	raw, err := os.ReadFile(rel("../../site-orig/assets/facets.js"))
	if err != nil {
		t.Skip("site-orig absent")
	}
	py := strings.TrimPrefix(string(raw), "/*")
	// extrait l'objet JSON : de "window.MEDIATEC_FACETS = " jusqu'au dernier ";"
	i := strings.Index(py, "window.MEDIATEC_FACETS = ")
	j := strings.LastIndex(py, ";")
	var want any
	if err := json.Unmarshal([]byte(py[i+len("window.MEDIATEC_FACETS = "):j]), &want); err != nil {
		t.Fatal(err)
	}
	var got any
	if err := json.Unmarshal([]byte(s.FacetsJS()[strings.Index(s.FacetsJS(), "{"):strings.LastIndex(s.FacetsJS(), "}")+1]), &got); err != nil {
		t.Fatal(err)
	}
	wb, _ := json.Marshal(want)
	gb, _ := json.Marshal(got)
	if string(wb) != string(gb) {
		t.Fatalf("facettes différentes:\nwant=%s\ngot=%s", wb, gb)
	}
}

// extractCards retire les cartes produit (<div class="card" ...>...</div>,
// profondeur équilibrée) : renvoie la page sans cartes + slugs dans l'ordre.
// Sert aux pages dont seul l'ordre des cartes peut changer (tri stock d'abord).
func extractCards(html string) (string, []string) {
	var out strings.Builder
	var slugs []string
	reSlug := regexp.MustCompile(`data-slug="([^"]+)"`)
	i := 0
	for {
		j := strings.Index(html[i:], `<div class="card"`)
		if j < 0 {
			out.WriteString(html[i:])
			break
		}
		j += i
		out.WriteString(html[i:j])
		depth := 0
		k := j
		for k < len(html) {
			if strings.HasPrefix(html[k:], "<div") {
				depth++
				k += 4
			} else if strings.HasPrefix(html[k:], "</div>") {
				depth--
				k += 6
				if depth == 0 {
					break
				}
			} else {
				k++
			}
		}
		if m := reSlug.FindStringSubmatch(html[j:k]); m != nil {
			slugs = append(slugs, m[1])
		}
		out.WriteString("<!--card-->")
		i = k
	}
	return out.String(), slugs
}

// assertSameCards : hors-cartes strictement identique + même multijeu de cartes.
func assertSameCards(t *testing.T, got, want string) {
	t.Helper()
	// Hub de types : le produit représentant est choisi par ordre alphabétique ;
	// les corrections de noms du 2026-11-25 changent donc légitimement l'image
	// et l'alt affichés. Neutralisé ici seulement (pas dans normPage, pour ne
	// pas toucher aux empreintes de TestPariteSansDesign).
	got = reHubImg.ReplaceAllString(got, "${1}@@${2}@@${3}")
	want = reHubImg.ReplaceAllString(want, "${1}@@${2}@@${3}")
	gs, gslugs := extractCards(got)
	ws, wslugs := extractCards(want)
	if gs != ws {
		diffFirst(t, gs, ws)
	}
	if len(gslugs) != len(wslugs) {
		t.Fatalf("cartes : got %d, want %d", len(gslugs), len(wslugs))
	}
	counts := map[string]int{}
	for _, s := range wslugs {
		counts[s]++
	}
	for _, s := range gslugs {
		counts[s]--
		if counts[s] < 0 {
			t.Fatalf("carte inattendue ou en trop : %s", s)
		}
	}
}

func diffFirst(t *testing.T, got, want string) {
	t.Helper()
	gl, wl := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(gl) && i < len(wl); i++ {
		if gl[i] != wl[i] {
			ga, wa := gl[i], wl[i]
			j := 0
			for j < len(ga) && j < len(wa) && ga[j] == wa[j] {
				j++
			}
			lo := j - 60
			if lo < 0 {
				lo = 0
			}
			hi := j + 120
			if hi > len(ga) {
				hi = len(ga)
			}
			hi2 := j + 120
			if hi2 > len(wa) {
				hi2 = len(wa)
			}
			t.Fatalf("ligne %d diffère à %d:\n got: ...%s...\nwant: ...%s...", i+1, j, ga[lo:hi], wa[lo:hi2])
		}
	}
	t.Fatalf("longueurs différentes: got %d lignes, want %d lignes", len(gl), len(wl))
}
