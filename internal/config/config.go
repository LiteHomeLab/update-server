package config

import (
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server             ServerConfig   `yaml:"server"`
	Database           DatabaseConfig `yaml:"database"`
	Storage            StorageConfig  `yaml:"storage"`
	API                APIConfig      `yaml:"api"`
	Logger             LoggerConfig   `yaml:"logger"`
	Crypto             CryptoConfig   `yaml:"crypto"`
	ServerURL          string         `yaml:"serverUrl"`        // 客户端连接的服务器地址
	AdminInitialized   bool           `yaml:"adminInitialized"` // 管理员是否已初始化
	ClientsDirectory   string         `yaml:"clientsDirectory"` // 客户端工具目录
	Admin              AdminConfig    `yaml:"admin"`            // 管理员配置
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

// AdminConfig 管理员配置
type AdminConfig struct {
	Username string `yaml:"username"`
	Password string `yaml:"password"`
}

// Load 从 YAML 文件加载配置
func Load(path string) (*Config, error) {
	// 创建默认配置
	config := &Config{
		Server: ServerConfig{
			Port: 8080,
			Host: "0.0.0.0",
		},
		Database: DatabaseConfig{
			Path: "./data/versions.db",
		},
		Storage: StorageConfig{
			BasePath:     "./data/packages",
			MaxFileSize:  536870912, // 512MB
		},
		Logger: LoggerConfig{
			Level:  "info",
			Output: "both",
		},
		ClientsDirectory: "./clients",
	}

	// 加载配置文件（如果存在）
	if _, err := os.Stat(path); err == nil {
		data, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}
		if err := yaml.Unmarshal(data, config); err != nil {
			return nil, err
		}
	}

	// 确保路径是绝对路径
	if !filepath.IsAbs(config.Database.Path) {
		config.Database.Path = absPath(config.Database.Path)
	}
	if !filepath.IsAbs(config.Storage.BasePath) {
		config.Storage.BasePath = absPath(config.Storage.BasePath)
	}
	if !filepath.IsAbs(config.Logger.FilePath) {
		config.Logger.FilePath = absPath(config.Logger.FilePath)
	}
	if !filepath.IsAbs(config.ClientsDirectory) {
		config.ClientsDirectory = absPath(config.ClientsDirectory)
	}

	return config, nil
}

// LoadConfig 从 YAML 文件加载配置（兼容性别名）
func LoadConfig(path string) (*Config, error) {
	return Load(path)
}

func absPath(p string) string {
	abs, err := filepath.Abs(p)
	if err != nil {
		return p
	}
	return abs
}
