package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

const TEST_PORT = 18080

func main() {
	if len(os.Args) < 2 {
		fmt.Println("Usage: test_server.go [start|stop|create-data]")
		return
	}

	command := os.Args[1]

	switch command {
	case "start":
		startTestServer()
	case "stop":
		stopTestServer()
	case "create-data":
		createTestData()
	default:
		fmt.Println("Unknown command:", command)
	}
}

func startTestServer() {
	cmd := exec.Command("go", "run", "../main.go")
	cmd.Dir = ".."
	cmd.Env = append(os.Environ(),
		fmt.Sprintf("SERVER_PORT=%d", TEST_PORT),
		"DB_PATH=./tests/test_data/versions.db",
		"STORAGE_PATH=./tests/test_data/packages",
	)

	if err := cmd.Start(); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}

	pidFile := filepath.Join("tests", "test_server.pid")
	if err := os.WriteFile(pidFile, []byte(fmt.Sprintf("%d", cmd.Process.Pid)), 0644); err != nil {
		log.Printf("Warning: failed to write PID file: %v", err)
	}

	fmt.Printf("Test server started on port %d (PID: %d)\n", TEST_PORT, cmd.Process.Pid)
	time.Sleep(2 * time.Second)

	cmd.Wait()
}

func stopTestServer() {
	pidFile := filepath.Join("tests", "test_server.pid")
	data, err := os.ReadFile(pidFile)
	if err != nil {
		log.Printf("No PID file found: %v", err)
		return
	}

	var pid int
	fmt.Sscanf(string(data), "%d", &pid)

	process, err := os.FindProcess(pid)
	if err != nil {
		log.Printf("Failed to find process: %v", err)
		return
	}

	if err := process.Kill(); err != nil {
		log.Printf("Failed to kill process: %v", err)
		return
	}

	os.Remove(pidFile)
	fmt.Println("Test server stopped")
}

func createTestData() {
	dirs := []string{
		"test_data/packages/testapp/stable/1.0.0",
		"test_data/packages/testapp/stable/2.0.0",
		"test_data/packages/testapp/beta/2.1.0-beta",
	}

	for _, dir := range dirs {
		if err := os.MkdirAll(filepath.Join("tests", dir), 0755); err != nil {
			log.Printf("Failed to create directory %s: %v", dir, err)
		}
	}

	createTestFile("tests/test_data/packages/testapp/stable/1.0.0/testapp.zip", 10*1024*1024)
	createTestFile("tests/test_data/packages/testapp/stable/2.0.0/testapp.zip", 15*1024*1024)
	createTestFile("tests/test_data/packages/testapp/beta/2.1.0-beta/testapp.zip", 12*1024*1024)

	fmt.Println("Test data created")
}

func createTestFile(path string, size int64) {
	file, err := os.Create(path)
	if err != nil {
		log.Printf("Failed to create file %s: %v", path, err)
		return
	}
	defer file.Close()

	buffer := make([]byte, 1024*1024)
	for i := 0; i < len(buffer); i++ {
		buffer[i] = byte(i % 256)
	}

	written := int64(0)
	for written < size {
		toWrite := size - written
		if toWrite > int64(len(buffer)) {
			toWrite = int64(len(buffer))
		}

		n, err := file.Write(buffer[:toWrite])
		if err != nil {
			log.Printf("Failed to write to %s: %v", path, err)
			return
		}

		written += int64(n)
	}

	fmt.Printf("Created %s (%d bytes)\n", path, size)
}
