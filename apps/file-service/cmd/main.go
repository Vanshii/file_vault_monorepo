package main

import (
    "log"
    "net/http"

    "file-service/config"
    "file-service/database"
    "file-service/models"
    "file-service/routes"
)

func main() {
    config.LoadEnv()
    database.Init()

    _, err := database.DB.Exec(models.FileTableMigration())
    if err != nil {
        log.Fatal("Failed migration:", err)
    }

    r := routes.Init()

    log.Println("File service started on :8001")
    log.Fatal(http.ListenAndServe(":8001", r))
}
