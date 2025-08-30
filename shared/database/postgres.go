package database

import (
	"database/sql"
	"fmt"
	"os"

	"meal-prep/shared/logging"

	_ "github.com/lib/pq"
)

type DB struct {
	*sql.DB
}

func NewPostgresConnection() (*DB, error) {
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}

	psqlInfo := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		host, port, user, password, dbname)

	logging.Logger.Debug("Connecting to database", "host", host, "port", port, "database", dbname)

	db, err := sql.Open("postgres", psqlInfo)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logging.Logger.Info("Successfully connected to PostgreSQL")
	return &DB{db}, nil
}
