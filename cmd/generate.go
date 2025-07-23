package cmd

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"

	"github.com/mbrt/gencmd/config"
	"github.com/mbrt/gencmd/ctrl"
)

var firstOnly bool

// generateCmd represents the generate command
var generateCmd = &cobra.Command{
	Use:   "generate [flags] <prompt...>",
	Short: "Non-interactive generation of commands from natural language prompt",
	Long: `Generate shell commands without the interactive UI.

This command takes a natural language prompt and generates one or more shell commands.
The prompt can be provided as multiple arguments or as a single quoted string.`,
	Example: `  gencmd generate list all files in current directory
  gencmd generate --first "list all processes"`,
	Args: cobra.MinimumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		if err := runGenerate(cmd, args); err != nil {
			fmt.Fprintf(cmd.OutOrStderr(), "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

func init() {
	rootCmd.AddCommand(generateCmd)
	generateCmd.Flags().BoolVarP(&firstOnly, "first", "f", false, "Select and output only the first generated command")
}

func runGenerate(cmd *cobra.Command, args []string) error {
	// Join all arguments with spaces to form the prompt
	prompt := strings.Join(args, " ")

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(cmd.OutOrStderr(), missingCfgMsg, err)
		return fmt.Errorf("failed to load configuration")
	}

	// Generate commands
	controller := ctrl.New(cfg)
	commands, err := controller.GenerateCommands(prompt)
	if err != nil {
		return fmt.Errorf("generating commands: %w", err)
	}
	if len(commands) == 0 {
		return fmt.Errorf("no commands generated")
	}
	if firstOnly {
		// If --first is specified, only return the first command
		commands = commands[:1]
	}
	// Output results
	for _, command := range commands {
		cmd.Println(command)
	}

	return nil
}
