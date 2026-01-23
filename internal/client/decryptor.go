package client

import (
	"crypto/aes"
	"crypto/cipher"
	"encoding/base64"
	"fmt"
	"io"
	"os"
)

// Decryptor 文件解密器
type Decryptor struct {
	key []byte
}

// NewDecryptor 创建解密器
func NewDecryptor(base64Key string) (*Decryptor, error) {
	key, err := base64.StdEncoding.DecodeString(base64Key)
	if err != nil {
		return nil, fmt.Errorf("failed to decode encryption key: %w", err)
	}

	if len(key) != 16 && len(key) != 24 && len(key) != 32 {
		return nil, fmt.Errorf("invalid key length: expected 16, 24, or 32 bytes, got %d", len(key))
	}

	return &Decryptor{key: key}, nil
}

// DecryptFile 解密文件（CTR 模式）
func (d *Decryptor) DecryptFile(srcPath, dstPath string) error {
	// 读取加密文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return fmt.Errorf("failed to open source file: %w", err)
	}
	defer srcFile.Close()

	// 读取全部内容到临时文件
	tempPath := dstPath + ".tmp"
	dstFile, err := os.Create(tempPath)
	if err != nil {
		return fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dstFile.Close()

	// 创建 AES 解密器
	block, err := aes.NewCipher(d.key)
	if err != nil {
		return fmt.Errorf("failed to create cipher: %w", err)
	}

	// 读取 IV（前 16 字节）
	iv := make([]byte, aes.BlockSize)
	if _, err := io.ReadFull(srcFile, iv); err != nil {
		return fmt.Errorf("failed to read IV: %w", err)
	}

	// 创建流解密器
	stream := cipher.NewCTR(block, iv)
	reader := &cipher.StreamReader{S: stream, R: srcFile}

	// 解密并写入临时文件
	if _, err := io.Copy(dstFile, reader); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to decrypt: %w", err)
	}

	// 关闭文件
	dstFile.Close()

	// 替换原文件
	if err := os.Rename(tempPath, dstPath); err != nil {
		os.Remove(tempPath)
		return fmt.Errorf("failed to rename decrypted file: %w", err)
	}

	return nil
}
