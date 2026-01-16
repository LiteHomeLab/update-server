package main

import (
	"os"
	"testing"
)

func TestCalculateSHA256(t *testing.T) {
	// 创建临时测试文件
	content := []byte("test content for sha256")
	tmpfile := "C:/tmp/test-sha256.bin"
	if err := os.WriteFile(tmpfile, content, 0644); err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpfile)

	hash, err := CalculateSHA256(tmpfile)
	if err != nil {
		t.Fatalf("Failed to calculate SHA256: %v", err)
	}

	// 已知的正确哈希值
	expected := "47914c8afb6da51b436bca58d0fd288d7cd3ea252f778b57617b86f12306c20f"
	if hash != expected {
		t.Errorf("Expected hash %s, got %s", expected, hash)
	}
}

func TestCalculateSHA256NonExistentFile(t *testing.T) {
	_, err := CalculateSHA256("C:/tmp/nonexistent-file-12345.bin")
	if err == nil {
		t.Error("Expected error for nonexistent file")
	}
}
