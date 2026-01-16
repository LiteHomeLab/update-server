package client

import "time"

// UpdateInfo 更新信息
type UpdateInfo struct {
	Version      string    `json:"version"`
	Channel      string    `json:"channel"`
	FileName     string    `json:"fileName"`
	FileSize     int64     `json:"fileSize"`
	FileHash     string    `json:"fileHash"`
	ReleaseNotes string    `json:"releaseNotes"`
	PublishDate  time.Time `json:"publishDate"`
	Mandatory    bool      `json:"mandatory"`
	DownloadCount int      `json:"downloadCount"`
}

// DownloadProgress 下载进度
type DownloadProgress struct {
	Version    string
	Downloaded int64
	Total      int64
	Percentage float64
	Speed      float64 // bytes/second
}

// ProgressCallback 进度回调函数
type ProgressCallback func(DownloadProgress)

// UpdateError 更新错误
type UpdateError struct {
	Code    string
	Message string
	Err     error
}

func (e *UpdateError) Error() string {
	return e.Message
}
