package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"time"
)

// UpdateChecker 更新检查器
type UpdateChecker struct {
	config     *Config
	httpClient *http.Client
	jsonOutput bool
}

// NewUpdateChecker 创建更新检查器
func NewUpdateChecker(config *Config, jsonOutput bool) *UpdateChecker {
	return &UpdateChecker{
		config:     config,
		jsonOutput: jsonOutput,
		httpClient: &http.Client{
			Timeout: config.GetTimeout(),
		},
	}
}

// CheckResult 检查结果
type CheckResult struct {
	HasUpdate      bool   `json:"hasUpdate"`
	CurrentVersion string `json:"currentVersion,omitempty"`
	LatestVersion  string `json:"latestVersion"`
	DownloadURL    string `json:"downloadUrl,omitempty"`
	FileSize       int64  `json:"fileSize,omitempty"`
	ReleaseNotes   string `json:"releaseNotes,omitempty"`
	PublishDate    string `json:"publishDate,omitempty"`
	Mandatory      bool   `json:"mandatory"`
}

// Check 检查更新并输出结果
func (c *UpdateChecker) Check(currentVersion string) error {
	info, err := c.CheckUpdate(currentVersion)
	if err != nil {
		return c.outputError(err)
	}

	if info == nil {
		// No update available
		if c.jsonOutput {
			result := &CheckResult{
				HasUpdate:     false,
				LatestVersion: currentVersion,
			}
			return json.NewEncoder(os.Stdout).Encode(result)
		}
		fmt.Printf("✓ Already up to date (version %s)\n", currentVersion)
		return nil
	}

	return c.outputResult(info, currentVersion)
}

// CheckUpdate 检查是否有新版本（internal method）
func (c *UpdateChecker) CheckUpdate(currentVersion string) (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/api/programs/%s/versions/latest?channel=%s",
		c.config.ServerURL, c.config.GetProgramID(), "stable") // TODO: support channel

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return nil, &UpdateError{
			Code:    "NETWORK_ERROR",
			Message: fmt.Sprintf("Failed to connect to server: %v", err),
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, &UpdateError{
			Code:    "NO_VERSION",
			Message: "No version found for this program",
		}
	}

	if resp.StatusCode != http.StatusOK {
		return nil, &UpdateError{
			Code:    "SERVER_ERROR",
			Message: fmt.Sprintf("Server returned status %d", resp.StatusCode),
		}
	}

	var info UpdateInfo
	if err := json.NewDecoder(resp.Body).Decode(&info); err != nil {
		return nil, &UpdateError{
			Code:    "PARSE_ERROR",
			Message: "Failed to parse response",
			Err:     err,
		}
	}

	// Check if version is newer
	if currentVersion != "" && CompareVersions(info.Version, currentVersion) <= 0 {
		return nil, nil
	}

	return &info, nil
}

// outputResult 输出检查结果
func (c *UpdateChecker) outputResult(info *UpdateInfo, currentVersion string) error {
	if c.jsonOutput {
		result := &CheckResult{
			HasUpdate:     true,
			CurrentVersion: currentVersion,
			LatestVersion:  info.Version,
			FileSize:      info.FileSize,
			ReleaseNotes:  info.ReleaseNotes,
			PublishDate:   info.PublishDate.Format(time.RFC3339),
			Mandatory:     info.Mandatory,
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	// Human-readable output
	fmt.Printf("✓ Connected to update server\n")
	if currentVersion != "" {
		fmt.Printf("  Current version: %s\n", currentVersion)
	}
	fmt.Printf("  Latest version: %s\n", info.Version)
	if currentVersion != "" && CompareVersions(info.Version, currentVersion) > 0 {
		fmt.Printf("\n  New version available!\n")
	}
	fmt.Printf("\n  Version details:\n")
	fmt.Printf("    Size: %.1f MB\n", float64(info.FileSize)/1024/1024)
	fmt.Printf("    Published: %s\n", info.PublishDate.Format("2006-01-02"))
	fmt.Printf("    Mandatory: %s\n", map[bool]string{true: "Yes", false: "No"}[info.Mandatory])
	if info.ReleaseNotes != "" {
		fmt.Printf("    Release notes:\n      %s\n", info.ReleaseNotes)
	}

	return nil
}

// outputError 输出错误信息
func (c *UpdateChecker) outputError(err error) error {
	if c.jsonOutput {
		result := map[string]interface{}{
			"success": false,
			"error":   err.Error(),
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	// Human-readable error
	fmt.Printf("✗ Failed to check for updates\n")
	if ue, ok := err.(*UpdateError); ok {
		fmt.Printf("  Error: %s\n", ue.Message)
	} else {
		fmt.Printf("  Error: %v\n", err)
	}

	return err
}
