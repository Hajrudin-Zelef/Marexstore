// Package site : assemblage des pages fidèles (chrome partagé + corps
// statiques verbatim + îles dynamiques) et chargement du catalogue depuis
// SQLite (rempli par l'import depuis produits-v3.json).
package site

import (
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
)

type Site struct {
	Store  *catalog.Store
	Bodies map[string]string // corps statiques (avec {{.Prefix}} / {{.CSRF}})
	Top    string
	Drawer string
	Foot   string
}

func fill(s, prefix, csrf, widgets string) string {
	s = strings.ReplaceAll(s, "{{.Prefix}}", prefix)
	s = strings.ReplaceAll(s, "{{.CSRF}}", csrf)
	s = strings.ReplaceAll(s, "{{.Widgets}}", widgets)
	return s
}

func LoadStatic(dir string) (top, drawer, foot string, bodies map[string]string, err error) {
	t, d, f, b, _, e := LoadStaticDB(nil, dir)
	return t, d, f, b, e
}

// FileEditAllowed : l'éditeur de fichiers (/admin/fichiers, T3) est-il actif ?
// Kill-switch : DISALLOW_FILE_EDIT=1 l'interdit (lecture seule).
func FileEditAllowed() bool {
	return os.Getenv("DISALLOW_FILE_EDIT") != "1"
}

// LoadStaticDB : charge les gabarits statiques avec la chaîne de repli
// design_files(actif=1) → disque → défaut Go (chaîne vide).
// Retourne aussi les fichiers tombés sur le défaut Go (à journaliser).
// Une base nil = disque seul (compatibilité LoadStatic).
// Les pages présentes uniquement en base (T3) sont aussi énumérées.
func LoadStaticDB(dbase *sql.DB, dir string) (top, drawer, foot string, bodies map[string]string, fallbacks []string, err error) {
	read := func(n string) string {
		// 1. design_files actif (T3) — seulement si la base est disponible.
		if dbase != nil {
			var c string
			if e := dbase.QueryRow(`SELECT contenu FROM design_files WHERE nom=? AND actif=1 ORDER BY version DESC LIMIT 1`, n).Scan(&c); e == nil {
				return c
			}
		}
		// 2. disque.
		if b, e := os.ReadFile(filepath.Join(dir, n)); e == nil {
			return string(b)
		}
		// 3. défaut Go (chaîne vide) — le site démarre quand même, dégradé.
		fallbacks = append(fallbacks, n)
		return ""
	}
	top, drawer, foot = read("_top.html"), read("_drawer.html"), read("_foot.html")
	bodies = map[string]string{}
	entries, e := os.ReadDir(dir)
	if e != nil {
		// Disque illisible : le chrome est déjà chargé (ou en défaut Go).
		return top, drawer, foot, bodies, fallbacks, nil
	}
	seen := map[string]bool{}
	for _, en := range entries {
		n := en.Name()
		if strings.HasPrefix(n, "_") || !strings.HasSuffix(n, ".html") {
			continue
		}
		seen[n] = true
		bodies[strings.TrimSuffix(n, ".html")] = read(n)
	}
	// Pages présentes uniquement en base (créées via /admin/fichiers, T3).
	if dbase != nil {
		rows, e := dbase.Query(`SELECT DISTINCT nom FROM design_files WHERE nom NOT LIKE '\_%' AND nom LIKE '%.html' AND actif=1`)
		if e == nil {
			defer rows.Close()
			for rows.Next() {
				var n string
				if err := rows.Scan(&n); err == nil && !seen[n] {
					bodies[strings.TrimSuffix(n, ".html")] = read(n)
				}
			}
		}
	}
	return top, drawer, foot, bodies, fallbacks, nil
}

