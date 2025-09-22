## Table: `files`

### Columns

- **id** (`SERIAL PRIMARY KEY`): Unique identifier for each file.
- **filename** (`VARCHAR(255) NOT NULL`): Original name of the uploaded file.
- **uploader** (`VARCHAR(100) NOT NULL`): Username of the uploader.
- **size** (`BIGINT NOT NULL`): File size in bytes.
- **mime_type** (`VARCHAR(100) NOT NULL`): MIME type of the file.
- **content_hash** (`VARCHAR(64) UNIQUE NOT NULL`): SHA-256 hash for deduplication.
- **upload_date** (`TIMESTAMPTZ DEFAULT now()`): Timestamp when the file was uploaded.
- **reference_count** (`INT DEFAULT 1`): Number of references to the file.
- **download_count** (`INT DEFAULT 0`): Total number of downloads.
- **public_link** (`VARCHAR(255) UNIQUE NULLABLE`): Randomized link for public sharing.
- **is_public** (`BOOLEAN DEFAULT FALSE`): Indicates whether the file is publicly accessible.

### Indexes

- Primary key on `id`.
- Unique index on `content_hash` for deduplication.
- GIN index on `filename` with trigram operations for fast substring search.
- Index on `uploader` for user-specific queries.

### Purpose

This schema is designed to:

- Support **deduplication** of files via `content_hash`.
- Track **user ownership** through the `uploader` field.
- Enable **efficient searching** using GIN + trigram indexes.
- Allow **secure public sharing** via `public_link` and `is_public`.
- Provide analytics such as **download counts** and **reference counts**.
