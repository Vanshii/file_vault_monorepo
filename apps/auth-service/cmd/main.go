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
    config.LoadEnv()
    database.Init()

    // Run migration for user table
    _, err := database.DB.Exec(models.UserTableMigration())
    if err != nil {
        log.Fatal("Failed migration:", err)
    }

    r := routes.Init()

	  c := cors.New(cors.Options{
        AllowedOrigins:   []string{"http://localhost:3000"},
        AllowedMethods:   []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
        AllowedHeaders:   []string{"Authorization", "Content-Type"},
        AllowCredentials: true,
    })

    handler := c.Handler(r)
    log.Println("Server started on :8000")
    log.Fatal(http.ListenAndServe(":8000", handler))
}
