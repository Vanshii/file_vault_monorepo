package routes

import (
    "net/http"

    "github.com/gorilla/mux"
    "file-service/controllers"
    "file-service/middleware"
)

func Init() *mux.Router {
    r := mux.NewRouter()

    // r.HandleFunc("/upload", controllers.UploadFile).Methods("POST")
    // r.Handle("/files", middleware.JWTAuth(http.HandlerFunc(controllers.ListFiles))).Methods("GET")


	 r.Handle("/upload", middleware.JWTAuth(http.HandlerFunc(controllers.UploadFile))).Methods("POST")

    r.Handle("/files", middleware.JWTAuth(http.HandlerFunc(controllers.ListFiles))).Methods("GET")

    r.Handle("/files/search", middleware.JWTAuth(http.HandlerFunc(controllers.SearchFiles))).Methods("GET")

    r.Handle("/files/{id}/download", middleware.JWTAuth(http.HandlerFunc(controllers.DownloadFile))).Methods("GET")

    r.Handle("/files/{id}", middleware.JWTAuth(http.HandlerFunc(controllers.DeleteFile))).Methods("DELETE")

    r.Handle("/files/{id}/share", middleware.JWTAuth(http.HandlerFunc(controllers.ShareFilePublic))).Methods("POST")

    r.HandleFunc("/public/{link}/download", controllers.DownloadPublicFile).Methods("GET") // Public route, no auth
    return r
}
