package main

import (
    "log"
    "net/http"

    "auth-service/config"
    "auth-service/database"
    "auth-service/models"
    "auth-service/routes"

    "github.com/rs/cors"
)

func main() {
    // Load environment variables/configuration
    config.LoadEnv()

    // Initialize database connection
    database.Init()

    // Run migration for User table
    _, err := database.DB.Exec(models.UserTableMigration())
    if err != nil {
        log.Fatal("Failed migration:", err)
    }

    // Setup HTTP routes with handlers (register/login/protected)
    routes.SetupRoutes()

    // Setup CORS middleware only here
    c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:5173"}, // Your frontend origin here
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    })

    // Use the default HTTP mux wrapped with the CORS handler
    handler := c.Handler(http.DefaultServeMux)

    log.Println("Server started on :8000")
    log.Fatal(http.ListenAndServe(":8000", handler))
}
