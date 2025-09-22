# File Service

A backend service for managing file uploads, downloads, search, deletion, and public sharing with JWT-secured authentication.

## Features

- Upload multiple files with SHA-256 deduplication
- List user files with metadata and ownership security
- Download files with download count tracking
- Delete files with reference counting and access control
- Search files by filename, MIME type, size, and date filters
- Generate public share links for unauthenticated access

## Architecture

<img width="600" height="600" alt="image" src="https://github.com/user-attachments/assets/2320946a-1243-4cd7-8917-aadc8d8ca404" />


- **Frontend (File Vault)**: React + Tailwind application providing user interface for uploading, viewing, downloading, and sharing files.
- **Auth Service**: Go-based microservice handling user registration, login, JWT authentication, and access control.
- **File Service**: Go-based backend managing file storage, metadata, deduplication, search, sharing, and download tracking.
- **Database (PostgreSQL)**: Stores user accounts, file metadata, reference counts, download counts, and public sharing links.
- **Communication**: Frontend communicates with Auth Service for authentication and File Service for file operations. Services are JWT-secured.
- **Storage**: Files are stored on the server or cloud storage with SHA-256 content hash to prevent duplication.

## Setup Instructions

### Prerequisites

- Go 1.19+ installed
- PostgreSQL 12+ installed and running
- `psql` CLI tool to manage database
- Node.js and npm installed
- Git (optional)

### Database Setup

- Create database: `createdb file_service_db`
- Connect and enable trigram extension:
  - `psql -d file_service_db`
  - `CREATE EXTENSION IF NOT EXISTS pg_trgm;`
  - `\q`
- Ensure migration creates tables and indexes (done automatically on app start or manually):
  - Create `files` table:
    - `id SERIAL PRIMARY KEY`
    - `filename VARCHAR(255) NOT NULL`
    - `uploader VARCHAR(100) NOT NULL`
    - `size BIGINT NOT NULL`
    - `mime_type VARCHAR(100) NOT NULL`
    - `content_hash VARCHAR(64) NOT NULL UNIQUE`
    - `upload_date TIMESTAMP WITH TIME ZONE DEFAULT now()`
    - `reference_count INT NOT NULL DEFAULT 1`
    - `download_count INT NOT NULL DEFAULT 0`
    - `public_link VARCHAR(255) UNIQUE`
    - `is_public BOOLEAN DEFAULT FALSE`
  - Create indexes:
    - `CREATE INDEX IF NOT EXISTS idx_filename ON files USING gin (filename gin_trgm_ops);`
    - `CREATE INDEX IF NOT EXISTS idx_content_hash ON files(content_hash);`
    - `CREATE INDEX IF NOT EXISTS idx_uploader ON files(uploader);`

### Environment Variables

- `DB_HOST=localhost`
- `DB_PORT=5432`
- `DB_USER=your_user`
- `DB_PASSWORD=your_password`
- `DB_NAME=file_service_db`
- `JWT_SECRET=your_jwt_secret`

### Build and Run

- **Auth Service**
  - Change directory: `cd auth-service`
  - Install dependencies: `go mod tidy`
  - Build application: `go build ./cmd/main.go`
  - Run application: `./main`

- **File Service**
  - Change directory: `cd ../file-service`
  - Install dependencies: `go mod tidy`
  - Build application: `go build ./cmd/main.go`
  - Run application: `./main`
  - Service listens on default port `8001`

- **Frontend (File Vault)**
  - Change directory: `cd ../file_vault_frontend`
  - Install dependencies: `npm install`
  - Run development server: `npm run dev`

## Usage

- Test the APIs using tools like `curl` or Postman with JWT authentication.
- Use the frontend to interact with uploaded files, view metadata, and access public links.

## License

- MIT License

## Support

- For questions or support, contact Vanshika at `vanshika.2022@vitstudent.ac.in`
