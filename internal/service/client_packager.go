package service

import (
	"crypto/sha256"
	"encoding/hex"
	"archive/zip"
	"bytes"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"docufiller-update-server/internal/models"
)

// ClientPackagerConfig 客户端打包配置
type ClientPackagerConfig struct {
	ServerURL   string
	ProgramID   string
	Token       string
	EncryptionKey string
}

// ClientPackagerResult 客户端打包结果
type ClientPackagerResult struct {
	PackagePath   string
	PackageSize   int64
	Checksum      string
	ProgramName   string
}

// ClientPackager 客户端打包器
type ClientPackager struct {
	programService *ProgramService
}

// NewClientPackager 创建客户端打包器
func NewClientPackager(programService *ProgramService) *ClientPackager {
	return &ClientPackager{
		programService: programService,
	}
}

// GeneratePublishClient 生成发布客户端包（exe + 配置文件 + README）
func (p *ClientPackager) GeneratePublishClient(programID, outputDir string) (*ClientPackagerResult, error) {
	// 获取程序信息
	program, err := p.programService.GetProgramByID(programID)
	if err != nil {
		return nil, fmt.Errorf("获取程序信息失败: %v", err)
	}

	// 获取程序的 Token 和密钥
	uploadToken, err := p.programService.tokenService.GetToken(programID, "upload", "system")
	if err != nil {
		return nil, fmt.Errorf("获取上传 Token 失败: %v", err)
	}

	// 打包配置
	config := ClientPackagerConfig{
		ServerURL:      "http://localhost:8080", // TODO: 从配置获取
		ProgramID:     programID,
		Token:         uploadToken.TokenValue,
		EncryptionKey: program.EncryptionKey,
	}

	// 创建临时目录
	tempDir := filepath.Join(outputDir, "temp_"+programID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 复制发布客户端可执行文件
	publishClientSrc := filepath.Join("./data/clients", "publish-client.exe")
	publishClientDst := filepath.Join(tempDir, "update-admin.exe")
	if err := copyFile(publishClientSrc, publishClientDst); err != nil {
		return nil, fmt.Errorf("复制发布客户端失败: %v", err)
	}

	// 生成配置文件
	configContent := generateConfigFile(config)
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configContent, 0644); err != nil {
		return nil, fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 生成 README 文件
	readmeContent, err := p.generateReadmeFile(config, "publish", program)
	if err != nil {
		return nil, fmt.Errorf("生成 README 文件失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), readmeContent, 0644); err != nil {
		return nil, fmt.Errorf("写入 README 文件失败: %v", err)
	}

	// 生成版本信息文件
	versionContent, err := p.generateVersionFile(config, "publish")
	if err != nil {
		return nil, fmt.Errorf("生成版本信息文件失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "version.txt"), versionContent, 0644); err != nil {
		return nil, fmt.Errorf("写入版本信息文件失败: %v", err)
	}

	// 创建 ZIP 文件
	zipPath := filepath.Join(outputDir, fmt.Sprintf("%s-client-publish.zip", program.Name))
	if err := createZipFile(zipPath, tempDir); err != nil {
		return nil, fmt.Errorf("创建 ZIP 文件失败: %v", err)
	}

	// 计算文件信息
	fileInfo, err := os.Stat(zipPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	checksum, err := p.calculateChecksum(zipPath)
	if err != nil {
		return nil, fmt.Errorf("计算校验和失败: %v", err)
	}

	return &ClientPackagerResult{
		PackagePath: zipPath,
		PackageSize: fileInfo.Size(),
		Checksum: checksum,
		ProgramName: program.Name,
	}, nil
}

// GenerateUpdateClient 生成更新客户端包
func (p *ClientPackager) GenerateUpdateClient(programID, outputDir string) (*ClientPackagerResult, error) {
	// 获取程序信息
	program, err := p.programService.GetProgramByID(programID)
	if err != nil {
		return nil, fmt.Errorf("获取程序信息失败: %v", err)
	}

	// 获取程序的 Token 和密钥
	downloadToken, err := p.programService.tokenService.GetToken(programID, "download", "system")
	if err != nil {
		return nil, fmt.Errorf("获取下载 Token 失败: %v", err)
	}

	// 打包配置
	config := ClientPackagerConfig{
		ServerURL:      "http://localhost:8080", // TODO: 从配置获取
		ProgramID:     programID,
		Token:         downloadToken.TokenValue,
		EncryptionKey: program.EncryptionKey,
	}

	// 创建临时目录
	tempDir := filepath.Join(outputDir, "temp_"+programID)
	if err := os.MkdirAll(tempDir, 0755); err != nil {
		return nil, fmt.Errorf("创建临时目录失败: %v", err)
	}
	defer os.RemoveAll(tempDir)

	// 复制更新客户端可执行文件（如果存在）
	updateClientSrc := filepath.Join("./data/clients", "update-client.exe")
	if _, err := os.Stat(updateClientSrc); err == nil {
		updateClientDst := filepath.Join(tempDir, "update-client.exe")
		if err := copyFile(updateClientSrc, updateClientDst); err != nil {
			return nil, fmt.Errorf("复制更新客户端失败: %v", err)
		}
	}
	// 如果更新客户端不存在，会在后续生成说明文档时提示

	// 生成配置文件
	configContent := generateConfigFile(config)
	if err := os.WriteFile(filepath.Join(tempDir, "config.yaml"), configContent, 0644); err != nil {
		return nil, fmt.Errorf("写入配置文件失败: %v", err)
	}

	// 生成 README 文件
	readmeContent, err := p.generateReadmeFile(config, "update", program)
	if err != nil {
		return nil, fmt.Errorf("生成 README 文件失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "README.md"), readmeContent, 0644); err != nil {
		return nil, fmt.Errorf("写入 README 文件失败: %v", err)
	}

	// 生成版本信息文件
	versionContent, err := p.generateVersionFile(config, "update")
	if err != nil {
		return nil, fmt.Errorf("生成版本信息文件失败: %v", err)
	}
	if err := os.WriteFile(filepath.Join(tempDir, "version.txt"), versionContent, 0644); err != nil {
		return nil, fmt.Errorf("写入版本信息文件失败: %v", err)
	}

	// 创建 ZIP 文件
	zipPath := filepath.Join(outputDir, fmt.Sprintf("%s-client-update.zip", program.Name))
	if err := createZipFile(zipPath, tempDir); err != nil {
		return nil, fmt.Errorf("创建 ZIP 文件失败: %v", err)
	}

	// 计算文件信息
	fileInfo, err := os.Stat(zipPath)
	if err != nil {
		return nil, fmt.Errorf("获取文件信息失败: %v", err)
	}

	checksum, err := p.calculateChecksum(zipPath)
	if err != nil {
		return nil, fmt.Errorf("计算校验和失败: %v", err)
	}

	return &ClientPackagerResult{
		PackagePath: zipPath,
		PackageSize: fileInfo.Size(),
		Checksum: checksum,
		ProgramName: program.Name,
	}, nil
}

// generateConfigFile 生成配置文件
func generateConfigFile(config ClientPackagerConfig) []byte {
	content := fmt.Sprintf(`# Update Server Configuration

server:
  url: "%s"
  timeout: 30

program:
  id: "%s"

auth:
  token: "%s"
  encryption_key: "%s"

logging:
  level: info
  file: "client.log"
`, config.ServerURL, config.ProgramID, config.Token, config.EncryptionKey)
	return []byte(content)
}

// generateReadmeFile 生成 README 文件
func (p *ClientPackager) generateReadmeFile(config ClientPackagerConfig, clientType string, program *models.Program) ([]byte, error) {
	var readmeTemplate string
	switch clientType {
	case "publish":
		readmeTemplate = "# {{programName}} 管理工具\n\n用于 {{programName}} 更新服务器的命令行管理工具。\n\n## 配置\n\n### 自动配置（推荐）\n工具启动时会自动读取当前目录的 config.yaml 文件。\n\n### 环境变量（可选）\n```\nsetx UPDATE_SERVER_URL \"{{.ServerURL}}\"\nsetx UPDATE_TOKEN \"{{.Token}}\"\n```\n\n## 使用示例\n\n### 上传新版本\n```\n{{clientExe}} upload --program-id {{.ProgramID}} --channel stable --version 1.0.0 --file your-app.zip --notes \"发布说明\"\n```\n\n### 列出版本\n```\n{{clientExe}} list --program-id {{.ProgramID}}\n```\n\n## 支持的命令\n\n| 命令 | 说明 |\n|------|------|\n| upload | 上传新版本 |\n| list | 列出版本信息 |\n| delete | 删除指定版本 |\n\n## 更多信息\n- 项目主页: https://github.com/LiteHomeLab/update-server\n- 文档: https://github.com/LiteHomeLab/update-server/docs\n"
	case "update":
		readmeTemplate = "# {{programName}} 更新客户端\n\n{{programName}} 的自动更新客户端工具。\n\n## 功能特点\n\n- 自动检查更新\n- 支持静默更新\n- 支持强制更新\n- 支持增量更新\n- 支持代理配置\n\n## 配置\n\n自动读取 config.yaml 配置文件。\n\n## 使用方法\n\n### 命令行启动\n```\n{{clientExe}} --program-id {{.ProgramID}}\n```\n\n### 系统托盘\n双击启动后会在系统托盘运行，自动检查更新。\n\n## 日志文件\n- 日志位置: client.log\n- 更新日志: updates.log\n\n## 版本信息\n- 当前版本: 1.0.0\n- 支持平台: Windows x64\n"
	default:
		readmeTemplate = "# Update Server Client\n\n通用更新客户端工具。\n\n## 配置\n读取 config.yaml 配置文件。\n\n## 使用\n```\n{{clientExe}} --help\n```\n"
	}

	tmpl, err := template.New("readme").Parse(readmeTemplate)
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = tmpl.Execute(&buffer, struct {
		ClientPackagerConfig
		ClientType string
		ProgramName string
		ClientExe  string
	}{
		ClientPackagerConfig: config,
		ClientType:          clientType,
		ProgramName:        program.Name,
		ClientExe:         func() string {
			switch clientType {
			case "publish":
				return "update-admin.exe"
			case "update":
				return "update-client.exe"
			default:
				return "client.exe"
			}
		}(),
	})
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}

// generateVersionFile 生成版本信息文件
func (p *ClientPackager) generateVersionFile(config ClientPackagerConfig, clientType string) ([]byte, error) {
	versionInfo := fmt.Sprintf("Update Server Client Package\n========================\n\nProgram ID: %s\nClient Type: %s\nServer URL: %s\nGenerated At: %s\n\nPackage Contents:\n- config.yaml: Configuration file\n- README.md: User documentation\n- update-client.exe: Client executable\n- version.txt: This version info\n\nChecksum: %s\nPackage Size: %s bytes\n",
		config.ProgramID,
		clientType,
		config.ServerURL,
		"2024-01-01T00:00:00Z", // TODO: 使用当前时间
		"<checksum-placeholder>",
		"<size-placeholder>",
	)

	return []byte(versionInfo), nil
}

// calculateChecksum 计算文件的 SHA256 校验和
func (p *ClientPackager) calculateChecksum(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

// createZipFile 创建 ZIP 文件
func createZipFile(zipPath, sourceDir string) error {
	zipFile, err := os.Create(zipPath)
	if err != nil {
		return err
	}
	defer zipFile.Close()

	zipWriter := zip.NewWriter(zipFile)
	defer zipWriter.Close()

	err = filepath.Walk(sourceDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 跳过目录
		if info.IsDir() {
			return nil
		}

		// 创建 ZIP 文件头
		header, err := FileHeaderByName(strings.TrimPrefix(path, sourceDir+string(os.PathSeparator)))
		if err != nil {
			return err
		}

		// 打开文件
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		defer file.Close()

		// 写入文件头
		writer, err := zipWriter.CreateHeader(header)
		if err != nil {
			return err
		}

		// 复制文件内容
		_, err = io.Copy(writer, file)
		return err
	})

	return err
}

// FileHeaderByName 创建 ZIP 文件头
func FileHeaderByName(name string) (*zip.FileHeader, error) {
	header := &zip.FileHeader{
		Name:   name,
		Method: zip.Deflate,
	}

	// 设置适当的文件权限
	if filepath.Ext(name) == ".exe" {
		header.SetMode(0755)
	} else {
		header.SetMode(0644)
	}

	return header, nil
}

// copyFile 复制文件
func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return err
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, sourceFile)
	if err != nil {
		return err
	}

	// 复制文件权限
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return err
	}

	return os.Chmod(dst, sourceInfo.Mode())
}