package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	ServerPort    int
	DatabaseURL   string
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	RequestIDKey  string
}

func Load() *Config {
	return &Config{
		ServerPort:   getEnvInt("SERVER_PORT", 8080),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/userage?sslmode=disable"),
		ReadTimeout:  time.Duration(getEnvInt("READ_TIMEOUT", 10)) * time.Second,
		WriteTimeout: time.Duration(getEnvInt("WRITE_TIMEOUT", 10)) * time.Second,
		RequestIDKey: "X-Request-ID",
	}
}

func getEnv(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return fallback
}
