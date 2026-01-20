package handler

import (
	"net/http"
	"update-server/internal/service"
	"github.com/gin-gonic/gin"
)

type AdminHandler struct {
	programService *service.ProgramService
	versionService *service.VersionService
	tokenService   *service.TokenService
	setupService   *service.SetupService
}

func NewAdminHandler(
	programService *service.ProgramService,
	versionService *service.VersionService,
	tokenService *service.TokenService,
	setupService *service.SetupService,
) *AdminHandler {
	return &AdminHandler{
		programService: programService,
		versionService: versionService,
		tokenService:   tokenService,
		setupService:   setupService,
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
	uploadToken, _ := h.tokenService.GetToken(programID, "upload")
	downloadToken, _ := h.tokenService.GetToken(programID, "download")

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

	if err := h.programService.Delete(programID); err != nil {
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

	if err := h.versionService.Delete(programID, version); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"success": true})
}