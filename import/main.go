package main

// Import fidèle : produits-v3.json -> SQLite via internal/catalog
// (mêmes dérivations que build_site_v3.py : CAT_MAP, live_map, facettes,
// index de recherche). Idempotent (INSERT OR REPLACE par slug).
import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	marexdb "github.com/Hajrudin-Zelef/Marexstore/internal/db"
	_ "modernc.org/sqlite"
)

func mustJSON(v any) string {
	b, _ := json.Marshal(v)
	return string(b)
}

func main() {
	jf := flag.String("json", "catalogdata/produits-v3.json", "chemin produits-v3.json")
	dbp := flag.String("db", "store.db", "chemin sqlite")
	mig1 := flag.String("sql", "db/migrations/001_init.sql", "migration 001")
	mig2 := flag.String("sql2", "db/migrations/002_fidelite.sql", "migration 002")
	flag.Parse()

	list, err := catalog.LoadRaw(*jf)
	if err != nil {
		log.Fatal(err)
	}
	catalog.ApplyTaxonomy(list)

	d, err := sql.Open("sqlite", *dbp+"?cache=shared")
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	for _, mig := range []string{*mig1, *mig2} {
		b, err := os.ReadFile(mig)
		if err != nil {
			log.Fatal(err)
		}
		if _, err := d.Exec(string(b)); err != nil {
			log.Printf("migration %s partielle (%v) — colonnes manquantes ajoutees si besoin", mig, err)
			break
		}
	}
	if err := marexdb.EnsureSchema(d, "db/migrations"); err != nil {
		log.Fatal(err)
	}
	tx, _ := d.Begin()
	n := 0
	for _, p := range list {
		if p.Slug == "" || p.Nom == "" {
			continue
		}
		facets := catalog.FacetValues(p)
		si := catalog.SearchIndex(p)
		var prix, pb, pa any
		if p.Prix != nil {
			prix = *p.Prix
		}
		if p.PrixBoutique != nil {
			pb = *p.PrixBoutique
		}
		if p.PrixAvant != nil {
			pa = *p.PrixAvant
		}
		var stock any
		if p.Stock != nil {
			stock = *p.Stock
		}
		devise := p.Devise
		if devise == "" {
			devise = "F CFA"
		}
		dispo := p.Dispo
		if dispo == "" {
			dispo = "en stock"
		}
		specs := mustJSON(p.SpecList())
		if specs == "null" {
			specs = "[]"
		}
		_, err := tx.Exec(`INSERT INTO products(slug,nom,marque,prix,prix_boutique,prix_avant,devise,dispo,
			categorie,accroche,description,image,specs_json,pointsforts_json,consommables_json,store_url,ref,ref_fabricant,stock,
			main,main_slug,sub,sub_slug,live_top,live_sub,facets_json,search_index)
			VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)
			ON CONFLICT(slug) DO UPDATE SET nom=excluded.nom, marque=excluded.marque,
			prix=excluded.prix, prix_boutique=excluded.prix_boutique, prix_avant=excluded.prix_avant,
			devise=excluded.devise, dispo=excluded.dispo, categorie=excluded.categorie,
			accroche=excluded.accroche, description=excluded.description, image=excluded.image,
			specs_json=excluded.specs_json, pointsforts_json=excluded.pointsforts_json, consommables_json=excluded.consommables_json, store_url=excluded.store_url,
			ref=excluded.ref, ref_fabricant=excluded.ref_fabricant, stock=excluded.stock,
			main=excluded.main, main_slug=excluded.main_slug, sub=excluded.sub, sub_slug=excluded.sub_slug,
			live_top=excluded.live_top, live_sub=excluded.live_sub,
			facets_json=excluded.facets_json, search_index=excluded.search_index`,
			p.Slug, p.Nom, p.Marque, prix, pb, pa, devise, dispo, p.Categorie,
			p.Accroche, p.Description, p.Image, specs, mustJSON(p.PointsForts), mustJSON(p.Consommables), p.StoreURL,
			p.Ref, p.RefFabricant, stock, p.Main, p.MainSlug, p.Sub, p.SubSlug,
			p.LiveTop, p.LiveSub, mustJSON(facets), si)
		if err != nil {
			log.Printf("skip %s: %v", p.Slug, err)
			continue
		}
		n++
	}
	if err := tx.Commit(); err != nil {
		log.Fatal(err)
	}
	// L'index de recherche FTS5 est reconstruit à chaque import (il était
	// créé par le schéma mais jamais rempli).
	if _, err := d.Exec(`DELETE FROM products_fts`); err == nil {
		if _, err := d.Exec(`INSERT INTO products_fts(slug, nom, marque, categorie, sous_categorie, accroche, description, search_index)
			SELECT slug, nom, marque, categorie, sous_categorie, accroche, description, search_index FROM products`); err != nil {
			log.Printf("fts: %v", err)
		}
	}
	fmt.Printf("import fidèle OK : %d produits -> %s\n", n, *dbp)
}
