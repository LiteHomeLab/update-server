#!/bin/bash

# Storage Migration Script for DocuFiller Update Server
# This script migrates existing package files to the new multi-program structure

echo "Starting storage migration..."
echo "================================"

# Define paths
OLD_BASE="./data/packages"
NEW_BASE="./data/packages"

# Step 1: Create new directory structure
echo ""
echo "[Step 1] Creating new directory structure..."
mkdir -p "$NEW_BASE/docufiller/stable"
mkdir -p "$NEW_BASE/docufiller/beta"
echo "Created directories:"
echo "  - $NEW_BASE/docufiller/stable"
echo "  - $NEW_BASE/docufiller/beta"

# Step 2: Migrate stable versions
echo ""
echo "[Step 2] Migrating stable versions..."
if [ -d "$OLD_BASE/stable" ]; then
    # Check if there are files to migrate
    file_count=$(find "$OLD_BASE/stable" -type f | wc -l)
    if [ "$file_count" -gt 0 ]; then
        echo "Found $file_count file(s) in stable channel"

        find "$OLD_BASE/stable" -type f -print0 | while IFS= read -r -d '' file; do
            filename=$(basename "$file")
            # Extract version from filename (remove extension)
            version="${filename%.*}"

            # Create version directory and move file
            target_dir="$NEW_BASE/docufiller/stable/$version"
            mkdir -p "$target_dir"
            mv "$file" "$target_dir/"
            echo "  Migrated: $filename -> docufiller/stable/$version/"
        done

        # Remove old directory if empty
        rmdir "$OLD_BASE/stable" 2>/dev/null && echo "  Removed old stable directory"
        echo "Done migrating stable versions"
    else
        echo "  No files found in stable channel"
    fi
else
    echo "  Old stable directory not found, skipping"
fi

# Step 3: Migrate beta versions
echo ""
echo "[Step 3] Migrating beta versions..."
if [ -d "$OLD_BASE/beta" ]; then
    # Check if there are files to migrate
    file_count=$(find "$OLD_BASE/beta" -type f | wc -l)
    if [ "$file_count" -gt 0 ]; then
        echo "Found $file_count file(s) in beta channel"

        find "$OLD_BASE/beta" -type f -print0 | while IFS= read -r -d '' file; do
            filename=$(basename "$file")
            # Extract version from filename (remove extension)
            version="${filename%.*}"

            # Create version directory and move file
            target_dir="$NEW_BASE/docufiller/beta/$version"
            mkdir -p "$target_dir"
            mv "$file" "$target_dir/"
            echo "  Migrated: $filename -> docufiller/beta/$version/"
        done

        # Remove old directory if empty
        rmdir "$OLD_BASE/beta" 2>/dev/null && echo "  Removed old beta directory"
        echo "Done migrating beta versions"
    else
        echo "  No files found in beta channel"
    fi
else
    echo "  Old beta directory not found, skipping"
fi

# Summary
echo ""
echo "================================"
echo "Storage migration completed!"
echo ""
echo "New directory structure:"
echo "  data/packages/docufiller/stable/"
echo "  data/packages/docufiller/beta/"
echo ""
echo "Next steps:"
echo "1. Run: go run scripts/migrate.go (to migrate database)"
echo "2. Verify the package files are in the correct locations"
echo "3. Test the download endpoint"
