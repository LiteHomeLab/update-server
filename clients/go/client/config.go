package client

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

// Config 客户端配置
type Config struct {
	Server   ServerConfig   `yaml:"server"`
	Program  ProgramConfig  `yaml:"program"`
	Auth     AuthConfig     `yaml:"auth"`
	Download DownloadConfig `yaml:"download"`
	Logging  LoggingConfig  `yaml:"logging"`

	// Deprecated fields for backward compatibility
	ServerURL  string        `yaml:"-"` // 旧字段，保留以兼容
	ProgramID  string        `yaml:"-"` // 旧字段，保留以兼容
	Channel    string        `yaml:"-"` // 旧字段，保留以兼容
	Timeout    time.Duration `yaml:"-"` // 旧字段，保留以兼容
	MaxRetries int           `yaml:"-"` // 旧字段，保留以兼容
	SavePath   string        `yaml:"-"` // 旧字段，保留以兼容
}

type ServerConfig struct {
	URL     string `yaml:"url"`
	Timeout int    `yaml:"timeout"`
}

type ProgramConfig struct {
	ID             string `yaml:"id"`
	CurrentVersion string `yaml:"current_version"`
}

type AuthConfig struct {
	Token         string `yaml:"token"`
	EncryptionKey string `yaml:"encryption_key"`
}

type DownloadConfig struct {
	SavePath   string `yaml:"save_path"`
	Naming     string `yaml:"naming"` // version | date | simple
	Keep       int    `yaml:"keep"`   // for date mode
	AutoVerify bool   `yaml:"auto_verify"`
}

type LoggingConfig struct {
	Level string `yaml:"level"`
	File  string `yaml:"file"`
}

// DefaultConfig returns default configuration
func DefaultConfig() *Config {
	return &Config{
		Server: ServerConfig{
			URL:     "http://localhost:8080",
			Timeout: 30,
		},
		Program: ProgramConfig{
			ID: "",
		},
		Download: DownloadConfig{
			SavePath:   "./updates",
			Naming:     "version",
			Keep:       3,
			AutoVerify: true,
		},
		Logging: LoggingConfig{
			Level: "info",
			File:  "update-client.log",
		},
		// Backward compatibility defaults
		ServerURL:  "http://localhost:8080",
		Channel:    "stable",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		SavePath:   "./updates",
	}
}

// LoadConfig loads configuration from YAML file
func LoadConfig(path string) (*Config, error) {
	cfg := DefaultConfig()

	if path == "" {
		return cfg, nil
	}

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	if err := yaml.Unmarshal(data, cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Sync new struct with old fields for backward compatibility
	if cfg.Server.URL != "" {
		cfg.ServerURL = cfg.Server.URL
	}
	if cfg.Program.ID != "" {
		cfg.ProgramID = cfg.Program.ID
	}
	if cfg.Server.Timeout > 0 {
		cfg.Timeout = time.Duration(cfg.Server.Timeout) * time.Second
	}
	if cfg.Download.SavePath != "" {
		cfg.SavePath = cfg.Download.SavePath
	}

	return cfg, nil
}

// GetTimeout returns the timeout duration
func (c *Config) GetTimeout() time.Duration {
	if c.Timeout > 0 {
		return c.Timeout
	}
	return 30 * time.Second
}

// GetProgramID returns the program ID (supports both old and new config)
func (c *Config) GetProgramID() string {
	if c.Program.ID != "" {
		return c.Program.ID
	}
	return c.ProgramID
}

// GetSavePath returns the save path (supports both old and new config)
func (c *Config) GetSavePath() string {
	if c.Download.SavePath != "" {
		return c.Download.SavePath
	}
	return c.SavePath
}

