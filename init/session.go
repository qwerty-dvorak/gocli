package init

import (
    "database/sql"
    "log"
    "os"
    "path/filepath"

    "github.com/joho/godotenv"
    _ "github.com/lib/pq"
)

func NewSession() (*sql.DB, error) {
    parentDir := filepath.Dir("..")
    err := godotenv.Load(filepath.Join(parentDir, ".env"))
    if err != nil {
        log.Fatalf("Error loading .env file from parent directory: %v", err)
    }
    databaseURL := os.Getenv("DATABASE_URL")
    if databaseURL == "" {
        log.Fatalf("DATABASE_URL not set in .env file")
    }
    return sql.Open("postgres", databaseURL)
}