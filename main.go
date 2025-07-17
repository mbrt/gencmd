package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/adrg/xdg"
	"github.com/joho/godotenv"
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

func readChoiceFromFzf(choices []string) (string, error) {
	cmd := exec.Command("fzf", "--print-query")
	stdin, _ := cmd.StdinPipe()
	go func() {
		for _, choice := range choices {
			fmt.Fprintln(stdin, choice)
		}
		stdin.Close()
	}()

	out, err := cmd.Output()
	if err != nil {
		return "", fmt.Errorf("running fzf: %w", err)
	}
	// fzf with --print-query returns the query on the first line and the
	// selection on the second line
	lines_out := strings.Split(strings.TrimSpace(string(out)), "\n")
	if len(lines_out) == 0 {
		return "", fmt.Errorf("no selection made")
	}

	// If there's only one line, it means no selection was made. Use the query
	if len(lines_out) == 1 {
		return lines_out[0], nil
	}
	// If there are two lines, return the selection.
	return lines_out[1], nil
}

func readFromFzf() (string, bool, error) {
	// Load the history from the JSONL file. On error, ignore
	entries, _ := loadHistory()

	if len(entries) == 0 {
		// If no history is available, prompt the user for a new command
		return "", false, fmt.Errorf("no history available, please enter a new command")
	}

	// Build fzf input
	var lines []string
	for _, e := range entries {
		lines = append(lines, fmt.Sprintf("%s -> %s", e.Prompt, e.Command))
	}
	selection, err := readChoiceFromFzf(lines)
	if err != nil {
		return "", false, err
	}
	// Try to find the selected line in the history
	for i, line := range lines {
		if selection == line {
			return entries[i].Command, false, nil
		}
	}
	// If selection doesn't match any history entry, use the query as a new
	// prompt
	return selection, true, nil
}

func readCommand() (prompt string, isNew bool, err error) {
	res, isNew, err := readFromFzf()
	if err == nil {
		return res, isNew, nil
	}

	// Fallback to reading from stdin
	fmt.Print("Enter command: ")
	scanner := bufio.NewScanner(os.Stdin)
	if scanner.Scan() {
		return scanner.Text(), true, nil
	}
	if err := scanner.Err(); err != nil {
		return "", false, fmt.Errorf("reading command from stdin: %w", err)
	}
	return "", false, fmt.Errorf("no command entered")
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
