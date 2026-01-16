package main

import (
	"os"
)

type Config struct {
	ServerURL string
	Token     string
	ProgramID string
}

func LoadConfig(serverURL, token, programID string) *Config {
	cfg := &Config{
		ServerURL: serverURL,
		Token:     token,
		ProgramID: programID,
	}

	// 从环境变量读取默认值
	if cfg.ServerURL == "" {
		cfg.ServerURL = os.Getenv("UPDATE_SERVER_URL")
	}
	if cfg.Token == "" {
		cfg.Token = os.Getenv("UPDATE_TOKEN")
	}

	// ProgramID 必须通过参数指定,不使用环境变量

	return cfg
}
