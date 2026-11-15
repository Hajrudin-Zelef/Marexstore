// Fichiers réels du site pour le mode Code du Studio (phases C2 lecture,
// C4 écriture / T3).
//
// GET /admin/api/fichiers (auth admin) expose le contenu RÉEL des fichiers
// que le Studio permet de consulter en mode Code — liste fermée (addendum) :
// templates/static/*.html (chrome + pages) + static/app.js, static/style.css,
// static/theme.js. Le faux `theme.css` du prototype disparaît : il n'existe
// pas, le thème vit dans static/theme.js + static/style.css.
//
// POST /admin/api/fichiers (auth admin + CSRF, C4) enregistre une nouvelle
// version d'un fichier : {nom, contenu} en JSON. Le contenu est stocké en
// base (design_files, versionné, append-only) — PAS sur disque : le
// conteneur est read_only et ./static est monté ro. Un gabarit enregistré
// prend effet au prochain reload() (LoadStaticDB lit la base d'abord) ;
// un fichier static/* est servi depuis la base par StaticFile (fallback
// disque). Rollback = activer une ancienne version (voir les endpoints
// /admin/api/fichiers/versions).
//
// Source du contenu (la même que le site) :
//   - templates/static/* : a.Top / a.Drawer / a.Foot / a.Bodies, chargés au
//     démarrage par reload() (design_files actif d'abord, puis disque).
//   - static/* : base d'abord (version active, si elle existe), puis disque
//     (ces fichiers ne passent pas par LoadStaticDB).
//
// Sécurité : kill-switch DISALLOW_FILE_EDIT=1 (lecture seule), liste fermée
// des noms (design.FichierNom, pas de traversée), limite 2 Mo, audit de
// chaque enregistrement / activation. Aucune validation de contenu : le code
// du site contient légitimement <script> et {{...}}.
// La réponse GET fait ~220 Ko (app.js + style.css dominent) ; le Studio ne
// la charge que quand l'onglet Code est ouvert, pas au chargement de la page.

package handlers

import (
	"encoding/json"
	"io"
	"mime"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/Hajrudin-Zelef/Marexstore/internal/design"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
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

// contenuStatique : contenu actuel d'un fichier static/* — version base
// (si elle existe) puis disque. C'est ce que le site sert (StaticFile) et
// ce que le Studio affiche (GET /admin/api/fichiers).
func (a *App) contenuStatique(nom string) (string, bool) {
	if a.DB != nil {
		if c, _, found, err := design.LoadFichier(a.DB, nom); err == nil && found {
			return c, true
		}
	}
	if b, err := os.ReadFile(repoFile(nom)); err == nil {
		return string(b), true
	}
	return "", false
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
	// static/* : base d'abord (version active), puis disque.
	for _, nom := range []string{"static/app.js", "static/style.css", "static/theme.js"} {
		if c, ok := a.contenuStatique(nom); ok {
			out = append(out, fichierCode{Nom: nom, Lang: langOf(nom), Lignes: strings.Count(c, "\n") + 1, Contenu: c})
		}
	}
	return out
}

// AdminFichiers : GET (liste) + POST (enregistre) /admin/api/fichiers.
func (a *App) AdminFichiers(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	switch r.Method {
	case http.MethodGet:
		a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "fichiers": a.buildFichiers()})
	case http.MethodPost:
		a.adminFichiersEnregistrer(w, r)
	default:
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
	}
}

