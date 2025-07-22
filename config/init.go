package config

import (
	_ "embed"
	"errors"
	"fmt"
	"maps"
	"os"
	"path/filepath"
	"strings"

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

// SaveProviderEnv saves the environment variables for the LLM provider to the .env file.
func SaveProviderEnv(name string, env map[string]string) error {
	var requiredOptions []ProviderOption
	var unwantedOptions []string
	for _, opts := range ProvidersInitOptions() {
		if opts.ID == name {
			requiredOptions = opts.Options
			// If the provider has fixed environment variables, add them to the
			// env map.
			maps.Copy(env, opts.FixedEnv)
			continue
		}
		// Remove any options that are not for the selected provider.
		// This ensures that only the relevant options are set.
		for _, opt := range opts.Options {
			unwantedOptions = append(unwantedOptions, opt.EnvVar)
		}
		for k := range opts.FixedEnv {
			unwantedOptions = append(unwantedOptions, k)
		}
	}

	// Validation
	if len(requiredOptions) == 0 {
		return fmt.Errorf("unknown provider: %s", name)
	}
	for _, opt := range requiredOptions {
		if val, ok := env[opt.EnvVar]; !ok || val == "" {
			return fmt.Errorf("missing required environment variable: %s", opt.EnvVar)
		}
	}

	// Load existing .env file contents
	envPath, err := xdg.ConfigFile("gencmd/.env")
	if err != nil {
		return fmt.Errorf("failed to get .env file path: %w", err)
	}
	envLines := readEnvFileContents(envPath)

	// Change options one by one
	for _, opt := range unwantedOptions {
		envLines = unsetOption(envLines, opt)
	}
	for k, v := range env {
		envLines = setOption(envLines, strings.TrimSpace(k), strings.TrimSpace(v))
	}
	// Write the updated .env file
	f, err := os.OpenFile(envPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("failed to open .env file for writing: %w", err)
	}
	defer f.Close()

	if _, err := f.WriteString(strings.Join(envLines, "\n")); err != nil {
		return fmt.Errorf("failed to write .env file: %w", err)
	}
	return nil
}

// ProvidersInitOptions returns the initialization options for each provider.
func ProvidersInitOptions() []ProviderDoc {
	return []ProviderDoc{
		{
			ID:   "googleai",
			Name: "Google Gemini AI",
			URL:  "https://aistudio.google.com/apikey",
			Options: []ProviderOption{
				{
					Name:        "Google Gemini API Key",
					EnvVar:      "GEMINI_API_KEY",
					Description: "API key for Google Gemini",
				},
			},
		},
		{
			ID:   "vertexai",
			Name: "Google Vertex AI",
			URL:  "https://cloud.google.com/vertex-ai/generative-ai/docs/start/api-keys",
			Options: []ProviderOption{
				{
					Name:        "Google Cloud Project",
					EnvVar:      "GOOGLE_CLOUD_PROJECT",
					Description: "ID of the Google Cloud project (e.g. my-project-12345)",
				},
				{
					Name:        "Google Cloud Location",
					EnvVar:      "GOOGLE_CLOUD_LOCATION",
					Description: "Region for Google Cloud services, e.g., us-central1",
				},
			},
			FixedEnv: map[string]string{
				"GOOGLE_GENAI_USE_VERTEXAI": "True",
			},
		},
		{
			ID:   "openai",
			Name: "OpenAI",
			URL:  "https://platform.openai.com/api-keys",
			Options: []ProviderOption{
				{
					Name:        "OpenAI API Key",
					EnvVar:      "OPENAI_API_KEY",
					Description: "API key for OpenAI",
				},
			},
		},
		{
			ID:   "anthropic",
			Name: "Anthropic",
			URL:  "https://console.anthropic.com/settings/keys",
			Options: []ProviderOption{
				{
					Name:        "Anthropic API Key",
					EnvVar:      "ANTHROPIC_API_KEY",
					Description: "API key for Anthropic",
				},
			},
		},
	}
}

type ProviderDoc struct {
	ID      string
	Name    string
	URL     string
	Options []ProviderOption
	// Fixed environment variables for the provider
	FixedEnv map[string]string
}

type ProviderOption struct {
	Name        string
	EnvVar      string
	Description string
}

// CfgDir returns the configuration directory for gencmd.
func Dir() string {
	return filepath.Join(xdg.ConfigHome, "gencmd")
}

// ConfigPaths returns the paths to the configuration files.
func ConfigPaths() []string {
	var paths []string
	dir := Dir()
	for _, cfg := range defaultConfigs {
		paths = append(paths, filepath.Join(dir, cfg.Name))
	}
	return paths
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

func readEnvFileContents(envPath string) []string {
	file, err := os.ReadFile(envPath)
	if err != nil {
		return nil
	}
	lines := string(file)
	return strings.Split(lines, "\n")
}

func setOption(lines []string, key, value string) []string {
	// If the key already exists, update its value.
	// Otherwise, append a new line.
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") {
			trimmedLine = strings.TrimSpace(strings.TrimPrefix(trimmedLine, "#"))
		}
		parts := strings.SplitN(trimmedLine, "=", 2)
		lineKey := strings.TrimSpace(parts[0])

		if lineKey == key {
			lines[i] = fmt.Sprintf("%s=%q", key, value)
			return lines
		}
	}
	// If the key doesn't exist, append it.
	return append(lines, fmt.Sprintf("%s=%q", key, value))
}

func unsetOption(lines []string, key string) []string {
	// Comment out the line with the specified key if it's present and not already commented out.
	for i, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		if strings.HasPrefix(trimmedLine, "#") {
			continue
		}
		parts := strings.SplitN(trimmedLine, "=", 2)
		lineKey := strings.TrimSpace(parts[0])

		if lineKey == key {
			lines[i] = "#" + line
		}
	}
	return lines
}
