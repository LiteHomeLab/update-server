package integration

import (
	"archive/zip"
	"bytes"
	"os"
	"path/filepath"
	"testing"

	"docufiller-update-server/internal/config"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
	"docufiller-update-server/tests/helpers"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClientPackager_GeneratePublishClient(t *testing.T) {
	// Setup test server
	srv := helpers.SetupTestServerWithProgram(t)
	defer srv.Close()

	// Create encryption key for the program
	encryptionKey, err := srv.ProgramService.GenerateEncryptionKey()
	require.NoError(t, err)
	keyRecord := &models.EncryptionKey{
		ProgramID: srv.TestProgramID,
		KeyData:   encryptionKey,
	}
	err = srv.DB.Create(keyRecord).Error
	require.NoError(t, err)

	// Create packager with config
	cfg := &config.Config{
		ServerURL: "http://test-server:8080",
	}
	packager := service.NewClientPackager(srv.ProgramService, cfg)

	// Create mock publish client file for testing (required by GeneratePublishClient)
	clientsDir := filepath.Join("data", "clients")
	os.MkdirAll(clientsDir, 0755)
	mockClientPath := filepath.Join(clientsDir, "publish-client.exe")
	mockClientContent := []byte("mock publish client")
	err = os.WriteFile(mockClientPath, mockClientContent, 0644)
	require.NoError(t, err, "Failed to create mock client file")
	defer os.RemoveAll(clientsDir) // Clean up

	// Generate publish client
	tempDir := t.TempDir()
	result, err := packager.GeneratePublishClient(srv.TestProgramID, tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.PackagePath)
	assert.Greater(t, result.PackageSize, int64(0))
	assert.NotEmpty(t, result.Checksum)

	// Verify ZIP file exists
	_, err = os.Stat(result.PackagePath)
	require.NoError(t, err, "ZIP file should exist")

	// Verify ZIP file contents
	zipReader, err := zip.OpenReader(result.PackagePath)
	require.NoError(t, err)
	defer zipReader.Close()

	var configFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "config.yaml" {
			configFile = f
			break
		}
	}
	require.NotNil(t, configFile, "config.yaml should be in the package")

	configReader, err := configFile.Open()
	require.NoError(t, err)
	defer configReader.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(configReader)
	require.NoError(t, err)

	configContent := buf.String()
	assert.Contains(t, configContent, "http://test-server:8080", "Server URL should be from config")
	assert.Contains(t, configContent, srv.TestProgramID, "Program ID should be in config")
	assert.Contains(t, configContent, srv.TestUploadToken, "Upload token should be in config")
	assert.Contains(t, configContent, encryptionKey, "Encryption key should be in config")
	assert.NotContains(t, configContent, "encryption_key: \"\"", "Encryption key should not be empty")
	assert.NotContains(t, configContent, "http://localhost:8080", "Should not contain hardcoded localhost:8080")
}

func TestClientPackager_GenerateUpdateClient(t *testing.T) {
	// Setup test server
	srv := helpers.SetupTestServerWithProgram(t)
	defer srv.Close()

	// Create encryption key
	encryptionKey, err := srv.ProgramService.GenerateEncryptionKey()
	require.NoError(t, err)
	keyRecord := &models.EncryptionKey{
		ProgramID: srv.TestProgramID,
		KeyData:   encryptionKey,
	}
	err = srv.DB.Create(keyRecord).Error
	require.NoError(t, err)

	// Create packager with config
	cfg := &config.Config{
		ServerURL: "http://test-update-server:9090",
	}
	packager := service.NewClientPackager(srv.ProgramService, cfg)

	// Generate update client
	tempDir := t.TempDir()
	result, err := packager.GenerateUpdateClient(srv.TestProgramID, tempDir)
	require.NoError(t, err)
	assert.NotEmpty(t, result.PackagePath)
	assert.Greater(t, result.PackageSize, int64(0))
	assert.NotEmpty(t, result.Checksum)

	// Verify ZIP contents
	zipReader, err := zip.OpenReader(result.PackagePath)
	require.NoError(t, err)
	defer zipReader.Close()

	var configFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "config.yaml" {
			configFile = f
			break
		}
	}
	require.NotNil(t, configFile, "config.yaml should be in the package")

	configReader, err := configFile.Open()
	require.NoError(t, err)
	defer configReader.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(configReader)
	require.NoError(t, err)

	configContent := buf.String()
	assert.Contains(t, configContent, "http://test-update-server:9090", "Server URL should be from config")
	assert.Contains(t, configContent, srv.TestDownloadToken, "Download token should be in config")
	assert.Contains(t, configContent, encryptionKey, "Encryption key should be in config")
	assert.NotContains(t, configContent, "encryption_key: \"\"", "Encryption key should not be empty")
	assert.NotContains(t, configContent, "http://localhost:8080", "Should not contain hardcoded localhost:8080")
}

func TestClientPackager_EmptyServerURL_Fallback(t *testing.T) {
	// Setup test server
	srv := helpers.SetupTestServerWithProgram(t)
	defer srv.Close()

	// Create encryption key
	encryptionKey, err := srv.ProgramService.GenerateEncryptionKey()
	require.NoError(t, err)
	keyRecord := &models.EncryptionKey{
		ProgramID: srv.TestProgramID,
		KeyData:   encryptionKey,
	}
	err = srv.DB.Create(keyRecord).Error
	require.NoError(t, err)

	// Create packager with empty ServerURL (should fallback to localhost:8080)
	cfg := &config.Config{
		ServerURL: "", // Empty - should use fallback
	}
	packager := service.NewClientPackager(srv.ProgramService, cfg)

	// Create mock publish client file for testing
	clientsDir := filepath.Join("data", "clients")
	os.MkdirAll(clientsDir, 0755)
	mockClientPath := filepath.Join(clientsDir, "publish-client.exe")
	mockClientContent := []byte("mock publish client")
	err = os.WriteFile(mockClientPath, mockClientContent, 0644)
	require.NoError(t, err, "Failed to create mock client file")
	defer os.RemoveAll(clientsDir) // Clean up

	// Generate publish client
	tempDir := t.TempDir()
	result, err := packager.GeneratePublishClient(srv.TestProgramID, tempDir)
	require.NoError(t, err)

	// Verify fallback server URL is used
	zipReader, err := zip.OpenReader(result.PackagePath)
	require.NoError(t, err)
	defer zipReader.Close()

	var configFile *zip.File
	for _, f := range zipReader.File {
		if f.Name == "config.yaml" {
			configFile = f
			break
		}
	}
	require.NotNil(t, configFile)

	configReader, err := configFile.Open()
	require.NoError(t, err)
	defer configReader.Close()

	buf := new(bytes.Buffer)
	_, err = buf.ReadFrom(configReader)
	require.NoError(t, err)

	configContent := buf.String()
	// When ServerURL is empty, should fallback to localhost:8080
	assert.Contains(t, configContent, "http://localhost:8080", "Should use fallback server URL when config is empty")
}
