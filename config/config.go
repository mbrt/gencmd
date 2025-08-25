package config

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

const defaultPromptTemplate = `You are a command line expert. Generate up to 5 shell command alternatives that implement the following description:
{{.UserInput}}
`

// Load reads the configuration from the default path "config.yaml" in the
// user's XDG data directory.
func Load() (Config, error) {
	path, err := configPath("config.yaml")
	if err != nil {
		return DefaultFromEnv(), err
	}
	return LoadFrom(path)
}

// LoadFrom reads the configuration from the specified YAML file path.
func LoadFrom(path string) (Config, error) {
	// Load the configuration from the specified path.
	res := DefaultFromEnv()

	f, err := os.Open(path)
	if err != nil {
		return res, fmt.Errorf("failed to open config file: %w", err)
	}
	defer f.Close()

	decoder := yaml.NewDecoder(f)
	if err := decoder.Decode(&res); err != nil {
		return res, fmt.Errorf("failed to decode config file: %w", err)
	}
	res.cfgPath = path
	return res, nil
}

// Default returns a default configuration with sensible defaults.
func Default() Config {
	return Config{
		LLM: LLMConfig{
			PromptTemplate: defaultPromptTemplate,
		},
	}
}

// DefaultFromEnv returns a default configuration based on environment
// variables.
func DefaultFromEnv() Config {
	cfg := Default()

	// Provider
	if v, ok := os.LookupEnv("GOOGLE_GENAI_USE_VERTEXAI"); ok && strings.ToLower(v) == "true" {
		cfg.LLM.Provider = "vertexai"
	} else if _, ok := os.LookupEnv("GEMINI_API_KEY"); ok {
		cfg.LLM.Provider = "googleai"
	} else if _, ok := os.LookupEnv("OPENAI_API_KEY"); ok {
		cfg.LLM.Provider = "openai"
	} else if _, ok := os.LookupEnv("ANTHROPIC_API_KEY"); ok {
		cfg.LLM.Provider = "anthropic"
	} else if _, ok := os.LookupEnv("OLLAMA_HOST"); ok {
		cfg.LLM.Provider = "ollama"
	}

	// Model name
	switch cfg.LLM.Provider {
	case "googleai", "vertexai":
		cfg.LLM.ModelName = "gemini-2.0-flash-lite"
	case "openai":
		cfg.LLM.ModelName = "gpt-4o-mini"
	case "anthropic":
		cfg.LLM.ModelName = "claude-3-5-haiku-latest"
	case "ollama":
		cfg.LLM.ModelName = "gemma-3"
	}

	cfg.envPath, _ = configPath(".env")
	return cfg
}

// Config represents the configuration structure for the application.
type Config struct {
	LLM     LLMConfig `yaml:"llm"`
	cfgPath string
	envPath string
}

func (c Config) String() string {
	var buf strings.Builder
	if c.cfgPath != "" {
		buf.WriteString(fmt.Sprintf("# Configuration file: %s\n", c.cfgPath))
	}
	if c.envPath != "" {
		buf.WriteString(fmt.Sprintf("# Environment file: %s\n", c.envPath))
	}
	if envs := collectSetEnvVars(); len(envs) > 0 {
		buf.WriteString(fmt.Sprintf("# Set environment variables: %s\n", strings.Join(envs, ", ")))
	}

	// Marshal and print the configuration
	yamlData, _ := yaml.Marshal(c)
	buf.Write(yamlData)
	return buf.String()
}

// LLMConfig represents the configuration for the Language Model.
type LLMConfig struct {
	Provider       string        `yaml:"provider"`
	ModelName      string        `yaml:"modelName"`
	PromptTemplate string        `yaml:"promptTemplate"`
	OpenAI         *OpenAIConfig `yaml:"openai,omitempty"`
}

// OpenAIConfig represents the configuration for OpenAI LLMs.
type OpenAIConfig struct {
	BaseURL string `yaml:"baseUrl"`
}

func collectSetEnvVars() []string {
	var res []string
	for _, provider := range ProvidersInitOptions() {
		for k := range provider.FixedEnv {
			if _, ok := os.LookupEnv(k); ok {
				res = append(res, k)
			}
		}
		for _, opt := range provider.Options {
			if _, ok := os.LookupEnv(opt.EnvVar); ok {
				res = append(res, opt.EnvVar)
			}
		}
	}
	return res
}
