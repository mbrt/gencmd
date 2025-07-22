package cmd

import (
	"fmt"

	"github.com/mbrt/gencmd/config"
	"github.com/mbrt/gencmd/ui"
	"github.com/spf13/cobra"
)

const initMessage = `
To finish the setup, open the .env file and set the API key for
your LLM here: %s/.env

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
		}
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}

func runInit(cmd *cobra.Command) error {
	created, err := config.InitConfig()
	if err != nil {
		return err
	}
	cmd.Println("Created config files:")
	if len(created) == 0 {
		cmd.Println(" - None: all files already exist.")
	}
	for _, path := range created {
		cmd.Println(" -", path)
	}
	cfgDir := config.Dir()
	cmd.Printf(initMessage, cfgDir, cfgDir, cfgDir)

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
	return nil
}
