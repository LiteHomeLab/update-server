package service

import (
	"docufiller-update-server/internal/models"
	"errors"

	"gorm.io/gorm"
)

type ProgramService struct {
	db *gorm.DB
}

func NewProgramService(db *gorm.DB) *ProgramService {
	return &ProgramService{db: db}
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
