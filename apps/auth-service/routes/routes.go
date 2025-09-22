package routes

import (
    "net/http"
    "auth-service/controllers" // Import by package name, not filename
)

func SetupRoutes() {
    http.HandleFunc("/register", controllers.WithCORS(controllers.Register))
    http.HandleFunc("/login", controllers.WithCORS(controllers.Login))
    http.HandleFunc("/protected", controllers.WithCORS(controllers.Protected))
}
