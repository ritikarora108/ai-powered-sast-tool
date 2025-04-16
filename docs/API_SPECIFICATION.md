# AI-Powered SAST Tool API Specification

This document provides a comprehensive overview of the API endpoints available in the AI-Powered SAST Tool.

## Base URLs

- **Production**: `https://api.example.com`
- **Development**: `http://localhost:8080`

## Authentication

Most API endpoints require authentication. The API uses JWT-based authentication.

### Authentication Flow

1. Users authenticate via Google OAuth
2. The system exchanges the Google token for a custom JWT token
3. This JWT token must be included in the `Authorization` header for authenticated requests

**Example**:

```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

## API Endpoints

### Health Check

#### GET /health

Check if the API is operational.

**Response**: 200 OK

```
OK
```

### Authentication

#### GET /auth/google

Initiates Google OAuth flow.

**Response**: Redirects to Google authentication page

#### GET /auth/google/callback

Callback URL for Google OAuth.

**Response**: Redirects to frontend with authentication token

#### POST /auth/token

Exchange a Google token for a JWT token.

**Request Body**:

```json
{
  "token": "google-oauth-token"
}
```

**Response**: 200 OK

```json
{
  "token": "jwt-token",
  "user": {
    "id": "user-id",
    "name": "User Name",
    "email": "user@example.com",
    "picture": "https://example.com/avatar.jpg"
  }
}
```

### Public Scanning

These endpoints don't require authentication and can be used for scanning public repositories.

#### POST /scan

Scan a public GitHub repository.

**Request Body**:

```json
{
  "repo_url": "https://github.com/owner/repo"
}
```

**Response**: 202 Accepted

```json
{
  "id": "scan-id",
  "status": "scan_initiated",
  "run_id": "workflow-run-id"
}
```

#### GET /scan/{id}/status

Get the status of a scan.

**Parameters**:

- `id`: Scan ID

**Response**: 200 OK

```json
{
  "scan_id": "scan-id",
  "status": "in_progress|completed|failed|canceled|timed_out",
  "progress": 75
}
```

#### GET /scan/{id}/results

Get the results of a completed scan.

**Parameters**:

- `id`: Scan ID

**Response**: 200 OK

```json
{
  "scan_id": "scan-id",
  "status": "completed",
  "vulnerabilities_count": 5,
  "vulnerabilities_by_category": {
    "Injection": [
      {
        "id": "vuln-id",
        "type": "SQL Injection",
        "severity": "high",
        "file_path": "src/database.js",
        "line_number": 42,
        "description": "Potential SQL injection vulnerability",
        "remediation": "Use parameterized queries"
      }
    ],
    "Broken Access Control": []
  }
}
```

#### GET /scan/{id}/debug

Get detailed debug information about a scan workflow.

**Parameters**:

- `id`: Scan ID

**Response**: 200 OK

```json
{
  "workflow_id": "workflow-id",
  "status": "COMPLETED",
  "execution_time": "120s",
  "activities": [
    {
      "name": "CloneRepository",
      "status": "completed",
      "start_time": "2023-07-01T12:00:00Z",
      "end_time": "2023-07-01T12:00:10Z"
    }
  ]
}
```

### Repositories (Authenticated)

These endpoints require authentication and are available under two routes:

- `/repositories` (direct routes)
- `/api/repositories` (API prefix routes)

Both routes provide the same functionality but are organized differently.

#### POST /repositories

Create a new repository.

**Request Body**:

```json
{
  "repo_url": "https://github.com/owner/repo"
}
```

**Response**: 201 Created

```json
{
  "id": "repo-id",
  "owner": "owner",
  "name": "repo",
  "url": "https://github.com/owner/repo",
  "clone_url": "https://github.com/owner/repo.git",
  "created_at": "2023-07-01T12:00:00Z",
  "updated_at": "2023-07-01T12:00:00Z",
  "last_scan_at": null,
  "status": "created"
}
```

#### GET /repositories

List repositories.

**Response**: 200 OK

```json
[
  {
    "id": "repo-id",
    "owner": "owner",
    "name": "repo",
    "url": "https://github.com/owner/repo",
    "clone_url": "https://github.com/owner/repo.git",
    "created_at": "2023-07-01T12:00:00Z",
    "updated_at": "2023-07-01T12:00:00Z",
    "last_scan_at": "2023-07-01T13:00:00Z",
    "status": "scanned"
  }
]
```

#### GET /repositories/{id}

Get details about a specific repository.

**Parameters**:

- `id`: Repository ID

**Response**: 200 OK

```json
{
  "id": "repo-id",
  "owner": "owner",
  "name": "repo",
  "url": "https://github.com/owner/repo",
  "clone_url": "https://github.com/owner/repo.git",
  "created_at": "2023-07-01T12:00:00Z",
  "updated_at": "2023-07-01T12:00:00Z",
  "last_scan_at": "2023-07-01T13:00:00Z",
  "status": "scanned"
}
```

#### POST /repositories/{id}/scan

Initiate a scan for a repository.

**Parameters**:

- `id`: Repository ID

**Response**: 202 Accepted

```json
{
  "id": "repo-id",
  "status": "scan_initiated",
  "run_id": "workflow-run-id"
}
```

#### GET /repositories/{id}/vulnerabilities

Get vulnerabilities for a specific repository.

**Parameters**:

- `id`: Repository ID

**Response**: 200 OK

```json
[
  {
    "id": "vuln-id",
    "repository_id": "repo-id",
    "type": "SQL Injection",
    "category": "Injection",
    "severity": "high",
    "file_path": "src/database.js",
    "line_number": 42,
    "code_snippet": "db.query('SELECT * FROM users WHERE id = ' + userId)",
    "description": "Potential SQL injection vulnerability",
    "remediation": "Use parameterized queries instead of string concatenation",
    "confidence": 0.95,
    "created_at": "2023-07-01T13:00:00Z"
  }
]
```

### User Information (Authenticated)

#### GET /api/users/me

Get the current user's profile information.

**Response**: 200 OK

```json
{
  "id": "user-id",
  "name": "User Name",
  "email": "user@example.com",
  "picture": "https://example.com/avatar.jpg",
  "created_at": "2023-06-01T12:00:00Z"
}
```

## Error Responses

All endpoints may return the following error responses:

### 400 Bad Request

Returned when the request is malformed or missing required parameters.

```json
{
  "error": "Invalid request: Missing required parameter 'repo_url'"
}
```

### 401 Unauthorized

Returned when authentication is required but not provided or invalid.

```json
{
  "error": "Unauthorized"
}
```

### 403 Forbidden

Returned when the authenticated user doesn't have permission to access the resource.

```json
{
  "error": "Forbidden: Insufficient permissions"
}
```

### 404 Not Found

Returned when the requested resource doesn't exist.

```json
{
  "error": "Resource not found"
}
```

### 500 Internal Server Error

Returned when an unexpected error occurs on the server.

```json
{
  "error": "Internal server error"
}
```

## Rate Limiting

The API implements rate limiting to prevent abuse. Clients may receive a `429 Too Many Requests` response if they exceed the limits.

**Headers**:

```
X-RateLimit-Limit: 100
X-RateLimit-Remaining: 99
X-RateLimit-Reset: 1625097600
```

## Webhook Notifications (Future Enhancement)

Future versions will support webhook notifications for scan completion and other events.


