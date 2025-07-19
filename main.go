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

func loadEnv() {
	// Load the environment variables from the XDG configuration directory.
	// Ignore errors, so that the program can still run if the file is not
	// found.
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
		if !errors.Is(err, ui.UserCancelErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
