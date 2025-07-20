package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// Version set by goreleaser at build time.
var version = "dev"

// versionCmd represents the version command
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Print gencmd version and exit",
	Run: func(*cobra.Command, []string) {
		fmt.Println(version)
	},
}

func init() {
	rootCmd.AddCommand(versionCmd)
}
