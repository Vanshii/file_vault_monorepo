package controllers

import (
    "database/sql"
    "encoding/json"
    "fmt"
    "io"
    "mime"
    "net/http"
    "os"
    "path"
    "path/filepath"
    "time"

    "github.com/gorilla/mux"

    "file-service/database"
    "file-service/models"
    "file-service/utils"
)

const UploadPath = "./uploads"

// UploadFile handles file uploads with deduplication and quota (simplified quota management)
func UploadFile(w http.ResponseWriter, r *http.Request) {
    err := r.ParseMultipartForm(50 << 20) // 50MB per request limit
    if err != nil {
        http.Error(w, "Error parsing form data", http.StatusBadRequest)
        return
    }

    files := r.MultipartForm.File["files"]
    if len(files) == 0 {
        http.Error(w, "No files uploaded", http.StatusBadRequest)
        return
    }

    // Ideally get uploader from JWT context, here use header for demo:
    uploader := r.Header.Get("Uploader")
    if uploader == "" {
        http.Error(w, "Uploader info missing", http.StatusUnauthorized)
        return
    }

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

        // Compute SHA256
        hash, err := utils.ComputeSHA256(f)
        if err != nil {
            http.Error(w, "Error hashing file", http.StatusInternalServerError)
            f.Close()
            return
        }
        f.Close()

        // Check for existing file with hash
        var existingFile models.File
        err = database.DB.QueryRow("SELECT id, filename, uploader, size, mime_type, content_hash, upload_date, reference_count, download_count, is_public, public_link FROM files WHERE content_hash = $1", hash).Scan(
            &existingFile.ID, &existingFile.Filename, &existingFile.Uploader,
            &existingFile.Size, &existingFile.MIMEType, &existingFile.ContentHash,
            &existingFile.UploadDate, &existingFile.ReferenceCount,
            &existingFile.DownloadCount, &existingFile.IsPublic, &existingFile.PublicLink,
        )

        if err != nil && err != sql.ErrNoRows {
            http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)

            return
        }

        if err == nil {
            // Duplicate - increment reference count
            _, err = database.DB.Exec("UPDATE files SET reference_count = reference_count + 1 WHERE id = $1", existingFile.ID)
            if err != nil {
                http.Error(w, "DB error updating reference count", http.StatusInternalServerError)
                return
            }

            // You should update user's quota usage here (omitted for brevity)

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

        // Determine MIME type based on extension or headers (basic)
        ext := filepath.Ext(fileHeader.Filename)
        mimeType := mime.TypeByExtension(ext)
        if mimeType == "" {
            mimeType = fileHeader.Header.Get("Content-Type")
        }

        _, err = database.DB.Exec(
    `INSERT INTO files (filename, uploader, size, mime_type, content_hash, upload_date, reference_count, download_count, is_public, public_link) 
     VALUES ($1, $2, $3, $4, $5, $6, 1, 0, FALSE, NULL)`,
    fileHeader.Filename, uploader, fileHeader.Size, mimeType, hash, time.Now(),
)
        if err != nil {
            http.Error(w, "DB error inserting file", http.StatusInternalServerError)
            return
        }

        uploadedFiles = append(uploadedFiles, models.File{
            Filename:       fileHeader.Filename,
            Uploader:       uploader,
            Size:           fileHeader.Size,
            MIMEType:       mimeType,
            ContentHash:    hash,
            UploadDate:     time.Now(),
            ReferenceCount: 1,
            DownloadCount:  0,
            IsPublic:       false,
        })
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(uploadedFiles)
}

// ListFiles lists all files by the logged-in user
func ListFiles(w http.ResponseWriter, r *http.Request) {
    user, ok := r.Context().Value("username").(string)
    if !ok || user == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    rows, err := database.DB.Query(
        "SELECT id, filename, uploader, size, mime_type, content_hash, upload_date, reference_count, download_count, is_public, public_link FROM files WHERE uploader = $1",
        user,
    )
    if err != nil {
       http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)

        return
    }
    defer rows.Close()

    var files []models.File
    for rows.Next() {
        var f models.File
        err := rows.Scan(&f.ID, &f.Filename, &f.Uploader, &f.Size, &f.MIMEType, &f.ContentHash, &f.UploadDate, &f.ReferenceCount, &f.DownloadCount, &f.IsPublic, &f.PublicLink)
        if err != nil {
            http.Error(w, "DB error reading file", http.StatusInternalServerError)
            return
        }
        files = append(files, f)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}

// DownloadFile streams the file content and increments download count
func DownloadFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileID := vars["id"]

    user, ok := r.Context().Value("username").(string)
    if !ok || user == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var f models.File
    err := database.DB.QueryRow(
        "SELECT filename, content_hash, uploader, is_public FROM files WHERE id = $1",
        fileID,
    ).Scan(&f.Filename, &f.ContentHash, &f.Uploader, &f.IsPublic)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    if !f.IsPublic && f.Uploader != user {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    go func() {
        _, _ = database.DB.Exec("UPDATE files SET download_count = download_count + 1 WHERE id = $1", fileID)
    }()

    filePath := path.Join(UploadPath, f.ContentHash)
    http.ServeFile(w, r, filePath)
}

