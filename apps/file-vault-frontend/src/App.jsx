import React, { useState, useRef, useCallback, useEffect } from 'react';
import {
  Upload,
  Search,
  Download,
  Share2,
  Trash2,
  File,
  Image,
  FileText,
  User,
  LogOut,
  AlertCircle,
  CheckCircle,
} from 'lucide-react';

import './app.css';

const AUTH_BASE_URL = 'http://localhost:8000';
const FILE_BASE_URL = 'http://localhost:8001';

const resolveApiBaseUrl = (endpoint) => {
  if (
    endpoint.startsWith('/register') ||
    endpoint.startsWith('/login') ||
    endpoint.startsWith('/protected')
  ) {
    return AUTH_BASE_URL;
  }
  return FILE_BASE_URL;
};

const decodeJWT = (token) => {
  if (!token) return null;
  try {
    const base64Payload = token.split('.')[1];
    const payload = atob(base64Payload.replace(/-/g, '+').replace(/_/g, '/'));
    return JSON.parse(decodeURIComponent(escape(payload)));
  } catch {
    return null;
  }
};

const api = {
  getToken: () => localStorage.getItem('jwt_token'),
  setToken: (token) => localStorage.setItem('jwt_token', token),
  removeToken: () => localStorage.removeItem('jwt_token'),

  request: async (endpoint, options = {}) => {
    const token = api.getToken();
    const headers = {
      'Content-Type': 'application/json',
      ...options.headers,
    };
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const config = {
      ...options,
      headers,
    };
    const baseUrl = resolveApiBaseUrl(endpoint);
    const response = await fetch(`${baseUrl}${endpoint}`, config);
    if (response.status === 401) {
      api.removeToken();
      throw new Error('Unauthorized');
    }
    return response;
  },

  uploadFiles: async (files) => {
    const formData = new FormData();
    Array.from(files).forEach((file) => {
      formData.append('files', file);
    });
    const token = api.getToken();
    const decoded = decodeJWT(token);
    const username = decoded?.username || 'unknown';

    const headers = {
      Authorization: `Bearer ${token}`,
      Uploader: username,
    };

    const response = await fetch(`${FILE_BASE_URL}/upload`, {
      method: 'POST',
      headers,
      body: formData,
    });

    if (!response.ok) {
      const errText = await response.text();
      throw new Error(errText || 'Upload failed');
    }
    return response.json();
  },

  getFiles: async () => {
    const response = await api.request('/files');
    if (!response.ok) {
      throw new Error('Failed to fetch files');
    }
    return response.json();
  },

  searchFiles: async (params) => {
    const queryString = new URLSearchParams(params).toString();
    const response = await api.request(`/files/search?${queryString}`);
    if (!response.ok) {
      throw new Error('Search failed');
    }
    return response.json();
  },

  downloadFile: async (fileId) => {
    const token = api.getToken();
    const headers = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const response = await fetch(`${FILE_BASE_URL}/files/${fileId}/download`, {
      method: 'GET',
      headers,
    });
    if (!response.ok) {
      throw new Error('Download failed');
    }
    return response;
  },

  deleteFile: async (fileId) => {
    const token = api.getToken();
    const headers = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const response = await fetch(`${FILE_BASE_URL}/files/${fileId}`, {
      method: 'DELETE',
      headers,
    });
    if (!response.ok) {
      throw new Error('Delete failed');
    }
  },

  shareFile: async (fileId) => {
    const token = api.getToken();
    const headers = {};
    if (token) {
      headers['Authorization'] = `Bearer ${token}`;
    }
    const response = await fetch(`${FILE_BASE_URL}/files/${fileId}/share`, {
      method: 'POST',
      headers,
    });
    if (!response.ok) {
      throw new Error('Share failed');
    }
    return response.json();
  },

  // Admin-only API calls
  getAdminFiles: async () => {
    const response = await api.request('/admin/files');
    if (!response.ok) {
      throw new Error('Failed to fetch admin files');
    }
    return response.json();
  },
};

const loginUser = async (username, password) => {
  const response = await fetch(`${AUTH_BASE_URL}/login`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, password }),
  });

  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || 'Login failed');
  }
  const data = await response.json();
  api.setToken(data.token);
  return data;
};

