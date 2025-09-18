package utils

import (
    "time"
    "github.com/golang-jwt/jwt/v5"
    "file-service/config"
)

type Claims struct {
    Username string `json:"username"`
    jwt.RegisteredClaims
}

func GenerateJWT(username string) (string, error) {
    expTime := time.Now().Add(24 * time.Hour)
    claims := &Claims{
        Username: username,
        RegisteredClaims: jwt.RegisteredClaims{
            ExpiresAt: jwt.NewNumericDate(expTime),
        },
    }
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
    return token.SignedString([]byte(config.GetEnv("JWT_SECRET")))
}

func ValidateToken(tokenStr string) (*Claims, error) {
    claims := &Claims{}
    token, err := jwt.ParseWithClaims(tokenStr, claims, func(token *jwt.Token) (interface{}, error) {
        return []byte(config.GetEnv("JWT_SECRET")), nil
    })
    if err != nil || !token.Valid {
        return nil, err
    }
    return claims, nil
}
