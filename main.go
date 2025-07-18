package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/joho/godotenv"

	"github.com/mbrt/gencmd/ctrl"
	"github.com/mbrt/gencmd/ui"
)

const model = "gemini-2.0-flash-lite"

var userCancelErr = errors.New("user cancelled")

func loadEnv() {
	// Load the environment variables from the XDG configuration directory or
	// the default .env file in the current directory. Ignore errors, so that
	// the program can still run if the file is not found.

	// First load the current directory config, then the XDG config, as the
	// first should win against the second.
	_ = godotenv.Load()
	envPath, err := xdg.ConfigFile("gencmd/.env")
	if err == nil {
		_ = godotenv.Load(envPath)
	}
}

func run() error {
	return ui.RunUI(ctrl.New())
}

func main() {
	loadEnv()
	if err := run(); err != nil {
		// Do not print the error if the user cancelled the operation
		if !errors.Is(err, userCancelErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
