package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"time"
)

type UpdateAdmin struct {
	serverURL string
	token     string
	client    *http.Client
}

func NewUpdateAdmin(serverURL, token string) *UpdateAdmin {
	return &UpdateAdmin{
		serverURL: serverURL,
		token:     token,
		client: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type VersionInfo struct {
	Version     string    `json:"version"`
	Channel     string    `json:"channel"`
	FileName    string    `json:"fileName"`
	FileSize    int64     `json:"fileSize"`
	PublishDate time.Time `json:"publishDate"`
}

func (a *UpdateAdmin) UploadVersion(programID, channel, version, filePath, notes string, mandatory bool) error {
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)

	writer.WriteField("channel", channel)
	writer.WriteField("version", version)
	writer.WriteField("notes", notes)
	writer.WriteField("mandatory", fmt.Sprintf("%v", mandatory))

	part, err := writer.CreateFormFile("file", file.Name())
	if err != nil {
		return fmt.Errorf("failed to create form file: %w", err)
	}

	io.Copy(part, file)
	writer.Close()

	url := fmt.Sprintf("%s/api/version/%s/upload", a.serverURL, programID)
	req, err := http.NewRequest("POST", url, &body)
	if err != nil {
		return err
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("upload failed with status %d", resp.StatusCode)
	}

	fmt.Printf("Version %s/%s/%s uploaded successfully\n", programID, channel, version)
	return nil
}

func (a *UpdateAdmin) DeleteVersion(programID, channel, version string) error {
	url := fmt.Sprintf("%s/api/version/%s/%s/%s", a.serverURL, programID, channel, version)
	req, err := http.NewRequest("DELETE", url, nil)
	if err != nil {
		return err
	}

	req.Header.Set("Authorization", "Bearer "+a.token)

	resp, err := a.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("delete failed with status %d", resp.StatusCode)
	}

	fmt.Printf("Version %s/%s/%s deleted successfully\n", programID, channel, version)
	return nil
}

func (a *UpdateAdmin) ListVersions(programID, channel string) ([]VersionInfo, error) {
	url := fmt.Sprintf("%s/api/version/%s/list?channel=%s", a.serverURL, programID, channel)

	resp, err := a.client.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("list failed with status %d", resp.StatusCode)
	}

	var result struct {
		Versions []VersionInfo `json:"versions"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return result.Versions, nil
}
