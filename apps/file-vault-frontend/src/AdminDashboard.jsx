import React, { useState, useEffect } from "react";
import { Download, Trash2, Share2, LogOut } from "lucide-react";
import { api, decodeJWT } from "./api"; // import your api and decode helper
import { Notification } from "./Notification"; // import your Notification component

export function AdminDashboard({ onNavigate }) {
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState(null);

  // Load all files on mount
  useEffect(() => {
    loadFiles();
  }, []);

  async function loadFiles() {
    setLoading(true);
    try {
      const response = await api.request("/admin/files");
      if (!response.ok) throw new Error("Failed to load files");
      const data = await response.json();
      setFiles(data);
    } catch (err) {
      setNotification({ type: "error", message: err.message });
    }
    setLoading(false);
  }

  function showNotification(type, msg) {
    setNotification({ type, message: msg });
  }

  async function handleDelete(file) {
    if (!window.confirm(`Delete file ${file.filename}?`)) return;
    try {
      await api.deleteFile(file.id);
      setFiles((f) => f.filter((file) => file.id !== file.id));
      showNotification("success", "File deleted");
    } catch (err) {
      showNotification("error", "Delete failed");
    }
  }

  async function handleShare(file) {
    try {
      const shareData = await api.shareFile(file.id);
      navigator.clipboard.writeText(shareData.url);
      showNotification("success", "Share link copied to clipboard");
    } catch (err) {
      showNotification("error", "Share failed");
    }
  }

  function handleLogout() {
    api.removeToken();
    onNavigate("landing");
  }

  return (
    <div style={{ padding: 20 }}>
      <header style={{ display: "flex", justifyContent: "space-between", marginBottom: 20 }}>
        <h1>Admin Dashboard</h1>
        <button onClick={handleLogout}>Logout</button>
      </header>

      {loading && <p>Loading files…</p>}

      {!loading && files.length === 0 && <p>No files found.</p>}

      {!loading && files.length > 0 && (
        <table border="1" cellPadding="8" cellSpacing="0" width="100%">
          <thead>
            <tr>
              <th>Filename</th>
              <th>Uploader</th>
              <th>Size</th>
              <th>Downloads</th>
              <th>Actions</th>
            </tr>
          </thead>
          <tbody>
            {files.map((file) => (
              <tr key={file.id}>
                <td>{file.filename}</td>
                <td>{file.uploader}</td>
                <td>{(file.size / 1024).toFixed(1)} KB</td>
                <td>{file.download_count}</td>
                <td>
                  <button onClick={() => handleShare(file)} title="Share">
                    <Share2 size={16} />
                  </button>
                  <button onClick={() => handleDelete(file)} title="Delete" style={{ color: "red" }}>
                    <Trash2 size={16} />
                  </button>
                </td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      {notification && (
        <Notification type={notification.type} message={notification.message} onClose={() => setNotification(null)} />
      )}
    </div>
  );
}
