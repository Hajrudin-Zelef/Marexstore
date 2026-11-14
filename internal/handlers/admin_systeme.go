package handlers

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/catalog"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/Hajrudin-Zelef/Marexstore/internal/site"
)

// /admin/systeme : état du backend, lecture seule (+ 2 actions).
// Version, comptes, taille DB/WAL, dernier backup, FTS, maintenance,
// kill-switch édition de fichiers. Boutons : Sauvegarder (téléchargement
// de la base, VACUUM INTO — le conteneur est en lecture seule, la sauvegarde
// planifiée reste le cron hôte tools/backup.sh) et Vérifier l'intégrité.

func fmtBytes(n int64) string {
	switch {
	case n < 1024:
		return fmt.Sprintf("%d o", n)
	case n < 1024*1024:
		return fmt.Sprintf("%.1f Ko", float64(n)/1024)
	default:
		return fmt.Sprintf("%.1f Mo", float64(n)/(1024*1024))
	}
}

func (a *App) AdminSysteme(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method == http.MethodPost {
		if !a.checkAdminCSRF(r) {
			http.Error(w, "refusé", http.StatusForbidden)
			return
		}
		switch r.FormValue("action") {
		case "integrite":
			var res string
			if err := a.DB.QueryRow(`PRAGMA integrity_check`).Scan(&res); err != nil || res != "ok" {
				_, _ = a.DB.Exec(`INSERT OR REPLACE INTO ops_jobs(nom,derniere_ok,detail) VALUES('integrity', datetime('now','localtime'), ?)`, "KO : "+res)
				a.auditAdmin(r, "systeme.integrite", "KO")
				http.Redirect(w, r, "/admin/systeme?err=V%C3%A9rification+int%C3%A9grit%C3%A9+%C3%A9chou%C3%A9e", http.StatusSeeOther)
				return
			}
			_, _ = a.DB.Exec(`INSERT OR REPLACE INTO ops_jobs(nom,derniere_ok,detail) VALUES('integrity', datetime('now','localtime'), 'OK')`)
			a.auditAdmin(r, "systeme.integrite", "OK")
			http.Redirect(w, r, "/admin/systeme?msg=Int%C3%A9grit%C3%A9+OK", http.StatusSeeOther)
			return
		case "sauvegarder":
			a.systemeBackup(w, r)
			return
		}
		http.Redirect(w, r, "/admin/systeme", http.StatusSeeOther)
		return
	}

	// --- Lecture seule : état du backend ---
	c := a.counts()
	var nItems int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM order_items`).Scan(&nItems)
	var nProd, nFTS int
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products`).Scan(&nProd)
	_ = a.DB.QueryRow(`SELECT COUNT(*) FROM products_fts`).Scan(&nFTS)
	fts := fmt.Sprintf("%d / %d", nFTS, nProd)
	if nFTS == nProd {
		fts += " — synchronisé"
	} else {
		fts += " — DÉSYNCHRONISÉ (lancer EnsureFTS)"
	}
	var dbSize, walSize int64
	if fi, err := os.Stat(a.Cfg.DBPath); err == nil {
		dbSize = fi.Size()
	}
	if fi, err := os.Stat(a.Cfg.DBPath + "-wal"); err == nil {
		walSize = fi.Size()
	}
	var lastBackup, lastBackupDetail string
	_ = a.DB.QueryRow(`SELECT derniere_ok, detail FROM ops_jobs WHERE nom='backup'`).Scan(&lastBackup, &lastBackupDetail)
	if lastBackup == "" {
		lastBackup = "jamais"
	}
	var lastIntegrity string
	_ = a.DB.QueryRow(`SELECT derniere_ok FROM ops_jobs WHERE nom='integrity'`).Scan(&lastIntegrity)
	if lastIntegrity == "" {
		lastIntegrity = "jamais"
	}
	var lastProd string
	_ = a.DB.QueryRow(`SELECT MAX(created_at) FROM products`).Scan(&lastProd)
	var maint, maintMsg string
	_ = a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom='maintenance'`).Scan(&maint)
	_ = a.DB.QueryRow(`SELECT valeur FROM settings WHERE nom='maintenance_message'`).Scan(&maintMsg)
	maintStr := "désactivée"
	if maint == "1" {
		maintStr = "ACTIVÉE"
		if maintMsg != "" {
			maintStr += " — " + maintMsg
		}
	}
	themeStr := "aucun thème publié (défaut)"
	if !catalog.SiteTheme.IsDefault() {
		themeStr = fmt.Sprintf("publié, version %d", catalog.SiteTheme.Version)
	}
	editStr := "autorisée"
	if !site.FileEditAllowed() {
		editStr = "INTERDITE (DISALLOW_FILE_EDIT=1)"
	}
	row := func(k, v string) string {
		return `<tr><td>` + render.Esc(k) + `</td><td><strong>` + render.Esc(v) + `</strong></td></tr>`
	}
	a.write(w, 200, a.adminPage("Système", "systeme",
		`<div class="mx-postbox"><h2>État du backend</h2><div class="in"><div class="mx-tablewrap"><table class="mx-table">`+
			row("Moteur Go", runtime.Version())+
			row("Thème publié", themeStr)+
			row("Produits", fmt.Sprint(c["produits"])+fmt.Sprintf(" (lignes de commande : %d)", nItems))+
			row("Clients", fmt.Sprint(c["clients"]))+
			row("Commandes", fmt.Sprint(c["commandes"]))+
			row("Base de données", a.Cfg.DBPath+" — "+fmtBytes(dbSize)+" (WAL : "+fmtBytes(walSize)+")")+
			row("Dernier backup", lastBackup+func() string {
				if lastBackupDetail != "" {
					return " — " + lastBackupDetail
				}
				return ""
			}())+
			row("Dernière vérification d'intégrité", lastIntegrity)+
			row("Dernier produit ajouté", lastProd)+
			row("Recherche (FTS)", fts)+
			row("Maintenance", maintStr)+
			row("Édition de fichiers (/admin/fichiers, T3)", editStr)+
			`</table></div></div></div>
