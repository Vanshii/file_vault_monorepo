package models

import "time"

type File struct {
    ID             int       `json:"id"`
    Filename       string    `json:"filename"`
    Uploader       string    `json:"uploader"`
    Size           int64     `json:"size"`
    MIMEType       string    `json:"mime_type"`
    ContentHash    string    `json:"content_hash"`
    UploadDate     time.Time `json:"upload_date"`
    ReferenceCount int       `json:"reference_count"`
}

func FileTableMigration() string {
    return `
    CREATE TABLE IF NOT EXISTS files (
        id SERIAL PRIMARY KEY,
        filename VARCHAR(255) NOT NULL,
        uploader VARCHAR(100) NOT NULL,
        size BIGINT NOT NULL,
        mime_type VARCHAR(100) NOT NULL,
        content_hash VARCHAR(64) NOT NULL UNIQUE,
        upload_date TIMESTAMP WITH TIME ZONE DEFAULT now(),
        reference_count INT NOT NULL DEFAULT 1
    );
    `
}
