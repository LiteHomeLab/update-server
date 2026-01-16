package models

import (
	"time"

	"gorm.io/gorm"
)

type Version struct {
	gorm.Model
	ProgramID     string    `gorm:"type:varchar(50);index;not null;uniqueIndex:idx_program_version_channel" json:"programId"`
	Version       string    `gorm:"type:varchar(20);uniqueIndex:idx_program_version_channel" json:"version"`
	Channel       string    `gorm:"type:varchar(10);uniqueIndex:idx_program_version_channel" json:"channel"`
	FileName      string    `gorm:"type:varchar(255);not null" json:"fileName"`
	FilePath      string    `gorm:"type:varchar(500);not null" json:"filePath"`
	FileSize      int64     `json:"fileSize"`
	FileHash      string    `gorm:"type:varchar(64);not null" json:"fileHash"`
	ReleaseNotes  string    `gorm:"type:text" json:"releaseNotes"`
	PublishDate   time.Time `json:"publishDate"`
	DownloadCount int64     `gorm:"default:0" json:"downloadCount"`
	Mandatory     bool      `gorm:"default:false" json:"mandatory"`
}

func (Version) TableName() string {
	return "versions"
}

// BeforeCreate GORM hook - 为 ProgramID 设置默认值以保持向后兼容
func (v *Version) BeforeCreate(tx *gorm.DB) error {
	if v.ProgramID == "" {
		v.ProgramID = "docufiller" // 默认程序ID，向后兼容
	}
	return nil
}
