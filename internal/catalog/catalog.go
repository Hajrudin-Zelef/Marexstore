// Package catalog porte la logique de données de build_site_v3.py :
// taxonomie, live_map, prix, recherche, facettes. Les fonctions de rendu
// HTML sont dans internal/render. L'objectif est l'octet-identique avec
// le site statique v3.25 pour les mêmes fiches sources.
package catalog

import (
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
	"unicode"

	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

const (
	Tel   = "(+225) 07 05 08 75 07"
	Tel2  = "(+225) 27 20 24 33 19"
	Email = "sales@mediatec-ci.com"
	WA    = "2250705087507"
)

// esc reproduit html.escape(quote=True) de Python : & < > " '
// (Go html.EscapeString donnerait &#34;/&#39; — différent, donc fonction dédiée).
func Esc(s string) string {
	r := strings.NewReplacer(
		"&", "&amp;",
		"<", "&lt;",
		">", "&gt;",
		`"`, "&quot;",
		"'", "&#x27;",
	)
	return r.Replace(s)
}

// normName reproduit EXACTEMENT norm_name() de Python : minuscules +
// suppression des diacritiques, RIEN d'autre (tirets et espaces conservés).
func normName(s string) string {
	t, _, _ := transform.String(transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), strings.ToLower(s))
	return t
}

func NormName(s string) string { return normName(s) }

func Slugify(s string) string {
	t, _, _ := transform.String(transform.Chain(norm.NFKD, runes.Remove(runes.In(unicode.Mn)), norm.NFC), strings.ToLower(s))
	re := regexp.MustCompile(`[^a-z0-9]+`)
	return strings.Trim(re.ReplaceAllString(t, "-"), "-")
}

func FmtPrix(p *int64) string {
	if p == nil {
		return "Prix sur demande"
	}
	n := *p
	s := fmt.Sprintf("%d", n)
	var out []byte
	for i, c := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, byte(c))
	}
	return string(out) + " F CFA"
}

// ---------- Fiche ----------

type Spec struct {
	Label  string `json:"label"`
	Valeur string `json:"valeur"`
}

type RawProduct struct {
	Slug         string          `json:"slug"`
	Nom          string          `json:"nom"`
	Marque       string          `json:"marque"`
	Prix         *int64          `json:"prix"`
	PrixBoutique *int64          `json:"prix_boutique"`
	PrixAvant    *int64          `json:"prix_avant"`
	Devise       string          `json:"devise"`
	Dispo        string          `json:"disponibilite"`
	Categorie    string          `json:"categorie"`
	Accroche     string          `json:"accroche"`
	Description  string          `json:"description"`
	Image        string          `json:"image"`
	Specs        json.RawMessage `json:"specs"`
	Ref          string          `json:"ref"`
	RefFabricant string          `json:"reference_fabricant"`
	Stock        *int            `json:"stock"`
	StoreURL     string          `json:"store_url"`
	PointsForts  []string        `json:"points_forts"`
	Consommables []Consommable   `json:"consommables"`
	Options      []any           `json:"options"`
	// dérivés
	Main     string `json:"main"`
	MainSlug string `json:"main_slug"`
	Sub      string `json:"sub"`
	SubSlug  string `json:"sub_slug"`
	LiveTop  string `json:"live_top"`
	LiveSub  string `json:"live_sub"`
}

type Consommable struct {
	Nom  string `json:"nom"`
	Prix string `json:"prix"`
}

func (p *RawProduct) PrixAffiche() *int64 {
	if p.PrixBoutique != nil {
		return p.PrixBoutique
	}
	return p.Prix
}

func LoadRaw(path string) ([]*RawProduct, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var list []*RawProduct
	if err := json.Unmarshal(b, &list); err != nil {
		return nil, err
	}
	return list, nil
}

// ---------- Taxonomie ----------

var MainOrder = []string{"Ordinateurs", "Imprimantes & Consommables", "Moniteur & Écrans",
	"Réseaux", "Stockage", "Serveurs", "Accessoires"}

var MainSlug = map[string]string{}

func init() {
	for _, m := range MainOrder {
		MainSlug[m] = Slugify(m)
	}
}

