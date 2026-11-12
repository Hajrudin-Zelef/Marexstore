package catalog

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// ---------- Extracteurs de caractéristiques (port des _* de build_site_v3.py) ----------

func findString(re *regexp.Regexp, s string) string {
	m := re.FindStringSubmatch(s)
	if m == nil {
		return ""
	}
	return m[1]
}

var (
	reGoTo     = regexp.MustCompile(`(?i)(\d+[,.]?\d*)\s*(Go|To)\b`)
	rePouces   = regexp.MustCompile(`(?i)(\d+[,.]?\d*)\s*pouces?`)
	reProcFam  = regexp.MustCompile(`Ultra\s+(\d)|Core i(\d)|Ryzen (?:AI )?(\d)|Apple M(\d)`)
	reRTX      = regexp.MustCompile(`RTX\s*(\d{4})`)
	reRadeon   = regexp.MustCompile(`RX\s*(\d{4})`)
	reCapacite = regexp.MustCompile(`(\d[\d\s]*)\s*pages?`)
	reVitesse  = regexp.MustCompile(`(\d+)\s*ppm`)
	reResol    = regexp.MustCompile(`(\d+)\s*x\s*\d+`)
	reFeuilles = regexp.MustCompile(`([\d\s]+)\s*feuilles`)
	rePortsNb  = regexp.MustCompile(`(?i)(\d+)\s*(?:ports?|x\s)`)
	reLongueur = regexp.MustCompile(`(\d+[,.]?\d*)\s*m\b`)
	rePoids    = regexp.MustCompile(`(\d+[,.]?\d*)\s*kg`)
	rePoidsG   = regexp.MustCompile(`(\d+[,.]?\d*)\s*g\b`)
	reParen    = regexp.MustCompile(`\s*\([^)]*\)\s*`)
)

func goTo(text string) string {
	m := reGoTo.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return m[1] + " " + m[2]
}

func pouces(text string) string {
	m := rePouces.FindStringSubmatch(text)
	if m == nil {
		return ""
	}
	return m[1] + " pouces"
}

func resolution(sd map[string]string) string {
	t := sd["Écran"] + " " + sd["Résolution native"]
	has := func(pat string) bool { m, _ := regexp.MatchString(pat, t); return m }
	switch {
	case has(`4K|3840`):
		return "4K UHD"
	case has(`2,?5K|WQXGA|2560`):
		return "2,5K (WQXGA)"
	case has(`WUXGA|1920\s*x\s*1200`):
		return "WUXGA"
	case has(`Full HD|FHD|1920\s*x\s*1080`):
		return "Full HD"
	case has(`1366|WXGA`):
		return "HD (WXGA)"
	case has(`1280\s*x\s*1024|SXGA`):
		return "SXGA"
	}
	return ""
}

func processeur(p *RawProduct, sd map[string]string) string {
	t := sd["Processeur"] + " " + p.Nom
	var fams []string
	for _, m := range reProcFam.FindAllStringSubmatch(t, -1) {
		var v string
		switch {
		case m[1] != "":
			v = "Intel Core Ultra " + m[1]
		case m[2] != "":
			v = "Intel Core i" + m[2]
		case m[3] != "":
			if strings.Contains(m[0], "AI") {
				v = "AMD Ryzen AI " + m[3]
			} else {
				v = "AMD Ryzen " + m[3]
			}
		default:
			v = "Apple M" + m[4]
		}
		found := false
		for _, f := range fams {
			if f == v {
				found = true
				break
			}
		}
		if !found {
			fams = append(fams, v)
		}
	}
	if len(fams) == 0 {
		return ""
	}
	if len(fams) == 1 {
		return fams[0]
	}
	return fams[0] + " / " + fams[len(fams)-1]
}

func gpu(p *RawProduct, sd map[string]string) string {
	t := sd["Carte graphique"]
	if t == "" {
		t = sd["Graphique"]
	}
	if m := reRTX.FindStringSubmatch(t); m != nil {
		return "NVIDIA GeForce RTX " + m[1]
	}
	switch {
	case strings.Contains(t, "Arc"):
		return "Intel Arc"
	case strings.Contains(t, "Iris Xe"):
		return "Intel Iris Xe"
	case strings.Contains(t, "UHD"):
		return "Intel UHD"
	case strings.Contains(t, "Radeon"):
		if m := reRadeon.FindStringSubmatch(t); m != nil {
			return "AMD Radeon RX " + m[1]
		}
		return "AMD Radeon"
	case strings.Contains(t, "Apple"):
		return "Apple GPU"
	}
	return ""
}

