package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	HTTPAddr       string
	DatabaseURL    string
	JWTSecret      []byte
	JWTExpiration  time.Duration
	MigrationsPath string
}

func Load() (Config, error) {
	httpAddr := os.Getenv("HTTP_ADDR")
	if httpAddr == "" {
		httpAddr = ":8080"
	}
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET is required")
	}
	hours := 168
	if v := os.Getenv("JWT_EXPIRATION_HOURS"); v != "" {
		n, err := strconv.Atoi(v)
		if err != nil || n <= 0 {
			return Config{}, fmt.Errorf("JWT_EXPIRATION_HOURS must be a positive integer")
		}
		hours = n
	}
	migrationsPath := os.Getenv("MIGRATIONS_PATH")
	if migrationsPath == "" {
		migrationsPath = "migrations"
	}
	return Config{
		HTTPAddr:       httpAddr,
		DatabaseURL:    dbURL,
		JWTSecret:      []byte(secret),
		JWTExpiration:  time.Duration(hours) * time.Hour,
		MigrationsPath: migrationsPath,
	}, nil
}
