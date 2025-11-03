package storage

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	
	
)

type Storage struct {
	DB *pgxpool.Pool
}

func New(DatabaseURL string) (*Storage, error) {

	cfg, err := pgxpool.ParseConfig(DatabaseURL)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

	db, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(ctx); err != nil {
		return nil, err
	}

	log.Println("[INFO] connected to PostgreSQL")

	return &Storage{DB: db}, nil
}

func (s *Storage) Close() {
	s.DB.Close()
}