package service

import (
	"crypto/rand"
	"docufiller-update-server/internal/models"
	"encoding/base64"
	"errors"

	"gorm.io/gorm"
)

type ProgramService struct {
	db           *gorm.DB
	tokenService *TokenService
}

func NewProgramService(db *gorm.DB) *ProgramService {
	return &ProgramService{
		db:           db,
		tokenService: NewTokenService(db),
	}
}

// CreateProgram 创建程序
func (s *ProgramService) CreateProgram(program *models.Program) error {
	return s.db.Create(program).Error
}

// GetProgramByID 获取程序
func (s *ProgramService) GetProgramByID(programID string) (*models.Program, error) {
	var program models.Program
	err := s.db.Where("program_id = ?", programID).First(&program).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("program not found")
		}
		return nil, err
	}
	return &program, nil
}

// ListPrograms 列出所有程序
func (s *ProgramService) ListPrograms() ([]models.Program, error) {
	var programs []models.Program
	err := s.db.Find(&programs).Error
	return programs, err
}

// UpdateProgram 更新程序
func (s *ProgramService) UpdateProgram(program *models.Program) error {
	return s.db.Save(program).Error
}

// DeleteProgram 删除程序（软删除）
func (s *ProgramService) DeleteProgram(programID string) error {
	return s.db.Where("program_id = ?", programID).Delete(&models.Program{}).Error
}

// CreateProgramWithOptions 创建程序并生成密钥和Token
func (s *ProgramService) CreateProgramWithOptions(req CreateProgramRequest) (*CreateProgramResponse, error) {
	var response CreateProgramResponse

	// 使用事务确保原子性
	err := s.db.Transaction(func(tx *gorm.DB) error {
		// 创建程序
		program := &models.Program{
			ProgramID:   req.ProgramID,
			Name:        req.Name,
			Description: req.Description,
			IsActive:    true,
		}
		if err := tx.Create(program).Error; err != nil {
			return err
		}
		response.Program = program

		// 生成加密密钥
		encryptionKey, err := s.GenerateEncryptionKey()
		if err != nil {
			return err
		}
		keyRecord := &models.EncryptionKey{
			ProgramID: program.ProgramID,
			KeyData:   encryptionKey,
		}
		if err := tx.Create(keyRecord).Error; err != nil {
			return err
		}
		response.EncryptionKey = encryptionKey

		// 生成上传Token
		_, uploadToken, err := s.tokenService.GenerateToken(program.ProgramID, "upload", "system")
		if err != nil {
			return err
		}
		response.UploadToken = uploadToken

		// 生成下载Token
		_, downloadToken, err := s.tokenService.GenerateToken(program.ProgramID, "download", "system")
		if err != nil {
			return err
		}
		response.DownloadToken = downloadToken

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &response, nil
}

// GenerateEncryptionKey 生成32字节随机密钥
func (s *ProgramService) GenerateEncryptionKey() (string, error) {
	key := make([]byte, 32)
	if _, err := rand.Read(key); err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(key), nil
}

// GetProgramEncryptionKey 获取程序的加密密钥
func (s *ProgramService) GetProgramEncryptionKey(programID string) (string, error) {
	var key models.EncryptionKey
	err := s.db.Where("program_id = ?", programID).First(&key).Error
	if err != nil {
		return "", err
	}
	return key.KeyData, nil
}

// RegenerateEncryptionKey 重新生成加密密钥
func (s *ProgramService) RegenerateEncryptionKey(programID string) (string, error) {
	newKey, err := s.GenerateEncryptionKey()
	if err != nil {
		return "", err
	}

	err = s.db.Model(&models.EncryptionKey{}).
		Where("program_id = ?", programID).
		Update("key_data", newKey).Error
	if err != nil {
		return "", err
	}

	return newKey, nil
}

type CreateProgramRequest struct {
	ProgramID   string
	Name        string
	Description string
}

type CreateProgramResponse struct {
	Program       *models.Program
	EncryptionKey string
	UploadToken   string
	DownloadToken string
}
