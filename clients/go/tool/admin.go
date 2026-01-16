package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

// UpdateAdmin manages update server operations
type UpdateAdmin struct {
	serverURL string
	token     string
	client    *http.Client
}

// VersionInfo represents version information from the server
type VersionInfo struct {
	Version      string    `json:"version"`
	Channel      string    `json:"channel"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	FileHash     string    `json:"fileHash"`
	ReleaseNotes string    `json:"releaseNotes"`
	PublishDate  time.Time `json:"publishDate"`
	Mandatory    bool      `json:"mandatory"`
}

// UploadProgress tracks upload progress
type UploadProgress struct {
	Uploaded   int64
	Total      int64
	Percentage float64
}

// ProgressCallback is called during upload to report progress
type ProgressCallback func(UploadProgress)

// NewUpdateAdmin creates a new UpdateAdmin instance
func NewUpdateAdmin(serverURL, token string) *UpdateAdmin {
	return &UpdateAdmin{
		serverURL: serverURL,
		token:     token,
		client: &http.Client{
			Timeout: 300 * time.Second, // 5 minutes for large file uploads
		},
	}
}

// ListVersions fetches all versions for a program
func (a *UpdateAdmin) ListVersions(programID, channel string) ([]VersionInfo, error) {
	url := fmt.Sprintf("%s/api/programs/%s/versions?channel=%s", a.serverURL, programID, channel)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("list versions failed with status %d", resp.StatusCode)
	}

	var versions []VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versions); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return versions, nil
}

// DeleteVersion deletes a specific version
func (a *UpdateAdmin) DeleteVersion(programID, version string) error {
	url := fmt.Sprintf("%s/api/programs/%s/versions/%s", a.serverURL, programID, version)

	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("delete version failed with status %d", resp.StatusCode)
	}

	return nil
}

// progressWriter wraps an io.Writer to track upload progress
type progressWriter struct {
	writer   io.Writer
	total    int64
	written  int64
	callback ProgressCallback
}

// Write implements io.Writer while tracking progress
func (pw *progressWriter) Write(p []byte) (int, error) {
	n, err := pw.writer.Write(p)
	if err != nil {
		return n, err
	}

	pw.written += int64(n)
	if pw.callback != nil {
		pw.callback(UploadProgress{
			Uploaded:    pw.written,
			Total:       pw.total,
			Percentage: float64(pw.written) / float64(pw.total) * 100,
		})
	}

	return n, nil
}

// UploadVersion uploads a file with progress tracking
func (a *UpdateAdmin) UploadVersion(programID, channel, version, filePath, notes string, mandatory bool, callback ProgressCallback) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}
	fileSize := fileInfo.Size()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	// Write form fields
	if err := writer.WriteField("channel", channel); err != nil {
		return fmt.Errorf("failed to write channel field: %w", err)
	}
	if err := writer.WriteField("version", version); err != nil {
		return fmt.Errorf("failed to write version field: %w", err)
	}
	if err := writer.WriteField("notes", notes); err != nil {
		return fmt.Errorf("failed to write notes field: %w", err)
	}
	if err := writer.WriteField("mandatory", fmt.Sprintf("%v", mandatory)); err != nil {
		return fmt.Errorf("failed to write mandatory field: %w", err)
	}

	// Create file form field
	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	// Copy file content with progress tracking
	pw := &progressWriter{
		writer:   part,
		total:    fileSize,
		callback: callback,
	}

	if _, err := io.Copy(pw, file); err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	if err := writer.Close(); err != nil {
		return fmt.Errorf("failed to close multipart writer: %w", err)
	}

	url := fmt.Sprintf("%s/api/programs/%s/versions", a.serverURL, programID)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	return nil
}

// VerifyUpload verifies the uploaded version by checking SHA256 hash
func (a *UpdateAdmin) VerifyUpload(programID, channel, version, expectedHash string) error {
	url := fmt.Sprintf("%s/api/programs/%s/versions/%s/%s", a.serverURL, programID, channel, version)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to execute request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("verification failed: status %d", resp.StatusCode)
	}

	var versionInfo VersionInfo
	if err := json.NewDecoder(resp.Body).Decode(&versionInfo); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}

	if versionInfo.FileHash != expectedHash {
		return fmt.Errorf("SHA256 mismatch: expected %s, got %s", expectedHash, versionInfo.FileHash)
	}

	return nil
}

// UploadVersionWithVerify uploads a file and verifies the SHA256 hash
func (a *UpdateAdmin) UploadVersionWithVerify(programID, channel, version, filePath, notes string, mandatory bool, callback ProgressCallback) error {
	// Calculate local SHA256
	localHash, err := CalculateSHA256(filePath)
	if err != nil {
		return fmt.Errorf("failed to calculate SHA256: %w", err)
	}

	// Upload with progress tracking
	if err := a.UploadVersion(programID, channel, version, filePath, notes, mandatory, callback); err != nil {
		return fmt.Errorf("upload failed: %w", err)
	}

	// Verify upload
	if err := a.VerifyUpload(programID, channel, version, localHash); err != nil {
		return fmt.Errorf("upload verification failed: %w", err)
	}

	return nil
}
