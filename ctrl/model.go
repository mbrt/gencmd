package ctrl

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"text/template"

	"google.golang.org/genai"

	"github.com/mbrt/gencmd/config"
)

func NewModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	if cfg.AutoFromEnv {
		if hasOneOfEnvVars("GEMINI_API_KEY", "GOOGLE_API_KEY", "GOOGLE_GENAI_USE_VERTEXAI") {
			cfg.Provider = "google"
			if cfg.ModelName == "" {
				cfg.ModelName = "gemini-2.0-flash-lite"
			}
		}
	}

	switch cfg.Provider {
	case "google":
		return newGeminiModel(ctx, cfg)
	default:
		return nil, fmt.Errorf("unsupported model provider: %s", cfg.Provider)
	}
}

type Model interface {
	GenerateCommands(ctx context.Context, prompt string) ([]string, error)
}

func newGeminiModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	if cfg.ModelName == "" {
		cfg.ModelName = "gemini-2.0-flash-lite"
	}

	client, err := genai.NewClient(ctx, &genai.ClientConfig{})
	if err != nil {
		return nil, fmt.Errorf("failed to create Gemini client: %w", err)
	}

	return &geminiModel{
		client:    client,
		modelName: cfg.ModelName,
		template:  cfg.PromptTemplate,
	}, nil
}

type geminiModel struct {
	client    *genai.Client
	modelName string
	template  string
}

func (m geminiModel) GenerateCommands(ctx context.Context, prompt string) ([]string, error) {
	text, err := templatePrompt(m.template, prompt)
	if err != nil {
		return nil, fmt.Errorf("templating prompt: %w", err)
	}
	content := []*genai.Content{
		{
			Parts: []*genai.Part{{Text: text}},
			Role:  genai.RoleUser,
		},
	}
	config := &genai.GenerateContentConfig{
		ResponseMIMEType: "application/json",
		// See the OpenAPI specification for more details and examples:
		//   https://spec.openapis.org/oas/v3.0.3.html#schema-object
		ResponseSchema: &genai.Schema{
			Type:  "array",
			Items: &genai.Schema{Type: "string"},
		},
	}
	res, err := m.client.Models.GenerateContent(ctx, m.modelName, content, config)
	if err != nil {
		return nil, fmt.Errorf("generating command: %w", err)
	}
	jsonResp := res.Text()
	if jsonResp == "" {
		return nil, fmt.Errorf("no response from model")
	}
	var cmds []string
	err = json.Unmarshal([]byte(jsonResp), &cmds)
	return cmds, err
}

func hasOneOfEnvVars(names ...string) bool {
	for _, name := range names {
		if _, exists := os.LookupEnv(name); exists {
			return true
		}
	}
	return false
}

func templatePrompt(templateStr, prompt string) (string, error) {
	tmpl, err := template.New("prompt").Parse(templateStr)
	if err != nil {
		return "", fmt.Errorf("parsing template: %w", err)
	}

	data := struct{ UserInput string }{
		UserInput: prompt,
	}

	var buf bytes.Buffer
	err = tmpl.Execute(&buf, data)
	if err != nil {
		return "", fmt.Errorf("executing template: %w", err)
	}

	return buf.String(), nil
}
