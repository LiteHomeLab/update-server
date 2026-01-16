# DocuFiller Update Server API Documentation

## Overview

The DocuFiller Update Server provides a RESTful API for managing and distributing application updates. The server supports multiple programs, update channels (stable/beta), and token-based authentication.

**Base URL**: `http://localhost:8080/api`

**Authentication**: Bearer Token (required for upload and download endpoints)

---

## Response Format

All API responses follow this format:

**Success Response**:
```json
{
  "data": { ... }
}
```

**Error Response**:
```json
{
  "error": "Error message description"
}
```

---

## Public Endpoints

### Health Check

Check if the server is running.

**Endpoint**: `GET /api/health`

**Authentication**: None

**Response**:
```json
{
  "status": "ok"
}
```

**Example**:
```bash
curl http://localhost:8080/api/health
```

---

### Get Latest Version

Retrieve the latest version information for a specific program and channel.

**Endpoint**: `GET /api/programs/{programId}/versions/latest`

**Authentication**: None

**Path Parameters**:
- `programId` (string) - The program identifier

**Query Parameters**:
- `channel` (string, optional) - Update channel: `stable` or `beta`. Default: `stable`

**Response**:
```json
{
  "id": 1,
  "programId": "docufiller",
  "version": "1.2.0",
  "channel": "stable",
  "fileName": "DocuFiller-1.2.0.zip",
  "filePath": "./data/packages/docufiller/stable/1.2.0",
  "fileSize": 52428800,
  "fileHash": "a1b2c3d4e5f6...",
  "releaseNotes": "New features and bug fixes",
  "publishDate": "2024-01-15T10:30:00Z",
  "downloadCount": 1250,
  "mandatory": false
}
```

**Error Responses**:
- `404 Not Found` - No version found for the specified program and channel
- `500 Internal Server Error` - Server error

**Example**:
```bash
curl http://localhost:8080/api/programs/docufiller/versions/latest?channel=stable
```

---

### Get Version List

Retrieve all versions for a specific program, optionally filtered by channel.

**Endpoint**: `GET /api/programs/{programId}/versions`

**Authentication**: None

**Path Parameters**:
- `programId` (string) - The program identifier

**Query Parameters**:
- `channel` (string, optional) - Filter by channel: `stable` or `beta`. If omitted, returns all channels

**Response**:
```json
[
  {
    "id": 3,
    "programId": "docufiller",
    "version": "1.2.0",
    "channel": "stable",
    "fileName": "DocuFiller-1.2.0.zip",
    "fileSize": 52428800,
    "fileHash": "a1b2c3d4e5f6...",
    "releaseNotes": "New features and bug fixes",
    "publishDate": "2024-01-15T10:30:00Z",
    "downloadCount": 1250,
    "mandatory": false
  },
  {
    "id": 2,
    "programId": "docufiller",
    "version": "1.1.0",
    "channel": "stable",
    "fileName": "DocuFiller-1.1.0.zip",
    "fileSize": 51200000,
    "fileHash": "f6e5d4c3b2a1...",
    "releaseNotes": "Bug fixes",
    "publishDate": "2024-01-01T09:00:00Z",
    "downloadCount": 3400,
    "mandatory": false
  }
]
```

**Error Responses**:
- `500 Internal Server Error` - Server error

**Example**:
```bash
# Get all versions
curl http://localhost:8080/api/programs/docufiller/versions

# Get only stable versions
curl http://localhost:8080/api/programs/docufiller/versions?channel=stable
```

---

### Get Version Detail

Retrieve detailed information about a specific version.

**Endpoint**: `GET /api/programs/{programId}/versions/{channel}/{version}`

**Authentication**: None

**Path Parameters**:
- `programId` (string) - The program identifier
- `channel` (string) - Update channel: `stable` or `beta`
- `version` (string) - Version number (e.g., `1.2.0`)

**Response**:
```json
{
  "id": 3,
  "programId": "docufiller",
  "version": "1.2.0",
  "channel": "stable",
  "fileName": "DocuFiller-1.2.0.zip",
  "filePath": "./data/packages/docufiller/stable/1.2.0",
  "fileSize": 52428800,
  "fileHash": "a1b2c3d4e5f6...",
  "releaseNotes": "New features and bug fixes",
  "publishDate": "2024-01-15T10:30:00Z",
  "downloadCount": 1250,
  "mandatory": false
}
```