var catMap = map[string][2]string{
	"Ordinateurs portables":             {"Ordinateurs", "Portables"},
	"Ordinateurs de bureau":             {"Ordinateurs", "Tout-en-Un"},
	"Tablettes":                         {"Ordinateurs", "Tablettes"},
	"Clients légers / Thin clients":     {"Ordinateurs", "Clients légers"},
	"Imprimantes":                       {"Imprimantes & Consommables", "Imprimantes"},
	"Scanners":                          {"Imprimantes & Consommables", "Scanners"},
	"Toners & consommables":             {"Imprimantes & Consommables", "Toners"},
	"Cartouches d'encre":                {"Imprimantes & Consommables", "Cartouches d'encre"},
	"Kits de maintenance":               {"Imprimantes & Consommables", "Kits de maintenance"},
	"Têtes d'impression":                {"Imprimantes & Consommables", "Têtes d'impression"},
	"Consommables impression":           {"Imprimantes & Consommables", "Consommables"},
	"Rubans pour imprimantes de cartes": {"Imprimantes & Consommables", "Rubans"},
	"Écrans":                            {"Moniteur & Écrans", "Écrans"},
	"Cisco & réseau":                    {"Réseaux", "Cisco & réseau"},
	"Réseau & câbles":                   {"Réseaux", "Réseau & câbles"},
	"Réseau Legrand":                    {"Réseaux", "Réseau Legrand"},
	"Adaptateurs réseau":                {"Réseaux", "Adaptateurs"},
	"Téléphonie IP":                     {"Réseaux", "Téléphonie IP"},
	"Stockage":                          {"Stockage", "Disques & stockage"},
	"Stockage / NAS":                    {"Stockage", "NAS"},
	"Clés USB":                          {"Stockage", "Clés USB"},
	"Cartes mémoire":                    {"Stockage", "Cartes mémoire"},
	"Cartouches LTO":                    {"Stockage", "Cartouches LTO"},
	"Mémoire RAM":                       {"Stockage", "Mémoire RAM"},
	"Serveurs":                          {"Serveurs", "Serveurs rack"},
	"Câbles HDMI":                       {"Accessoires", "Câbles HDMI"},
	"Câbles RJ45":                       {"Accessoires", "Câbles RJ45"},
	"Kits clavier / souris":             {"Accessoires", "Kits clavier / souris"},
	"Souris":                            {"Accessoires", "Souris"},
	"Casques & audio":                   {"Accessoires", "Casques & audio"},
	"Audio-conférence":                  {"Accessoires", "Audio-conférence"},
	"Enceintes de soirée":               {"Accessoires", "Enceintes"},
	"Sacs & housses":                    {"Accessoires", "Sacs & housses"},
	"Pièces de rechange":                {"Accessoires", "Pièces de rechange"},
	"Énergie & protection":              {"Accessoires", "Énergie & protection"},
	"Accessoires scanners":              {"Accessoires", "Accessoires scanners"},
	"GPS / Navigation":                  {"Accessoires", "GPS"},
}

type LiveSub struct {
	Label string
	Slug  string
}
type LiveTop struct {
	Label string
	Slug  string
	Subs  []LiveSub
}

