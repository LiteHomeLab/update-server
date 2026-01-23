package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/spf13/cobra"
	"docufiller-update-server/internal/client"
)

var (
	cfgFile    string
	jsonOutput bool
	daemonMode bool
	daemonPort int
)

var rootCmd = &cobra.Command{
	Use:   "update-client",
	Short: "Update Client - Check and download updates from update server",
	Long:  `A command-line tool for checking and downloading application updates from the update server.`,
}

var checkCmd = &cobra.Command{
	Use:   "check [--current-version VERSION]",
	Short: "Check for updates",
	Long:  `Check if a new version is available on the update server.`,
	RunE:  runCheck,
}

var downloadCmd = &cobra.Command{
	Use:   "download --version VERSION [--output PATH]",
	Short: "Download an update",
	Long:  `Download a specific version from the update server.`,
	RunE:  runDownload,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "update-config.yaml", "config file")
	rootCmd.PersistentFlags().BoolVar(&jsonOutput, "json", false, "output JSON format")

	checkCmd.Flags().String("current-version", "", "current version (overrides config)")

	downloadCmd.Flags().String("output", "", "output file path")
	downloadCmd.Flags().String("version", "", "version to download")
	downloadCmd.Flags().BoolVar(&daemonMode, "daemon", false, "enable daemon mode (HTTP server)")
	downloadCmd.Flags().IntVar(&daemonPort, "port", 0, "HTTP server port (required with --daemon)")
	downloadCmd.MarkFlagRequired("version")

	rootCmd.AddCommand(checkCmd, downloadCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func runCheck(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := client.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get current version from flag or config
	currentVersion, _ := cmd.Flags().GetString("current-version")
	if currentVersion == "" {
		currentVersion = cfg.Program.CurrentVersion
	}

	// Create checker and run
	checker := client.NewUpdateChecker(cfg, jsonOutput)
	return checker.Check(currentVersion)
}

func runDownload(cmd *cobra.Command, args []string) error {
	// Load config
	cfg, err := client.LoadConfig(cfgFile)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Get flags
	version, _ := cmd.Flags().GetString("version")
	outputPath, _ := cmd.Flags().GetString("output")
	daemon, _ := cmd.Flags().GetBool("daemon")
	port, _ := cmd.Flags().GetInt("port")

	// 验证 Daemon 模式参数
	if daemon && port == 0 {
		return fmt.Errorf("--port is required when using --daemon")
	}

	// Daemon 模式
	if daemon {
		return runDaemonDownload(cfg, version, outputPath, port)
	}

	// 普通模式
	checker := client.NewUpdateChecker(cfg, jsonOutput)
	return checker.DownloadWithOutput(version, outputPath)
}

func runDaemonDownload(cfg *client.Config, version, outputPath string, port int) error {
	// 创建状态管理器
	state := client.NewDaemonState(version)

	// 创建并启动 Daemon 服务器
	server := client.NewDaemonServer(port, state)

	// 创建 checker 并设置 daemonState
	checker := client.NewUpdateChecker(cfg, false)
	checker.SetDaemonState(state)

	// 启动父进程监控
	parentPID := client.GetParentPID()
	go server.MonitorParentProcess(parentPID)

	// 在后台启动 HTTP 服务器
	serverErr := make(chan error, 1)
	go func() {
		serverErr <- server.Start()
	}()

	// 等待服务器启动或失败
	select {
	case err := <-serverErr:
		// 服务器启动失败
		return fmt.Errorf("failed to start daemon server: %w", err)
	case <-time.After(5 * time.Second):
		// 超时 - 假设服务器成功启动（Start() 是阻塞调用）
		log.Printf("Daemon server started on port %d\n", port)
	}

	// 执行下载
	downloadErr := checker.DownloadWithOutput(version, outputPath)

	// 等待 shutdown 信号
	<-server.Done()

	return downloadErr
}
