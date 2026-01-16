package main

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

var cfgFile string

var rootCmd = &cobra.Command{
	Use:   "update-admin",
	Short: "A CLI tool for managing DocuFiller updates",
	Long: `Update Admin is a CLI tool for managing versions and updates
for the DocuFiller application. It provides commands to upload, delete,
and query version information from the update server.`,
}

var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print the version number",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Println("Update Admin CLI v1.0.0")
	},
}

var uploadCmd = &cobra.Command{
	Use:   "upload [file]",
	Short: "Upload a version package to the server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channel, _ := cmd.Flags().GetString("channel")
		version, _ := cmd.Flags().GetString("version")
		server, _ := cmd.Flags().GetString("server")
		token, _ := cmd.Flags().GetString("token")

		fmt.Printf("Uploading %s...\n", args[0])
		fmt.Printf("Channel: %s, Version: %s\n", channel, version)
		fmt.Printf("Server: %s\n", server)

		err := UploadVersion(args[0], channel, version, server, token)
		if err != nil {
			fmt.Printf("Error uploading: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Upload completed successfully!")
	},
}

var deleteCmd = &cobra.Command{
	Use:   "delete [version]",
	Short: "Delete a version from the server",
	Args:  cobra.ExactArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		channel, _ := cmd.Flags().GetString("channel")
		server, _ := cmd.Flags().GetString("server")
		token, _ := cmd.Flags().GetString("token")

		fmt.Printf("Deleting version %s from channel %s...\n", args[0], channel)

		err := DeleteVersion(channel, args[0], server, token)
		if err != nil {
			fmt.Printf("Error deleting: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Deletion completed successfully!")
	},
}

var listCmd = &cobra.Command{
	Use:   "list",
	Short: "List all versions from the server",
	Run: func(cmd *cobra.Command, args []string) {
		channel, _ := cmd.Flags().GetString("channel")
		server, _ := cmd.Flags().GetString("server")

		fmt.Printf("Listing versions for channel: %s\n", channel)

		versions, err := ListVersions(channel, server)
		if err != nil {
			fmt.Printf("Error listing: %v\n", err)
			os.Exit(1)
		}

		for _, v := range versions {
			fmt.Printf("- %s\n", v)
		}
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
	rootCmd.AddCommand(uploadCmd)
	rootCmd.AddCommand(deleteCmd)
	rootCmd.AddCommand(listCmd)

	// Upload flags
	uploadCmd.Flags().StringP("channel", "c", "stable", "Release channel (stable or beta)")
	uploadCmd.Flags().StringP("version", "v", "", "Version number (required)")
	uploadCmd.Flags().StringP("server", "s", "http://localhost:8080", "Server URL")
	uploadCmd.Flags().StringP("token", "t", "", "Bearer token for authentication")
	uploadCmd.MarkFlagRequired("version")

	// Delete flags
	deleteCmd.Flags().StringP("channel", "c", "stable", "Release channel (stable or beta)")
	deleteCmd.Flags().StringP("server", "s", "http://localhost:8080", "Server URL")
	deleteCmd.Flags().StringP("token", "t", "", "Bearer token for authentication")

	// List flags
	listCmd.Flags().StringP("channel", "c", "stable", "Release channel (stable or beta)")
	listCmd.Flags().StringP("server", "s", "http://localhost:8080", "Server URL")
}

func Execute() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func main() {
	Execute()
}
