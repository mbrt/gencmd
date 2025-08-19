package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mbrt/gencmd/config"
)

// configCmd represents the config command
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Manage gencmd configuration",
	Long:  `Manage gencmd configuration files and settings.`,
}

// configShowCmd represents the config show command
var configShowCmd = &cobra.Command{
	Use:   "show",
	Short: "Show the computed configuration",
	Long: `Show the computed configuration with comments about file locations and set environment variables.

This command displays the final configuration that gencmd uses, computed from:
- Default values
- Environment variables
- Configuration file settings`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := runConfigShow(cmd); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configShowCmd)
}

func runConfigShow(cmd *cobra.Command) error {
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load configuration: %w", err)
	}
	_, err = fmt.Fprintf(cmd.OutOrStdout(), "%s", cfg.String())
	return err
}