<div class="mx-postbox"><h2>Actions</h2><div class="in">
<p class="mx-hint">La sauvegarde planifiée est le cron hôte (<code>0 2 * * * tools/backup.sh</code>). Ici, « Sauvegarder » télécharge une copie cohérente de la base (VACUUM INTO) ; « Vérifier l'intégrité » lance PRAGMA integrity_check.</p>
<form method="post" style="display:inline"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`"><input type="hidden" name="action" value="sauvegarder"><button class="mx-btn primary">Sauvegarder (télécharger)</button></form>
<form method="post" style="display:inline"><input type="hidden" name="csrf" value="`+a.adminCSRF()+`"><input type="hidden" name="action" value="integrite"><button class="mx-btn">Vérifier l'intégrité</button></form>
</div></div>`, r))
}

// systemeBackup : copie cohérente de la base (VACUUM INTO) servie en téléchargement.
func (a *App) systemeBackup(w http.ResponseWriter, r *http.Request) {
	tmp, err := os.CreateTemp("", "marexstore-backup-*.db")
	if err != nil {
		http.Error(w, "sauvegarde impossible", 500)
		return
	}
	tmpPath := tmp.Name()
	_ = tmp.Close()
	defer os.Remove(tmpPath)
	// VACUUM INTO : copie cohérente même si le serveur tourne. Le chemin
	// vient de os.CreateTemp (jamais d'entrée utilisateur) : guillemets doublés.
	q := strings.ReplaceAll(tmpPath, "'", "''")
	if _, err := a.DB.Exec(`VACUUM INTO '` + q + `'`); err != nil {
		http.Error(w, "sauvegarde impossible", 500)
		return
	}
	f, err := os.Open(tmpPath)
	if err != nil {
		http.Error(w, "sauvegarde impossible", 500)
		return
	}
	defer f.Close()
	fi, _ := f.Stat()
	_, _ = a.DB.Exec(`INSERT OR REPLACE INTO ops_jobs(nom,derniere_ok,detail) VALUES('backup', datetime('now','localtime'), 'manuel (téléchargement)')`)
	a.auditAdmin(r, "systeme.sauvegarder", "manuel (téléchargement)")
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Disposition", `attachment; filename="marexstore-`+time.Now().Format("20060102-150405")+`.db"`)
	w.Header().Set("Content-Length", strconv.FormatInt(fi.Size(), 10))
	_, _ = io.Copy(w, f)
}
