package utils

import (
    "crypto/rand"
    "encoding/base64"
    "strings"
)

// GenerateRandomString returns a URL-safe random string of length n
func GenerateRandomString(n int) string {
    b := make([]byte, n)
    _, err := rand.Read(b)
    if err != nil {
        return "" // or handle error
    }
    s := base64.URLEncoding.EncodeToString(b)
    // base64 may be longer than n because of encoding, so trim
    s = strings.TrimRight(s, "=")
    if len(s) > n {
        s = s[:n]
    }
    return s
}
