package routes

import (
    "net/http"

    "github.com/gorilla/mux"
    "file-service/controllers"
    "file-service/middleware"
)

func Init() *mux.Router {
    r := mux.NewRouter()

    r.HandleFunc("/upload", controllers.UploadFile).Methods("POST")
    r.Handle("/files", middleware.JWTAuth(http.HandlerFunc(controllers.ListFiles))).Methods("GET")

    return r
}
