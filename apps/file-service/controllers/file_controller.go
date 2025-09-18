package controllers

import (
    "database/sql"
    "encoding/json"
    
    "io"
    
    "mime"
    
    "net/http"
    "os"
    "path/filepath"
    
    "time"

    "file-service/database"
    "file-service/models"
    "file-service/utils"
)

const UploadPath = "./uploads"

func UploadFile(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(10 << 20) // limit 10MB
    if err != nil {
        http.Error(w, "Error parsing form data", http.StatusBadRequest)
        return
    }

    files := r.MultipartForm.File["files"]
    if len(files) == 0 {
        http.Error(w, "No files uploaded", http.StatusBadRequest)
        return
    }

    uploader := r.Header.Get("Uploader") // pass uploader info in header or JWT token ideally

    if _, err := os.Stat(UploadPath); os.IsNotExist(err) {
        os.MkdirAll(UploadPath, os.ModePerm)
    }

    var uploadedFiles []models.File

    for _, fileHeader := range files {
        f, err := fileHeader.Open()
        if err != nil {
            http.Error(w, "Could not open uploaded file", http.StatusInternalServerError)
            return
        }

        // Compute SHA256 hash
        hash, err := utils.ComputeSHA256(f)
        if err != nil {
            http.Error(w, "Error hashing file", http.StatusInternalServerError)
            return
        }
        f.Close()

        // Check if file with hash already exists in DB
        var existingFile models.File
        err = database.DB.QueryRow("SELECT id, filename, uploader, size, mime_type, content_hash, upload_date, reference_count FROM files WHERE content_hash = $1", hash).Scan(
            &existingFile.ID, &existingFile.Filename, &existingFile.Uploader,
            &existingFile.Size, &existingFile.MIMEType, &existingFile.ContentHash,
            &existingFile.UploadDate, &existingFile.ReferenceCount,
        )

        if err != nil && err != sql.ErrNoRows {
            http.Error(w, "DB error", http.StatusInternalServerError)
            return
        }

        if err == nil {
            // Duplicate file found, increment reference count
            _, err = database.DB.Exec("UPDATE files SET reference_count = reference_count + 1 WHERE id = $1", existingFile.ID)
            if err != nil {
                http.Error(w, "DB error updating reference count", http.StatusInternalServerError)
                return
            }
            uploadedFiles = append(uploadedFiles, existingFile)
            continue
        }

        // Save file to disk
        src, err := fileHeader.Open()
        if err != nil {
            http.Error(w, "Error reading file", http.StatusInternalServerError)
            return
        }
        defer src.Close()

        dstPath := filepath.Join(UploadPath, hash)
        dst, err := os.Create(dstPath)
        if err != nil {
            http.Error(w, "Could not save file", http.StatusInternalServerError)
            return
        }
        defer dst.Close()

        if _, err := io.Copy(dst, src); err != nil {
            http.Error(w, "Could not write file", http.StatusInternalServerError)
            return
        }

        // Validate MIME type matches file extension (basic)
        ext := filepath.Ext(fileHeader.Filename)
        mimeType := mime.TypeByExtension(ext)
        if mimeType == "" {
            mimeType = fileHeader.Header.Get("Content-Type")
        }

        // Insert metadata in DB
        _, err = database.DB.Exec(
            `INSERT INTO files (filename, uploader, size, mime_type, content_hash, upload_date, reference_count) 
             VALUES ($1, $2, $3, $4, $5, $6, $7)`,
            fileHeader.Filename, uploader, fileHeader.Size, mimeType, hash, time.Now(), 1,
        )
        if err != nil {
            http.Error(w, "DB error inserting file", http.StatusInternalServerError)
            return
        }

        uploadedFiles = append(uploadedFiles, models.File{
            Filename:    fileHeader.Filename,
            Uploader:    uploader,
            Size:        fileHeader.Size,
            MIMEType:    mimeType,
            ContentHash: hash,
            UploadDate:  time.Now(),
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(uploadedFiles)
}

func ListFiles(w http.ResponseWriter, r *http.Request) {
    rows, err := database.DB.Query("SELECT id, filename, uploader, size, mime_type, content_hash, upload_date, reference_count FROM files")
    if err != nil {
        http.Error(w, "DB error", http.StatusInternalServerError)
        return
    }
    defer rows.Close()

    var files []models.File
    for rows.Next() {
        var f models.File
        err := rows.Scan(&f.ID, &f.Filename, &f.Uploader, &f.Size, &f.MIMEType, &f.ContentHash, &f.UploadDate, &f.ReferenceCount)
        if err != nil {
            http.Error(w, "DB error reading file", http.StatusInternalServerError)
            return
        }
        files = append(files, f)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}
