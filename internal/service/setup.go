package service

import (
	"crypto/rand"
	"encoding/base64"
	"errors"
	"docufiller-update-server/internal/models"

	"gorm.io/gorm"
)

type SetupService struct {
	db *gorm.DB
}

func NewSetupService(db *gorm.DB) *SetupService {
	return &SetupService{db: db}
}

// IsInitialized 检查是否已初始化
func (s *SetupService) IsInitialized() (bool, error) {
	var count int64
	err := s.db.Model(&models.AdminUser{}).Count(&count).Error
	return count > 0, err
}

// CreateAdminUser 创建管理员用户
func (s *SetupService) CreateAdminUser(username, password string) (*models.AdminUser, error) {
	admin := &models.AdminUser{
		Username: username,
	}
	if err := admin.SetPassword(password); err != nil {
		return nil, err
	}

	if err := s.db.Create(admin).Error; err != nil {
		return nil, err
	}

	return admin, nil
}

// GenerateEncryptionKey 生成32字节随机密钥
func (s *SetupService) GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// InitializeServer 初始化服务器配置
func (s *SetupService) InitializeServer(req InitializeRequest) error {
	// 检查是否已初始化
	initialized, err := s.IsInitialized()
	if err != nil {
		return err
	}
	if initialized {
		return errors.New("server already initialized")
	}

	// 创建管理员
	_, err = s.CreateAdminUser(req.Username, req.Password)
	if err != nil {
		return err
	}

	// TODO: ServerURL 将在后续任务中保存到配置文件
	// 这是为了让客户端工具下载时能自动填充正确的服务器地址

	return nil
}

type InitializeRequest struct {
	Username  string
	Password  string
	ServerURL string
}
