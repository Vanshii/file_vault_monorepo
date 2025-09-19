package models

import "time"
import "database/sql"

type File struct {
    ID             int       `json:"id"`
    Filename       string    `json:"filename"`
    Uploader       string    `json:"uploader"`
    Size           int64     `json:"size"`
    MIMEType       string    `json:"mime_type"`
    ContentHash    string    `json:"content_hash"`
    UploadDate     time.Time `json:"upload_date"`
    ReferenceCount int       `json:"reference_count"`
    DownloadCount  int       `json:"download_count"`  // New: number of downloads
   PublicLink     sql.NullString `json:"public_link"` // New: unique public URL token
    IsPublic       bool      `json:"is_public"`       // New: whether file is publicly shared
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
        reference_count INT NOT NULL DEFAULT 1,
        download_count INT NOT NULL DEFAULT 0,
        public_link VARCHAR(255) UNIQUE,
        is_public BOOLEAN DEFAULT FALSE
    );

    CREATE INDEX IF NOT EXISTS idx_filename ON files USING gin (filename gin_trgm_ops);
    CREATE INDEX IF NOT EXISTS idx_content_hash ON files(content_hash);
    CREATE INDEX IF NOT EXISTS idx_uploader ON files(uploader);
    `
}
