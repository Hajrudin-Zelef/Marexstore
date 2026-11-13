package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"strconv"
)

type Config struct {
	Port          string
	DBPath        string
	SessionSecret string
	BaseURL       string
	WhatsApp      string
	AdminUser     string
	AdminPassHash string // bcrypt hash, vide = admin désactivé
	Env           string
}

func Load() Config {
	// Secret de session : jamais de valeur connue par défaut. Sans .env,
	// un secret éphémère est généré (les sessions ne survivent pas au redémarrage).
	secret := os.Getenv("SESSION_SECRET")
	if secret == "" {
		b := make([]byte, 32)
		if _, err := rand.Read(b); err == nil {
			secret = base64.RawURLEncoding.EncodeToString(b)
			log.Println("marexstore: SESSION_SECRET absent — secret éphémère généré (sessions perdues au redémarrage)")
		}
	}
	return Config{
		Port:          envOr("PORT", "8080"),
		DBPath:        envOr("DB_PATH", "store.db"),
		SessionSecret: secret,
		BaseURL:       envOr("BASE_URL", "http://localhost:8080"),
		WhatsApp:      envOr("WHATSAPP_NUMBER", "2250705087507"),
		AdminUser:     envOr("ADMIN_USER", "admin"),
		AdminPassHash: os.Getenv("ADMIN_PASS_HASH"),
		Env:           envOr("APP_ENV", "prod"),
	}
}

func envOr(k, d string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return d
}

func EnvInt(k string, d int) int {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	n, err := strconv.Atoi(v)
	if err != nil {
		return d
	}
	return n
}
