package config

import (
	"errors"
	"os"
	"strconv"

	"github.com/joho/godotenv"
)

type Config struct {
	Addr          string
	DBDSN         string
	SessionSecret string
	CookieSecure  bool
}

func Load() (Config, error) {
	_ = godotenv.Load()
	secure, err := strconv.ParseBool(env("COOKIE_SECURE", "false"))
	if err != nil {
		return Config{}, errors.New("COOKIE_SECURE harus true atau false")
	}
	cfg := Config{
		Addr:          env("ADDR", ":8080"),
		DBDSN:         os.Getenv("DB_DSN"),
		SessionSecret: os.Getenv("SESSION_SECRET"),
		CookieSecure:  secure,
	}
	if cfg.DBDSN == "" {
		return cfg, errors.New("DB_DSN belum diatur")
	}
	if len(cfg.SessionSecret) < 32 {
		return cfg, errors.New("SESSION_SECRET minimal 32 karakter")
	}
	return cfg, nil
}

func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
