package client

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"time"
)

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
		c.config.ServerURL, c.config.ProgramID, c.config.Channel, version)

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