const signupUser = async (username, email, password) => {
  const response = await fetch(`${AUTH_BASE_URL}/register`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ username, email, password }),
  });
  if (!response.ok) {
    const text = await response.text();
    throw new Error(text || 'Signup failed');
  }
  return response.json();
};

const Notification = ({ message, type, onClose }) => {
  useEffect(() => {
    const timer = setTimeout(onClose, 3000);
    return () => clearTimeout(timer);
  }, [onClose]);
  return (
    <div
      className={`fixed top-4 right-4 p-4 rounded-lg shadow-lg z-50 flex items-center space-x-2 ${
        type === 'success' ? 'bg-green-600 text-white' : 'bg-red-600 text-white'
      }`}
      role="alert"
    >
      {type === 'success' ? (
        <CheckCircle className="w-5 h-5" />
      ) : (
        <AlertCircle className="w-5 h-5" />
      )}
      <span>{message}</span>
    </div>
  );
};


const LoginForm = ({ onLogin, onCancel }) => {
  const [username, setUsername] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setLoading(true);
    setError('');
    try {
      await loginUser(username, password);
      onLogin();
    } catch (err) {
      setError(err.message || 'Login failed. Please check your credentials.');
    }
    setLoading(false);
  };

  return (
    <div className="fixed inset-0 bg-grey bg-opacity-50 flex items-center justify-center z-50">
      <div className="relative z-20 w-11/12 max-w-sm bg-white/95 rounded-2xl shadow-xl border border-gray-200 px-8 py-10">
        <h2 className="text-2xl font-bold mb-6 text-center text-primary-700">
          Login to <span className="text-blue-700">File Vault</span>
        </h2>
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>
          <div className="mb-6">
            <label className="block text-gray-700 text-sm font-bold mb-2">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>
          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">{error}</div>
          )}
          <div className="flex space-x-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-lg font-medium hover:bg-gray-400 transition-colors"
            >
              {loading ? 'Logging in...' : 'Login'}
            </button>
            <button
              type="button"
              onClick={onCancel}
              className="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-lg font-medium hover:bg-gray-400 transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};


