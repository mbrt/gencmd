package config

import (
	"bytes"
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// Load reads the configuration from the default path "config.yaml" in the
// user's XDG data directory.
func Load() (Config, error) {
	path, err := configPath("config.yaml")
	if err != nil {
		return DefaultConfig(), err
	}
	return LoadFrom(path)
}

// LoadFrom reads the configuration from the specified YAML file path.
func LoadFrom(path string) (Config, error) {
	// Load the configuration from the specified path.
	res := DefaultConfig()

	f, err := os.Open(path)
	if err != nil {
		return res, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&res); err != nil {
		return res, fmt.Errorf("failed to decode config file: %w", err)
	}
	return res, nil
}

// DefaultConfig returns a default configuration with sensible defaults.
func DefaultConfig() Config {
	// Return a default configuration.
	cfg := Config{}
	decoder := yaml.NewDecoder(bytes.NewReader(defaultConfigYaml))
	if err := decoder.Decode(&cfg); err != nil {
		panic(fmt.Sprintf("failed to decode default config: %v", err))
	}
	return cfg
}

// Config represents the configuration structure for the application.
type Config struct {
	LLM LLMConfig `yaml:"llm"`
}

// LLMConfig represents the configuration for the Language Model.
type LLMConfig struct {
	AutoFromEnv    bool   `yaml:"auto_from_env,omitempty"`
	Provider       string `yaml:"provider,omitempty"`
	ModelName      string `yaml:"model_name,omitempty"`
	PromptTemplate string `yaml:"prompt_template,omitempty"`
}
