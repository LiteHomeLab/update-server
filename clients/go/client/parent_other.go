//go:build !windows

package client

import (
	"log"
	"time"
)

// monitorParentProcess 监控父进程存活状态（非 Windows 平台）
func (d *DaemonServer) monitorParentProcess(parentPID int) {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if !d.isParentAlive(parentPID) {
				log.Println("Parent process died, shutting down")
				d.Shutdown()
				return
			}
		case <-d.done:
			return
		}
	}
}

// isParentAlive 检查父进程是否存活（非 Windows）
func (d *DaemonServer) isParentAlive(pid int) bool {
	if pid == 0 {
		return true
	}
	// Unix: 检查进程是否存在
	// 简化实现
	return true
}

// GetParentPID 获取父进程 PID（非 Windows）
func GetParentPID() int {
	return 0
}