const SignupForm = ({ onSignup, onCancel }) => {
  const [username, setUsername] = useState('');
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      await signupUser(username, email, password);
      setLoading(false);
      onSignup();
    } catch (err) {
      setError(err.message);
      setLoading(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-grey bg-opacity-50 flex items-center justify-center z-50">
      <div className="bg-white rounded-lg p-8 max-w-md w-full mx-4 shadow-xl">
        <h2 className="text-2xl font-bold mb-6 text-center text-primary-700">
          Signup for <span className="text-blue-700">File Vault</span>
        </h2>
        <form onSubmit={handleSubmit}>
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">Username</label>
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>
          <div className="mb-4">
            <label className="block text-gray-700 text-sm font-bold mb-2">Email</label>
            <input
              type="email"
              value={email}
              onChange={(e) => setEmail(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>
          <div className="mb-6">
            <label className="block text-gray-700 text-sm font-bold mb-2">Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => setPassword(e.target.value)}
              className="w-full px-3 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
              required
            />
          </div>
          {error && (
            <div className="mb-4 p-3 bg-red-100 border border-red-400 text-red-700 rounded">{error}</div>
          )}
          <div className="flex space-x-4">
            <button
              type="submit"
              disabled={loading}
              className="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-lg font-medium hover:bg-gray-400 transition-colors"
            >
              {loading ? 'Signing up...' : 'Signup'}
            </button>
            <button
              type="button"
              onClick={onCancel}
              className="flex-1 bg-gray-300 text-gray-700 px-4 py-2 rounded-lg font-medium hover:bg-gray-400 transition-colors"
            >
              Cancel
            </button>
          </div>
        </form>
      </div>
    </div>
  );
};


const LandingPage = ({ onNavigate, setUserRole }) => {
  const [showLogin, setShowLogin] = useState(false);
  const [showSignup, setShowSignup] = useState(false);

  const handleLoginSuccess = () => {
    const token = api.getToken();
    const decoded = decodeJWT(token);
    setUserRole(decoded?.role || 'user');
    onNavigate(decoded?.role === 'admin' ? 'adminDashboard' : 'dashboard');
  };

  const handleSignupSuccess = () => {
    alert('Signup successful! Please login to continue.');
  };


return (
  <div className="fixed inset-0 flex items-center justify-center bg-gradient-to-br from-blue-100 via-white to-slate-200 p-4">
    {/* Background blurs */}
    <div className="absolute top-10 left-1/4 w-64 h-64 bg-blue-400 rounded-full opacity-10 blur-3xl"></div>
    <div className="absolute bottom-10 right-1/4 w-48 h-48 bg-indigo-400 rounded-full opacity-10 blur-3xl"></div>

    {/* Centered welcome card */}
    {!showLogin && !showSignup && (
      <div className="relative z-10 text-center w-11/12 max-w-[750px] px-16 py-20 bg-white/90 rounded-3xl shadow-2xl border border-gray-100">
        <div className="mb-8 flex justify-center">
          <div className="w-24 h-24 bg-gradient-to-br from-blue-600 to-indigo-700 rounded-2xl flex items-center justify-center shadow-xl">
            <Upload className="w-12 h-12 text-white" />
          </div>
        </div>
        <h1 className="text-6xl font-extrabold text-gray-900 mb-6 tracking-tight">
          Welcome to <br />
          <span className="block text-blue-700 mt-2">File Vault</span>
        </h1>
        <p className="text-xl text-gray-600 mb-14 leading-relaxed font-medium">
          Your secure cloud storage for all files.
        </p>
        <div className="flex flex-col sm:flex-row gap-4 justify-center items-center">
          <button
            onClick={() => setShowLogin(true)}
            className="px-8 py-4 bg-white text-blue-700 border-2 border-blue-700 rounded-full font-semibold text-lg hover:bg-blue-50 transition-all duration-300 shadow-lg hover:shadow-xl min-w-[140px]"
          >
            Login
          </button>
          <button
            onClick={() => setShowSignup(true)}
            className="px-8 py-4 bg-white text-blue-700 border-2 border-blue-700 rounded-full font-semibold text-lg hover:bg-blue-50 transition-all duration-300 shadow-lg hover:shadow-xl min-w-[140px]"
          >
            Signup
          </button>
        </div>
      </div>
    )}

    {/* Login / Signup forms - smaller width */}
    {showLogin && (
      <div className="relative z-20 w-11/12 max-w-md">
        <LoginForm
          onLogin={handleLoginSuccess}
          onCancel={() => setShowLogin(false)}
        />
      </div>
    )}

    {showSignup && (
      <div className="relative z-20 w-11/12 max-w-md">
        <SignupForm
          onSignup={handleSignupSuccess}
          onCancel={() => setShowSignup(false)}
        />
      </div>
    )}
  </div>
);
};


const Dashboard = ({ onNavigate }) => {
  const [searchTerm, setSearchTerm] = useState('');
  const [files, setFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [dragActive, setDragActive] = useState(false);
  const [notification, setNotification] = useState(null);
  const fileInputRef = useRef(null);

  useEffect(() => {
    loadFiles();
  }, []);

  const loadFiles = async () => {
    setLoading(true);
    try {
      const filesData = await api.getFiles();
      setFiles(filesData || []);
    } catch (error) {
      showNotification('Failed to load files', 'error');
      if (error.message === 'Unauthorized') {
        handleLogout();
      }
    }
    setLoading(false);
  };

  const showNotification = (message, type) => {
    setNotification({ message, type });
  };

  const handleLogout = () => {
    api.removeToken();
    onNavigate('landing');
  };

  const formatFileSize = (bytes) => {
    if (bytes === 0) return '0 Bytes';
    const k = 1024;
    const sizes = ['Bytes', 'KB', 'MB', 'GB'];
    const i = Math.floor(Math.log(bytes) / Math.log(k));
    return parseFloat((bytes / Math.pow(k, i)).toFixed(1)) + ' ' + sizes[i];
  };

  const getFileIcon = (filename, mimeType) => {
    if (mimeType?.startsWith('image/')) return Image;
    if (mimeType === 'application/pdf') return FileText;
    if (mimeType?.includes('spreadsheet') || filename?.endsWith('.xlsx') || filename?.endsWith('.xls'))
      return FileText;
    return File;
  };

  const getFileTypeColor = (mimeType) => {
    if (mimeType?.startsWith('image/')) return 'bg-blue-200 text-blue-700';
    if (mimeType === 'application/pdf') return 'bg-red-200 text-red-700';
    if (mimeType?.includes('spreadsheet')) return 'bg-green-200 text-green-700';
    return 'bg-gray-200 text-gray-700';
  };

  const formatDate = (dateString) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffTime = now - date;
    const diffDays = Math.floor(diffTime / (1000 * 60 * 60 * 24));
    if (diffDays === 0) return 'Today';
    if (diffDays === 1) return 'Yesterday';
    if (diffDays < 7) return `${diffDays} days ago`;
    return date.toLocaleDateString();
  };

  const handleDrag = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    if (e.type === 'dragenter' || e.type === 'dragover') {
      setDragActive(true);
    } else if (e.type === 'dragleave') {
      setDragActive(false);
    }
  }, []);

  const handleDrop = useCallback((e) => {
    e.preventDefault();
    e.stopPropagation();
    setDragActive(false);
    if (e.dataTransfer.files && e.dataTransfer.files[0]) {
      handleFileUpload(e.dataTransfer.files);
    }
  }, []);

  const handleFileUpload = async (fileList) => {
    setLoading(true);
    try {
      const uploadedFiles = await api.uploadFiles(fileList);
      const existingFileIds = new Set(files.map((f) => f.id));
      const uniqueNewFiles = uploadedFiles.filter((f) => !existingFileIds.has(f.id));
      setFiles((prev) => [...uniqueNewFiles, ...prev]);
      showNotification(`Successfully uploaded ${uniqueNewFiles.length} file(s)`, 'success');
    } catch (error) {
      showNotification(`Upload failed: ${error.message}`, 'error');
    }
    setLoading(false);
  };

  const handleFileInputChange = (e) => {
    if (e.target.files && e.target.files.length > 0) {
      handleFileUpload(e.target.files);
    }
  };

  const handleDownload = async (file) => {
    try {
      const response = await api.downloadFile(file.id);
      const blob = await response.blob();
      const url = window.URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = file.filename;
      document.body.appendChild(a);
      a.click();
      window.URL.revokeObjectURL(url);
      document.body.removeChild(a);
      showNotification('Download started', 'success');
    } catch (error) {
      showNotification('Download failed', 'error');
    }
  };

  const handleShare = async (file) => {
    try {
      const shareData = await api.shareFile(file.id);
      navigator.clipboard.writeText(shareData.url);
      showNotification('Share link copied to clipboard', 'success');
    } catch (error) {
      showNotification('Share failed', 'error');
    }
  };

  const handleDelete = async (file) => {
    if (window.confirm(`Are you sure you want to delete ${file.filename}?`)) {
      try {
        await api.deleteFile(file.id);
        setFiles((prev) => prev.filter((f) => f.id !== file.id));
        showNotification('File deleted successfully', 'success');
      } catch (error) {
        showNotification('Delete failed', 'error');
      }
    }
  };

  const handleSearch = async () => {
    if (!searchTerm.trim()) {
      loadFiles();
      return;
    }
    setLoading(true);
    try {
      const searchResults = await api.searchFiles({ filename: searchTerm });
      setFiles(searchResults || []);
    } catch (error) {
      showNotification('Search failed', 'error');
    }
    setLoading(false);
  };

  const filteredFiles = files.filter((file) =>
    file.filename.toLowerCase().includes(searchTerm.toLowerCase())
  );

  return (
    <div className="h-screen w-screen flex flex-col bg-gradient-to-br from-blue-50 to-gray-100 overflow-hidden">
      {/* Header */}
      <header className="w-full bg-blue-700 text-white px-6 py-4 shadow-lg">
        <div className="flex items-center justify-between">
          <div className="flex items-center space-x-3">
            <div className="w-8 h-8 bg-white bg-opacity-20 rounded flex items-center justify-center">
              <Upload className="w-5 h-5" />
            </div>
            <h1 className="text-xl font-bold tracking-wide">File Vault</h1>
          </div>
          <div className="flex items-center space-x-4">
            <div className="flex items-center space-x-2">
              <User className="w-5 h-5" />
              <span>User</span>
            </div>
            <button
              onClick={handleLogout}
              className="p-2 !bg-blue-700 rounded"
            >
              <LogOut className="w-5 h-5" />
            </button>
          </div>
        </div>
      </header>

      {/* Main Content */}
      <div className="flex-1 w-full flex">
        {/* Upload Section */}
        <div className="w-1/3 p-6 flex flex-col justify-center">
          <div
            className={`border-2 border-dashed rounded-lg p-8 text-center transition-colors ${
              dragActive ? 'border-blue-500 bg-blue-50' : 'border-gray-300 hover:border-gray-400'
            }`}
            onDragEnter={handleDrag}
            onDragLeave={handleDrag}
            onDragOver={handleDrag}
            onDrop={handleDrop}
          >
            <div className="mb-4">
              <Upload className="w-12 h-12 text-blue-500 mx-auto mb-4" />
              <p className="text-gray-600 mb-4 text-lg">Drag & drop files here, or...</p>
              <button
                onClick={() => fileInputRef.current?.click()}
                disabled={loading}
                className="bg-blue-700 text-grey px-6 py-3 rounded-lg font-bold hover:bg-blue-800 transition-colors disabled:opacity-50"
              >
                {loading ? 'Uploading...' : 'Upload File'}
              </button>
            </div>
            <input
              ref={fileInputRef}
              type="file"
              multiple
              onChange={handleFileInputChange}
              className="hidden"
            />
          </div>
        </div>

        {/* Files Section */}
        <div className="flex-1 p-6 overflow-auto">
          <div className="bg-white rounded-lg shadow-md h-full flex flex-col">
            {/* Files Header */}
            <div className="p-4 border-b border-gray-200">
              <div className="flex items-center justify-between mb-4">
                <h2 className="text-lg font-bold text-gray-800">My Files</h2>
                <div className="flex items-center space-x-2">
                  <button
                    onClick={loadFiles}
                    className="p-2 hover:bg-gray-100 rounded text-blue-700"
                    title="Refresh"
                  >
                    <Download className="w-4 h-4" />
                  </button>
                </div>
              </div>
              
              {/* Search and Filters */}
              <div className="flex items-center space-x-4">
                <div className="relative flex-1">
                  <Search className="w-4 h-4 text-gray-400 absolute left-3 top-1/2 transform -translate-y-1/2" />
                  <input
                    type="text"
                    placeholder="Search files..."
                    value={searchTerm}
                    onChange={(e) => setSearchTerm(e.target.value)}
                    onKeyDown={(e) => e.key === 'Enter' && handleSearch()}
                    className="w-full pl-10 pr-4 py-2 border border-gray-300 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
                  />
                </div>
                <div className="flex space-x-2">
                  <button className="px-3 py-2 text-sm bg-gray-100 text-gray-700 rounded shadow">Name</button>
                  <button className="px-3 py-2 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200">
                    Type
                  </button>
                  <button className="px-3 py-2 text-sm bg-gray-100 text-gray-700 rounded hover:bg-gray-200">
                    Date
                  </button>
                </div>
              </div>
            </div>

            {/* Files List */}
            <div className="p-4 flex-1 overflow-auto">
              {loading ? (
                <div className="text-center py-8">
                  <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-blue-700 mx-auto"></div>
                  <p className="text-blue-700 mt-2">Loading files...</p>
                </div>
              ) : filteredFiles.length === 0 ? (
                <div className="text-center py-8 text-gray-600">
                  {searchTerm ? 'No files found matching your search.' : 'No files uploaded yet.'}
                </div>
              ) : (
                <div className="space-y-2">
                  {filteredFiles.map((file) => {
                    const IconComponent = getFileIcon(file.filename, file.mime_type);
                    const colorClasses = getFileTypeColor(file.mime_type);
                    return (
                      <div
                        key={file.id}
                        className="flex items-center justify-between p-3 hover:bg-gray-50 rounded-lg group"
                      >
                        <div className="flex items-center space-x-3">
                          <div className={`w-10 h-10 rounded-lg flex items-center justify-center ${colorClasses}`}>
                            <IconComponent className="w-5 h-5" />
                          </div>
                          <div>
                            <p className="font-medium text-gray-900">{file.filename}</p>
                            <p className="text-sm text-gray-500">
                              Uploaded {formatDate(file.upload_date)} • Downloads: {file.download_count || 0}
                            </p>
                          </div>
                        </div>
                        <div className="flex items-center space-x-4">
                          <span className="text-sm text-gray-500 font-medium">{formatFileSize(file.size)}</span>
                          <div className="flex items-center space-x-1 opacity-0 group-hover:opacity-100 transition-opacity">
                            <button
                              onClick={() => handleDownload(file)}
                              className="p-1 hover:bg-gray-200 rounded"
                              title="Download"
                            >
                              <Download className="w-4 h-4 text-blue-700" />
                            </button>
                            <button
                              onClick={() => handleShare(file)}
                              className="p-1 hover:bg-gray-200 rounded"
                              title="Share"
                            >
                              <Share2 className="w-4 h-4 text-blue-700" />
                            </button>
                            <button
                              onClick={() => handleDelete(file)}
                              className="p-1 hover:bg-gray-200 rounded"
                              title="Delete"
                            >
                              <Trash2 className="w-4 h-4 text-red-700" />
                            </button>
                          </div>
                        </div>
                      </div>
                    );
                  })}
                </div>
              )}
            </div>
          </div>
        </div>
      </div>

      {/* Notification */}
      {notification && (
        <Notification 
          message={notification.message} 
          type={notification.type} 
          onClose={() => setNotification(null)} 
        />
      )}
    </div>
  );
};

