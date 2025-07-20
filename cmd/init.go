package cmd

import (
	"github.com/mbrt/gencmd/config"
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
	RunE: func(cmd *cobra.Command, args []string) error {
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
		return nil
	},
}

func init() {
	rootCmd.AddCommand(initCmd)
}
