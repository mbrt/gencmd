package cmd

import (
	"errors"
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"github.com/mbrt/gencmd/ui"
)

var demoCmd = &cobra.Command{
	Use:   "demo",
	Short: "Simulates the gencmd command, without requiring an LLM API",
	Run: func(cmd *cobra.Command, args []string) {
		err := ui.RunUI(ui.NewFakeController(), ui.Options{
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

func init() {
	rootCmd.AddCommand(demoCmd)

	demoCmd.Flags().StringVar(&ttyPath, "tty", "", "Path to the TTY device to use. Defaults to the current terminal.")
}
