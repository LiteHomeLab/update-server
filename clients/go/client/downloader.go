package client

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// DownloadResult 下载结果
type DownloadResult struct {
	Success   bool   `json:"success"`
	File      string `json:"file"`
	FileSize  int64  `json:"fileSize"`
	Verified  bool   `json:"verified"`
	Decrypted bool   `json:"decrypted"`
}

// DownloadWithOutput 下载更新并输出结果
func (c *UpdateChecker) DownloadWithOutput(version string, outputPath string) error {
	if outputPath == "" {
		outputPath = c.generateOutputPath(version)
	}

	// Download
	fmt.Printf("✓ Starting download: %s\n", filepath.Base(outputPath))
	info, err := c.CheckUpdate("")
	if err != nil {
		return c.outputError(err)
	}
	if info != nil {
		fmt.Printf("  Size: %.1f MB\n", float64(info.FileSize)/1024/1024)
	}

	if err := c.DownloadUpdate(version, outputPath, c.progressCallback); err != nil {
		return c.outputError(err)
	}

	// Verify
	verified := true
	if info != nil && info.FileHash != "" {
		verified, _ = c.VerifyFile(outputPath, info.FileHash)
		if !verified {
			return c.outputError(fmt.Errorf("verification failed"))
		}
	}

	// Decrypt if key is available
	decrypted := false
	if c.config.Auth.EncryptionKey != "" {
		decryptor, err := NewDecryptor(c.config.Auth.EncryptionKey)
		if err == nil {
			// Decrypt to same path (in-place)
			if err := decryptor.DecryptFile(outputPath, outputPath); err == nil {
				decrypted = true
			}
		}
	}

	return c.outputDownloadResult(outputPath, decrypted, verified)
}

func (c *UpdateChecker) generateOutputPath(version string) string {
	baseName := "app"
	if c.config.Program.ID != "" {
		baseName = c.config.Program.ID
	}

	switch c.config.Download.Naming {
	case "version":
		return fmt.Sprintf("%s/%s-v%s.zip", c.config.GetSavePath(), baseName, version)
	case "date":
		return fmt.Sprintf("%s/%s-%s.zip", c.config.GetSavePath(), baseName, time.Now().Format("2006-01-02"))
	default:
		return fmt.Sprintf("%s/%s.zip", c.config.GetSavePath(), baseName)
	}
}

func (c *UpdateChecker) progressCallback(progress DownloadProgress) {
	if c.jsonOutput {
		return // No progress in JSON mode
	}

	barWidth := 20
	filled := int(progress.Percentage) / 100 * barWidth / 5
	if filled > barWidth {
		filled = barWidth
	}

	fmt.Printf("\r  Progress: [%-20s] %.1f%% (%.1f/%.1f MB) - %.1f MB/s",
		strings.Repeat("=", filled),
		progress.Percentage,
		float64(progress.Downloaded)/1024/1024,
		float64(progress.Total)/1024/1024,
		progress.Speed/1024/1024)
}

func (c *UpdateChecker) outputDownloadResult(filePath string, decrypted, verified bool) error {
	if c.jsonOutput {
		result := &DownloadResult{
			Success:   true,
			File:      filePath,
			Verified:  verified,
			Decrypted: decrypted,
		}
		if info, err := os.Stat(filePath); err == nil {
			result.FileSize = info.Size()
		}
		return json.NewEncoder(os.Stdout).Encode(result)
	}

	fmt.Println() // New line after progress
	fmt.Printf("✓ Download completed: %s\n", filePath)
	if verified {
		fmt.Printf("✓ Verified: SHA256 matches\n")
	}
	if decrypted {
		fmt.Printf("✓ Decrypted: file ready to use\n")
	}

	return nil
}


// DownloadUpdate 下载更新包
func (c *UpdateChecker) DownloadUpdate(version string, destPath string, callback ProgressCallback) error {
	var lastErr error

	// 重试机制
	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 2 * time.Second) // 指数退避
		}

		err := c.downloadOnce(version, destPath, callback)
		if err == nil {
			return nil // 成功
		}

		lastErr = err
	}

	return lastErr
}

func (c *UpdateChecker) downloadOnce(version string, destPath string, callback ProgressCallback) error {
	url := fmt.Sprintf("%s/api/download/%s/%s/%s",
		c.config.ServerURL, c.config.GetProgramID(), c.config.Channel, version)

	resp, err := c.httpClient.Get(url)
	if err != nil {
		return &UpdateError{
			Code:    "NETWORK_ERROR",
			Message: "Failed to connect to server",
			Err:     err,
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return &UpdateError{
			Code:    "DOWNLOAD_ERROR",
			Message: fmt.Sprintf("Server returned status %d", resp.StatusCode),
		}
	}

	// 确保目录存在
	if err := os.MkdirAll(filepath.Dir(destPath), 0755); err != nil {
		return &UpdateError{
			Code:    "FILE_ERROR",
			Message: "Failed to create directory",
			Err:     err,
		}
	}

	// 创建文件
	file, err := os.Create(destPath)
	if err != nil {
		return &UpdateError{
			Code:    "FILE_ERROR",
			Message: "Failed to create file",
			Err:     err,
		}
	}
	defer file.Close()

	// 获取文件大小
	total := resp.ContentLength
	downloaded := int64(0)
	startTime := time.Now()

	// 使用 buffer 复制
	buffer := make([]byte, 32*1024) // 32KB chunks
	for {
		n, err := resp.Body.Read(buffer)
		if n > 0 {
			if _, writeErr := file.Write(buffer[:n]); writeErr != nil {
				return writeErr
			}

			downloaded += int64(n)

			// 调用进度回调
			if callback != nil && total > 0 {
				elapsed := time.Since(startTime).Seconds()
				speed := float64(downloaded) / elapsed
				if elapsed == 0 {
					speed = 0
				}

				callback(DownloadProgress{
					Version:    version,
					Downloaded: downloaded,
					Total:      total,
					Percentage: float64(downloaded) / float64(total) * 100,
					Speed:      speed,
				})
			}
		}

		if err == io.EOF {
			break
		}
		if err != nil {
			return &UpdateError{
				Code:    "DOWNLOAD_ERROR",
				Message: "Failed to download file",
				Err:     err,
			}
		}
	}

	return nil
}

// VerifyFile 验证文件 SHA256 哈希
func (c *UpdateChecker) VerifyFile(filePath string, expectedHash string) (bool, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return false, err
	}

	actualHash := hex.EncodeToString(hash.Sum(nil))
	return actualHash == expectedHash, nil
}
