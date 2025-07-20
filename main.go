package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/joho/godotenv"

	"github.com/mbrt/gencmd/ctrl"
	"github.com/mbrt/gencmd/ui"
)

const model = "gemini-2.0-flash-lite"

var (
	ttyPath = flag.String("tty", "", "Path to the TTY device to use for the UI. If not set, the program will use the current terminal.")
)

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
	// TODO: Add a fallback for when we don't have a terminal
	return ui.RunUI(ctrl.New(), ui.Options{
		TtyPath: *ttyPath,
	})
}

func main() {
	loadEnv()
	flag.Parse()

	if err := run(); err != nil {
		// Do not print the error if the user cancelled the operation
		if !errors.Is(err, ui.UserCancelErr) {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		}
		os.Exit(1)
	}
}
