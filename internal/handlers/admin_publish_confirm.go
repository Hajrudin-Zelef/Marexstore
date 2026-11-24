package handlers

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// Jeton de confirmation de publication : preuve qu'un humain a saisi le mot
// de passe admin juste avant de publier. Sans état (HMAC + échéance), valable
// 3 minutes, lié au compte admin. Le mot de passe ne circule jamais dans la
// requête de publication elle-même.

const publishConfirmTTL = 3 * time.Minute

func (a *App) publishToken(exp int64) string {
	mac := hmac.New(sha256.New, []byte(a.Cfg.SessionSecret))
	fmt.Fprintf(mac, "publish|%s|%d", a.Cfg.AdminUser, exp)
	return strconv.FormatInt(exp, 10) + "." + base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}

func (a *App) publishTokenValid(tok string) bool {
	parts := strings.SplitN(strings.TrimSpace(tok), ".", 2)
	if len(parts) != 2 {
		return false
	}
	exp, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil || time.Now().Unix() > exp {
		return false
	}
	want := a.publishToken(exp)
	return hmac.Equal([]byte(want), []byte(tok))
}

// AdminPublishConfirm : POST {password} → {token, expires_in}. Vérifie le mot
// de passe admin (bcrypt) ; un échec ne révèle rien d'autre que le refus.
func (a *App) AdminPublishConfirm(w http.ResponseWriter, r *http.Request) {
	if !a.adminAuth(w, r) {
		return
	}
	if r.Method != http.MethodPost {
		http.Error(w, "méthode non autorisée", http.StatusMethodNotAllowed)
		return
	}
	if !a.checkAdminCSRF(r) {
		http.Error(w, "refusé", http.StatusForbidden)
		return
	}
	var body struct {
		Password string `json:"password"`
	}
	_ = json.NewDecoder(io.LimitReader(r.Body, 4096)).Decode(&body)
	if strings.TrimSpace(body.Password) == "" || !adminPassOK(a.Cfg.AdminPassHash, body.Password) {
		// Petit délai : rend le tâtonnement coûteux sans bloquer l'admin.
		time.Sleep(300 * time.Millisecond)
		http.Error(w, "mot de passe incorrect", http.StatusUnauthorized)
		return
	}
	exp := time.Now().Add(publishConfirmTTL).Unix()
	a.auditAdmin(r, "design.confirmer", "publication autorisée")
	a.writeJSON(w, http.StatusOK, map[string]any{
		"ok":         true,
		"token":      a.publishToken(exp),
		"expires_in": int(publishConfirmTTL.Seconds()),
	})
}

// requirePublishToken : 403 si l'en-tête X-Publish-Confirm est absent/invalide.
func (a *App) requirePublishToken(w http.ResponseWriter, r *http.Request) bool {
	if a.publishTokenValid(r.Header.Get("X-Publish-Confirm")) {
		return true
	}
	http.Error(w, "confirmation requise : saisissez le mot de passe admin pour publier", http.StatusForbidden)
	return false
}
