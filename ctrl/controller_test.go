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

func TestDeleteHistory(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.jsonl")
	rejectedPath := filepath.Join(tempDir, "rejected.jsonl")

	// Create a controller
	controller := &Controller{
		historyPath:  historyPath,
		rejectedPath: rejectedPath,
	}

	// Add some initial entries
	entries := []HistoryEntry{
		{Prompt: "p1", Command: "c1"},
		{Prompt: "p2", Command: "c2"},
		{Prompt: "p3", Command: "c3"},
	}

	for _, entry := range entries {
		err := controller.UpdateHistory(entry.Prompt, entry.Command)
		require.NoError(t, err)
	}

	// Load initial history to verify it's there
	initialHistory := controller.LoadHistory()
	require.Equal(t, []HistoryEntry{
		{Prompt: "p3", Command: "c3"},
		{Prompt: "p2", Command: "c2"},
		{Prompt: "p1", Command: "c1"},
	}, initialHistory)

	// Delete the middle entry
	entryToDelete := HistoryEntry{Prompt: "p2", Command: "c2"}
	err := controller.DeleteHistory(entryToDelete)
	require.NoError(t, err)

	// Verify the entry was removed from history
	updatedHistory := controller.LoadHistory()
	require.Equal(t, []HistoryEntry{
		{Prompt: "p3", Command: "c3"},
		{Prompt: "p1", Command: "c1"},
	}, updatedHistory)

	// Verify the deleted entry was logged to rejected.jsonl
	rejectedController := &Controller{historyPath: rejectedPath}
	rejected := rejectedController.LoadHistory()
	assert.Equal(t, []HistoryEntry{{Prompt: "p2", Command: "c2"}}, rejected)
}

func TestDeleteHistory_NonExistentEntry(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.jsonl")
	rejectedPath := filepath.Join(tempDir, "rejected.jsonl")

	// Create a controller
	controller := &Controller{
		historyPath:  historyPath,
		rejectedPath: rejectedPath,
	}

	// Add some initial entries
	entries := []HistoryEntry{
		{Prompt: "p1", Command: "c1"},
		{Prompt: "p2", Command: "c2"},
	}

	for _, entry := range entries {
		err := controller.UpdateHistory(entry.Prompt, entry.Command)
		require.NoError(t, err)
	}

	// Try to delete an entry that doesn't exist
	nonExistentEntry := HistoryEntry{Prompt: "nonexistent", Command: "cmd"}
	err := controller.DeleteHistory(nonExistentEntry)
	require.NoError(t, err) // Should not error, just be a no-op for history

	// Verify history is unchanged
	history := controller.LoadHistory()
	require.Len(t, history, 2)

	// Verify the non-existent entry was still logged to rejected.jsonl
	rejectedController := &Controller{historyPath: rejectedPath}
	rejectedEntries := rejectedController.LoadHistory()
	require.Len(t, rejectedEntries, 1)
	assert.Equal(t, nonExistentEntry, rejectedEntries[0])
}

func TestDeleteHistory_EmptyHistory(t *testing.T) {
	// Create temporary files
	tempDir := t.TempDir()
	historyPath := filepath.Join(tempDir, "history.jsonl")
	rejectedPath := filepath.Join(tempDir, "rejected.jsonl")

	// Create a controller with empty history
	controller := &Controller{
		historyPath:  historyPath,
		rejectedPath: rejectedPath,
	}

	// Try to delete from empty history
	entryToDelete := HistoryEntry{Prompt: "p1", Command: "c1"}
	err := controller.DeleteHistory(entryToDelete)
	require.NoError(t, err)

	// Verify history remains empty
	history := controller.LoadHistory()
	assert.Len(t, history, 0)

	// Verify the entry was logged to rejected.jsonl
	rejectedController := &Controller{historyPath: rejectedPath}
	rejected := rejectedController.LoadHistory()
	assert.Equal(t, []HistoryEntry{{Prompt: "p1", Command: "c1"}}, rejected)
}
