// Outil : régénère le panneau du méga-menu pour _top.html (pages statiques)
// à partir des vraies données (même DB que le serveur), afin que le menu des
// pages statiques reste identique au menu Go des pages dynamiques.
// Usage : go run ./tools/megapanel --db data/store.db > /tmp/megapanel.html
package main

import (
	"database/sql"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
	_ "modernc.org/sqlite"
)

func main() {
	dbp := flag.String("db", "data/store.db", "chemin store.db")
	flag.Parse()
	d, err := sql.Open("sqlite", *dbp+"?mode=ro")
	if err != nil {
		log.Fatal(err)
	}
	defer d.Close()
	st, err := site.LoadStore(d)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Fprint(os.Stdout, render.MegaPanel(st, "{{.Prefix}}"))
}
