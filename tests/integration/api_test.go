package integration

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"docufiller-update-server/tests/helpers"
)

// TestHealthCheck tests the health check endpoint
func TestHealthCheck(t *testing.T) {
	srv := helpers.SetupTestServer(t)
	defer srv.Close()

	req, _ := http.NewRequest("GET", "/api/health", nil)
	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.JSONEq(t, `{"status":"ok"}`, w.Body.String())
}

// TestAdminLoginAPI tests admin login with various credentials
func TestAdminLoginAPI(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	tests := []struct {
		name       string
		username   string
		password   string
		wantStatus int
		wantSuccess bool
	}{
		{
			name:       "valid credentials",
			username:   srv.AdminUser.Username,
			password:   "TestPassword123!",
			wantStatus: http.StatusOK,
			wantSuccess: true,
		},
		{
			name:       "invalid username",
			username:   "nonexistent",
			password:   "TestPassword123!",
			wantStatus: http.StatusUnauthorized,
			wantSuccess: false,
		},
		{
			name:       "invalid password",
			username:   srv.AdminUser.Username,
			password:   "wrongpassword",
			wantStatus: http.StatusUnauthorized,
			wantSuccess: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body := fmt.Sprintf(`{"username":"%s","password":"%s"}`, tt.username, tt.password)
			req := httptest.NewRequest("POST", "/api/admin/login", strings.NewReader(body))
			req.Header.Set("Content-Type", "application/json")

			w := httptest.NewRecorder()
			srv.Router.ServeHTTP(w, req)

			assert.Equal(t, tt.wantStatus, w.Code)

			if tt.wantSuccess {
				var response map[string]interface{}
				json.Unmarshal(w.Body.Bytes(), &response)
				assert.True(t, response["success"].(bool))
			}
		})
	}
}

// TestGetStats tests the stats endpoint
func TestGetStats(t *testing.T) {
	srv := helpers.SetupTestServerWithAdmin(t)
	defer srv.Close()

	req := httptest.NewRequest("GET", "/api/admin/stats", nil)
	req.Header.Set("Authorization", "Bearer "+srv.AdminToken)

	w := httptest.NewRecorder()
	srv.Router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var stats map[string]interface{}
	json.Unmarshal(w.Body.Bytes(), &stats)

	// Verify stats structure
	assert.Contains(t, stats, "totalPrograms")
	assert.Contains(t, stats, "totalVersions")
}
