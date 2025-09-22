package controllers

import (
    "encoding/json"
    "net/http"
    "strings"
    "auth-service/database"
    "auth-service/models"
    "auth-service/utils"
)

// CORS middleware wrapper function
func WithCORS(handler http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        w.Header().Set("Access-Control-Allow-Origin", "http://localhost:5173")
        w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
        w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

        if r.Method == http.MethodOptions {
            w.WriteHeader(http.StatusNoContent)
            return
        }

        handler(w, r)
    }
}

type AuthRequest struct {
    Username string `json:"username"`
    Email    string `json:"email"`
    Password string `json:"password"`
}

func Register(w http.ResponseWriter, r *http.Request) {
    var req AuthRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }
    hashedPassword, err := utils.HashPassword(req.Password)
    if err != nil {
        http.Error(w, "Error hashing password", http.StatusInternalServerError)
        return
    }

    var userID int
    err = database.DB.QueryRow(
        "INSERT INTO users (username, email, password) VALUES ($1, $2, $3) RETURNING id",
        req.Username, req.Email, hashedPassword).Scan(&userID)
    if err != nil {
        http.Error(w, "User already exists or DB error", http.StatusConflict)
        return
    }
    w.WriteHeader(http.StatusCreated)
    w.Write([]byte(`{"message":"Registration successful"}`))
}

func Login(w http.ResponseWriter, r *http.Request) {
    var req AuthRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        http.Error(w, "Bad request", http.StatusBadRequest)
        return
    }

    row := database.DB.QueryRow("SELECT id, username, password FROM users WHERE username = $1", req.Username)
    var user models.User
    err := row.Scan(&user.ID, &user.Username, &user.Password)
    if err != nil {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }
    if !utils.CheckPasswordHash(req.Password, user.Password) {
        http.Error(w, "Invalid credentials", http.StatusUnauthorized)
        return
    }

    token, err := utils.GenerateJWT(user.Username)
    if err != nil {
        http.Error(w, "Could not generate token", http.StatusInternalServerError)
        return
    }
    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{"token": token})
}

func Protected(w http.ResponseWriter, r *http.Request) {
    token := r.Header.Get("Authorization")
    if token == "" {
        http.Error(w, "Missing token", http.StatusUnauthorized)
        return
    }
    token = strings.TrimPrefix(token, "Bearer ")
    claims, err := utils.ValidateToken(token)
    if err != nil {
        http.Error(w, "Invalid token", http.StatusUnauthorized)
        return
    }
    w.Write([]byte(`{"message":"Welcome ` + claims.Username + `"}`))
}
