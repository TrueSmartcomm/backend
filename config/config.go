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
    //через env файл
    if err := godotenv.Load(); err != nil {
        log.Println("[WARN] .env file not found")
    }


    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        
        dbUser := os.Getenv("DB_USER")
        dbPassword := os.Getenv("DB_PASSWORD")
        dbHost := os.Getenv("DB_HOST")
        dbPort := os.Getenv("DB_PORT")
        dbName := os.Getenv("DB_NAME")
        
        databaseURL = fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
            dbUser, dbPassword, dbHost, dbPort, dbName)
    }

    cfg := &Config{
        Port:        getEnv("PORT", "8080"),
        DatabaseURL: databaseURL,
    }

    // Валидация
    if cfg.Port == "" {
        return nil, fmt.Errorf("PORT is required")
    }
    if cfg.DatabaseURL == "" {
        log.Println("[WARN] DATABASE_URL is empty - database connection may fail")
    }

    return cfg, nil
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func MustLoad() *Config {
    cfg, err := Load()
    if err != nil {
        log.Fatalf("config error: %v", err)
    }
    return cfg
}