func os_(p *RawProduct, sd map[string]string) string {
	t := sd["Système d'exploitation"]
	var vals []string
	for _, v := range []string{"Windows 11", "Windows 10", "macOS", "ChromeOS", "Linux", "Android", "iOS", "FreeDOS", "Sans OS"} {
		if strings.Contains(t, v) {
			vals = append(vals, v)
		}
	}
	return strings.Join(vals, " / ")
}

func wifi(p *RawProduct, sd map[string]string) string {
	t := sd["Connectivité"] + " " + sd["Connexion"]
	switch {
	case strings.Contains(t, "Wi-Fi 7") || strings.Contains(t, "802.11be"):
		return "Wi-Fi 7"
	case strings.Contains(t, "Wi-Fi 6E"):
		return "Wi-Fi 6E"
	case strings.Contains(t, "Wi-Fi 6") || strings.Contains(t, "802.11ax"):
		return "Wi-Fi 6"
	case strings.Contains(t, "Wi-Fi 5") || strings.Contains(t, "802.11ac"):
		return "Wi-Fi 5"
	}
	return ""
}

func poids(p *RawProduct, sd map[string]string) string {
	t := sd["Poids"]
	var kg float64
	if m := rePoids.FindStringSubmatch(t); m != nil {
		kg, _ = strconv.ParseFloat(strings.ReplaceAll(m[1], ",", "."), 64)
	} else if m := rePoidsG.FindStringSubmatch(t); m != nil {
		g, _ := strconv.ParseFloat(strings.ReplaceAll(m[1], ",", "."), 64)
		kg = g / 1000
	} else {
		return ""
	}
	switch {
	case kg < 1:
		return "Moins de 1 kg"
	case kg < 1.5:
		return "1 - 1,5 kg"
	case kg < 2:
		return "1,5 - 2 kg"
	default:
		return "Plus de 2 kg"
	}
}

func connecteurs(p *RawProduct, sd map[string]string) []string {
	t := sd["Ports"] + " " + sd["Connectivité"] + " " + sd["Entrées vidéo"]
	var out []string
	add := func(pat, lab string) {
		m, _ := regexp.MatchString(`(?i)`+pat, t)
		if m {
			for _, e := range out {
				if e == lab {
					return
				}
			}
			out = append(out, lab)
		}
	}
	add("Thunderbolt", "Thunderbolt")
	add("HDMI", "HDMI")
	add("DisplayPort", "DisplayPort")
	add(`USB Type-C|USB-C|Type-C`, "USB-C")
	add(`USB Type-A|Type-A`, "USB-A")
	add(`RJ-45|Ethernet`, "Ethernet")
	add("microSD", "microSD")
	add("VGA", "VGA")
	add("jack audio", "Jack audio")
	if len(out) == 0 {
		return nil
	}
	return out
}

func activite(p *RawProduct, sd map[string]string) string {
	nom := p.Nom
	if p.LiveSub == "laptop-gaming" {
		return "Gaming"
	}
	m, _ := regexp.MatchString(`(?i)gaming|rog|tuf`, nom)
	if m {
		return "Gaming"
	}
	m, _ = regexp.MatchString(`(?i)elitebook|probook|thinkpad|latitude|vostro`, nom)
	if m {
		return "Bureautique"
	}
	m, _ = regexp.MatchString(`(?i)zenbook`, nom)
	if m {
		return "Création / Polyvalent"
	}
	return "Polyvalent"
}

func couleur(p *RawProduct, sd map[string]string) string {
	return strings.TrimSpace(reParen.ReplaceAllString(sd["Couleur"], ""))
}

func longueur(p *RawProduct, sd map[string]string) string {
	if m := reLongueur.FindStringSubmatch(sd["Longueur"]); m != nil {
		return m[1] + " m"
	}
	return ""
}

