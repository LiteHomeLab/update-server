package handler

import (
	"fmt"
	"net/http"
	"docufiller-update-server/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	programService      *service.ProgramService
	versionService     *service.VersionService
	tokenService       *service.TokenService
	clientPackagerService *service.ClientPackager
}

func NewAdminHandler(
	programService *service.ProgramService,
	versionService *service.VersionService,
	tokenService *service.TokenService,
	clientPackagerService *service.ClientPackager,
) *AdminHandler {
	return &AdminHandler{
		programService:      programService,
		versionService:     versionService,
		tokenService:       tokenService,
		clientPackagerService: clientPackagerService,
	}
}

// GetStats 获取统计数据
func (h *AdminHandler) GetStats(c *gin.Context) {
	// TODO: 实现统计逻辑
	c.JSON(http.StatusOK, gin.H{
		"totalPrograms": 0,
		"totalVersions": 0,
		"totalDownloads": 0,
	})
}

// ListPrograms 列出所有程序
func (h *AdminHandler) ListPrograms(c *gin.Context) {
	programs, err := h.programService.ListAll()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, programs)
}

// CreateProgram 创建新程序
func (h *AdminHandler) CreateProgram(c *gin.Context) {
	var req service.CreateProgramRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	result, err := h.programService.CreateProgramWithOptions(req)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

// GetProgramDetail 获取程序详情
func (h *AdminHandler) GetProgramDetail(c *gin.Context) {
	programID := c.Param("programId")

	program, err := h.programService.GetByProgramID(programID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "程序不存在"})
		return
	}

	encryptionKey, _ := h.programService.GetProgramEncryptionKey(programID)
	uploadTokenObj, _ := h.tokenService.GetToken(programID, "upload", "admin")
	downloadTokenObj, _ := h.tokenService.GetToken(programID, "download", "admin")

	uploadToken := ""
	downloadToken := ""
	if uploadTokenObj != nil {
		uploadToken = uploadTokenObj.TokenValue
	}
	if downloadTokenObj != nil {
		downloadToken = downloadTokenObj.TokenValue
	}

	c.JSON(http.StatusOK, gin.H{
		"program":       program,
		"encryptionKey": encryptionKey,
		"uploadToken":   uploadToken,
		"downloadToken": downloadToken,
	})
}

// DeleteProgram 删除程序
func (h *AdminHandler) DeleteProgram(c *gin.Context) {
	programID := c.Param("programId")

	if err := h.programService.DeleteProgram(programID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// ListVersions 列出版本
func (h *AdminHandler) ListVersions(c *gin.Context) {
	programID := c.Param("programId")

	versions, err := h.versionService.ListByProgramID(programID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, versions)
}

// DeleteVersion 删除版本
func (h *AdminHandler) DeleteVersion(c *gin.Context) {
	programID := c.Param("programId")
	version := c.Param("version")

	if err := h.versionService.DeleteVersion(programID, "", version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// DownloadPublishClient 下载发布客户端包
func (h *AdminHandler) DownloadPublishClient(c *gin.Context) {
	programID := c.Param("programId")

	// 生成发布客户端包
	result, err := h.clientPackagerService.GeneratePublishClient(programID, "./temp")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回文件下载
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"",
		fmt.Sprintf("%s-client-publish.zip", result.ProgramName)))
	c.Header("Content-Type", "application/zip")
	c.File(result.PackagePath)
}

// DownloadUpdateClient 下载更新客户端包
func (h *AdminHandler) DownloadUpdateClient(c *gin.Context) {
	programID := c.Param("programId")

	// 生成更新客户端包
	result, err := h.clientPackagerService.GenerateUpdateClient(programID, "./temp")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// 返回文件下载
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"%s\"",
		fmt.Sprintf("%s-client-update.zip", result.ProgramName)))
	c.Header("Content-Type", "application/zip")
	c.File(result.PackagePath)
}

// RegenerateToken 重新生成指定类型的 Token
func (h *AdminHandler) RegenerateToken(c *gin.Context) {
	programID := c.Param("programId")
	tokenType := c.Query("type")

	if tokenType != "upload" && tokenType != "download" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid token type, must be 'upload' or 'download'"})
		return
	}

	token, tokenValue, err := h.tokenService.RegenerateToken(programID, tokenType, "admin")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"tokenID": token.TokenID,
		"token":   tokenValue,
		"type":    tokenType,
	})
}

// RegenerateEncryptionKey 重新生成加密密钥
func (h *AdminHandler) RegenerateEncryptionKey(c *gin.Context) {
	programID := c.Param("programId")

	newKey, err := h.programService.RegenerateEncryptionKey(programID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"encryptionKey": newKey,
	})
}