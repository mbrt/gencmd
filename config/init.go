package config

import (
	_ "embed"
	"errors"
	"os"
	"path/filepath"

	"github.com/adrg/xdg"
)

var (
	//go:embed key-bindings.bash
	keyBindingsBash []byte
	//go:embed key-bindings.zsh
	keyBindingsZsh []byte
	//go:embed default-config.yaml
	defaultConfigYaml []byte
	//go:embed default-dotenv
	defaultDotenv []byte
)

var defaultConfigs = []configFile{
	{
		Name:    "key-bindings.bash",
		Content: keyBindingsBash,
	},
	{
		Name:    "key-bindings.zsh",
		Content: keyBindingsZsh,
	},
	{
		Name:    "config.yaml",
		Content: defaultConfigYaml,
	},
	{
		Name:    ".env",
		Content: defaultDotenv,
	},
}

// InitConfig initializes the configuration for gencmd on a new system.
func InitConfig() ([]string, error) {
	var errs error

	// Initialize config paths.
	for i, cfg := range defaultConfigs {
		fullPath, err := configPath(cfg.Name)
		if err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		defaultConfigs[i].FullPath = fullPath
	}

	var createdPaths []string

	// Create configs if they do not exist.
	for _, cfg := range defaultConfigs {
		if _, err := os.Stat(cfg.FullPath); err == nil {
			continue
		}
		if err := os.WriteFile(cfg.FullPath, cfg.Content, 0o600); err != nil {
			errs = errors.Join(errs, err)
			continue
		}
		createdPaths = append(createdPaths, cfg.FullPath)
	}

	return createdPaths, errs
}

// CfgDir returns the configuration directory for gencmd.
func Dir() string {
	return filepath.Join(xdg.ConfigHome, "gencmd")
}

func configPath(name string) (string, error) {
	// Return the full path to the config file.
	return xdg.ConfigFile(filepath.Join("gencmd", name))
}

type configFile struct {
	Name     string
	Content  []byte
	FullPath string
}
