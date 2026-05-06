package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	JWT    JWTConfig
}

type ServerConfig struct {
	Host string
	Port string
}

type DBConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	DBName   string
	SSLMode  string
}

type JWTConfig struct {
	Secret          string
	AccessTokenTTL  time.Duration
	RefreshTokenTTL time.Duration
}

func (d DBConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		d.Host, d.Port, d.User, d.Password, d.DBName, d.SSLMode,
	)
}

func Load() (*Config, error) {
	accessTTL, err := strconv.Atoi(getEnv("JWT_ACCESS_TTL_MINUTES", "60"))
	if err != nil {
		return nil, fmt.Errorf("JWT_ACCESS_TTL_MINUTES: %w", err)
	}
	refreshTTL, err := strconv.Atoi(getEnv("JWT_REFRESH_TTL_HOURS", "720"))
	if err != nil {
		return nil, fmt.Errorf("JWT_REFRESH_TTL_HOURS: %w", err)
	}

	return &Config{
		Server: ServerConfig{
			Host: getEnv("SERVER_HOST", "0.0.0.0"),
			Port: getEnv("SERVER_PORT", "8080"),
		},
		DB: DBConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnv("DB_PORT", "5435"),
			User:     getEnv("DB_USER", "postgres"),
			Password: getEnv("DB_PASSWORD", "postgres"),
			DBName:   getEnv("DB_NAME", "postgres"),
			SSLMode:  getEnv("DB_SSLMODE", "disable"),
		},
		JWT: JWTConfig{
			Secret:          getEnv("JWT_SECRET", "TimofeyRomanov"),
			AccessTokenTTL:  time.Duration(accessTTL) * time.Minute,
			RefreshTokenTTL: time.Duration(refreshTTL) * time.Hour,
		},
	}, nil
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
