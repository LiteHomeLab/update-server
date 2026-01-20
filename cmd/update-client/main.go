package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"github.com/LiteHomeLab/update-server/clients/go/client"
)

var (
	cfgFile    string
	jsonOutput bool
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

	// Get version flag
	version, _ := cmd.Flags().GetString("version")
	outputPath, _ := cmd.Flags().GetString("output")

	// Create checker and run
	checker := client.NewUpdateChecker(cfg, jsonOutput)
	return checker.DownloadWithOutput(version, outputPath)
}
