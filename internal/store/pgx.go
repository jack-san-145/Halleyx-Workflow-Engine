package store

import (
	"context"
	"log"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

var pool *pgxpool.Pool

func ConnectPostgres(dsn string) error {
	var err error
	pool, err = pgxpool.New(context.Background(), dsn)
	if err != nil {
		return err
	}
	// simple ping check
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := pool.Ping(ctx); err != nil {
		return err
	}
	log.Println("Connected to Postgres")
	return nil
}

func GetPool() *pgxpool.Pool {
	return pool
}
