package main

import (
	"github.com/adrg/xdg"
	"github.com/joho/godotenv"

	"github.com/mbrt/gencmd/cmd"
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

func main() {
	loadEnv()
	cmd.Execute()
}
