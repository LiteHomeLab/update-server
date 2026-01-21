# Project Restructure Design

## Overview

Restructure the update-server project to follow standard Go layout, rename programs to match their functionality, and remove deprecated code.

## Current Structure

```
update-server/
├── main.go                     # Server in root
├── cmd/
│   ├── update-client/
│   ├── gen-token/
│   └── test-server/
├── clients/go/
│   ├── admin/                  # Publisher tool (misplaced)
│   ├── client/                 # Shared library
│   └── tool/                   # Deprecated code
└── build-all.bat
```

**Problems:**
1. `update-admin.exe` name doesn't match function (publishing)
2. `clients/go/admin/` is not in standard `cmd/` location
3. `clients/go/tool/` contains deprecated duplicate code
4. `main.go` should be in `cmd/` for consistency

## Target Structure

```
update-server/
├── cmd/
│   ├── update-server/          # Moved from root
│   │   └── main.go
│   ├── update-publisher/       # Renamed from admin
│   │   ├── main.go
│   │   ├── admin.go
│   │   ├── go.mod
│   │   └── update-publisher.usage.txt
│   ├── update-client/          # Unchanged
│   │   ├── main.go
│   │   ├── update-client.config.yaml
│   │   └── go.mod
│   ├── gen-token/              # Unchanged
│   └── test-server/            # Unchanged
│
├── internal/                   # Server internal packages
│   ├── config/
│   ├── handler/
│   ├── middleware/
│   ├── models/
│   └── service/
│
├── clients/go/client/          # Shared library only
│   ├── config.go
│   ├── checker.go
│   ├── downloader.go
│   └── ...
│
├── build-all.bat               # Updated paths
├── go.work                     # Updated module paths
├── config.yaml                 # Server config
└── main.go                     # Will be deleted after move
```

## Module and Output Names

| Source Path | Module Name | Output Binary | Notes |
|-------------|-------------|---------------|-------|
| `cmd/update-server/` | docufiller-update-server | `update-server.exe` | Unchanged |
| `cmd/update-publisher/` | update-publisher | `update-publisher.exe` | Renamed |
| `cmd/update-client/` | update-client | `update-client.exe` | Unchanged |
| `cmd/gen-token/` | gen-token | `gen-token.exe` | Unchanged |
| `cmd/test-server/` | test-server | `test-server.exe` | Unchanged |

## Migration Steps

### 1. File Movements

```bash
# Move server
git mv main.go cmd/update-server/main.go

# Move and rename publisher
git mv clients/go/admin/main.go cmd/update-publisher/main.go
git mv clients/go/admin/admin.go cmd/update-publisher/admin.go
git mv clients/go/admin/go.mod cmd/update-publisher/go.mod
git mv clients/go/admin/publish-client.usage.txt cmd/update-publisher/update-publisher.usage.txt

# Remove deprecated code
git rm -r clients/go/admin/
git rm -r clients/go/tool/
```

### 2. File Modifications

**go.work:**
```go
go 1.24.12

use (
    ./cmd/update-server
    ./cmd/update-publisher
    ./cmd/update-client
    ./internal/...
    ./clients/go/client
)
```

**cmd/update-publisher/main.go:**
- Change `rootCmd.Use` from `"update-admin"` to `"update-publisher"`

**cmd/update-publisher/go.mod:**
- Change module from `github.com/LiteHomeLab/update-admin` to `update-publisher`

**cmd/update-publisher/update-publisher.usage.txt:**
- Update all command examples from `publish-client.exe` to `update-publisher.exe`

**build-all.bat:**
- Update all paths to reflect new structure
- Change `update-admin.exe` to `update-publisher.exe`
- Change `publish-client.usage.txt` to `update-publisher.usage.txt`

### 3. Build Script Update

```batch
[1/6] Creating output directories...
[2/6] Building Update Server...
      cd cmd/update-server
      go build -o bin/update-server.exe .

[3/6] Building Update Publisher...
      cd cmd/update-publisher
      go build -o bin/clients/update-publisher.exe .

[4/6] Building Update Client...
      cd cmd/update-client
      go build -o bin/clients/update-client.exe .

[5/6] Copying executables to deployment directory...

[6/6] Copying configuration files...
```

### 4. Verification

```bash
# Build verification
cd cmd/update-server && go build .
cd cmd/update-publisher && go build .
cd cmd/update-client && go build .

# Full build
build-all.bat

# Test
update-publisher.exe --help
update-client.exe check --config update-client.config.yaml

# Workspace sync
go work sync
```

## Deployment Structure

```
data/clients/
├── update-publisher.exe           # Renamed
├── update-client.exe
├── update-publisher.usage.txt     # Renamed
└── update-client.config.yaml
```

## Important Notes

1. **Import Paths:** The update-client imports `github.com/LiteHomeLab/update-server/clients/go/client` - verify after restructuring

2. **Relative Paths:** Server uses `./config.yaml` and `./data/` - ensure it runs from project root

3. **Git History:** Use `git mv` to preserve file history

4. **Module Sync:** Run `go work sync` after updating go.work

5. **Testing:** Test all executables after migration
