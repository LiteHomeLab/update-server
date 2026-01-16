package service

import (
	"testing"
)

func TestCryptoService_EncryptDecrypt(t *testing.T) {
	masterKey := "test-master-key-32-bytes-long-!!"
	cryptoSvc := NewCryptoService(masterKey)

	plaintext := []byte("Hello, World!")
	programID := "test-program"

	// 测试加密
	encrypted, err := cryptoSvc.Encrypt(plaintext, programID)
	if err != nil {
		t.Fatalf("Encrypt failed: %v", err)
	}

	if !encrypted.Encrypted {
		t.Error("Encrypted flag should be true")
	}

	// 测试解密
	decrypted, err := cryptoSvc.Decrypt(encrypted, programID)
	if err != nil {
		t.Fatalf("Decrypt failed: %v", err)
	}

	if string(decrypted) != string(plaintext) {
		t.Errorf("Decrypted text mismatch: got %s, want %s", decrypted, plaintext)
	}
}