// Produit DB -> RawProduct (toutes colonnes).
func rowToProduct(cols map[string]any) *catalog.RawProduct {
	str := func(k string) string {
		if v, ok := cols[k].(string); ok {
			return v
		}
		return ""
	}
	p := &catalog.RawProduct{
		Slug:         str("slug"),
		Nom:          str("nom"),
		Marque:       str("marque"),
		Devise:       str("devise"),
		Dispo:        str("dispo"),
		Categorie:    str("categorie"),
		Accroche:     str("accroche"),
		Description:  str("description"),
		Image:        str("image"),
		Ref:          str("ref"),
		RefFabricant: str("ref_fabricant"),
		Main:         str("main"),
		MainSlug:     str("main_slug"),
		Sub:          str("sub"),
		SubSlug:      str("sub_slug"),
		LiveTop:      str("live_top"),
		LiveSub:      str("live_sub"),
	}
	if v, ok := cols["prix"].(int64); ok {
		p.Prix = &v
	}
	if v, ok := cols["prix_boutique"].(int64); ok {
		p.PrixBoutique = &v
	}
	if v, ok := cols["prix_avant"].(int64); ok {
		p.PrixAvant = &v
	}
	if v, ok := cols["stock"].(int64); ok {
		n := int(v)
		p.Stock = &n
	}
	var specs []catalog.Spec
	_ = json.Unmarshal([]byte(str("specs_json")), &specs)
	p.Specs, _ = json.Marshal(specs)
	var pf []string
	_ = json.Unmarshal([]byte(str("pointsforts_json")), &pf)
	p.PointsForts = pf
	var cons []catalog.Consommable
	_ = json.Unmarshal([]byte(str("consommables_json")), &cons)
	p.Consommables = cons
	p.StoreURL = str("store_url")
	return p
}

func LoadStore(db *sql.DB) (*catalog.Store, error) {
	rows, err := db.Query(`SELECT slug,nom,marque,prix,prix_boutique,prix_avant,devise,dispo,
		categorie,accroche,description,image,specs_json,pointsforts_json,consommables_json,store_url,ref,ref_fabricant,stock,
		main,main_slug,sub,sub_slug,live_top,live_sub FROM products ORDER BY rowid`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	names, _ := rows.Columns()
	var list []*catalog.RawProduct
	for rows.Next() {
		vals := make([]any, len(names))
		ptrs := make([]any, len(names))
		for i := range vals {
			ptrs[i] = &vals[i]
		}
		if err := rows.Scan(ptrs...); err != nil {
			return nil, err
		}
		cols := map[string]any{}
		for i, n := range names {
			switch v := vals[i].(type) {
			case []byte:
				cols[n] = string(v)
			case int64:
				cols[n] = v
			case nil:
			default:
				cols[n] = v
			}
		}
		list = append(list, rowToProduct(cols))
	}
	return catalog.NewStore(list), rows.Err()
}

// ThemeFromDB : lit les réglages theme_* (table settings) vers catalog.Theme.
// Absents = thème par défaut (sortie identique au site d'origine).
func ThemeFromDB(db *sql.DB) catalog.Theme {
	t := catalog.DefaultTheme()
	get := func(k string) string {
		var v string
		if err := db.QueryRow(`SELECT valeur FROM settings WHERE nom=?`, k).Scan(&v); err != nil {
			return ""
		}
		return v
	}
	if v := get("theme_couleur"); v != "" {
		t.Couleur = v
	}
	if v := get("theme_accent"); v != "" {
		t.Accent = v
	}
	if v := get("theme_mode"); v != "" {
		t.Mode = v
	}
	if v := get("theme_police"); v != "" {
		t.Police = v
	}
	if get("theme_annonce_on") == "1" {
		t.AnnonceOn = true
	}
	if v := get("theme_annonce_texte"); v != "" {
		t.AnnonceTexte = v
	}
	if v := get("theme_sections"); v != "" {
		var ss []string
		for _, s := range strings.Split(v, ",") {
			if s = strings.TrimSpace(s); s != "" {
				ss = append(ss, s)
			}
		}
		if len(ss) > 0 {
			t.Sections = ss
		}
	}
	if v := get("theme_css"); v != "" {
		t.CSS = v
	}
	// P1 — version du thème publié (le serveur gagne). Absente = jamais publié.
	if v := get("theme_version"); v != "" {
		if n, err := strconv.Atoi(strings.TrimSpace(v)); err == nil && n > 0 {
			t.Version = n
		}
	}
	return t
}