func technologie(p *RawProduct, sd map[string]string) string {
	t := sd["Technologie"]
	tl := strings.ToLower(t)
	switch {
	case strings.Contains(tl, "sublimation"):
		return "Sublimation thermique"
	case strings.Contains(tl, "retransfert") || strings.Contains(tl, "hdp"):
		return "Retransfert (cartes)"
	case strings.Contains(tl, "laser") || strings.Contains(tl, "jetintelligence") || strings.Contains(tl, "microfin"):
		if strings.Contains(tl, "couleur") {
			return "Laser couleur"
		}
		return "Laser"
	case strings.Contains(tl, "inkjet") || strings.Contains(tl, "jet d'encre"):
		return "Jet d'encre"
	}
	if len(t) > 40 {
		return t[:40]
	}
	return t
}

func typeCartouche(p *RawProduct, sd map[string]string) string {
	v := sd["Type de cartouche"]
	l := strings.ToLower(v)
	if strings.Contains(l, "haute") {
		return "Haute capacité"
	}
	if strings.Contains(l, "standard") {
		return "Capacité standard"
	}
	return v
}

func rendement(p *RawProduct, sd map[string]string) string {
	m := reCapacite.FindStringSubmatch(sd["Rendement"])
	if m == nil {
		return ""
	}
	n, _ := strconv.Atoi(strings.ReplaceAll(strings.ReplaceAll(m[1], " ", ""), "\u00a0", ""))
	switch {
	case n < 2000:
		return "Moins de 2 000 pages"
	case n < 5000:
		return "2 000 - 5 000 pages"
	case n < 10000:
		return "5 000 - 10 000 pages"
	default:
		return "Plus de 10 000 pages"
	}
}

func typeImprimante(p *RawProduct, sd map[string]string) string {
	t := strings.ToLower(sd["Type"] + " " + p.Nom)
	switch {
	case strings.Contains(t, "scanner") || strings.Contains(t, "alaris"):
		return "Scanner"
	case strings.Contains(t, "matricielle"):
		return "Imprimante matricielle"
	case strings.Contains(t, "multifonction") || strings.Contains(t, "mfp"):
		return "Multifonction"
	case strings.Contains(t, "laser"):
		return "Imprimante laser"
	case strings.Contains(t, "jet d'encre") || strings.Contains(t, "ecotank"):
		return "Imprimante jet d'encre"
	}
	return ""
}

func vitesse(p *RawProduct, sd map[string]string) string {
	v := sd["Vitesse"]
	if v == "" {
		v = sd["Vitesse d'impression"]
	}
	if m := reVitesse.FindStringSubmatch(v); m != nil {
		return m[1] + " ppm"
	}
	return ""
}

func resolutionImp(p *RawProduct, sd map[string]string) string {
	v := sd["Résolution"]
	if v == "" {
		v = sd["Résolution d'impression"]
	}
	if m := reResol.FindStringSubmatch(v); m != nil {
		return m[1] + " dpi"
	}
	return ""
}

func formatPapier(p *RawProduct, sd map[string]string) string {
	t := sd["Type"] + " " + sd["Vitesse"] + " " + p.Nom
	if strings.Contains(t, "A3") {
		return "A3"
	}
	if strings.Contains(t, "A4") {
		return "A4"
	}
	return ""
}

func rectoVerso(p *RawProduct, sd map[string]string) string {
	t := strings.ToLower(sd["Recto-verso"])
	if strings.Contains(t, "automatique") {
		return "Automatique"
	}
	if strings.Contains(t, "manuel") {
		return "Manuel"
	}
	return ""
}

// group1000 reproduit f"{n:,}".replace(",", " ") de Python.
func group1000(n int) string {
	s := strconv.Itoa(n)
	var out []byte
	for i := range s {
		if i > 0 && (len(s)-i)%3 == 0 {
			out = append(out, ' ')
		}
		out = append(out, s[i])
	}
	return string(out)
}

func capacitePapier(p *RawProduct, sd map[string]string) string {
	m := reFeuilles.FindStringSubmatch(sd["Capacité papier"])
	if m == nil {
		return ""
	}
	n, _ := strconv.Atoi(strings.ReplaceAll(strings.ReplaceAll(m[1], " ", ""), "\u00a0", ""))
	return group1000(n) + " feuilles"
}

