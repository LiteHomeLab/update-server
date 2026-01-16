package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var (
	serverURL string
	token     string
	programID string
)

var rootCmd = &cobra.Command{
	Use:   "update-admin",
	Short: "DocuFiller Update Server Admin Tool",
}

var uploadCmd = &cobra.Command{
	Use:   "upload --channel <stable|beta> --version <version> --file <path>",
	Short: "Upload a new version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		channel, _ := cmd.Flags().GetString("channel")
		version, _ := cmd.Flags().GetString("version")
		filePath, _ := cmd.Flags().GetString("file")
		notes, _ := cmd.Flags().GetString("notes")
		mandatory, _ := cmd.Flags().GetBool("mandatory")

		admin := NewUpdateAdmin(serverURL, token)
		if err := admin.UploadVersion(programID, channel, version, filePath, notes, mandatory); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete --channel <stable|beta> --version <version>",
	Short: "Delete a version",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		channel, _ := cmd.Flags().GetString("channel")
		version, _ := cmd.Flags().GetString("version")

		admin := NewUpdateAdmin(serverURL, token)
		if err := admin.DeleteVersion(programID, channel, version); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var listCmd = &cobra.Command{
	Use:   "list [--channel <stable|beta>]",
	Short: "List versions",
	Args:  cobra.NoArgs,
	Run: func(cmd *cobra.Command, args []string) {
		channel, _ := cmd.Flags().GetString("channel")

		admin := NewUpdateAdmin(serverURL, token)
		versions, err := admin.ListVersions(programID, channel)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}

		for _, v := range versions {
			fmt.Printf("%s (%s) - %s - %d bytes\n", v.Version, v.Channel, v.PublishDate.Format("2006-01-02"), v.FileSize)
		}
	},
}

func init() {
	rootCmd.PersistentFlags().StringVar(&serverURL, "server", "http://localhost:8080", "Server URL")
	rootCmd.PersistentFlags().StringVar(&token, "token", "", "API token (required)")
	rootCmd.PersistentFlags().StringVar(&programID, "program-id", "", "Program ID (required)")
	rootCmd.MarkPersistentFlagRequired("token")
	rootCmd.MarkPersistentFlagRequired("program-id")

	uploadCmd.Flags().String("channel", "", "Channel (stable/beta)")
	uploadCmd.Flags().String("version", "", "Version number")
	uploadCmd.Flags().String("file", "", "File path")
	uploadCmd.Flags().String("notes", "", "Release notes")
	uploadCmd.Flags().Bool("mandatory", false, "Mandatory update")
	uploadCmd.MarkFlagRequired("channel")
	uploadCmd.MarkFlagRequired("version")
	uploadCmd.MarkFlagRequired("file")

	deleteCmd.Flags().String("channel", "", "Channel (stable/beta)")
	deleteCmd.Flags().String("version", "", "Version number")
	deleteCmd.MarkFlagRequired("channel")
	deleteCmd.MarkFlagRequired("version")

	listCmd.Flags().String("channel", "", "Channel filter (stable/beta)")

	rootCmd.AddCommand(uploadCmd, deleteCmd, listCmd)
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
