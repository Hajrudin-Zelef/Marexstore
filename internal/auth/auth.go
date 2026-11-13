package auth

import (
	"crypto/rand"
	"crypto/sha256"
	"crypto/subtle"
	"database/sql"
	"encoding/base64"
	"errors"
	"net/http"
	"time"

	"golang.org/x/crypto/bcrypt"
)

const cookieName = "marex_session"

// HashPassword bcrypt(12) : robuste, coût maîtrisé.
func HashPassword(pw string) (string, error) {
	h, err := bcrypt.GenerateFromPassword([]byte(pw), 12)
	return string(h), err
}

func CheckPassword(hash, pw string) bool {
	return bcrypt.CompareHashAndPassword([]byte(hash), []byte(pw)) == nil
}

func newToken() string {
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return base64.RawURLEncoding.EncodeToString(b)
}

// CreateSession persiste le token hashé (SHA256) pour éviter vol en clair en DB.
func CreateSession(d *sql.DB, userID int64, r *http.Request, ttl time.Duration) (string, error) {
	tok := newToken()
	sum := sha256.Sum256([]byte(tok))
	_, err := d.Exec(`INSERT INTO sessions(token,user_id,expires_at,ip,ua) VALUES(?,?,?,?,?)`,
		hexKey(sum[:]), userID, time.Now().Add(ttl).UTC().Format(time.RFC3339),
		r.Header.Get("X-Forwarded-For"), r.UserAgent())
	return tok, err
}

func hexKey(b []byte) string {
	const h = "0123456789abcdef"
	out := make([]byte, len(b)*2)
	for i, v := range b {
		out[i*2] = h[v>>4]
		out[i*2+1] = h[v&0x0f]
	}
	return string(out)
}

func SetCookie(w http.ResponseWriter, tok string, ttl time.Duration, secure bool) {
	http.SetCookie(w, &http.Cookie{
		Name:     cookieName,
		Value:    tok,
		Path:     "/",
		HttpOnly: true,
		Secure:   secure,
		SameSite: http.SameSiteLaxMode,
		MaxAge:   int(ttl.Seconds()),
	})
}

func ClearCookie(w http.ResponseWriter) {
	http.SetCookie(w, &http.Cookie{Name: cookieName, Path: "/", MaxAge: -1})
}

// DeleteSession supprime la session en base (le logout ne doit pas se
// contenter d'effacer le cookie : le token resterait valable 30 jours).
func DeleteSession(d *sql.DB, r *http.Request) {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return
	}
	sum := sha256.Sum256([]byte(c.Value))
	_, _ = d.Exec(`DELETE FROM sessions WHERE token=?`, hexKey(sum[:]))
}

// CurrentUserID renvoie nil si anonyme (compte optionnel).
func CurrentUserID(d *sql.DB, r *http.Request) *int64 {
	c, err := r.Cookie(cookieName)
	if err != nil || c.Value == "" {
		return nil
	}
	sum := sha256.Sum256([]byte(c.Value))
	var uid int64
	var exp string
	err = d.QueryRow(`SELECT user_id, expires_at FROM sessions WHERE token=?`, hexKey(sum[:])).Scan(&uid, &exp)
	if err != nil {
		return nil
	}
	t, err := time.Parse(time.RFC3339, exp)
	if err != nil || time.Now().After(t) {
		return nil
	}
	return &uid
}

// CSRF : token par session, comparaison constant-time.
func CSRFToken(secret, session string) string {
	sum := sha256.Sum256([]byte(secret + "|" + session))
	return base64.RawURLEncoding.EncodeToString(sum[:])
}

func CheckCSRF(secret, session, got string) bool {
	want := CSRFToken(secret, session)
	return subtle.ConstantTimeCompare([]byte(want), []byte(got)) == 1
}

var ErrInvalid = errors.New("identifiants invalides")
