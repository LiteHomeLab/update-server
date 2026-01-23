package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"docufiller-update-server/internal/models"
	"docufiller-update-server/tests/helpers"
)

// TestCreateProgram tests creating a new program
func TestCreateProgram(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	// Create program
	payload := map[string]interface{}{
		"programId":   "test-app-001",
		"name":        "TestApp",
		"description": "Test application",
	}
	body, _ := json.Marshal(payload)

	req := httptest.NewRequest("POST", "/api/admin/programs", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Response structure: {Program: {...}, EncryptionKey: "...", UploadToken: "...", DownloadToken: "..."}
	programObj, ok := response["Program"].(map[string]interface{})
	assert.True(t, ok, "Program should be in response")

	// The JSON field is camelCase: programId
	programID, ok := programObj["programId"].(string)
	assert.True(t, ok, "programId should be in Program object")
	assert.NotEmpty(t, programID)

	// Verify database
	var program models.Program
	srv.DB.Where("program_id = ?", programID).First(&program)
	assert.Equal(t, "TestApp", program.Name)

	// Verify auto-generated tokens
	uploadToken, err := srv.TokenService.GetToken(programID, "upload", "system")
	assert.NoError(t, err)
	assert.NotEmpty(t, uploadToken.TokenValue)

	downloadToken, err := srv.TokenService.GetToken(programID, "download", "system")
	assert.NoError(t, err)
	assert.NotEmpty(t, downloadToken.TokenValue)

	// Verify encryption key from response
	encryptionKey, ok := response["EncryptionKey"].(string)
	assert.True(t, ok, "EncryptionKey should be in response")
	assert.NotEmpty(t, encryptionKey)
}

// TestListPrograms tests listing programs
func TestListPrograms(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	token := srv.AdminToken

	// Create multiple programs
	programID1 := helpers.CreateTestProgram(t, srv, "App1", "Description 1")
	programID2 := helpers.CreateTestProgram(t, srv, "App2", "Description 2")
	_ = helpers.CreateTestProgram(t, srv, "App3", "Description 3")

	// Query list
	req := httptest.NewRequest("GET", "/api/admin/programs", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response []map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	assert.GreaterOrEqual(t, len(response), 3)

	// Verify our programs are in the list
	programIDs := make(map[string]bool)
	for _, p := range response {
		if pid, ok := p["programId"].(string); ok {
			programIDs[pid] = true
		}
	}

	assert.True(t, programIDs[programID1])
	assert.True(t, programIDs[programID2])
}

// TestGetProgramDetail tests getting a specific program
func TestGetProgramDetail(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "DetailApp", "Test for detail view")

	req := httptest.NewRequest("GET", "/api/admin/programs/"+programID, nil)
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var response map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &response)

	// Response format: {program: {...}, encryptionKey: "...", uploadToken: "...", downloadToken: "..."}
	programObj, ok := response["program"].(map[string]interface{})
	assert.True(t, ok, "program should be in response")

	assert.Equal(t, programID, programObj["programId"])
	assert.Equal(t, "DetailApp", programObj["name"])
}

// TestDeleteProgram tests soft deleting a program
func TestDeleteProgram(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	programID := helpers.CreateTestProgram(t, srv, "DeleteApp", "To be deleted")

	// Delete program
	req := httptest.NewRequest("DELETE", "/api/admin/programs/"+programID, nil)
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	// Verify soft delete
	var program models.Program
	err := srv.DB.Unscoped().Where("program_id = ?", programID).First(&program).Error
	assert.NoError(t, err)
	assert.NotNil(t, program.DeletedAt)

	// Should not be found in normal queries
	err = srv.DB.Where("program_id = ?", programID).First(&program).Error
	assert.Error(t, err)
}
