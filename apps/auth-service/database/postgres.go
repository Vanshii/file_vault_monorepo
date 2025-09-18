package database

import (
    "database/sql"
    "fmt"
    "log"
    _ "github.com/lib/pq"
    "auth-service/config"
)

var DB *sql.DB

func Init() {
    config.LoadEnv()

    connStr := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
        config.GetEnv("DB_HOST"),
        config.GetEnv("DB_PORT"),
        config.GetEnv("DB_USER"),
        config.GetEnv("DB_PASSWORD"),
        config.GetEnv("DB_NAME"),
    )
    var err error
    DB, err = sql.Open("postgres", connStr)
    if err != nil {
        log.Fatal("Failed to connect to DB:", err)
    }
    if err = DB.Ping(); err != nil {
        log.Fatal("Failed to ping DB:", err)
    }
    log.Println("Connected to DB successfully")
}