func typeCable(p *RawProduct, sd map[string]string) string {
	nom := p.Nom
	for _, pr := range [][2]string{{"HDMI", "HDMI"}, {"RJ45|Ethernet", "Ethernet (RJ45)"}, {"USB-C", "USB-C"}, {"DisplayPort", "DisplayPort"}, {"VGA", "VGA"}, {"fibre optique", "Fibre optique"}, {"USB", "USB"}} {
		m, _ := regexp.MatchString(`(?i)`+pr[0], nom)
		if m {
			return pr[1]
		}
	}
	return ""
}

func typeDisque(p *RawProduct, sd map[string]string) string {
	nom := strings.ToUpper(p.Nom + " " + sd["Type"])
	if strings.Contains(nom, "SSD") {
		return "SSD"
	}
	if strings.Contains(nom, "HDD") || strings.Contains(nom, "DISQUE DUR") {
		return "Disque dur (HDD)"
	}
	return ""
}

func iface(p *RawProduct, sd map[string]string) string {
	t := sd["Interface"]
	switch {
	case strings.Contains(t, "USB 3"):
		return "USB 3.x"
	case strings.Contains(t, "USB"):
		return "USB"
	case strings.Contains(t, "SAS"):
		return "SAS"
	case strings.Contains(t, "SATA"):
		return "SATA"
	case strings.Contains(t, "Micro-USB"):
		return "Micro-USB"
	}
	if len(t) > 30 {
		return t[:30]
	}
	return t
}

func typeReseau(p *RawProduct, sd map[string]string) string {
	t := strings.ToLower(sd["Type"] + " " + p.Nom)
	has := func(s string) bool { return strings.Contains(t, s) }
	switch {
	case has("switch") || has("commutateur"):
		return "Switch"
	case has("routeur") || has("router") || has("isr"):
		return "Routeur"
	case has("point d'accès") || has("aironet") || has("access point"):
		return "Point d'accès"
	case has("téléphone") || has("téléphonie"):
		return "Téléphonie IP"
	case has("appliance") || has("meraki") || has("sécurité"):
		return "Sécurité / SD-WAN"
	case has("antenne"):
		return "Antenne"
	case has("module"):
		return "Module réseau"
	case has("câble") || has("cable"):
		return "Câble"
	case has("rack") || has("montage"):
		return "Accessoire rack"
	case has("adaptateur"):
		return "Adaptateur réseau"
	}
	return ""
}

func poe(p *RawProduct, sd map[string]string) string {
	var vals []string
	for _, v := range sd {
		vals = append(vals, v)
	}
	if strings.Contains(strings.ToLower(p.Nom+" "+strings.Join(vals, " ")), "poe") {
		return "Oui"
	}
	return "Non"
}

func portsNb(p *RawProduct, sd map[string]string) string {
	v := sd["Ports"]
	if v == "" {
		v = sd["Ports Ethernet"]
	}
	if m := rePortsNb.FindStringSubmatch(v); m != nil {
		n, _ := strconv.Atoi(m[1])
		if n == 1 {
			return "1 port"
		}
		return fmt.Sprintf("%d ports", n)
	}
	return ""
}

func connexion(p *RawProduct, sd map[string]string) string {
	var keys []string
	for k := range sd {
		keys = append(keys, k)
	}
	var vals []string
	for _, v := range sd {
		vals = append(vals, v)
	}
	txt := strings.ToLower(p.Nom + " " + strings.Join(keys, " ") + " " + strings.Join(vals, " "))
	switch {
	case strings.Contains(txt, "sans fil") || strings.Contains(txt, "wireless") || strings.Contains(txt, "bluetooth") || strings.Contains(txt, "portée sans fil"):
		return "Sans fil"
	case strings.Contains(txt, "filaire") || strings.Contains(txt, "longueur des câbles") || strings.Contains(txt, "usb"):
		return "Filaire (USB)"
	}
	return ""
}

func dalle(p *RawProduct, sd map[string]string) string {
	t := sd["Type de dalle"]
	if t == "" {
		t = sd["Écran"]
	}
	for _, d := range []string{"OLED", "AMOLED", "Mini LED", "IPS", "VA", "TN"} {
		if strings.Contains(t, d) {
			return d
		}
	}
	return ""
}

