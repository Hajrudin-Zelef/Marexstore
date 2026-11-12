package config

import (
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
	return Config{
		Port:          envOr("PORT", "8080"),
		DBPath:        envOr("DB_PATH", "store.db"),
		SessionSecret: envOr("SESSION_SECRET", "change-me-en-prod-32-caracteres-min"),
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
