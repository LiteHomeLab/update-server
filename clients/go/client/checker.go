package client

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// UpdateChecker 更新检查器
type UpdateChecker struct {
	config     *Config
	httpClient *http.Client
}

// NewUpdateChecker 创建更新检查器
func NewUpdateChecker(config *Config) *UpdateChecker {
	return &UpdateChecker{
		config: config,
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
	}
}

// CheckUpdate 检查是否有新版本
func (c *UpdateChecker) CheckUpdate(currentVersion string) (*UpdateInfo, error) {
	url := fmt.Sprintf("%s/api/programs/%s/versions/latest?channel=%s",
		c.config.ServerURL, c.config.ProgramID, c.config.Channel)

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

	if CompareVersions(info.Version, currentVersion) <= 0 {
		return nil, nil
	}

	return &info, nil
}