func tactile(p *RawProduct, sd map[string]string) string {
	t := sd["Écran"]
	if t == "" {
		return ""
	}
	if strings.Contains(strings.ToLower(t), "tactile") {
		return "Oui"
	}
	return "Non"
}

type facetRule struct {
	Label string
	Slug  string
	Tops  map[string]bool // nil = tous
	Subs  map[string]bool // nil = tous
	Fn    func(*RawProduct, map[string]string) any
}

func set(ss ...string) map[string]bool {
	m := map[string]bool{}
	for _, s := range ss {
		m[s] = true
	}
	return m
}

var laptops = set("laptots", "casi-neuf")

var FacetRules = []facetRule{
	{"Taille de l'écran", "ecran-taille", map[string]bool{"laptots": true, "casi-neuf": true, "ordinateurs-de-bureau-neufs": true, "telephones-tablettes": true}, nil, func(f *RawProduct, sd map[string]string) any {
		return pouces(sd["Écran"] + " " + sd["Taille d'écran"])
	}},
	{"Processeur", "processeur", map[string]bool{"laptots": true, "casi-neuf": true, "serveur": true, "ordinateurs-de-bureau-neufs": true}, nil, func(f *RawProduct, sd map[string]string) any { return processeur(f, sd) }},
	{"Chipset graphique", "gpu", laptops, nil, func(f *RawProduct, sd map[string]string) any { return gpu(f, sd) }},
	{"Système d'exploitation", "os", map[string]bool{"laptots": true, "casi-neuf": true, "telephones-tablettes": true}, nil, func(f *RawProduct, sd map[string]string) any { return os_(f, sd) }},
	{"Taille de la mémoire (Go)", "memoire", map[string]bool{"laptots": true, "casi-neuf": true, "serveur": true}, nil, func(f *RawProduct, sd map[string]string) any { return goTo(sd["Mémoire"] + " " + sd["Mémoire vive"]) }},
	{"Disque fourni", "stockage", laptops, nil, func(f *RawProduct, sd map[string]string) any { return goTo(sd["Stockage"]) }},
	{"Résolution Max", "resolution", map[string]bool{"laptots": true, "casi-neuf": true, "ordinateurs-de-bureau-neufs": true}, nil, func(f *RawProduct, sd map[string]string) any { return resolution(sd) }},
	{"Type de Dalle", "dalle", map[string]bool{"laptots": true, "casi-neuf": true, "ordinateurs-de-bureau-neufs": true}, nil, func(f *RawProduct, sd map[string]string) any { return dalle(f, sd) }},
	{"Écran tactile", "tactile", laptops, nil, func(f *RawProduct, sd map[string]string) any { return tactile(f, sd) }},
	{"Norme Wi-Fi", "wifi", laptops, nil, func(f *RawProduct, sd map[string]string) any { return wifi(f, sd) }},
	{"Poids", "poids", laptops, nil, func(f *RawProduct, sd map[string]string) any { return poids(f, sd) }},
	{"Connecteur(s) disponible(s)", "connecteurs", map[string]bool{"laptots": true, "casi-neuf": true, "ordinateurs-de-bureau-neufs": true}, nil, func(f *RawProduct, sd map[string]string) any { return connecteurs(f, sd) }},
	{"Type d'activités", "activite", laptops, nil, func(f *RawProduct, sd map[string]string) any { return activite(f, sd) }},
	{"Couleur", "couleur", set("accessoires-informatiques"), nil, func(f *RawProduct, sd map[string]string) any { return couleur(f, sd) }},
	{"Technologie", "technologie", set("accessoires-informatiques"), nil, func(f *RawProduct, sd map[string]string) any { return technologie(f, sd) }},
	{"Type de cartouche", "type-cartouche", nil, set("cartouche-dencre"), func(f *RawProduct, sd map[string]string) any { return typeCartouche(f, sd) }},
	{"Gamme", "gamme", set("accessoires-informatiques"), nil, func(f *RawProduct, sd map[string]string) any { v := sd["Gamme"]; return v }},
	{"Rendement", "rendement", nil, set("cartouche-dencre"), func(f *RawProduct, sd map[string]string) any { return rendement(f, sd) }},
	{"Type", "type-imprimante", set("imprimantes"), nil, func(f *RawProduct, sd map[string]string) any { return typeImprimante(f, sd) }},
	{"Vitesse d'impression", "vitesse", set("imprimantes"), nil, func(f *RawProduct, sd map[string]string) any { return vitesse(f, sd) }},
	{"Résolution d'impression", "resolution-imp", set("imprimantes"), nil, func(f *RawProduct, sd map[string]string) any { return resolutionImp(f, sd) }},
	{"Format papier", "format", set("imprimantes"), nil, func(f *RawProduct, sd map[string]string) any { return formatPapier(f, sd) }},
	{"Recto-verso", "recto-verso", set("imprimantes"), nil, func(f *RawProduct, sd map[string]string) any { return rectoVerso(f, sd) }},
	{"Capacité papier", "capacite-papier", set("imprimantes"), nil, func(f *RawProduct, sd map[string]string) any { return capacitePapier(f, sd) }},
	{"Longueur", "longueur", nil, set("cables-adaptateurs"), func(f *RawProduct, sd map[string]string) any { return longueur(f, sd) }},
	{"Type de câble", "type-cable", nil, set("cables-adaptateurs"), func(f *RawProduct, sd map[string]string) any { return typeCable(f, sd) }},
	{"Capacité", "capacite", nil, set("disques-durs-ssd-externes", "cles-usb-cartes-memoire"), func(f *RawProduct, sd map[string]string) any { return goTo(sd["Capacité"]) }},
	{"Type de disque", "type-disque", nil, set("disques-durs-ssd-externes"), func(f *RawProduct, sd map[string]string) any { return typeDisque(f, sd) }},
	{"Interface", "interface", nil, set("disques-durs-ssd-externes", "cles-usb-cartes-memoire"), func(f *RawProduct, sd map[string]string) any { return iface(f, sd) }},
	{"Type d'équipement", "type-reseau", set("materiels-reseaux"), nil, func(f *RawProduct, sd map[string]string) any { return typeReseau(f, sd) }},
	{"PoE", "poe", set("materiels-reseaux"), nil, func(f *RawProduct, sd map[string]string) any { return poe(f, sd) }},
	{"Nombre de ports", "ports-nb", set("materiels-reseaux"), nil, func(f *RawProduct, sd map[string]string) any { return portsNb(f, sd) }},
	{"Connexion", "connexion", nil, set("claviers", "souris", "casques-ecouteurs"), func(f *RawProduct, sd map[string]string) any { return connexion(f, sd) }},
	{"Entrées vidéo", "entrees-video", set("ordinateurs-de-bureau-neufs"), nil, func(f *RawProduct, sd map[string]string) any {
		var out []string
		for _, x := range []string{"DisplayPort", "HDMI", "VGA"} {
			if strings.Contains(sd["Entrées vidéo"], x) {
				out = append(out, x)
			}
		}
		if len(out) == 0 {
			return nil
		}
		return out
	}},
}

