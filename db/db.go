package db

import (
	"context"
	"log"

	"github.com/jackc/pgx/v4/pgxpool"
)

func NewConnectionPool(databaseURL string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.Connect(context.Background(), databaseURL)
	if err != nil {
		return nil, err
	}

	log.Println("Connected to database")
	return pool, nil
}
