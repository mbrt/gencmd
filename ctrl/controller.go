package ctrl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/adrg/xdg"

	"github.com/mbrt/gencmd/config"
)

func New(cfg config.Config) *Controller {
	hpath, _ := xdg.DataFile("gencmd/history.jsonl")
	return &Controller{
		historyPath: hpath,
		cfg:         cfg,
	}
}

type Controller struct {
	historyPath string
	cfg         config.Config
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
	// Reverse the order to have the most recent entries first.
	for i, j := 0, len(entries)-1; i < j; i, j = i+1, j-1 {
		entries[i], entries[j] = entries[j], entries[i]
	}
	return entries
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
	model, err := NewModel(ctx, c.cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("creating model: %w", err)
	}
	return model.GenerateCommands(ctx, prompt)
}

type HistoryEntry struct {
	Prompt  string `json:"prompt"`
	Command string `json:"command"`
}
