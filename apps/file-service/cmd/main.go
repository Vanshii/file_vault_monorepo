package main

import (
    "log"
    "net/http"

    "file-service/config"
    "file-service/database"
    "file-service/models"
    "file-service/routes"

    "github.com/rs/cors"
)

func main() {
    config.LoadEnv()
    database.Init()

    // Run DB migrations for files table
    _, err := database.DB.Exec(models.FileTableMigration())
    if err != nil {
        log.Fatal("Failed migration:", err)
    }

    r := routes.Init()

    // Setup CORS to allow frontend calls from http://localhost:3000
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    })

    handler := c.Handler(r)

    log.Println("File service started on :8001 with CORS")
    log.Fatal(http.ListenAndServe(":8001", handler))
}
