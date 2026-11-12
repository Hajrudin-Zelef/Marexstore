package main

// Import V1 : lit sources/produits-v3.json (ou /tmp/opencode/ref/...) -> SQLite.
// Usage : go run ./import --json <fichier> --db store.db --sql db/migrations/001_init.sql
// Idempotent : INSERT OR REPLACE par slug.

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
	"unicode"

	_ "modernc.org/sqlite"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type Spec struct {
	Label  string `json:"label"`
	Valeur string `json:"valeur"`
}

type P struct {
	Slug        string `json:"slug"`
	Nom         string `json:"nom"`
	Marque      string `json:"marque"`
	Prix        *int64 `json:"prix"`
	Devise      string `json:"devise"`
	Dispo       string `json:"disponibilite"`
	Categorie   string `json:"categorie"`
	Accroche    string `json:"accroche"`
	Description string `json:"description"`
	Image       string `json:"image"`
	Specs       any    `json:"specs"`
	Stock       int    `json:"stock"`
}

func normStr(s string) string {
	t := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	o, _, _ := transform.String(t, strings.ToLower(s))
	var b strings.Builder
	for _, r := range o {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else {
			b.WriteRune(' ')
		}
	}
	return strings.Join(strings.Fields(b.String()), " ")
}

func main() {
	jf := flag.String("json", "/tmp/opencode/ref/sources/produits-v3.json", "chemin produits-v3.json")
	dbp := flag.String("db", "store.db", "chemin sqlite")
	sqlp := flag.String("sql", "db/migrations/001_init.sql", "migration")
	flag.Parse()

	raw, err := os.ReadFile(*jf)
	if err != nil {
		log.Fatal(err)
	}
	var list []P
	if err := json.Unmarshal(raw, &list); err != nil {
		log.Fatal(err)
	}
	mig, err := os.ReadFile(*sqlp)
	if err != nil {
		log.Fatal(err)
	}
	d, err := sql.Open("sqlite", *dbp+"?cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	if _, err := d.Exec(string(mig)); err != nil {
		log.Fatal(err)
	}
	tx, _ := d.Begin()
	n := 0
	for _, p := range list {
		if p.Slug == "" || p.Nom == "" {
			continue
		}
		var specsJSON string
		switch v := p.Specs.(type) {
		case []any:
			b, _ := json.Marshal(v)
			specsJSON = string(b)
		case map[string]any:
			var arr []Spec
			for k, val := range v {
				arr = append(arr, Spec{Label: k, Valeur: fmt.Sprint(val)})
			}
			b, _ := json.Marshal(arr)
			specsJSON = string(b)
		default:
			specsJSON = "[]"
		}
		devise := p.Devise
		if devise == "" {
			devise = "F CFA"
		}
		dispo := p.Dispo
		if dispo == "" {
			dispo = "en stock"
		}
		si := normStr(p.Nom + " " + p.Marque + " " + p.Categorie + " " + p.Accroche + " " + p.Description)
		var prix any
		if p.Prix != nil {
			prix = *p.Prix
		}
		_, err := tx.Exec(`INSERT INTO products(slug,nom,marque,prix,devise,dispo,categorie,image,accroche,description,specs_json,search_index,stock)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(slug) DO UPDATE SET nom=excluded.nom, marque=excluded.marque, prix=excluded.prix,
			categorie=excluded.categorie, image=excluded.image, accroche=excluded.accroche,
			description=excluded.description, specs_json=excluded.specs_json, search_index=excluded.search_index, stock=excluded.stock`,
			p.Slug, p.Nom, p.Marque, prix, devise, dispo, p.Categorie, p.Image, p.Accroche, p.Description, specsJSON, si, p.Stock)
		if err != nil {
			log.Printf("skip %s: %v", p.Slug, err)
			continue
		}
		_, _ = tx.Exec(`INSERT INTO products_fts(slug,nom,marque,categorie,sous_categorie,accroche,description,search_index)
			VALUES(?,?,?,?,?,?,?,?)`, p.Slug, p.Nom, p.Marque, p.Categorie, "", p.Accroche, p.Description, si)
		n++
	}
	_ = tx.Commit()
	fmt.Printf("import OK : %d produits -> %s\n", n, *dbp)
}
