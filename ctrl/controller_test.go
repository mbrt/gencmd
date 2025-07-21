package ctrl

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadHistory(t *testing.T) {
	// Create a temporary history file
	historyPath := filepath.Join(t.TempDir(), "history.jsonl")

	// Write some history entries with duplicates
	entries := []HistoryEntry{
		{Prompt: "p1", Command: "c1"},
		{Prompt: "p2", Command: "c2"},
		{Prompt: "p1", Command: "c1"}, // Duplicate
		{Prompt: "p3", Command: "c3"},
	}

	// Create a controller and write the entries to the history file
	controller := &Controller{historyPath: historyPath}
	for _, entry := range entries {
		err := controller.UpdateHistory(entry.Prompt, entry.Command)
		require.NoError(t, err)
	}

	// Load the history
	loadedEntries := controller.LoadHistory()

	// Check that duplicates are removed and the order is correct
	expectedEntries := []HistoryEntry{
		{Prompt: "p3", Command: "c3"},
		{Prompt: "p1", Command: "c1"},
		{Prompt: "p2", Command: "c2"},
	}
	assert.Equal(t, expectedEntries, loadedEntries)
}
