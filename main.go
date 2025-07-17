package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/adrg/xdg"
	"github.com/joho/godotenv"
	"github.com/mbrt/gencmd/ui"
	"google.golang.org/genai"
)

const model = "gemini-2.0-flash-lite"

func loadEnv() {
	// Load the environment variables from the XDG configuration directory or
	// the default .env file in the current directory. Ignore errors, so that
	// the program can still run if the file is not found.
	envPath, err := xdg.ConfigFile("gencmd/.env")
	if err == nil {
		_ = godotenv.Load(envPath)
	}
	_ = godotenv.Load()
}

type historyEntry struct {
	Prompt  string `json:"prompt"`
	Command string `json:"command"`
}

func loadHistory() ([]historyEntry, error) {
	histPath, err := xdg.DataFile("gencmd/history.jsonl")
	if err != nil {
		return nil, err
	}
	file, err := os.Open(histPath)
	if err != nil {
		return nil, fmt.Errorf("opening history file: %w", err)
	}
	defer file.Close()

	var entries []historyEntry
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		var entry historyEntry
		if err := json.Unmarshal(scanner.Bytes(), &entry); err != nil {
			continue // Skip malformed entries
		}
		entries = append(entries, entry)
	}
	return entries, nil
}

func updateHistory(prompt, command string) error {
	histPath, err := xdg.DataFile("gencmd/history.jsonl")
	if err != nil {
		return fmt.Errorf("getting history file path: %w", err)
	}
	file, err := os.OpenFile(histPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("opening history file: %w", err)
	}
	defer file.Close()

	entry := historyEntry{Prompt: prompt, Command: command}
	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling history entry: %w", err)
	}
	if _, err := file.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("writing to history file: %w", err)
	}
	return nil
}

func readCommand() (prompt string, isNew bool, err error) {
	entries, err := loadHistory()
	if err != nil {
		// Ignore errors, but we won't have history
		entries = []historyEntry{}
	}

	// Convert to the UI's history entry type
	uiEntries := make([]ui.HistoryEntry, len(entries))
	for i, e := range entries {
		uiEntries[i] = ui.HistoryEntry{
			Prompt:  e.Prompt,
			Command: e.Command,
		}
	}

	return ui.RunUI(uiEntries)
}

func generateCommands(prompt string) ([]string, error) {
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

func run() error {
	prompt, isNew, err := readCommand()
	if err != nil {
		return fmt.Errorf("reading command: %w", err)
	}
	if !isNew {
		fmt.Println(prompt)
		return nil
	}
	cmds, err := generateCommands(prompt)
	if err != nil {
		return fmt.Errorf("generating commands: %w", err)
	}
	updateHistory(prompt, cmds[0])
	fmt.Printf("Generated commands: %v\n", cmds)
	return nil
}

func main() {
	loadEnv()
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
