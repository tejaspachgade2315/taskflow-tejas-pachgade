package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	JWTSecret     string
	JWTExpiry     time.Duration
	BcryptCost    int
	AllowedOrigin string
	MigrationDir  string
}

func Load() (Config, error) {
	expiryHours, err := getIntEnv("JWT_EXP_HOURS", 24)
	if err != nil {
		return Config{}, fmt.Errorf("parse JWT_EXP_HOURS: %w", err)
	}

	bcryptCost, err := getIntEnv("BCRYPT_COST", 12)
	if err != nil {
		return Config{}, fmt.Errorf("parse BCRYPT_COST: %w", err)
	}

	if bcryptCost < 12 {
		return Config{}, fmt.Errorf("BCRYPT_COST must be >= 12")
	}

	cfg := Config{
		Port:          getEnv("APP_PORT", "8080"),
		DatabaseURL:   os.Getenv("DATABASE_URL"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTExpiry:     time.Duration(expiryHours) * time.Hour,
		BcryptCost:    bcryptCost,
		AllowedOrigin: getEnv("ALLOWED_ORIGIN", "http://localhost:5173"),
		MigrationDir:  getEnv("MIGRATIONS_DIR", "./migrations"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	if cfg.JWTSecret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func getIntEnv(key string, fallback int) (int, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return 0, err
	}
	return parsed, nil
}
