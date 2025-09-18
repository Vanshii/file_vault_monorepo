package routes

import (
    "net/http"
    "github.com/gorilla/mux"
    "auth-service/controllers"
    "auth-service/middleware"
)

func Init() *mux.Router {
    r := mux.NewRouter()
    r.HandleFunc("/register", controllers.Register).Methods("POST")
    r.HandleFunc("/login", controllers.Login).Methods("POST")
    r.Handle("/protected", middleware.JWTAuth(http.HandlerFunc(controllers.Protected))).Methods("GET")
    return r
}
