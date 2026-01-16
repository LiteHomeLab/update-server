package service

import (
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"io"

	"golang.org/x/crypto/hkdf"
)

type CryptoService struct {
	masterKey []byte
}

type EncryptedData struct {
	Encrypted  bool   `json:"encrypted"`
	Algorithm  string `json:"algorithm"`
	IV         string `json:"iv"`
	Ciphertext string `json:"ciphertext"`
	Tag        string `json:"tag"`
}

func NewCryptoService(masterKey string) *CryptoService {
	return &CryptoService{
		masterKey: []byte(masterKey),
	}
}

// DeriveKey 使用 HKDF 从 masterKey 派生程序专用密钥
func (s *CryptoService) DeriveKey(programID string) ([]byte, error) {
	salt := []byte("update-server-salt-" + programID)
	hash := sha256.New

	hkdf := hkdf.New(hash, s.masterKey, salt, nil)

	key := make([]byte, 32) // AES-256
	if _, err := io.ReadFull(hkdf, key); err != nil {
		return nil, err
	}

	return key, nil
}

// Encrypt 加密数据
func (s *CryptoService) Encrypt(plaintext []byte, programID string) (*EncryptedData, error) {
	key, err := s.DeriveKey(programID)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, err
	}

	ciphertext := gcm.Seal(nonce, nonce, plaintext, nil)

	// 分离 nonce 和 ciphertext
	iv := base64.StdEncoding.EncodeToString(nonce)
	tagAndCiphertext := ciphertext[gcm.NonceSize():]

	return &EncryptedData{
		Encrypted:  true,
		Algorithm:  "AES-256-GCM",
		IV:         iv,
		Ciphertext: base64.StdEncoding.EncodeToString(tagAndCiphertext),
	}, nil
}

// Decrypt 解密数据
func (s *CryptoService) Decrypt(data *EncryptedData, programID string) ([]byte, error) {
	if !data.Encrypted || data.Algorithm != "AES-256-GCM" {
		return nil, errors.New("invalid encrypted data format")
	}

	key, err := s.DeriveKey(programID)
	if err != nil {
		return nil, err
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	iv, err := base64.StdEncoding.DecodeString(data.IV)
	if err != nil {
		return nil, err
	}

	ciphertext, err := base64.StdEncoding.DecodeString(data.Ciphertext)
	if err != nil {
		return nil, err
	}

	// 重组 nonce 和 ciphertext
	combined := append(iv, ciphertext...)

	plaintext, err := gcm.Open(nil, combined[:gcm.NonceSize()], combined[gcm.NonceSize():], nil)
	if err != nil {
		return nil, fmt.Errorf("decryption failed: %w", err)
	}

	return plaintext, nil
}
