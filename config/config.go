package config

import (
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port        string
	DatabaseURL string
}


func Load() (*Config, error) {
	
	_ = godotenv.Load()

	cfg := &Config{
		Port:        getEnv("PORT", "8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
	}

	// Мини-валидация.
	if cfg.Port == "" {
		return nil, fmt.Errorf("PORT is required")
	}
	if cfg.DatabaseURL == "" {
		log.Println("[WARN] DATABASE_URL is empty")
	}

	return cfg, nil
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}


func MustLoad() *Config {
	cfg, err := Load()
	if err != nil {
		log.Fatalf("config error: %v", err)
	}
	return cfg
}