const AdminDashboard = ({ onNavigate }) => {
  const [adminFiles, setAdminFiles] = useState([]);
  const [loading, setLoading] = useState(false);
  const [notification, setNotification] = useState(null);

  useEffect(() => {
    loadAdminFiles();
  }, []);

  const loadAdminFiles = async () => {
    setLoading(true);
    try {
      const data = await api.getAdminFiles();
      setAdminFiles(data);
    } catch (error) {
      setNotification({ message: error.message, type: 'error' });
    }
    setLoading(false);
  };

  const handleLogout = () => {
    api.removeToken();
    onNavigate('landing');
  };

  const showNotification = (message, type) => {
    setNotification({ message, type });
  };

  return (
    <div className="min-h-screen flex flex-col bg-gray-100">
      <header className="bg-gray-900 text-white p-4 flex justify-between items-center shadow">
        <h1 className="text-2xl font-bold">Admin Panel</h1>
        <button
          className="bg-red-600 px-4 py-2 rounded hover:bg-red-700"
          onClick={handleLogout}
        >
          Logout
        </button>
      </header>
      <main className="p-4 flex-1 overflow-auto">
        <h2 className="text-lg font-semibold mb-4">All Files</h2>
        {loading ? (
          <p>Loading files...</p>
        ) : !adminFiles.length ? (
          <p>No files found.</p>
        ) : (
          <table className="w-full border-collapse border border-gray-300">
            <thead>
              <tr>
                <th className="border border-gray-300 p-2">Filename</th>
                <th className="border border-gray-300 p-2">Uploader</th>
                <th className="border border-gray-300 p-2">Size (bytes)</th>
                <th className="border border-gray-300 p-2">Downloads</th>
              </tr>
            </thead>
            <tbody>
              {adminFiles.map((file) => (
                <tr key={file.id} className="border border-gray-300">
                  <td className="border border-gray-300 p-2">{file.filename}</td>
                  <td className="border border-gray-300 p-2">{file.uploader}</td>
                  <td className="border border-gray-300 p-2">{file.size}</td>
                  <td className="border border-gray-300 p-2">{file.download_count}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
        {notification && (
          <Notification
            message={notification.message}
            type={notification.type}
            onClose={() => setNotification(null)}
          />
        )}
      </main>
    </div>
  );
};



const App = () => {
 const [currentPage, setCurrentPage] = useState('landing');
  const [userRole, setUserRole] = useState(null);

  useEffect(() => {
    const token = api.getToken();
    if (token) {
      const decoded = decodeJWT(token);
      console.log(decoded.role); 
      setUserRole(decoded?.role || 'user');
      // console.log(role);
      setCurrentPage(decoded?.role === 'admin' ? 'adminDashboard' : 'dashboard');
    }
  }, []);

  return (
    <>
      {currentPage === 'landing' && (
        <LandingPage onNavigate={setCurrentPage} setUserRole={setUserRole} />
      )}
      {currentPage === 'dashboard' && <Dashboard onNavigate={setCurrentPage} />}
      {currentPage === 'adminDashboard' && <AdminDashboard onNavigate={setCurrentPage} />}
    </>
  );
};

export default App;
