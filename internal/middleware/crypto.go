package middleware

import (
	"bytes"
	"encoding/json"
	"io"

	"github.com/gin-gonic/gin"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/service"
)

type CryptoMiddleware struct {
	cryptoSvc *service.CryptoService
}

func NewCryptoMiddleware(cryptoSvc *service.CryptoService) *CryptoMiddleware {
	return &CryptoMiddleware{
		cryptoSvc: cryptoSvc,
	}
}

// Process handles encryption and decryption of requests and responses
func (m *CryptoMiddleware) Process() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Read request body
		var bodyBytes []byte
		if c.Request.Body != nil {
			var err error
			bodyBytes, err = io.ReadAll(c.Request.Body)
			c.Request.Body.Close()
			if err != nil {
				logger.Warnf("Failed to read request body: %v", err)
				// Continue with empty body
			}
		}

		// 2. Try to parse as encrypted format
		var encryptedData service.EncryptedData
		isEncrypted := json.Unmarshal(bodyBytes, &encryptedData) == nil && encryptedData.Encrypted

		// 3. Get programID from path or query
		programID := c.Param("programId")
		if programID == "" {
			programID = c.Query("programId")
		}

		// 4. If encrypted request, decrypt it
		if isEncrypted && programID != "" {
			logger.Debugf("Decrypting request for program: %s", programID)
			plaintext, err := m.cryptoSvc.Decrypt(&encryptedData, programID)
			if err != nil {
				logger.Warnf("Decryption failed for program %s: %v", programID, err)
				c.JSON(400, gin.H{"error": "decryption failed"})
				c.Abort()
				return
			}

			// Replace request body with decrypted data
			bodyBytes = plaintext
		}

		// 5. Always restore the request body (whether decrypted or original)
		if bodyBytes != nil {
			c.Request.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		}

		// 6. Continue processing with custom response writer
		writer := &cryptoResponseWriter{
			ResponseWriter: c.Writer,
			cryptoSvc:      m.cryptoSvc,
			programID:      programID,
			shouldEncrypt:  isEncrypted,
		}
		c.Writer = writer

		c.Next()
	}
}

type cryptoResponseWriter struct {
	gin.ResponseWriter
	cryptoSvc     *service.CryptoService
	programID     string
	shouldEncrypt bool
	written       bool
}

func (w *cryptoResponseWriter) Write(data []byte) (int, error) {
	// If not encrypted or no programID, write original data
	if !w.shouldEncrypt || w.programID == "" {
		return w.ResponseWriter.Write(data)
	}

	// Only encrypt once
	if w.written {
		return w.ResponseWriter.Write(data)
	}
	w.written = true

	logger.Debugf("Encrypting response for program: %s", w.programID)

	// Encrypt response
	encrypted, err := w.cryptoSvc.Encrypt(data, w.programID)
	if err != nil {
		logger.Errorf("Encryption failed for program %s: %v", w.programID, err)
		// Return original data if encryption fails
		return w.ResponseWriter.Write(data)
	}

	// Write encrypted data as JSON
	jsonData, err := json.Marshal(encrypted)
	if err != nil {
		logger.Errorf("Failed to marshal encrypted data: %v", err)
		return w.ResponseWriter.Write(data)
	}

	// Set content type if not already set
	if w.Header().Get("Content-Type") == "" {
		w.Header().Set("Content-Type", "application/json")
	}

	return w.ResponseWriter.Write(jsonData)
}

func (w *cryptoResponseWriter) WriteString(s string) (int, error) {
	return w.Write([]byte(s))
}
