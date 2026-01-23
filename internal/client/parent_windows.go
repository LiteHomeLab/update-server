//go:build windows

package client

import (
	"log"
	"syscall"
	"time"
)

const (
	STILL_ACTIVE = 259 // Windows process exit code indicating process is still running
)

// MonitorParentProcess 监控父进程存活状态
func (d *DaemonServer) MonitorParentProcess(parentPID int) {
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

// isParentAlive 检查父进程是否存活
func (d *DaemonServer) isParentAlive(pid int) bool {
	if pid == 0 {
		return true
	}

	handle, err := syscall.OpenProcess(syscall.PROCESS_QUERY_INFORMATION, false, uint32(pid))
	if err != nil {
		return false
	}
	defer syscall.CloseHandle(handle)

	var exitCode uint32
	err = syscall.GetExitCodeProcess(handle, &exitCode)
	if err != nil {
		return false
	}

	return exitCode == STILL_ACTIVE
}

// GetParentPID 获取父进程 PID
func GetParentPID() int {
	// Windows: 获取当前进程的父进程 PID
	// 使用 NtQueryInformationProcess 或其他 API
	// 简化实现：返回 0 表示未实现
	return 0
}
