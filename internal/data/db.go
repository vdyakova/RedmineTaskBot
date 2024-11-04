package data

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"log"
	"os"
)

var pool *pgxpool.Pool

func init() {
	var err error
	db := os.Getenv("DATABASE_URL")
	pool, err = pgxpool.New(context.Background(), db)
	if err != nil {
		log.Fatal(err)
	}
}

func GetDB() *pgxpool.Pool {
	return pool
}
