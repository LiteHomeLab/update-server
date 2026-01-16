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
func (s *VersionService) GetLatestVersion(channel string) (*models.Version, error) {
	var version models.Version
	err := s.db.Where("channel = ?", channel).Order("publish_date DESC").First(&version).Error
	return &version, err
}

// GetVersionList 获取版本列表
func (s *VersionService) GetVersionList(channel string) ([]models.Version, error) {
	var versions []models.Version
	query := s.db.Order("publish_date DESC")
	if channel != "" {
		query = query.Where("channel = ?", channel)
	}
	err := query.Find(&versions).Error
	return versions, err
}

// GetVersion 获取指定版本
func (s *VersionService) GetVersion(channel, version string) (*models.Version, error) {
	var v models.Version
	err := s.db.Where("channel = ? AND version = ?", channel, version).First(&v).Error
	return &v, err
}

// CreateVersion 创建新版本
func (s *VersionService) CreateVersion(version *models.Version) error {
	return s.db.Create(version).Error
}

// DeleteVersion 删除版本
func (s *VersionService) DeleteVersion(channel, version string) error {
	return s.db.Where("channel = ? AND version = ?", channel, version).Delete(&models.Version{}).Error
}

// IncrementDownloadCount 增加下载计数
func (s *VersionService) IncrementDownloadCount(id uint) error {
	return s.db.Model(&models.Version{}).Where("id = ?", id).UpdateColumn("download_count", gorm.Expr("download_count + ?", 1)).Error
}

// GetStorageService 返回存储服务
func (s *VersionService) GetStorageService() *StorageService {
	return s.storageSvc
}
