package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Database DatabaseConfig `yaml:"database"`
	Storage  StorageConfig  `yaml:"storage"`
	API      APIConfig      `yaml:"api"`
	Logger   LoggerConfig   `yaml:"logger"`
	Crypto   CryptoConfig   `yaml:"crypto"`
}

type ServerConfig struct {
	Port int    `yaml:"port"`
	Host string `yaml:"host"`
}

type DatabaseConfig struct {
	Path string `yaml:"path"`
}

type StorageConfig struct {
	BasePath    string `yaml:"basePath"`
	MaxFileSize int64  `yaml:"maxFileSize"`
}

type APIConfig struct {
	UploadToken string `yaml:"uploadToken"`
	CorsEnable  bool   `yaml:"corsEnable"`
}

type LoggerConfig struct {
	Level      string `yaml:"level"`
	Output     string `yaml:"output"`
	FilePath   string `yaml:"filePath"`
	MaxSize    int64  `yaml:"maxSize"`
	MaxBackups int    `yaml:"maxBackups"`
	MaxAge     int    `yaml:"maxAge"`
	Compress   bool   `yaml:"compress"`
}

type CryptoConfig struct {
	MasterKey string `yaml:"masterKey"`
}

// Load 从 YAML 文件加载配置
func Load(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}

	// 确保路径是绝对路径
	if !filepath.IsAbs(cfg.Database.Path) {
		cfg.Database.Path = absPath(cfg.Database.Path)
	}
	if !filepath.IsAbs(cfg.Storage.BasePath) {
		cfg.Storage.BasePath = absPath(cfg.Storage.BasePath)
	}
	if !filepath.IsAbs(cfg.Logger.FilePath) {
		cfg.Logger.FilePath = absPath(cfg.Logger.FilePath)
	}

	return &cfg, nil
}

func absPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}