**Error Responses**:
- `404 Not Found` - Version not found
- `500 Internal Server Error` - Server error

**Example**:
```bash
curl http://localhost:8080/api/programs/docufiller/versions/stable/1.2.0
```

---

## Protected Endpoints

### Upload New Version

Upload a new version package for a program.

**Endpoint**: `POST /api/programs/{programId}/versions`

**Authentication**: Required (Bearer Token with upload permission)

**Path Parameters**:
- `programId` (string) - The program identifier

**Request**: `multipart/form-data`

**Form Fields**:
- `file` (file, required) - The update package file (zip)
- `channel` (string, required) - Update channel: `stable` or `beta`
- `version` (string, required) - Version number (e.g., `1.2.0`)
- `notes` (string, optional) - Release notes
- `mandatory` (boolean, optional) - Whether this update is mandatory. Default: `false`

**Headers**:
```
Authorization: Bearer {upload_token}
Content-Type: multipart/form-data
```

**Response**:
```json
{
  "message": "Version uploaded successfully",
  "version": {
    "id": 4,
    "programId": "docufiller",
    "version": "1.3.0",
    "channel": "stable",
    "fileName": "DocuFiller-1.3.0.zip",
    "filePath": "./data/packages/docufiller/stable/1.3.0",
    "fileSize": 53000000,
    "fileHash": "x9y8z7w6v5u4...",
    "releaseNotes": "Major new release",
    "publishDate": "2024-01-20T14:00:00Z",
    "downloadCount": 0,
    "mandatory": false
  }
}
```

**Error Responses**:
- `400 Bad Request` - Missing required fields or file
- `401 Unauthorized` - Missing or invalid authorization token
- `403 Forbidden` - Insufficient permissions for the program
- `500 Internal Server Error` - Failed to save file or create version record

**Example**:
```bash
curl -X POST http://localhost:8080/api/programs/docufiller/versions \
  -H "Authorization: Bearer your-upload-token" \
  -F "file=@/path/to/DocuFiller-1.3.0.zip" \
  -F "channel=stable" \
  -F "version=1.3.0" \
  -F "notes=Major new release" \
  -F "mandatory=false"
```

---

### Delete Version

Delete a specific version and its associated file.

**Endpoint**: `DELETE /api/programs/{programId}/versions/{channel}/{version}`

**Authentication**: Required (Bearer Token with upload permission)

**Path Parameters**:
- `programId` (string) - The program identifier
- `channel` (string) - Update channel: `stable` or `beta`
- `version` (string) - Version number (e.g., `1.2.0`)

**Headers**:
```
Authorization: Bearer {upload_token}
```

**Response**:
```json
{
  "message": "Version deleted successfully"
}
```

**Error Responses**:
- `401 Unauthorized` - Missing or invalid authorization token
- `403 Forbidden` - Insufficient permissions
- `500 Internal Server Error` - Failed to delete version

**Example**:
```bash
curl -X DELETE "http://localhost:8080/api/programs/docufiller/versions/stable/1.2.0" \
  -H "Authorization: Bearer your-upload-token"
```

---

### Download Version Package

Download the update package file for a specific version.

**Endpoint**: `GET /api/programs/{programId}/download/{channel}/{version}`

**Authentication**: Required (Bearer Token with download permission)

**Path Parameters**:
- `programId` (string) - The program identifier
- `channel` (string) - Update channel
- `version` (string) - Version number

**Headers**:
```
Authorization: Bearer {download_token}
```

**Response**: Binary file stream

**Error Responses**:
- `401 Unauthorized` - Missing or invalid authorization token
- `403 Forbidden` - Insufficient permissions for the program
- `404 Not Found` - Version or file not found

**Example**:
```bash
curl -O http://localhost:8080/api/programs/docufiller/download/stable/1.2.0 \
  -H "Authorization: Bearer your-download-token"
```

---

## Program Management Endpoints

### Create Program

Create a new program in the system.

**Endpoint**: `POST /api/programs`

**Authentication**: Required (Bearer Token with upload permission)

**Request Body**:
```json
{
  "id": "new-app",
  "name": "New Application",
  "description": "Description of the application"
}
```

