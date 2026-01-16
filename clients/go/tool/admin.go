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
)

// UploadVersion uploads a version package to the update server
func UploadVersion(filePath, channel, version, serverURL, token string) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("failed to get file info: %w", err)
	}

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	part, err := writer.CreateFormFile("file", filepath.Base(filePath))
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	_, err = io.Copy(part, file)
	if err != nil {
		return fmt.Errorf("failed to copy file content: %w", err)
	}

	_ = writer.WriteField("channel", channel)
	_ = writer.WriteField("version", version)
	_ = writer.WriteField("release_notes", "New release")
	_ = writer.WriteField("mandatory", "false")

	err = writer.Close()
	if err != nil {
		return fmt.Errorf("failed to close writer: %w", err)
	}

	url := fmt.Sprintf("%s/api/version/upload", serverURL)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("upload failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// DeleteVersion deletes a version from the update server
func DeleteVersion(channel, version, serverURL, token string) error {
	url := fmt.Sprintf("%s/api/version/%s/%s", serverURL, channel, version)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	if token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
	}

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK && resp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}

// ListVersions retrieves all versions from the update server
func ListVersions(channel, serverURL string) ([]string, error) {
	url := fmt.Sprintf("%s/api/version/list?channel=%s", serverURL, channel)

	resp, err := http.Get(url)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("list failed with status %d: %s", resp.StatusCode, string(respBody))
	}

	var result struct {
		Success bool     `json:"success"`
		Data    []string `json:"data"`
	}

	err = json.NewDecoder(resp.Body).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	if !result.Success {
		return nil, fmt.Errorf("server returned unsuccessful response")
	}

	return result.Data, nil
}
