package client

import "time"

// Config 客户端配置
type Config struct {
	ServerURL   string        // 服务器地址
	ProgramID   string        // 程序 ID
	Channel     string        // 发布渠道: stable 或 beta
	Timeout     time.Duration // 请求超时时间
	MaxRetries  int           // 最大重试次数
	SavePath    string        // 下载保存路径
}

// DefaultConfig 返回默认配置
func DefaultConfig() *Config {
	return &Config{
		ServerURL:  "http://localhost:8080",
		Channel:    "stable",
		Timeout:    30 * time.Second,
		MaxRetries: 3,
		SavePath:   "./updates",
	}
}
