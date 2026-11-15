// Fichiers réels du site pour le mode Code du Studio (phase C2, lecture
// seule).
//
// GET /admin/api/fichiers (auth admin) expose le contenu RÉEL des fichiers
// que le Studio permet de consulter en mode Code — liste fermée (addendum) :
// templates/static/*.html (chrome + pages) + static/app.js, static/style.css,
// static/theme.js. Le faux `theme.css` du prototype disparaît : il n'existe
// pas, le thème vit dans static/theme.js + static/style.css.
//
// Source du contenu (la même que le site) :
//   - templates/static/* : a.Top / a.Drawer / a.Foot / a.Bodies, chargés au
//     démarrage par reload() (design_files actif d'abord, puis disque).
//   - static/* : lecture disque (ces fichiers ne passent pas par
//     design_files). Le chemin est résolu depuis la racine du dépôt
//     (le serveur démarre depuis là ; repli "../../" pour les tests).
//
// Lecture seule : aucune écriture (c'est le T3 / C4, sur feu vert explicite).
// La réponse fait ~220 Ko (app.js + style.css dominent) ; le Studio ne la
// charge que quand l'onglet Code est ouvert, pas au chargement de la page.

package handlers

import (
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type fichierCode struct {
	Nom     string `json:"nom"`     // chemin affiché, ex. templates/static/_top.html
	Lang    string `json:"lang"`    // html | css | js (pour la coloration)
	Lignes  int    `json:"lignes"`  // nombre de lignes
	Contenu string `json:"contenu"` // contenu réel du fichier
}

// repoFile résout un chemin relatif à la racine du dépôt : le serveur
// démarre depuis la racine, les tests depuis internal/handlers/.
func repoFile(rel string) string {
	if _, err := os.Stat(rel); err == nil {
		return rel
	}
	return filepath.Join("..", "..", rel)
}

func langOf(nom string) string {
	switch {
	case strings.HasSuffix(nom, ".html"):
		return "html"
	case strings.HasSuffix(nom, ".css"):
		return "css"
	case strings.HasSuffix(nom, ".js"):
		return "js"
	default:
		return "txt"
	}
}

// buildFichiers assemble la liste fermée des fichiers du mode Code.
func (a *App) buildFichiers() []fichierCode {
	out := []fichierCode{}
	// templates/static/* : chrome + pages, depuis l'App (déjà chargée).
	if a.Top != "" {
		out = append(out, fichierCode{Nom: "templates/static/_top.html", Lang: "html", Lignes: strings.Count(a.Top, "\n") + 1, Contenu: a.Top})
	}
	if a.Drawer != "" {
		out = append(out, fichierCode{Nom: "templates/static/_drawer.html", Lang: "html", Lignes: strings.Count(a.Drawer, "\n") + 1, Contenu: a.Drawer})
	}
	if a.Foot != "" {
		out = append(out, fichierCode{Nom: "templates/static/_foot.html", Lang: "html", Lignes: strings.Count(a.Foot, "\n") + 1, Contenu: a.Foot})
	}
	noms := make([]string, 0, len(a.Bodies))
	for k := range a.Bodies {
		noms = append(noms, k)
	}
	sort.Strings(noms)
	for _, k := range noms {
		c := a.Bodies[k]
		if c == "" {
			continue
		}
		out = append(out, fichierCode{Nom: "templates/static/" + k + ".html", Lang: "html", Lignes: strings.Count(c, "\n") + 1, Contenu: c})
	}
	// static/* : lecture disque.
	for _, nom := range []string{"static/app.js", "static/style.css", "static/theme.js"} {
		if b, err := os.ReadFile(repoFile(nom)); err == nil {
			c := string(b)
			out = append(out, fichierCode{Nom: nom, Lang: langOf(nom), Lignes: strings.Count(c, "\n") + 1, Contenu: c})
		}
	}
	return out
}

// AdminFichiers : GET /admin/api/fichiers — les vrais fichiers du site.
func (a *App) AdminFichiers(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "fichiers": a.buildFichiers()})
}
