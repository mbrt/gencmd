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
	rpath, _ := xdg.DataFile("gencmd/rejected.jsonl")
	return &Controller{
		historyPath:  hpath,
		rejectedPath: rpath,
		cfg:          cfg,
	}
}

type Controller struct {
	historyPath  string
	rejectedPath string
	cfg          config.Config
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

	// Remove duplicates, keeping the most recent entry.
	seen := make(map[HistoryEntry]bool)
	var result []HistoryEntry
	for _, entry := range entries {
		if !seen[entry] {
			result = append(result, entry)
			seen[entry] = true
		}
	}

	return result
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

func (c *Controller) DeleteHistory(entry HistoryEntry) error {
	if c.historyPath == "" {
		return fmt.Errorf("history path is not set")
	}
	if c.rejectedPath == "" {
		return fmt.Errorf("rejected path is not set")
	}

	// First, log the deleted entry to rejected.jsonl
	rejectedFile, err := os.OpenFile(c.rejectedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("opening rejected file: %w", err)
	}
	defer rejectedFile.Close()

	rejectedData, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling rejected entry: %w", err)
	}
	if _, err := rejectedFile.WriteString(string(rejectedData) + "\n"); err != nil {
		return fmt.Errorf("writing to rejected file: %w", err)
	}

	// Read current history (if file exists)
	historyFile, err := os.Open(c.historyPath)
	if err != nil {
		if os.IsNotExist(err) {
			// History file doesn't exist, nothing to delete from
			// Still log to rejected file though
			return nil
		}
		return fmt.Errorf("opening history file for reading: %w", err)
	}
	defer historyFile.Close()

	var entries []HistoryEntry
	scanner := bufio.NewScanner(historyFile)
	for scanner.Scan() {
		var histEntry HistoryEntry
		if err := json.Unmarshal(scanner.Bytes(), &histEntry); err != nil {
			continue // Skip malformed entries
		}
		// Keep entries that don't match the one to delete
		if histEntry != entry {
			entries = append(entries, histEntry)
		}
	}
	if err := scanner.Err(); err != nil {
		return fmt.Errorf("reading history file: %w", err)
	}

	// Rewrite history file without the deleted entry
	newHistoryFile, err := os.OpenFile(c.historyPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0o600)
	if err != nil {
		return fmt.Errorf("opening history file for writing: %w", err)
	}
	defer newHistoryFile.Close()

	for _, histEntry := range entries {
		data, err := json.Marshal(histEntry)
		if err != nil {
			return fmt.Errorf("marshalling history entry: %w", err)
		}
		if _, err := newHistoryFile.WriteString(string(data) + "\n"); err != nil {
			return fmt.Errorf("writing to history file: %w", err)
		}
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
