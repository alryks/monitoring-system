package config

import (
	"os"
)

type Config struct {
	DatabaseURL string
	JWTSecret   string
	Port        string
	AdminUser   string
	AdminPass   string
}

func Load() *Config {

	databaseURL := getEnv("POSTGRES_URL", "")

	return &Config{
		DatabaseURL: databaseURL,
		JWTSecret:   getEnv("JWT_SECRET", "my-secret"),
		Port:        "8000",
		AdminUser:   getEnv("USER", "admin"),
		AdminPass:   getEnv("PASSWORD", "admin"),
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
