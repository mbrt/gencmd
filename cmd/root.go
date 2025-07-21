package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mbrt/gencmd/config"
	"github.com/mbrt/gencmd/ctrl"
	"github.com/mbrt/gencmd/ui"
)

var ttyPath string

const missingCfgMsg = `WARNING: Error loading configuration: %v
Please run "gencmd init" to create a default configuration.`

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "gencmd",
	Short: "gencmd is a bash command generator from natural language descriptions",
	Long: `
This tool generates shell commands based on natural language
prompts by using a large language model (LLM). It depends on
having access to a compatible LLM API, such as Google Gemini.`,
	CompletionOptions: cobra.CompletionOptions{DisableDefaultCmd: true},
	Run: func(cmd *cobra.Command, args []string) {
		cfg, err := config.Load()
		if err != nil {
			fmt.Fprintf(os.Stderr, missingCfgMsg, err)
		}
		// TODO: Add a fallback for when we don't have a terminal
		err = ui.RunUI(ctrl.New(cfg), ui.Options{
			TtyPath: ttyPath,
		})
		if err != nil {
			// Do not print the error if the user cancelled the operation.
			if !errors.Is(err, ui.UserCancelErr) {
				fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			}
			os.Exit(1)
		}
	},
}

// Execute the root command and its subcommands.
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// We get to this point only if the flag usage is invalid. No need to
		// print the error message, as cobra will do that for us.
		os.Exit(1)
	}
}

func init() {
	rootCmd.Flags().StringVar(&ttyPath, "tty", "", "Path to the TTY device to use. Defaults to the current terminal.")
}
