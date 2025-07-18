package ctrl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"google.golang.org/genai"
)

const model = "gemini-2.0-flash-lite"

func New() *Controller {
	hpath, _ := xdg.DataFile("gencmd/history.jsonl")
	return &Controller{
		historyPath: hpath,
	}
}

type Controller struct {
	historyPath string
}

func (c *Controller) LoadHistory() []HistoryEntry {
	if c.historyPath == "" {
		return nil
	}
	file, err := os.Open(c.historyPath)
	if err != nil {
		return nil
	}
	defer file.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed entries
		}
		entries = append(entries, entry)
	}
	return entries
}

func (c *Controller) OutputCommand(command string) {
	if command != "" {
		fmt.Println(command)
	}
}

func (c *Controller) UpdateHistory(prompt, command string) error {
	if c.historyPath == "" {
		return fmt.Errorf("history path is not set")
	}
	file, err := os.OpenFile(c.historyPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("opening history file: %w", err)
	}
	defer file.Close()

	entry := HistoryEntry{Prompt: prompt, Command: command}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling history entry: %w", err)
	}
	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("writing to history file: %w", err)
	}
	return nil
}

func (c *Controller) GenerateCommands(prompt string) ([]string, error) {
	ctx := context.Background()
	client, err := genai.NewClient(ctx, &genai.ClientConfig{})
	if err != nil {
		return nil, fmt.Errorf("initializing GenAI client: %w", err)
	}

	content := []*genai.Content{
		{
			Parts: []*genai.Part{
				{
					Text: "You are a command line expert. Generate 5 shell command alternatives that implement the following description:\n" + prompt,
				},
			},
			Role: genai.RoleUser,
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
	res, err := client.Models.GenerateContent(ctx, model, content, config)
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

type HistoryEntry struct {
	Prompt  string `json:"prompt"`
	Command string `json:"command"`
}
