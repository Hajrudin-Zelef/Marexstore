package handlers

import (
	"crypto/sha256"
	"crypto/subtle"
	"sync"
	"time"

	"github.com/Hajrudin-Zelef/Marexstore/internal/auth"
)

const adminAuthTTL = 10 * time.Minute

var adminAuthCache struct {
	sync.Mutex
	sum []byte
	exp time.Time
}

// adminPassOK : bcrypt (coût 12) seulement si ce couple hash+mot de passe
// n'a pas déjà été validé dans les dernières minutes.
func adminPassOK(hash, pass string) bool {
	sum := sha256.Sum256([]byte(hash + "\x00" + pass))
	now := time.Now()
	adminAuthCache.Lock()
	hit := adminAuthCache.sum != nil && now.Before(adminAuthCache.exp) &&
		subtle.ConstantTimeCompare(adminAuthCache.sum, sum[:]) == 1
	adminAuthCache.Unlock()
	if hit {
		return true
	}
	if !auth.CheckPassword(hash, pass) {
		return false
	}
	adminAuthCache.Lock()
	adminAuthCache.sum = sum[:]
	adminAuthCache.exp = now.Add(adminAuthTTL)
	adminAuthCache.Unlock()
	return true
}
