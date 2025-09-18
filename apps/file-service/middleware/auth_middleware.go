package middleware

import (
    "net/http"
    "strings"
    "file-service/utils"
)

func JWTAuth(next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "No Authorization header provided", http.StatusUnauthorized)
            return
        }
        token := strings.TrimPrefix(authHeader, "Bearer ")
        _, err := utils.ValidateToken(token)
        if err != nil {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }
        next.ServeHTTP(w, r)
    })
}
