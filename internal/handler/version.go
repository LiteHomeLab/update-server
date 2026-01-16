package handler

import (
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"docufiller-update-server/internal/logger"
	"docufiller-update-server/internal/models"
	"docufiller-update-server/internal/service"
	"gorm.io/gorm"
)

type VersionHandler struct {
	versionSvc *service.VersionService
}

func NewVersionHandler(db *gorm.DB) *VersionHandler {
	storageSvc := service.NewStorageService("./data/packages")
	return &VersionHandler{
		versionSvc: service.NewVersionService(db, storageSvc),
	}
}

// GetLatestVersion 获取最新版本
func (h *VersionHandler) GetLatestVersion(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.DefaultQuery("channel", "stable")

	logger.Debugf("Get latest version request, program: %s, channel: %s", programID, channel)

	version, err := h.versionSvc.GetLatestVersion(programID, channel)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "No version found"})
		} else {
			logger.Errorf("Failed to get latest version: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(200, version)
}

// GetVersionList 获取版本列表
func (h *VersionHandler) GetVersionList(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.Query("channel")

	logger.Debugf("Get version list request, program: %s, channel: %s", programID, channel)

	versions, err := h.versionSvc.GetVersionList(programID, channel)
	if err != nil {
		logger.Errorf("Failed to get version list: %v", err)
		c.JSON(500, gin.H{"error": "Internal server error"})
		return
	}

	c.JSON(200, versions)
}

// GetVersionDetail 获取版本详情
func (h *VersionHandler) GetVersionDetail(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.Param("channel")
	version := c.Param("version")

	logger.Debugf("Get version detail: %s/%s/%s", programID, channel, version)

	v, err := h.versionSvc.GetVersion(programID, channel, version)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Version not found"})
		} else {
			logger.Errorf("Failed to get version: %v", err)
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	c.JSON(200, v)
}

// UploadVersion 上传新版本
func (h *VersionHandler) UploadVersion(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.PostForm("channel")
	version := c.PostForm("version")
	notes := c.PostForm("notes")
	mandatory, _ := strconv.ParseBool(c.PostForm("mandatory"))

	if programID == "" || channel == "" || version == "" {
		c.JSON(400, gin.H{"error": "programId, channel and version are required"})
		return
	}

	fileHeader, err := c.FormFile("file")
	if err != nil {
		c.JSON(400, gin.H{"error": "file is required"})
		return
	}

	logger.Infof("Upload request: %s/%s/%s, file: %s", programID, channel, version, fileHeader.Filename)

	// 打开文件
	file, err := fileHeader.Open()
	if err != nil {
		logger.Errorf("Failed to open uploaded file: %v", err)
		c.JSON(500, gin.H{"error": "Failed to process file"})
		return
	}
	defer file.Close()

	// 保存文件
	fileName, fileSize, fileHash, err := h.versionSvc.GetStorageService().SaveFile(programID, channel, version, file)
	if err != nil {
		logger.Errorf("Failed to save file: %v", err)
		c.JSON(500, gin.H{"error": "Failed to save file"})
		return
	}

	// 创建版本记录
	v := &models.Version{
		ProgramID:    programID,
		Version:      version,
		Channel:      channel,
		FileName:     fileName,
		FilePath:     filepath.Join("./data/packages", programID, channel, version),
		FileSize:     fileSize,
		FileHash:     fileHash,
		ReleaseNotes: notes,
		PublishDate:  time.Now(),
		Mandatory:    mandatory,
	}

	if err := h.versionSvc.CreateVersion(v); err != nil {
		logger.Errorf("Failed to create version record: %v", err)
		c.JSON(500, gin.H{"error": "Failed to create version"})
		return
	}

	logger.Infof("Version uploaded successfully: %s/%s/%s", programID, channel, version)
	c.JSON(http.StatusOK, gin.H{"message": "Version uploaded successfully", "version": v})
}

// DeleteVersion 删除版本
func (h *VersionHandler) DeleteVersion(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.Query("channel")
	version := c.Param("version")

	if channel == "" {
		channel = "stable" // 默认通道
	}

	logger.Infof("Delete request: %s/%s/%s", programID, channel, version)

	// 删除文件
	if err := h.versionSvc.GetStorageService().DeleteFile(programID, channel, version); err != nil {
		logger.Warnf("Failed to delete file: %v", err)
	}

	// 删除记录
	if err := h.versionSvc.DeleteVersion(programID, channel, version); err != nil {
		logger.Errorf("Failed to delete version: %v", err)
		c.JSON(500, gin.H{"error": "Failed to delete version"})
		return
	}

	c.JSON(200, gin.H{"message": "Version deleted successfully"})
}

// DownloadFile 下载文件
func (h *VersionHandler) DownloadFile(c *gin.Context) {
	programID := c.Param("programId")
	channel := c.Param("channel")
	version := c.Param("version")

	logger.Debugf("Download request: %s/%s/%s", programID, channel, version)

	v, err := h.versionSvc.GetVersion(programID, channel, version)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(404, gin.H{"error": "Version not found"})
		} else {
			c.JSON(500, gin.H{"error": "Internal server error"})
		}
		return
	}

	filePath := h.versionSvc.GetStorageService().GetFilePath(programID, channel, version)
	c.File(filePath)

	// 增加下载计数
	go h.versionSvc.IncrementDownloadCount(v.ID)
}
