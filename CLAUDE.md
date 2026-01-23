# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Language and Environment Settings

- **Language**: Use Chinese (中文) to answer questions and communicate with the user.
- **Development Platform**: Windows - This is a Windows development environment.

## Build and Development Commands

### Building the Project

```bash
make build
```

This compiles the server and outputs to `./bin/docufiller-update-server.exe`.

### Running the Server

```bash
make run
```

Or directly:

```bash
go run main.go
```

### Installing Dependencies

```bash
make install-deps
```

This runs `go mod tidy` and `go mod download`.

### Cleaning Build Artifacts

```bash
make clean
```

Note: The Makefile uses Unix-style commands (`rm`). On Windows, you need:
- MinGW/MSYS2 make, or
- Git Bash, or
- WSL

### Running Tests

```bash
go test ./...
```

### Code Formatting

```bash
go fmt ./...
gofmt -w .
```

## Project Architecture

This is a **DocuFiller Update Server** - a RESTful API server built with Go that provides automatic update functionality for the DocuFiller WPF application.

### High-Level Architecture

```
main.go (entry point)
    ↓
config.Load() → config.yaml
    ↓
logger.Init() → WQGroup/logger
    ↓
database.NewGORM() → SQLite (versions.db)
    ↓
Gin Router → API Routes
    ↓
Handlers → Services → Models
```

### Key Components

- **main.go**: Application entry point, initializes all subsystems (config, logger, database, routes)
- **internal/config**: YAML-based configuration loading
- **internal/logger**: Structured logging using WQGroup/logger with file rotation
- **internal/database**: GORM + SQLite for version metadata storage
- **internal/models**: GORM models (Version entity)
- **internal/handler**: HTTP request handlers for version management APIs
- **internal/service**: Business logic (version operations, file storage)
- **internal/middleware**: JWT Bearer Token authentication for protected endpoints

### Data Flow

1. **Upload Flow**: Client POSTs to `/api/version/upload` → Handler validates → Service stores file → Database record created
2. **Download Flow**: Client requests `/api/version/latest?channel={channel}` → Handler queries DB → Returns latest version info → Client downloads from `/api/download/{channel}/{version}`

### Directory Structure

- `config.yaml`: Server configuration (port, database path, API token, logging settings)
- `data/versions.db`: SQLite database (auto-created on first run)
- `data/packages/stable/`: Stable release packages
- `data/packages/beta/`: Beta release packages
- `logs/`: Rotated log files
- `tmp/`: Temporary test code and data files
- `docs/plans/`: Project planning documents

## General Development Rules

### Working Directory

- **Temporary files**: All temporary test code, data, and experimental files should be placed in the `tmp/` folder at the project root.
- **Project plans**: All planning documents should be placed in `docs/plans/` directory.

### Logging

- **Go logging**: Use `github.com/WQGroup/logger` for all Go projects. Query usage examples via Context7 if needed.

### Testing for Frontend-Backend Projects

- **Browser testing**: For projects with separated frontend and backend, use the `dev-browser` skill to test frontend-backend communication and interactions.

### Image Processing

- **MCP/Agent image capabilities**: When using screenshot MCP or Agent capabilities, ensure image dimensions are smaller than 1000x1000 pixels before submitting to recognition services.

## Script Development Rules

### Batch Script (.bat) Guidelines

- **No Chinese characters in BAT scripts**: All BAT scripts must use only English characters for comments, variable names, and file paths to avoid encoding issues on Windows.
- **Fix existing scripts first**: When asked to fix a script, modify the existing script rather than creating a new one, unless there's a compelling reason to start fresh.

### Configuration Files

- The main configuration is in `config.yaml`
- Key settings:
  - `server.port`: HTTP port (default: 8080)
  - `api.uploadToken`: Bearer token for upload authentication (MUST be changed in production)
  - `storage.maxFileSize`: Maximum upload size (default: 512MB)
  - `logger.level`: Log verbosity (trace/debug/info/warn/error)
  - `admin.username`: Admin username for management backend (default: "admin")
  - `admin.password`: Admin password for management backend (MUST be changed in production)

## API Endpoints Reference

### Public Endpoints
- `GET /api/health` - Health check
- `GET /api/programs/{programId}/versions/latest?channel={stable|beta}` - Get latest version
- `GET /api/programs/{programId}/versions?channel={stable|beta}` - List all versions
- `GET /api/programs/{programId}/download/{channel}/{version}` - Download package file

### Protected Endpoints (Bearer Token Required)
- `POST /api/programs/{programId}/versions` - Upload new version
- `DELETE /api/programs/{programId}/versions/{version}` - Delete a version

### Admin Backend Endpoints (Session Authentication Required)
- `GET /admin` - Admin dashboard page
- `GET /admin/login` - Login page
- `POST /admin/login` - Login API
- `POST /admin/logout` - Logout API
- `GET /api/admin/programs` - List all programs
- `POST /api/admin/programs` - Create new program
- `GET /api/admin/programs/{programId}` - Get program details
- `DELETE /api/admin/programs/{programId}` - Delete program
- `GET /api/admin/programs/{programId}/versions` - List versions
- `DELETE /api/admin/programs/{programId}/versions/{version}` - Delete version
- `POST /api/admin/programs/{programId}/tokens/regenerate?type={upload|download}` - Regenerate token
- `POST /api/admin/programs/{programId}/encryption/regenerate` - Regenerate encryption key
- `GET /api/admin/programs/{programId}/client/publish` - Download publish client
- `GET /api/admin/programs/{programId}/client/update` - Download update client

## Database

- **Type**: SQLite
- **Location**: `./data/versions.db` (configurable via `config.yaml`)
- **ORM**: GORM with auto-migration enabled
- **Models**: `Program`, `Version`, `Token`, `EncryptionKey` entities
  - Admin credentials are stored in `config.yaml`, not in database