**Response**:
```json
{
  "id": "new-app",
  "name": "New Application",
  "description": "Description of the application",
  "createdAt": "2024-01-20T14:00:00Z"
}
```

---

### List Programs

Get a list of all programs.

**Endpoint**: `GET /api/programs`

**Authentication**: Required (Bearer Token with upload permission)

**Response**:
```json
[
  {
    "id": "docufiller",
    "name": "DocuFiller",
    "description": "Document filler application",
    "createdAt": "2024-01-01T00:00:00Z"
  }
]
```

---

### Get Program

Get details of a specific program.

**Endpoint**: `GET /api/programs/{programId}`

**Authentication**: Required (Bearer Token with upload permission)

**Response**:
```json
{
  "id": "docufiller",
  "name": "DocuFiller",
  "description": "Document filler application",
  "createdAt": "2024-01-01T00:00:00Z"
}
```

---

## Legacy Endpoints (Deprecated)

For backward compatibility, the following legacy endpoints are still supported. They redirect to the new endpoints with `programId` set to `docufiller`.

**Warning**: These endpoints will respond with a deprecation warning header. Use the new endpoints instead.

| Legacy Endpoint | New Endpoint |
|----------------|--------------|
| `GET /api/version/latest?channel=stable` | `GET /api/programs/docufiller/versions/latest?channel=stable` |
| `GET /api/version/list?channel=stable` | `GET /api/programs/docufiller/versions?channel=stable` |
| `POST /api/version/upload` | `POST /api/programs/docufiller/versions` |
| `GET /api/download/{channel}/{version}` | `GET /api/programs/docufiller/download/{channel}/{version}` |

---

## Token Types and Permissions

The server uses three types of tokens with different permission levels:

### Admin Token
- Full access to all programs
- Can upload and delete versions
- Can create and manage programs
- Format: `admin` token type

### Upload Token
- Program-specific or global
- Can upload new versions
- Can delete versions
- Can view programs
- Format: `upload` token type with optional `programId`

### Download Token
- Program-specific
- Can download version packages
- Can view version information
- Format: `download` token type with required `programId`

---

## Error Codes

| HTTP Status | Error Message | Description |
|-------------|---------------|-------------|
| 400 | Bad Request | Invalid request parameters or missing required fields |
| 401 | Unauthorized | Missing or invalid authorization token |
| 403 | Forbidden | Insufficient permissions for the requested operation |
| 404 | Not Found | Requested resource (version, program, file) not found |
| 500 | Internal Server Error | Server-side error occurred |

---

## File Upload Limits

- Maximum file size: 512 MB (configurable via `storage.maxFileSize`)
- Supported file formats: ZIP archives
- File hash: SHA-256 calculated automatically on upload

---

## Rate Limiting

Currently, rate limiting is not implemented. Consider implementing rate limiting for production deployments.

---

## CORS

CORS is enabled by default. Configure via `api.corsEnable` in `config.yaml`.

---

## Examples with PowerShell

```powershell
# Get latest version
Invoke-RestMethod -Uri "http://localhost:8080/api/programs/docufiller/versions/latest?channel=stable" -Method GET

# Upload new version
$headers = @{
    "Authorization" = "Bearer your-upload-token"
}
$fields = @{
    "channel" = "stable"
    "version" = "1.3.0"
    "notes" = "Test release"
    "mandatory" = "false"
}
$file = @{
    "file" = "C:\path\to\DocuFiller-1.3.0.zip"
}
Invoke-RestMethod -Uri "http://localhost:8080/api/programs/docufiller/versions" -Method POST -Headers $headers -Form $fields -Files $file

# Download version
$output = "C:\Downloads\DocuFiller-1.2.0.zip"
Invoke-WebRequest -Uri "http://localhost:8080/api/programs/docufiller/download/stable/1.2.0" -Headers $headers -OutFile $output
```

---

## Version Comparison

Version numbers follow semantic versioning: `MAJOR.MINOR.PATCH`

When querying for the "latest" version, the server compares versions:
1. First by MAJOR (highest)
2. Then by MINOR (highest)
3. Then by PATCH (highest)

Example: `2.0.0` > `1.9.9` > `1.5.0` > `1.0.10`
