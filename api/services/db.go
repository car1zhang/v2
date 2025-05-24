package services

import (
    "os"
    "context"
    "log"

    "github.com/jackc/pgx/v5/pgxpool"
)

var DB *pgxpool.Pool

func InitializePhotosDB() {
    postgresURL := os.Getenv("PHOTOS_POSTGRES_URL")

    var err error
	DB, err = pgxpool.New(context.Background(), postgresURL)
	if err != nil {
        log.Fatal(err)
	}

    if err := DB.Ping(context.Background()); err != nil {
        log.Fatal(err)
	}

    log.Print("Successfully connected to photos database")
}
