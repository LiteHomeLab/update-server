package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	cfgServerURL string
	cfgToken     string
	cfgProgramID string
)

var rootCmd = &cobra.Command{
	Use:   "update-admin",
	Short: "DocuFiller Update Server Admin Tool",
	Long:  `A command-line tool for managing program versions on the update server.`,
}

func init() {
	rootCmd.PersistentFlags().StringVar(&cfgServerURL, "server", "", "Server URL (overrides UPDATE_SERVER_URL env var)")
	rootCmd.PersistentFlags().StringVar(&cfgToken, "token", "", "API token (overrides UPDATE_TOKEN env var)")
	rootCmd.PersistentFlags().StringVar(&cfgProgramID, "program-id", "", "Program ID (required)")
	rootCmd.MarkPersistentFlagRequired("program-id")

	// Add subcommands
	rootCmd.AddCommand(uploadCmd, listCmd, deleteCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

// upload command
var (
	uploadChannel   string
	uploadVersion   string
	uploadFile      string
	uploadNotes     string
	uploadMandatory bool
)

var uploadCmd = &cobra.Command{
	Use:   "upload --channel <stable|beta> --version <version> --file <path>",
	Short: "Upload a new version",
	Long:  `Upload a new version to the update server with progress display and automatic verification.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig(cfgServerURL, cfgToken, cfgProgramID)

		if uploadChannel == "" || uploadVersion == "" || uploadFile == "" {
			return fmt.Errorf("--channel, --version, and --file are required")
		}

		if _, err := os.Stat(uploadFile); os.IsNotExist(err) {
			return fmt.Errorf("file not found: %s", uploadFile)
		}

		admin := NewUpdateAdmin(cfg.ServerURL, cfg.Token)

		progressCallback := func(p UploadProgress) {
			fmt.Printf("\rProgress: %.1f%% (%d/%d bytes)", p.Percentage, p.Uploaded, p.Total)
		}

		fmt.Printf("Uploading %s/%s/%s...\n", cfgProgramID, uploadChannel, uploadVersion)

		if err := admin.UploadVersionWithVerify(cfg.ProgramID, uploadChannel, uploadVersion, uploadFile, uploadNotes, uploadMandatory, progressCallback); err != nil {
			fmt.Println()
			return err
		}

		fmt.Println("\n✓ Upload successful!")
		return nil
	},
}

func init() {
	uploadCmd.Flags().StringVar(&uploadChannel, "channel", "", "Channel (stable/beta)")
	uploadCmd.Flags().StringVar(&uploadVersion, "version", "", "Version number")
	uploadCmd.Flags().StringVar(&uploadFile, "file", "", "File path")
	uploadCmd.Flags().StringVar(&uploadNotes, "notes", "", "Release notes")
	uploadCmd.Flags().BoolVar(&uploadMandatory, "mandatory", false, "Mandatory update")

	uploadCmd.MarkFlagRequired("channel")
	uploadCmd.MarkFlagRequired("version")
	uploadCmd.MarkFlagRequired("file")
}

// list command
var (
	listChannel string
)

var listCmd = &cobra.Command{
	Use:   "list [--channel <stable|beta>]",
	Short: "List versions",
	Long:  `List all versions for a program, optionally filtered by channel.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig(cfgServerURL, cfgToken, cfgProgramID)
		admin := NewUpdateAdmin(cfg.ServerURL, cfg.Token)

		versions, err := admin.ListVersions(cfg.ProgramID, listChannel)
		if err != nil {
			return err
		}

		if len(versions) == 0 {
			fmt.Println("No versions found")
			return nil
		}

		fmt.Println("Version\tChannel\tSize\t\tDate\t\tMandatory")
		fmt.Println("-------\t-------\t----\t\t----\t\t---------")

		for _, v := range versions {
			sizeMB := float64(v.FileSize) / 1024 / 1024
			date := v.PublishDate.Format("2006-01-02")
			mandatory := "No"
			if v.Mandatory {
				mandatory = "Yes"
			}
			fmt.Printf("%s\t%s\t%.2f MB\t%s\t%s\n", v.Version, v.Channel, sizeMB, date, mandatory)
		}

		return nil
	},
}

func init() {
	listCmd.Flags().StringVar(&listChannel, "channel", "", "Channel filter (stable/beta)")
}

// delete command
var (
	deleteVersion string
)

var deleteCmd = &cobra.Command{
	Use:   "delete --version <version>",
	Short: "Delete a version",
	Long:  `Delete a specific version from the update server.`,
	Args:  cobra.NoArgs,
	RunE: func(cmd *cobra.Command, args []string) error {
		cfg := LoadConfig(cfgServerURL, cfgToken, cfgProgramID)

		if deleteVersion == "" {
			return fmt.Errorf("--version is required")
		}

		admin := NewUpdateAdmin(cfg.ServerURL, cfg.Token)

		fmt.Printf("Deleting %s/%s...\n", cfgProgramID, deleteVersion)

		if err := admin.DeleteVersion(cfg.ProgramID, deleteVersion); err != nil {
			return err
		}

		fmt.Println("✓ Delete successful!")
		return nil
	},
}

func init() {
	deleteCmd.Flags().StringVar(&deleteVersion, "version", "", "Version number")
	deleteCmd.MarkFlagRequired("version")
}
