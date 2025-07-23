package cmd

import (
	"fmt"
	"os"

	"github.com/mbrt/gencmd/config"
	"github.com/mbrt/gencmd/ui"
	"github.com/spf13/cobra"
)

var reset bool

const initMessage = `
To enable key bindings, add the following line to your shell:
source %s/key-bindings.bash 

(or, for zsh users):
source %s/key-bindings.zsh
`

// initCmd represents the init command
var initCmd = &cobra.Command{
	Use:   "init",
	Short: "Initializes gencmd",
	Long: `Initializes gencmd configuration, if not already done.

This command is safe to run multiple times, as it will not
overwrite existing configuration files.`,
	Run: func(cmd *cobra.Command, _ []string) {
		if err := runInit(cmd); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
	initCmd.Flags().BoolVar(&reset, "reset", false, "Reconfigure the provider.")
}

func runInit(cmd *cobra.Command) error {
	if !reset {
		// Determine whether the configuration already exists.
		cfg, err := config.Load()
		if err == nil && cfg.LLM.Provider != "" {
			cmd.Println("gencmd is already initialized.")
			printConfigPaths(cmd)
			cmd.Println("\nYou can run `gencmd init --reset` to reconfigure the provider.")
			return nil
		}
	}

	_, err := config.InitConfig()
	if err != nil {
		return err
	}

	// Initialize the llm providers.
	providers := config.ProvidersInitOptions()
	name, env, err := ui.SelectProvider(providers)
	if err != nil {
		return err
	}

	if err := config.SaveProviderEnv(name, env); err != nil {
		return err
	}

	cmd.Println("\nProvider configured successfully!")
	printConfigPaths(cmd)
	cmd.Printf(initMessage, config.Dir(), config.Dir())

	return nil
}

func printConfigPaths(cmd *cobra.Command) {
	cmd.Println("\nConfiguration files:")
	for _, cfg := range config.ConfigPaths() {
		cmd.Printf("- %s\n", cfg)
	}
}
