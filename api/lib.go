package main

import (
    "os"
    "log"
    "context"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
)

type Photo struct {
    ID              string      `json:"id"`
    Title           string      `json:"title"`
    Timestamp       *time.Time  `json:"timestamp,omitempty"`
    Precedence      *int        `json:"precedence,omitempty"`
}
type Collection struct {
    ID              string      `json:"id"`
    Title           string      `json:"title"`
    Precedence      *int        `json:"precedence,omitempty"`
}

var ctx = context.Background()

var db *pgxpool.Pool
func initializePhotosDB() error {
    postgresURL := os.Getenv("PHOTOS_POSTGRES_URL")

    var err error
	db, err = pgxpool.New(ctx, postgresURL)
	if err != nil {
        return err
	}

    if err := db.Ping(context.Background()); err != nil {
        return err
	}

    log.Print("Successfully connected to photos database")
    return nil
}
