# Data Migration Tools

This directory contains tools to migrate existing data to the new multi-program structure.

## Overview

The migration tools perform the following tasks:

1. **Database Migration** (`migrate.go`)
   - Creates the `docufiller` program record
   - Updates existing version records to associate them with the `docufiller` program
   - Generates initial upload and download tokens

2. **Storage Migration** (`migrate-storage.sh` / `migrate-storage.bat`)
   - Reorganizes package files from the old structure to the new multi-program structure
   - Old: `./data/packages/stable/` and `./data/packages/beta/`
   - New: `./data/packages/docufiller/stable/` and `./data/packages/docufiller/beta/`

## Prerequisites

- Go runtime installed
- Database file at `./data/versions.db`
- Existing package files in `./data/packages/` (optional)

## Usage

### Step 1: Database Migration

Run the Go migration script:

```bash
go run scripts/migrate.go
```

This will:
- Check if the `docufiller` program exists, create it if not
- Update all version records that don't have a `program_id` to use `docufiller`
- Generate upload and download tokens
- Display the generated tokens (save them securely!)

**Expected Output:**

```
Starting migration...

[Step 1] Creating docufiller program record...
✓ Created docufiller program

[Step 2] Updating existing version records...
✓ Updated X version records

[Step 3] Generating initial tokens...
✓ Generated upload token: <token_value>
✓ Generated download token: <token_value>

--------------------------------------------------
Migration completed successfully!
--------------------------------------------------

IMPORTANT: Save these tokens securely:
  Upload Token:   <upload_token>
  Download Token: <download_token>
```

### Step 2: Storage Migration

#### On Linux/macOS:

```bash
chmod +x scripts/migrate-storage.sh
./scripts/migrate-storage.sh
```

#### On Windows:

```cmd
scripts\migrate-storage.bat
```

This will:
- Create the new directory structure `./data/packages/docufiller/stable/` and `./data/packages/docufiller/beta/`
- Move existing package files to the new structure
- Remove old empty directories

**Expected Output:**

```
Starting storage migration...
================================

[Step 1] Creating new directory structure...
Created directories:
  - ./data/packages/docufiller/stable
  - ./data/packages/docufiller/beta

[Step 2] Migrating stable versions...
  Migrated: <filename> -> docufiller/stable/<version>/
  ...

[Step 3] Migrating beta versions...
  Migrated: <filename> -> docufiller/beta/<version>/
  ...

================================
Storage migration completed!
```

## Important Notes

### Token Security

The migration script generates two tokens:
- **Upload Token**: Required for uploading new versions via `/api/version/upload`
- **Download Token**: Required for downloading packages via `/api/download/{channel}/{version}`

**Store these tokens securely!** You will need them to access the API endpoints.

### Idempotency

Both migration scripts are idempotent - you can run them multiple times safely:
- The database migration checks for existing records before creating them
- The storage migration only migrates files that haven't been migrated yet

### Backup

Before running migration, it's recommended to:
1. Backup your database: `cp ./data/versions.db ./data/versions.db.backup`
2. Backup your packages: `cp -r ./data/packages ./data/packages.backup`

## Verification

After migration, verify the following:

1. **Database:**
   ```bash
   sqlite3 ./data/versions.db "SELECT * FROM programs;"
   sqlite3 ./data/versions.db "SELECT program_id, COUNT(*) FROM versions GROUP BY program_id;"
   sqlite3 ./data/versions.db "SELECT * FROM tokens;"
   ```

2. **Storage:**
   ```bash
   ls -R ./data/packages/docufiller/
   ```

3. **API:**
   ```bash
   # Test health check
   curl http://localhost:8080/api/health

   # Test version list with download token
   curl -H "Authorization: Bearer <download_token>" \
        http://localhost:8080/api/version/list?channel=stable
   ```

## Rollback

If you need to rollback:

1. Restore database from backup:
   ```bash
   cp ./data/versions.db.backup ./data/versions.db
   ```

2. Manually move files back:
   ```bash
   mv ./data/packages/docufiller/stable/*/* ./data/packages/stable/
   mv ./data/packages/docufiller/beta/*/* ./data/packages/beta/
   rm -rf ./data/packages/docufiller
   ```

## Troubleshooting

### "record not found" errors

These are expected during first run - they indicate that records don't exist yet and will be created.

### Permission denied

- Ensure the database file is writable: `chmod 644 ./data/versions.db`
- Ensure the packages directory is writable: `chmod 755 ./data/packages`

### Files not migrated

- Check that the old directory structure exists: `./data/packages/stable/` and `./data/packages/beta/`
- Verify file permissions allow moving

## Next Steps

After successful migration:

1. Update your client applications to use the new API endpoints
2. Use the generated tokens for authentication
3. Test the upload and download flows
4. Remove old backup files after verification

## Support

For issues or questions, please refer to the main project documentation.