var LiveCats = []LiveTop{
	{"Accessoires Informatiques", "accessoires-informatiques", []LiveSub{
		{"Autres accessoires informatiques", "autres-accessoires-informatiques"},
		{"Câbles & Adaptateurs", "cables-adaptateurs"},
		{"Cartouche d'encre", "cartouche-dencre"},
		{"Casques & Écouteurs", "casques-ecouteurs"},
		{"Chargeurs & Batteries", "chargeurs-batteries"},
		{"Claviers", "claviers"},
		{"Clés USB & Cartes mémoire", "cles-usb-cartes-memoire"},
		{"Disques durs & SSD externes", "disques-durs-ssd-externes"},
		{"Sacs & Housses pour laptop", "sacs-housses-pour-laptop"},
		{"Souris", "souris"},
		{"Supports & Stations d'accueil", "supports-stations-daccueil"},
		{"Tapis de souris", "tapis-de-souris"},
		{"Webcams", "webcams"},
	}},
	{"Accessoires Téléphones & Tablettes", "accessoires-telephones-tablettes", []LiveSub{
		{"Autres accessoires de telephones", "autres-accessoires-de-telephones"},
		{"Batteries externes (Power Bank)", "batteries-externes-power-bank"},
		{"Cartes mémoire & Clés USB OTG", "cartes-memoire-cles-usb-otg"},
		{"Chargeurs & Câbles", "chargeurs-cables"},
		{"Coques & Étuis", "coques-etuis"},
		{"Protections d'écran (verre trempé, film)", "protections-decran-verre-trempe-film"},
		{"Selfie sticks & Trépieds", "selfie-sticks-trepieds"},
		{"Stylets", "stylets"},
		{"Supports & Stations de charge", "supports-stations-de-charge"},
		{"Écouteurs & Oreillettes", "ecouteurs-oreillettes"},
	}},
	{"Caméras & Vidéosurveillances", "cameras-videosurveillances", []LiveSub{
		{"Caméra IP", "camera-ip"}, {"Caméra Wi-Fi", "camera-wi-fi"},
		{"Caméra analogique (CVI/AHD)", "camera-analogique-cvi-ahd"},
		{"Caméra solaire/autonome", "camera-solaire-autonome"},
		{"Dahua", "dahua"}, {"Ezviz", "ezviz"}, {"Hikvision", "hikvision"},
		{"Kit vidéosurveillance complet (caméras + DVR/NVR)", "kit-videosurveillance-complet-cameras-dvr-nvr"},
		{"TP-Link", "tp-link"}, {"Xiaomi", "xiaomi"},
	}},
	{"Imprimantes", "imprimantes", []LiveSub{
		{"Brother", "brother"}, {"Canon", "canon"}, {"Epson", "epson"},
		{"Jet d'encre", "jet-dencre"}, {"Laser", "laser"},
		{"Multifonction", "multifonction"}, {"HP", "hp-imprimantes"},
	}},
	{"Laptops neuf", "laptots", []LiveSub{
		{"Acer", "acer"}, {"Apple", "apple"}, {"Asus", "asus"}, {"Dell", "dell"},
		{"HP", "hp"}, {"Laptop Gaming", "laptop-gaming"}, {"Lenovo", "lenovo"}, {"MSI", "msi"},
	}},
	{"Matériels réseaux", "materiels-reseaux", []LiveSub{}},
	{"Ordinateurs de Bureau Neufs", "ordinateurs-de-bureau-neufs", []LiveSub{
		{"Acer", "acer-ordinateurs-de-bureau-neufs"},
		{"Apple (iMac / Mac mini)", "apple-imac-mac-mini"},
		{"Asus", "asus-ordinateurs-de-bureau-neufs"},
		{"Bureau All-in-One", "bureau-all-in-one"},
		{"Dell", "dell-ordinateurs-de-bureau-neufs"},
		{"Ecrans bureaux", "ecrans-bureaux"},
		{"HP", "hp-ordinateurs-de-bureau-neufs"},
		{"Lenovo", "lenovo-ordinateurs-de-bureau-neufs"},
	}},
	{"Reconditionné & Casi neuf", "casi-neuf", []LiveSub{
		{"Autres Appareils Reconditionnés", "autres-reconditionne"},
		{"Imprimantes casi neuf", "imprimantes-casi-neuf"},
		{"Laptops reconditionnés", "laptops-reconditionnes"},
		{"Ordinateurs Bureau casi neuf", "ordinateurs-bureau-casi-neuf"},
	}},
	{"Serveur", "serveur", []LiveSub{}},
	{"Téléphones & Tablettes", "telephones-tablettes", []LiveSub{
		{"Autres telephones", "autres-telephones"}, {"Huawei", "huawei"},
		{"Infinix", "infinix"}, {"Iphone", "iphone"}, {"Itel", "itel"},
		{"Samsung", "samsung"}, {"Smartphone", "smartphone"},
		{"Tablette", "tablette"}, {"Tecno", "tecno"},
		{"Xiaomi", "xiaomi-telephones-tablettes"},
	}},
	{"Uncategorized", "uncategorized", []LiveSub{}},
}

var LiveTops = map[string][]LiveSub{}
var SubLabels = map[[2]string]string{}

func init() {
	for _, t := range LiveCats {
		LiveTops[t.Slug] = t.Subs
		for _, s := range t.Subs {
			SubLabels[[2]string{t.Slug, s.Slug}] = s.Label
		}
	}
}

type MegaGroup struct {
	Title string
	Subs  []string
}

