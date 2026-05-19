package config

import "time"

type Config struct {
	ServerAddr string
	DSN        string
	JWTSecret  string
	AccessTTL  time.Duration
	RefreshTTL time.Duration
}

func Load() *Config {
	return &Config{
		ServerAddr: "0.0.0.0:8888",
		DSN:        "host=localhost port=5435 user=postgres password=1234 dbname=postgres sslmode=disable",
		JWTSecret:  "TimofeyRomanov",
		AccessTTL:  15 * time.Minute,
		RefreshTTL: 720 * time.Hour,
	}
}
