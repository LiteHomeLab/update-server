package helpers

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// assertFileExists checks if a file exists
func assertFileExists(t TestingT, path string) {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		t.Fatalf("File does not exist: %s", path)
	}
}

// assertFileContains checks if a file contains specific content
func assertFileContains(t TestingT, path, content string) {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	if !strings.Contains(string(data), content) {
		t.Fatalf("File %s does not contain expected content: %s", path, content)
	}
}

// assertZipStructure verifies a ZIP file contains expected files
func assertZipStructure(t TestingT, zipPath string, expectedFiles []string) {
	// Open ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip file %s: %v", zipPath, err)
	}
	defer reader.Close()

	// Collect actual files in ZIP
	actualFiles := make(map[string]bool)
	for _, f := range reader.File {
		actualFiles[f.Name] = true
	}

	// Check all expected files exist
	for _, expected := range expectedFiles {
		if !actualFiles[expected] {
			t.Fatalf("ZIP file %s missing expected file: %s\nAvailable files: %v",
				zipPath, expected, getMapKeys(actualFiles))
		}
	}
}

// unzipFile extracts a ZIP file to the destination directory
func unzipFile(t TestingT, zipPath, destDir string) {
	// Ensure destination directory exists
	if err := os.MkdirAll(destDir, 0755); err != nil {
		t.Fatalf("Failed to create destination directory %s: %v", destDir, err)
	}

	// Open ZIP file
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		t.Fatalf("Failed to open zip file %s: %v", zipPath, err)
	}
	defer reader.Close()

	// Extract each file
	for _, f := range reader.File {
		// Create destination file path
		destPath := filepath.Join(destDir, f.Name)

		// Ensure directory exists for the file
		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(destPath, 0755); err != nil {
				t.Fatalf("Failed to create directory %s: %v", destPath, err)
			}
			continue
		}

		// Create parent directory if needed
		destDir := filepath.Dir(destPath)
		if err := os.MkdirAll(destDir, 0755); err != nil {
			t.Fatalf("Failed to create parent directory %s: %v", destDir, err)
		}

		// Create destination file
		destFile, err := os.OpenFile(destPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			t.Fatalf("Failed to create file %s: %v", destPath, err)
		}

		// Copy file contents
		srcFile, err := f.Open()
		if err != nil {
			t.Fatalf("Failed to open file in zip %s: %v", f.Name, err)
			destFile.Close()
			continue
		}

		_, err = io.Copy(destFile, srcFile)
		srcFile.Close()
		destFile.Close()

		if err != nil {
			t.Fatalf("Failed to extract file %s: %v", f.Name, err)
		}
	}
}

// compareFileHash compares SHA256 hashes of two files
func compareFileHash(t TestingT, file1, file2 string) bool {
	hash1 := calculateFileHash(t, file1)
	hash2 := calculateFileHash(t, file2)

	if hash1 != hash2 {
		t.Fatalf("File hashes do not match:\n%s: %s\n%s: %s", file1, hash1, file2, hash2)
		return false
	}
	return true
}

// calculateFileHash calculates SHA256 hash of a file
func calculateFileHash(t TestingT, filePath string) string {
	file, err := os.Open(filePath)
	if err != nil {
		t.Fatalf("Failed to open file %s: %v", filePath, err)
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		t.Fatalf("Failed to calculate hash for %s: %v", filePath, err)
	}

	return hex.EncodeToString(hash.Sum(nil))
}

// getMapKeys returns the keys of a map as a slice
func getMapKeys(m map[string]bool) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// assertDirExists checks if a directory exists
func assertDirExists(t TestingT, path string) {
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("Directory does not exist: %s: %v", path, err)
	}
	if !info.IsDir() {
		t.Fatalf("Path exists but is not a directory: %s", path)
	}
}

// assertFileNotEmpty checks if a file exists and is not empty
func assertFileNotEmpty(t TestingT, path string) {
	info, err := os.Stat(path)
	if err != nil {
		t.Fatalf("File does not exist: %s: %v", path, err)
	}
	if info.Size() == 0 {
		t.Fatalf("File is empty: %s", path)
	}
}

// assertContains checks if a string contains a substring
func assertContains(t TestingT, s, substr string) {
	if !strings.Contains(s, substr) {
		t.Fatalf("String does not contain expected substring:\nExpected: %s\nActual: %s", substr, s)
	}
}

// assertNotEmpty checks if a string is not empty
func assertNotEmpty(t TestingT, s, name string) {
	if strings.TrimSpace(s) == "" {
		t.Fatalf("%s is empty", name)
	}
}

// createTestFileInDir creates a test file with content in a directory
func createTestFileInDir(t TestingT, dir, filename, content string) string {
	filePath := filepath.Join(dir, filename)
	if err := os.WriteFile(filePath, []byte(content), 0644); err != nil {
		t.Fatalf("Failed to create test file %s: %v", filePath, err)
	}
	return filePath
}

// readFileContent reads and returns file content as string
func readFileContent(t TestingT, path string) string {
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("Failed to read file %s: %v", path, err)
	}
	return string(data)
}

// removeAllTempDir removes a directory and all its contents
func removeAllTempDir(t TestingT, path string) {
	if err := os.RemoveAll(path); err != nil {
		// Silently ignore cleanup errors in tests
		_ = err
	}
}