var MegaGroups = map[string][]MegaGroup{
	"accessoires-informatiques": {
		{"Cartouches & encres", []string{"cartouche-dencre"}},
		{"Stockage", []string{"disques-durs-ssd-externes", "cles-usb-cartes-memoire"}},
		{"Périphériques", []string{"souris", "claviers", "casques-ecouteurs", "webcams", "tapis-de-souris"}},
		{"Câbles & alimentation", []string{"cables-adaptateurs", "chargeurs-batteries"}},
		{"Transport & supports", []string{"sacs-housses-pour-laptop", "supports-stations-daccueil"}},
		{"Autres", []string{"autres-accessoires-informatiques"}},
	},
	"accessoires-telephones-tablettes": {
		{"Protection", []string{"coques-etuis", "protections-decran-verre-trempe-film"}},
		{"Charge", []string{"chargeurs-cables", "batteries-externes-power-bank", "supports-stations-de-charge"}},
		{"Audio", []string{"ecouteurs-oreillettes"}},
		{"Stockage & divers", []string{"cartes-memoire-cles-usb-otg", "selfie-sticks-trepieds", "stylets", "autres-accessoires-de-telephones"}},
	},
	"cameras-videosurveillances": {
		{"Caméras", []string{"camera-ip", "camera-wi-fi", "camera-analogique-cvi-ahd", "camera-solaire-autonome", "kit-videosurveillance-complet-cameras-dvr-nvr"}},
		{"Marques", []string{"dahua", "ezviz", "hikvision", "tp-link", "xiaomi"}},
	},
	"imprimantes": {
		{"Technologies", []string{"laser", "jet-dencre", "multifonction"}},
		{"Marques", []string{"brother", "canon", "epson", "hp-imprimantes"}},
	},
	"laptots": {
		{"Marques", []string{"acer", "apple", "asus", "dell", "hp", "lenovo", "msi"}},
		{"Gaming", []string{"laptop-gaming"}},
	},
	"materiels-reseaux": {},
	"ordinateurs-de-bureau-neufs": {
		{"Types", []string{"bureau-all-in-one", "ecrans-bureaux"}},
		{"Marques", []string{"acer-ordinateurs-de-bureau-neufs", "apple-imac-mac-mini", "asus-ordinateurs-de-bureau-neufs", "dell-ordinateurs-de-bureau-neufs", "hp-ordinateurs-de-bureau-neufs", "lenovo-ordinateurs-de-bureau-neufs"}},
	},
	"casi-neuf": {
		{"Reconditionnés", []string{"laptops-reconditionnes", "ordinateurs-bureau-casi-neuf", "imprimantes-casi-neuf", "autres-reconditionne"}},
	},
	"serveur": {},
	"telephones-tablettes": {
		{"Appareils", []string{"smartphone", "tablette"}},
		{"Marques", []string{"huawei", "infinix", "iphone", "itel", "samsung", "tecno", "xiaomi-telephones-tablettes"}},
		{"Autres", []string{"autres-telephones"}},
	},
	"uncategorized": {},
}

var reconditionneSlugs = map[string]bool{"hp-elitebook-840g6": true}

// SpecList normalise specs (dict -> liste DANS L'ORDRE DU JSON,
// comme build_site_v3.py ; liste -> telle quelle).
func (p *RawProduct) SpecList() []Spec {
	raw := strings.TrimSpace(string(p.Specs))
	if raw == "" || raw == "null" {
		return nil
	}
	if strings.HasPrefix(raw, "[") {
		var out []Spec
		if err := json.Unmarshal(p.Specs, &out); err != nil {
			return nil
		}
		return out
	}
	// Objet : ordre du document via Decoder token par token.
	dec := json.NewDecoder(strings.NewReader(raw))
	tok, err := dec.Token()
	if err != nil || tok != json.Delim('{') {
		return nil
	}
	var out []Spec
	for dec.More() {
		k, err := dec.Token()
		if err != nil {
			return nil
		}
		var v string
		if err := dec.Decode(&v); err != nil {
			return nil
		}
		ks, _ := k.(string)
		out = append(out, Spec{Label: ks, Valeur: v})
	}
	return out
}

func specsDict(p *RawProduct) map[string]string {
	d := map[string]string{}
	for _, s := range p.SpecList() {
		if s.Label == "" || s.Valeur == "" {
			continue
		}
		k := strings.TrimSpace(s.Label)
		if _, ok := d[k]; !ok {
			d[k] = strings.TrimSpace(s.Valeur)
		}
	}
	return d
}

func IsReconditionne(p *RawProduct) bool {
	sd := specsDict(p)
	for k, v := range sd {
		if k == "État" && strings.Contains(v, "Reconditionn") {
			return true
		}
	}
	return strings.Contains(strings.ToLower(p.Nom+" "+p.Description), "reconditionn")
}

