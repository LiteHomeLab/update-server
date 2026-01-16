package service

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"docufiller-update-server/internal/logger"
)

type StorageService struct {
	basePath string
}

func NewStorageService(basePath string) *StorageService {
	return &StorageService{basePath: basePath}
}

// SaveFile 保存文件到指定路径
func (s *StorageService) SaveFile(programID, channel, version string, file io.Reader) (string, int64, string, error) {
	// 创建目录: data/packages/{programID}/{channel}/{version}/
	dir := filepath.Join(s.basePath, programID, channel, version)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return "", 0, "", err
	}

	// 计算哈希
	hash := sha256.New()
	tee := io.TeeReader(file, hash)

	// 保存文件
	fileName := fmt.Sprintf("%s-%s.zip", programID, version)
	filePath := filepath.Join(dir, fileName)

	f, err := os.Create(filePath)
	if err != nil {
		return "", 0, "", err
	}
	defer f.Close()

	size, err := io.Copy(f, tee)
	if err != nil {
		return "", 0, "", err
	}

	fileHash := hex.EncodeToString(hash.Sum(nil))

	logger.Infof("File saved: %s, size: %d, hash: %s", filePath, size, fileHash)

	return fileName, size, fileHash, nil
}

// DeleteFile 删除文件
func (s *StorageService) DeleteFile(programID, channel, version string) error {
	dir := filepath.Join(s.basePath, programID, channel, version)
	return os.RemoveAll(dir)
}

// GetFilePath 获取文件路径
func (s *StorageService) GetFilePath(programID, channel, version string) string {
	return filepath.Join(s.basePath, programID, channel, version, fmt.Sprintf("%s-%s.zip", programID, version))
}