// DeleteFile deletes or decrements reference count for a file
func DeleteFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileID := vars["id"]

    user, ok := r.Context().Value("username").(string)
    if !ok || user == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    var refCount int
    var uploader string
    var contentHash string
    err := database.DB.QueryRow(
        "SELECT reference_count, uploader, content_hash FROM files WHERE id = $1",
        fileID,
    ).Scan(&refCount, &uploader, &contentHash)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }

    if uploader != user {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    if refCount > 1 {
        _, err = database.DB.Exec("UPDATE files SET reference_count = reference_count - 1 WHERE id = $1", fileID)
        if err != nil {
            http.Error(w, "Failed to decrement reference count", http.StatusInternalServerError)
            return
        }
    } else {
        _, err = database.DB.Exec("DELETE FROM files WHERE id = $1", fileID)
        if err != nil {
            http.Error(w, "Failed to delete record", http.StatusInternalServerError)
            return
        }
        os.Remove(path.Join(UploadPath, contentHash))
    }

    w.WriteHeader(http.StatusNoContent)
}

// SearchFiles supports filtering by filename, mime type, size range, date range, uploader (logged-in user)
func SearchFiles(w http.ResponseWriter, r *http.Request) {
    user, ok := r.Context().Value("username").(string)
    if !ok || user == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    query := "SELECT id, filename, uploader, size, mime_type, content_hash, upload_date, reference_count FROM files WHERE uploader = $1"
    args := []interface{}{user}
    idx := 2

    q := r.URL.Query()

    if filename := q.Get("filename"); filename != "" {
        query += fmt.Sprintf(" AND filename ILIKE $%d", idx)
        args = append(args, "%"+filename+"%")
        idx++
    }
    if mime := q.Get("mime"); mime != "" {
        query += fmt.Sprintf(" AND mime_type = $%d", idx)
        args = append(args, mime)
        idx++
    }
    if sizeMin := q.Get("size_min"); sizeMin != "" {
        query += fmt.Sprintf(" AND size >= $%d", idx)
        args = append(args, sizeMin)
        idx++
    }
    if sizeMax := q.Get("size_max"); sizeMax != "" {
        query += fmt.Sprintf(" AND size <= $%d", idx)
        args = append(args, sizeMax)
        idx++
    }
    if dateStart := q.Get("date_start"); dateStart != "" {
        query += fmt.Sprintf(" AND upload_date >= $%d", idx)
        args = append(args, dateStart)
        idx++
    }
    if dateEnd := q.Get("date_end"); dateEnd != "" {
        query += fmt.Sprintf(" AND upload_date <= $%d", idx)
        args = append(args, dateEnd)
        idx++
    }

    rows, err := database.DB.Query(query, args...)
    if err != nil {
       http.Error(w, "DB error: "+err.Error(), http.StatusInternalServerError)

        return
    }
    defer rows.Close()

    var files []models.File
    for rows.Next() {
        var f models.File
        err := rows.Scan(&f.ID, &f.Filename, &f.Uploader, &f.Size, &f.MIMEType, &f.ContentHash, &f.UploadDate, &f.ReferenceCount)
        if err != nil {
            http.Error(w, "DB error reading files", http.StatusInternalServerError)
            return
        }
        files = append(files, f)
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(files)
}

// ShareFilePublic generates a public link string for sharing
func ShareFilePublic(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    fileID := vars["id"]

    user, ok := r.Context().Value("username").(string)
    if !ok || user == "" {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Verify file ownership
    var uploader string
    err := database.DB.QueryRow("SELECT uploader FROM files WHERE id = $1", fileID).Scan(&uploader)
    if err != nil {
        http.Error(w, "File not found", http.StatusNotFound)
        return
    }
    if uploader != user {
        http.Error(w, "Unauthorized", http.StatusUnauthorized)
        return
    }

    // Generate random public link (example: use UUID or random string)
    publicLink := utils.GenerateRandomString(20)

    _, err = database.DB.Exec(
        "UPDATE files SET public_link = $1, is_public = TRUE WHERE id = $2",
        publicLink, fileID,
    )
    if err != nil {
        http.Error(w, "Failed to update file for sharing", http.StatusInternalServerError)
        return
    }

    w.Header().Set("Content-Type", "application/json")
    json.NewEncoder(w).Encode(map[string]string{
        "public_link": publicLink,
        "url":        fmt.Sprintf("http://localhost:8001/public/%s/download", publicLink),
    })
}

// DownloadPublicFile serves a public file by link (no auth required), increments download count
func DownloadPublicFile(w http.ResponseWriter, r *http.Request) {
    vars := mux.Vars(r)
    publicLink := vars["link"]

    var f models.File
    err := database.DB.QueryRow(
        "SELECT filename, content_hash FROM files WHERE public_link = $1 AND is_public = TRUE",
        publicLink,
    ).Scan(&f.Filename, &f.ContentHash)
    if err != nil {
        http.Error(w, "Public file not found", http.StatusNotFound)
        return
    }

    go func() {
        _, _ = database.DB.Exec("UPDATE files SET download_count = download_count + 1 WHERE public_link = $1", publicLink)
    }()

    filePath := path.Join(UploadPath, f.ContentHash)
    http.ServeFile(w, r, filePath)
}