// LiveMap reproduit live_map() : (live_top, live_sub ou "").
func LiveMap(p *RawProduct) (string, string) {
	marque := strings.ToLower(strings.TrimSpace(p.Marque))
	nom := strings.ToLower(p.Nom)
	sub, main := p.Sub, p.Main
	brandLaptop := map[string]string{"hp": "hp", "lenovo": "lenovo", "dell": "dell", "asus": "asus", "acer": "acer", "apple": "apple", "msi": "msi"}
	brandDesk := map[string]string{"hp": "hp-ordinateurs-de-bureau-neufs", "lenovo": "lenovo-ordinateurs-de-bureau-neufs", "dell": "dell-ordinateurs-de-bureau-neufs", "asus": "asus-ordinateurs-de-bureau-neufs", "acer": "acer-ordinateurs-de-bureau-neufs", "apple": "apple-imac-mac-mini"}
	lower := strings.ToLower
	contains := strings.Contains
	has := func(s string, pats ...string) bool {
		for _, pat := range pats {
			if contains(s, pat) {
				return true
			}
		}
		return false
	}
	_ = lower
	switch {
	case sub == "Portables":
		if IsReconditionne(p) || reconditionneSlugs[p.Slug] {
			return "casi-neuf", "laptops-reconditionnes"
		}
		if has(nom, "gaming", "rog", "tuf") {
			return "laptots", "laptop-gaming"
		}
		return "laptots", brandLaptop[marque]
	case sub == "Tout-en-Un":
		return "ordinateurs-de-bureau-neufs", "bureau-all-in-one"
	case sub == "Clients légers":
		return "ordinateurs-de-bureau-neufs", brandDesk[marque]
	case sub == "Écrans":
		return "ordinateurs-de-bureau-neufs", "ecrans-bureaux"
	case sub == "Tablettes":
		return "telephones-tablettes", "tablette"
	case sub == "Téléphonie IP":
		return "materiels-reseaux", ""
	case sub == "Serveurs rack":
		return "serveur", ""
	case main == "Imprimantes & Consommables":
		if sub == "Toners" || sub == "Cartouches d'encre" {
			return "accessoires-informatiques", "cartouche-dencre"
		}
		if sub == "Imprimantes" {
			if contains(nom, "laser") {
				return "imprimantes", "laser"
			}
			if has(nom, "ecotank", "jet d'encre", "inkjet") {
				return "imprimantes", "jet-dencre"
			}
			if has(nom, "multifonction", "mfp") {
				return "imprimantes", "multifonction"
			}
			brands := map[string]string{"epson": "epson", "hp": "hp-imprimantes", "canon": "canon", "brother": "brother"}
			return "imprimantes", brands[marque]
		}
		if sub == "Scanners" {
			return "imprimantes", ""
		}
		return "accessoires-informatiques", "autres-accessoires-informatiques"
	case main == "Réseaux":
		return "materiels-reseaux", ""
	case sub == "Disques & stockage" || sub == "NAS":
		return "accessoires-informatiques", "disques-durs-ssd-externes"
	case sub == "Clés USB" || sub == "Cartes mémoire":
		return "accessoires-informatiques", "cles-usb-cartes-memoire"
	case sub == "Souris":
		return "accessoires-informatiques", "souris"
	case sub == "Kits clavier / souris":
		return "accessoires-informatiques", "claviers"
	case sub == "Casques & audio" || sub == "Audio-conférence":
		return "accessoires-informatiques", "casques-ecouteurs"
	case sub == "Câbles HDMI" || sub == "Câbles RJ45":
		return "accessoires-informatiques", "cables-adaptateurs"
	case sub == "Sacs & housses":
		return "accessoires-informatiques", "sacs-housses-pour-laptop"
	default:
		return "accessoires-informatiques", "autres-accessoires-informatiques"
	}
}

func ApplyTaxonomy(list []*RawProduct) {
	for _, p := range list {
		if ms, ok := catMap[p.Categorie]; ok {
			p.Main, p.Sub = ms[0], ms[1]
		} else {
			p.Main, p.Sub = "Accessoires", p.Categorie
		}
		p.MainSlug = Slugify(p.Main)
		p.SubSlug = Slugify(p.Sub)
		top, sub := LiveMap(p)
		p.LiveTop, p.LiveSub = top, sub
	}
}