// FacetValues reproduit facet_values() : {slug: valeur ou [valeurs]}.
func FacetValues(p *RawProduct) map[string]any {
	sd := specsDict(p)
	out := map[string]any{}
	for _, r := range FacetRules {
		if r.Tops != nil && !r.Tops[p.LiveTop] {
			continue
		}
		if r.Subs != nil && !r.Subs[p.LiveSub] {
			continue
		}
		v := r.Fn(p, sd)
		switch t := v.(type) {
		case string:
			if t != "" {
				out[r.Slug] = t
			}
		case []string:
			if len(t) > 0 {
				out[r.Slug] = t
			}
		}
	}
	return out
}

// LiveTopLabel : libellé du top pour SearchIndex.
func LiveTopLabel(slug string) string {
	for _, t := range LiveCats {
		if t.Slug == slug {
			return t.Label
		}
	}
	return ""
}

// SearchIndex reproduit search_index().
func SearchIndex(p *RawProduct) string {
	parts := []string{p.Nom, p.RefFabricant, p.Ref, p.Marque, p.Main, p.Sub,
		LiveTopLabel(p.LiveTop), SubLabels[[2]string{p.LiveTop, p.LiveSub}], p.Accroche, p.Description}
	var keep []string
	for _, s := range parts {
		if s != "" {
			keep = append(keep, s)
		}
	}
	return strings.Join(strings.Fields(normName(strings.Join(keep, " "))), " ")
}
