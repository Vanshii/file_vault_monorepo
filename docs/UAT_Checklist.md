# User Acceptance Testing (UAT) Checklist – File Service App

This checklist is intended to guide testing of the File Service application to ensure all functionalities meet the requirements and work as expected.

---

## 1. Authentication

- Verify user registration with **valid inputs**:
  - Ensure all required fields are validated.
  - Confirm successful creation of a new user account.
- Verify registration **fails** with:
  - Duplicate usernames.
  - Invalid data formats (e.g., incorrect email, weak password).
- Verify user login returns a **valid JWT token**.
- Verify access to **protected APIs** using a valid JWT token.
- Verify access is **denied** for requests with:
  - Missing JWT token.
  - Expired or invalid token.

---

## 2. File Upload

- Upload a **single file** successfully:
  - Verify the file is stored correctly.
  - Confirm metadata (size, MIME type, uploader) is saved.
- Upload **multiple files** in a single request successfully.
- Verify **duplicate file upload**:
  - Does not store duplicate content.
  - Increments `reference_count` correctly.
- Upload files with **various MIME types** and ensure metadata accuracy.
- Verify **error handling** for:
  - Invalid file types.
  - Exceeding size limits.
  - Missing file in the request.

---

## 3. File Listing

- Verify **listing files** returns **only the files uploaded by the authenticated user**.
- Check that **metadata fields** (filename, size, upload date, MIME type, reference count, download count) are correct.

---

## 4. File Download

- Successfully **download existing files**.
- Verify **download count** increments after a successful download.
- Verify **download access** is denied for:
  - Unauthorized users.
  - Files marked as private.

---

## 5. File Deletion

- Successfully **delete file** when the user is the owner.
- Verify **reference counting** behavior on deletion:
  - File is removed only when `reference_count` reaches zero.
- Verify **deletion is blocked** for unauthorized users.
- Verify **404 error** is returned for non-existing files.

---

## 6. File Search

- Search by **partial filename** returns relevant results.
- Filter files by **MIME type** successfully.
- Filter files by **size range** and **upload date range** correctly.
- Verify **access control** in search results:
  - Only files owned by the authenticated user are shown.

---

## 7. File Sharing

- Generate **public sharing link** successfully.
- Access and **download file via public link** without authentication.
- Verify **invalid or expired links** return appropriate error messages.

---

## 8. Performance & Stability

- Test **concurrent uploads** and downloads to ensure system handles multiple users simultaneously.
- Verify **no data corruption** or crashes occur under heavy load.
- Monitor **response times** and ensure API performance is acceptable.
- Check system stability over extended periods of use.

---

## Notes

- Record any **unexpected behavior** and capture **screenshots/logs** for reference.
- Ensure **all edge cases** are tested, such as very large files, empty files, or unusual MIME types.
- Confirm **security measures** (JWT authentication, public/private file access) are enforced consistently.
