package client

import (
	"encoding/base64"
	"testing"
)

func TestNewDecryptor_ValidKey(t *testing.T) {
	// 32-byte key encoded in base64
	key := base64.StdEncoding.EncodeToString([]byte("12345678901234567890123456789012"))
	_, err := NewDecryptor(key)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}
}

func TestNewDecryptor_InvalidBase64(t *testing.T) {
	_, err := NewDecryptor("invalid-base64!!!")
	if err == nil {
		t.Fatal("Expected error for invalid base64")
	}
}

func TestNewDecryptor_WrongLength(t *testing.T) {
	// Too short
	shortKey := base64.StdEncoding.EncodeToString([]byte("short"))
	_, err := NewDecryptor(shortKey)
	if err == nil {
		t.Fatal("Expected error for wrong key length")
	}
}

func TestNewDecryptor_Valid16ByteKey(t *testing.T) {
	// AES-128
	key := base64.StdEncoding.EncodeToString([]byte("1234567890123456"))
	_, err := NewDecryptor(key)
	if err != nil {
		t.Fatalf("Expected no error for 16-byte key, got %v", err)
	}
}

func TestNewDecryptor_Valid24ByteKey(t *testing.T) {
	// AES-192
	key := base64.StdEncoding.EncodeToString([]byte("123456789012345678901234"))
	_, err := NewDecryptor(key)
	if err != nil {
		t.Fatalf("Expected no error for 24-byte key, got %v", err)
	}
}
