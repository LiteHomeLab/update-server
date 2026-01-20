package handler

import (
	"net/http"
	"update-server/internal/service"
	"github.com/gin-gonic/gin"
)

type SetupHandler struct {
	setupService *service.SetupService
}

func NewSetupHandler(setupService *service.SetupService) *SetupHandler {
	return &SetupHandler{setupService: setupService}
}

// CheckInitStatus 检查初始化状态
func (h *SetupHandler) CheckInitStatus(c *gin.Context) {
	initialized, err := h.setupService.IsInitialized()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"initialized": initialized})
}

// Initialize 初始化服务器
func (h *SetupHandler) Initialize(c *gin.Context) {
	var req service.InitializeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.setupService.InitializeServer(req); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}