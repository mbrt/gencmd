package ctrl

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"text/template"

	"github.com/firebase/genkit/go/ai"
	"github.com/firebase/genkit/go/genkit"
	"github.com/firebase/genkit/go/plugins/compat_oai/anthropic"
	"github.com/firebase/genkit/go/plugins/compat_oai/openai"
	"github.com/firebase/genkit/go/plugins/googlegenai"
	"github.com/openai/openai-go/option"

	"github.com/mbrt/gencmd/config"
)

func NewModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	if cfg.PromptTemplate == "" {
		return Model{}, fmt.Errorf("prompt template is required")
	}
	if cfg.ModelName == "" {
		return Model{}, fmt.Errorf("model name is required")
	}

	switch cfg.Provider {
	case "googleai":
		return newGeminiModel(ctx, cfg)
	case "vertexai":
		return newVertexAIModel(ctx, cfg)
	case "openai":
		return newOpenAIModel(ctx, cfg)
	case "anthropic":
		return newAnthropicModel(ctx, cfg)
	default:
		return Model{}, fmt.Errorf("unsupported model provider: %s", cfg.Provider)
	}
}

// Model is the interface for generating commands based on a prompt.
type Model struct {
	client         *genkit.Genkit
	promptTemplate string
}

// GenerateCommands generates commands based on the provided prompt.
func (m Model) GenerateCommands(ctx context.Context, prompt string) ([]string, error) {
	text, err := templatePrompt(m.promptTemplate, prompt)
	if err != nil {
		return nil, fmt.Errorf("templating prompt: %w", err)
	}
	item, resp, err := genkit.GenerateData[[]string](ctx, m.client, ai.WithPrompt(text))
	if err != nil {
		return nil, fmt.Errorf("generating commands: %w", err)
	}
	if resp == nil || item == nil {
		return nil, fmt.Errorf("no response from model")
	}
	return *item, nil
}

func newGeminiModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.GoogleAI{}),
		genkit.WithDefaultModel("googleai/"+cfg.ModelName),
	)
	if err != nil {
		return Model{}, fmt.Errorf("initializing genkit: %w", err)
	}
	return Model{
		client:         g,
		promptTemplate: cfg.PromptTemplate,
	}, nil
}

func newVertexAIModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&googlegenai.VertexAI{}),
		genkit.WithDefaultModel("vertexai/"+cfg.ModelName),
	)
	if err != nil {
		return Model{}, fmt.Errorf("initializing genkit: %w", err)
	}
	return Model{
		client:         g,
		promptTemplate: cfg.PromptTemplate,
	}, nil
}

func newOpenAIModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&openai.OpenAI{}),
		genkit.WithDefaultModel("openai/"+cfg.ModelName),
	)
	if err != nil {
		return Model{}, fmt.Errorf("initializing genkit: %w", err)
	}
	return Model{
		client:         g,
		promptTemplate: cfg.PromptTemplate,
	}, nil
}

func newAnthropicModel(ctx context.Context, cfg config.LLMConfig) (Model, error) {
	g, err := genkit.Init(ctx,
		genkit.WithPlugins(&anthropic.Anthropic{
			Opts: []option.RequestOption{
				option.WithAPIKey(os.Getenv("ANTHROPIC_API_KEY")),
			},
		}),
		genkit.WithDefaultModel("anthropic/"+cfg.ModelName),
	)
	if err != nil {
		return Model{}, fmt.Errorf("initializing genkit: %w", err)
	}
	return Model{
		client:         g,
		promptTemplate: cfg.PromptTemplate,
	}, nil
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
