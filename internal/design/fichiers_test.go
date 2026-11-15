package design

import (
	"strings"
	"testing"
)

// C4 — noms de fichiers : liste fermée, pas de traversée.
func TestFichierNom(t *testing.T) {
	ok := map[string]string{
		"templates/static/_top.html":    "_top.html",
		"templates/static/_drawer.html": "_drawer.html",
		"templates/static/_foot.html":   "_foot.html",
		"templates/static/apropos.html": "apropos.html",
		"static/app.js":                 "static/app.js",
		"static/style.css":              "static/style.css",
		"static/theme.js":               "static/theme.js",
	}
	for in, want := range ok {
		got, good := FichierNom(in)
		if !good || got != want {
			t.Errorf("FichierNom(%q) = (%q,%v), attendu (%q,true)", in, got, good, want)
		}
	}
	ko := []string{
		"", "templates/static/_top.htm", "templates/static/../main.go",
		"templates/static/sub/page.html", "templates/static/_top.html/",
		"/etc/passwd", "static/theme.css", "static/app2.js", "images/logo.png",
		"templates/static/apropos.html ", "templates/static\\_top.html",
		"design.json", "templates/static/.html",
	}
	for _, in := range ko {
		if _, good := FichierNom(in); good {
			t.Errorf("FichierNom(%q) accepté à tort", in)
		}
	}
}

// C4 — sauvegarde + lecture + versions + activation (rollback).
func TestSaveFichier(t *testing.T) {
	db := openTestDB(t)
	nom := "templates/static/_top.html"
	v1, err := SaveFichier(db, nom, "<div>v1</div>", "admin")
	if err != nil {
		t.Fatalf("SaveFichier v1 : %v", err)
	}
	if v1 != 1 {
		t.Fatalf("v1 = %d, attendu 1", v1)
	}
	v2, err := SaveFichier(db, nom, "<div>v2</div>", "admin")
	if err != nil {
		t.Fatalf("SaveFichier v2 : %v", err)
	}
	if v2 != 2 {
		t.Fatalf("v2 = %d, attendu 2", v2)
	}
	c, v, found, err := LoadFichier(db, nom)
	if err != nil || !found || c != "<div>v2</div>" || v != 2 {
		t.Fatalf("LoadFichier = (%q,%d,%v,%v), attendu (v2,2,true,nil)", c, v, found, err)
	}
	// Un fichier jamais enregistré : absent, pas d'erreur.
	if _, _, found, err := LoadFichier(db, "static/app.js"); err != nil || found {
		t.Fatalf("LoadFichier app.js jamais enregistré = (found=%v, err=%v)", found, err)
	}
	// Versions : v2 active, v1 inactive.
	vs, err := ListFichierVersions(db, nom)
	if err != nil || len(vs) != 2 || !vs[0].Actif || vs[0].Version != 2 || vs[1].Actif || vs[1].Version != 1 {
		t.Fatalf("ListFichierVersions = %+v, err=%v", vs, err)
	}
	// Rollback : activer v1.
	if err := ActivateFichier(db, nom, 1); err != nil {
		t.Fatalf("ActivateFichier v1 : %v", err)
	}
	c, v, _, _ = LoadFichier(db, nom)
	if c != "<div>v1</div>" || v != 1 {
		t.Fatalf("après rollback : (%q,%d), attendu (v1,1)", c, v)
	}
	// Version inexistante : erreur.
	if err := ActivateFichier(db, nom, 99); err == nil {
		t.Fatal("ActivateFichier v99 accepté à tort")
	}
}

// C4 — un fichier statique (static/app.js) ne passe pas par LoadStaticDB :
// son nom en base est le chemin complet, sans collision avec les gabarits.
func TestSaveFichierStatique(t *testing.T) {
	db := openTestDB(t)
	v, err := SaveFichier(db, "static/app.js", "console.log(1);", "admin")
	if err != nil || v != 1 {
		t.Fatalf("SaveFichier app.js = (%d,%v)", v, err)
	}
	c, _, found, err := LoadFichier(db, "static/app.js")
	if err != nil || !found || c != "console.log(1);" {
		t.Fatalf("LoadFichier app.js = (%q,%v,%v)", c, found, err)
	}
}

// C4 — refus : nom hors liste, contenu trop gros. Le contenu n'est PAS
// validé (le code contient légitimement <script> et {{...}}).
func TestSaveFichierRefus(t *testing.T) {
	db := openTestDB(t)
	if _, err := SaveFichier(db, "../main.go", "x", "admin"); err == nil {
		t.Fatal("SaveFichier ../main.go accepté à tort")
	}
	if _, err := SaveFichier(db, "static/theme.css", "x", "admin"); err == nil {
		t.Fatal("SaveFichier theme.css accepté à tort")
	}
	gros := strings.Repeat("x", MaxFichierBytes+1)
	if _, err := SaveFichier(db, "static/app.js", gros, "admin"); err == nil {
		t.Fatal("SaveFichier trop gros accepté à tort")
	}
	// <script> et {{...}} sont légitimes dans un fichier de code.
	if _, err := SaveFichier(db, "static/app.js", "<script>{{.Prefix}}</script>", "admin"); err != nil {
		t.Fatalf("SaveFichier avec <script>/{{...}} refusé à tort : %v", err)
	}
}