// adminFichiersEnregistrer : POST /admin/api/fichiers — {nom, contenu}.
// Kill-switch, CSRF, liste fermée, limite 2 Mo, audit, reload() pour les
// gabarits (les fichiers static/* sont servis depuis la base à chaque
// requête, pas besoin de reload).
func (a *App) adminFichiersEnregistrer(w http.ResponseWriter, r *http.Request) {
	if !site.FileEditAllowed() {
		http.Error(w, "édition des fichiers interdite (DISALLOW_FILE_EDIT=1)", http.StatusForbidden)
		return
	}
	if !a.checkAdminCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	body, err := io.ReadAll(io.LimitReader(r.Body, design.MaxFichierBytes+1))
	if err != nil {
		http.Error(w, "lecture impossible", http.StatusBadRequest)
		return
	}
	var req struct {
		Nom     string `json:"nom"`
		Contenu string `json:"contenu"`
	}
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "JSON invalide", http.StatusBadRequest)
		return
	}
	if req.Nom == "" {
		http.Error(w, "nom manquant", http.StatusBadRequest)
		return
	}
	version, err := design.SaveFichier(a.DB, req.Nom, req.Contenu, a.Cfg.AdminUser)
	if err != nil {
		http.Error(w, "enregistrement impossible : "+err.Error(), http.StatusBadRequest)
		return
	}
	a.auditAdmin(r, "fichier.enregistrer", req.Nom)
	// Un gabarit enregistré prend effet immédiatement (reload lit la base).
	if strings.HasPrefix(req.Nom, "templates/static/") {
		a.reload()
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nom": req.Nom, "version": version, "msg": "Fichier enregistré"})
}

// AdminFichiersVersions : GET /admin/api/fichiers/versions?nom=<nom> —
// l'historique des versions d'un fichier (rollback 1 clic).
func (a *App) AdminFichiersVersions(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodGet {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	nom := r.URL.Query().Get("nom")
	if nom == "" {
		http.Error(w, "nom manquant", http.StatusBadRequest)
		return
	}
	versions, err := design.ListFichierVersions(a.DB, nom)
	if err != nil {
		http.Error(w, "lecture impossible : "+err.Error(), http.StatusBadRequest)
		return
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nom": nom, "versions": versions})
}

// AdminFichiersActiver : POST /admin/api/fichiers/versions/{n}/activer?nom=<nom>
// — active une ancienne version (rollback). Un gabarit réactivé prend effet
// immédiatement (reload) ; un fichier static/* est servi depuis la base à
// chaque requête.
func (a *App) AdminFichiersActiver(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	if !site.FileEditAllowed() {
		http.Error(w, "édition des fichiers interdite (DISALLOW_FILE_EDIT=1)", http.StatusForbidden)
		return
	}
	if !a.checkAdminCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	nom := r.URL.Query().Get("nom")
	n, err := strconv.Atoi(chi.URLParam(r, "n"))
	if err != nil || n < 1 || nom == "" {
		http.Error(w, "version ou nom invalide", http.StatusBadRequest)
		return
	}
	if err := design.ActivateFichier(a.DB, nom, n); err != nil {
		http.Error(w, "activation impossible : "+err.Error(), http.StatusBadRequest)
		return
	}
	a.auditAdmin(r, "fichier.activer", nom)
	if strings.HasPrefix(nom, "templates/static/") {
		a.reload()
	}
	a.writeJSON(w, http.StatusOK, map[string]any{"ok": true, "nom": nom, "version": n, "msg": "Version réactivée"})
}

// StaticFile : sert /static/* et /assets/* — version base d'abord (si elle
// existe), sinon le disque. La version base est servie avec
// Cache-Control: no-cache : le contenu a changé sans que ?v= ne bouge, les
// navigateurs doivent revalider. Le disque garde le cache long habituel.
func (a *App) StaticFile() http.Handler {
	disk := http.FileServer(http.Dir("static"))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rel := strings.TrimPrefix(r.URL.Path, "/")
		if rel != "" && a.DB != nil {
			nom := "static/" + rel
			if c, _, found, err := design.LoadFichier(a.DB, nom); err == nil && found {
				if ct := mime.TypeByExtension(filepath.Ext(rel)); ct != "" {
					w.Header().Set("Content-Type", ct)
				}
				w.Header().Set("Cache-Control", "no-cache")
				_, _ = io.WriteString(w, c)
				return
			}
		}
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		disk.ServeHTTP(w, r)
	})
}
