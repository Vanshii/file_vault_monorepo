package main

import (
    "log"
    "net/http"
    "auth-service/config"
    "auth-service/database"
    "auth-service/models"
    "auth-service/routes"
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
    log.Println("Server started on :8000")
    log.Fatal(http.ListenAndServe(":8000", r))
}
