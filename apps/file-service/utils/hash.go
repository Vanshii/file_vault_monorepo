package utils

import (
    "crypto/sha256"
    "encoding/hex"
    "io"
    "mime/multipart"
)

func ComputeSHA256(file multipart.File) (string, error) {
    hasher := sha256.New()
    if _, err := io.Copy(hasher, file); err != nil {
        return "", err
    }
    return hex.EncodeToString(hasher.Sum(nil)), nil
}
