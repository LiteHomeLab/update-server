package service

import (
	"docufiller-update-server/internal/models"
	"gorm.io/gorm"
)

type VersionService struct {
	db         *gorm.DB
	storageSvc *StorageService
}

func NewVersionService(db *gorm.DB, storageSvc *StorageService) *VersionService {
	return &VersionService{
		db:         db,
		storageSvc: storageSvc,
	}
}

// GetLatestVersion 获取最新版本
func (s *VersionService) GetLatestVersion(programID, channel string) (*models.Version, error) {
	var version models.Version
	err := s.db.Where("program_id = ? AND channel = ?", programID, channel).
		Order("publish_date DESC").
		First(&version).Error
	return &version, err
}

// GetVersionList 获取版本列表
func (s *VersionService) GetVersionList(programID, channel string) ([]models.Version, error) {
	var versions []models.Version
	query := s.db.Where("program_id = ?", programID)
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	err := query.Order("publish_date DESC").Find(&versions).Error
	return versions, err
}

// ListByProgramID 根据程序ID列出所有版本
func (s *VersionService) ListByProgramID(programID string) ([]models.Version, error) {
	return s.GetVersionList(programID, "")
}

// GetVersion 获取指定版本
func (s *VersionService) GetVersion(programID, channel, version string) (*models.Version, error) {
	var v models.Version
	err := s.db.Where("program_id = ? AND channel = ? AND version = ?", programID, channel, version).First(&v).Error
	return &v, err
}

// CreateVersion 创建新版本
func (s *VersionService) CreateVersion(version *models.Version) error {
	// 确保 ProgramID 不为空
	if version.ProgramID == "" {
		version.ProgramID = "docufiller"
	}
	return s.db.Create(version).Error
}

// DeleteVersion 删除版本（硬删除，允许重新上传相同版本）
func (s *VersionService) DeleteVersion(programID, channel, version string) error {
	return s.db.Unscoped().Where("program_id = ? AND channel = ? AND version = ?", programID, channel, version).Delete(&models.Version{}).Error
}

// IncrementDownloadCount 增加下载计数
func (s *VersionService) IncrementDownloadCount(id uint) error {
	return s.db.Model(&models.Version{}).Where("id = ?", id).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error
}

// GetStorageService 返回存储服务
func (s *VersionService) GetStorageService() *StorageService {
	return s.storageSvc
}
