package ctrl

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"

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
	entries := c.loadHistoryRaw()

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
	if c.historyPath == "" || c.rejectedPath == "" {
		return fmt.Errorf("history or rejected paths not set")
	}

	// First, log the deleted entry to rejected.jsonl
	if err := c.logRejected(entry); err != nil {
		return fmt.Errorf("logging rejected entry: %w", err)
	}

	// Read current history (if file exists)
	entries := c.loadHistoryRaw()
	if len(entries) == 0 {
		return nil
	}

	// Remove all instances of the entry
	entries = slices.DeleteFunc(entries, func(e HistoryEntry) bool {
		return e == entry
	})

	// Rewrite history file
	return c.rewriteHistory(entries)
}

func (c *Controller) GenerateCommands(prompt string) ([]string, error) {
	ctx := context.Background()
	model, err := NewModel(ctx, c.cfg.LLM)
	if err != nil {
		return nil, fmt.Errorf("creating model: %w", err)
	}
	return model.GenerateCommands(ctx, prompt)
}

func (c *Controller) loadHistoryRaw() []HistoryEntry {
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

func (c *Controller) rewriteHistory(entries []HistoryEntry) error {
	dir := filepath.Dir(c.historyPath)
	tmp, err := os.CreateTemp(dir, "tmp-history-*.jsonl")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // Clean up temp file

	// Write entries to the temporary file
	for _, entry := range entries {
		data, err := json.Marshal(entry)
		if err != nil {
			_ = tmp.Close()
			return fmt.Errorf("marshalling history entry: %w", err)
		}
		if _, err := tmp.WriteString(string(data) + "\n"); err != nil {
			_ = tmp.Close()
			return fmt.Errorf("writing to tmp history file: %w", err)
		}
	}

	// Flush
	if err := tmp.Sync(); err != nil {
		_ = tmp.Close()
		return fmt.Errorf("syncing tmp history file: %w", err)
	}
	if err := tmp.Close(); err != nil {
		return err
	}

	// Rename tmp into the new file (which is atomic)
	return os.Rename(tmp.Name(), c.historyPath)
}

func (c *Controller) logRejected(entry HistoryEntry) error {
	f, err := os.OpenFile(c.rejectedPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o600)
	if err != nil {
		return fmt.Errorf("opening rejected file: %w", err)
	}
	defer f.Close()

	data, err := json.Marshal(entry)
	if err != nil {
		return fmt.Errorf("marshalling rejected entry: %w", err)
	}
	if _, err := f.WriteString(string(data) + "\n"); err != nil {
		return fmt.Errorf("writing to rejected file: %w", err)
	}
	return nil
}

type HistoryEntry struct {
	Prompt  string `json:"prompt"`
	Command string `json:"command"`
}
