package handlers

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"net/http"
	"strconv"
	"strings"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
	"github.com/Hajrudin-Zelef/Marexstore/internal/render"
	"github.com/go-chi/chi/v5"
)

// Messages affichés sur la fiche client (clé whitelistée dans ?msg=).
var clientMsgs = map[string]string{
	"pw":     "Mot de passe modifié. Le client a été déconnecté de tous ses appareils.",
	"infos":  "Informations mises à jour.",
	"logout": "Toutes les sessions du client ont été supprimées.",
	"dup":    "Cet e-mail est déjà utilisé par un autre compte.",
	"badpw":  "Mot de passe trop court (8 caractères minimum).",
	"bad":    "Informations invalides (e-mail requis).",
}

// genPassword : mot de passe aléatoire sans caractères ambigus (0/O, 1/l/I).
func genPassword(n int) string {
	const alpha = "ABCDEFGHJKLMNPQRSTUVWXYZabcdefghijkmnpqrstuvwxyz23456789"
	var sb strings.Builder
	for i := 0; i < n; i++ {
		k, err := rand.Int(rand.Reader, big.NewInt(int64(len(alpha))))
		if err != nil {
			panic(err)
		}
		sb.WriteByte(alpha[k.Int64()])
	}
	return sb.String()
}

// setClientPassword : hash bcrypt + invalidation de toutes les sessions du client.
func (a *App) setClientPassword(id int64, pw string) bool {
	h, err := auth.HashPassword(pw)
	if err != nil {
		return false
	}
	if _, err := a.DB.Exec(`UPDATE users SET pass_hash=? WHERE id=?`, h, id); err != nil {
		return false
	}
	_, _ = a.DB.Exec(`DELETE FROM sessions WHERE user_id=?`, id)
	return true
}

// AdminClientAction : POST /admin/clients/{id} (password | generate | infos | logout).
func (a *App) AdminClientAction(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if !a.checkAdminCSRF(r) {
		http.Error(w, "CSRF invalide.", http.StatusForbidden)
		return
	}
	id, _ := strconv.ParseInt(chi.URLParam(r, "id"), 10, 64)
	var n int
	if err := a.DB.QueryRow(`SELECT COUNT(*) FROM users WHERE id=?`, id).Scan(&n); err != nil || n == 0 {
		http.NotFound(w, r)
		return
	}
	back := func(m string) {
		http.Redirect(w, r, fmt.Sprintf("/admin/clients/%d?msg=%s", id, m), http.StatusSeeOther)
	}
	switch r.FormValue("action") {
	case "password":
		pw := r.FormValue("password")
		if len(pw) < 8 {
			back("badpw")
			return
		}
		if !a.setClientPassword(id, pw) {
			http.Error(w, "Erreur serveur.", http.StatusInternalServerError)
			return
		}
		back("pw")
	case "generate":
		pw := genPassword(12)
		if !a.setClientPassword(id, pw) {
			http.Error(w, "Erreur serveur.", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Cache-Control", "no-store")
		a.write(w, 200, a.adminPage("Mot de passe généré", "clients",
			`<div class="mx-postbox"><h2>Nouveau mot de passe</h2><div class="in">`+
				`<p>Communiquez-le au client. Il ne sera plus affiché ensuite.</p>`+
				`<p style="font-size:22px;font-family:monospace;font-weight:700;user-select:all">`+render.Esc(pw)+`</p>`+
				`<p><a class="mx-btn" href="/admin/clients/`+strconv.FormatInt(id, 10)+`">Retour à la fiche</a></p>`+
				`</div></div>`, r))
	case "infos":
		em := strings.ToLower(strings.TrimSpace(r.FormValue("email")))
		nm := strings.TrimSpace(r.FormValue("nom"))
		tel := strings.TrimSpace(r.FormValue("tel"))
		if em == "" || !strings.Contains(em, "@") || len(em) > 254 || len(nm) > 120 || len(tel) > 40 {
			back("bad")
			return
		}
		if _, err := a.DB.Exec(`UPDATE users SET email=?, nom=?, tel=? WHERE id=?`, em, nm, tel, id); err != nil {
			if strings.Contains(err.Error(), "UNIQUE") {
				back("dup")
				return
			}
			http.Error(w, "Erreur serveur.", http.StatusInternalServerError)
			return
		}
		back("infos")
	case "logout":
		_, _ = a.DB.Exec(`DELETE FROM sessions WHERE user_id=?`, id)
		back("logout")
	default:
		http.Error(w, "Action inconnue.", http.StatusBadRequest)
	}
}

// clientActionsHTML : bloc d'administration du compte (fiche client).
func (a *App) clientActionsHTML(r *http.Request, id int64, em, nm, tel string) string {
	act := fmt.Sprintf("/admin/clients/%d", id)
	csrf := `<input type="hidden" name="csrf" value="` + render.Esc(a.adminCSRF()) + `">`
	in := `style="width:100%;box-sizing:border-box;margin:4px 0 10px;padding:8px;border:1px solid #c3c4c7;border-radius:6px"`
	msg := ""
	if m, ok := clientMsgs[r.URL.Query().Get("msg")]; ok {
		msg = `<div class="mx-badge ok" style="display:block;margin-bottom:12px;padding:8px 10px;white-space:normal">` + render.Esc(m) + `</div>`
	}
	return `<div class="mx-postbox" style="margin-top:16px"><h2>Administration du compte</h2><div class="in">` + msg +
		`<form method="post" action="` + act + `">` + csrf + `<input type="hidden" name="action" value="infos">` +
		`<label>Nom</label><input ` + in + ` name="nom" value="` + render.Esc(nm) + `">` +
		`<label>E-mail</label><input ` + in + ` type="email" name="email" value="` + render.Esc(em) + `" required>` +
		`<label>Téléphone</label><input ` + in + ` name="tel" value="` + render.Esc(tel) + `">` +
		`<button class="mx-btn">Enregistrer les infos</button></form>` +
		`<hr style="margin:16px 0;border:0;border-top:1px solid #e0e0e0">` +
		`<form method="post" action="` + act + `">` + csrf + `<input type="hidden" name="action" value="password">` +
		`<label>Nouveau mot de passe (8 caractères min.)</label><input ` + in + ` type="text" name="password" minlength="8" autocomplete="off" required>` +
		`<button class="mx-btn">Changer le mot de passe</button></form>` +
		`<form method="post" action="` + act + `" style="margin-top:10px">` + csrf + `<input type="hidden" name="action" value="generate">` +
		`<button class="mx-btn">Générer un mot de passe aléatoire</button></form>` +
		`<form method="post" action="` + act + `" style="margin-top:10px">` + csrf + `<input type="hidden" name="action" value="logout">` +
		`<button class="mx-btn">Déconnecter le client partout</button></form>` +
		`</div></div>`
}
