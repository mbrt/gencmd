package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

	"github.com/mbrt/gencmd/ctrl"
)

func TestNewPromptModel(t *testing.T) {
	km := DefaultKeyMap()
	history := []ctrl.HistoryEntry{
		{Prompt: "test prompt", Command: "test command"},
	}

	model := newPromptModel(km, history)

	if !model.historyVisible {
		t.Error("expected history to be visible initially")
	}

	if len(model.list.Items()) != len(history) {
		t.Errorf("expected %d history items, got %d", len(history), len(model.list.Items()))
	}

	if model.textInput.Value() != "" {
		t.Error("expected empty text input initially")
	}

	if model.textInput.Placeholder == "" {
		t.Error("expected placeholder text to be set")
	}
}

func TestPromptModelUpdate_WindowSizeMsg(t *testing.T) {
	km := DefaultKeyMap()
	model := newPromptModel(km, []ctrl.HistoryEntry{})

	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(windowMsg)

	if updatedModel.textInput.Width != 96 { // 100 - 4
		t.Errorf("expected text input width 96, got %d", updatedModel.textInput.Width)
	}
}

func TestPromptModelUpdate_KeyMsg(t *testing.T) {
	km := DefaultKeyMap()
	history := []ctrl.HistoryEntry{
		{Prompt: "first", Command: "cmd1"},
		{Prompt: "second", Command: "cmd2"},
	}
	model := newPromptModel(km, history)

	tests := []struct {
		name           string
		key            tea.KeyMsg
		expectCursor   int
		expectHistory  bool
		expectFiltered bool
	}{
		{
			name:          "up key moves cursor up",
			key:           tea.KeyMsg{Type: tea.KeyUp},
			expectCursor:  -1, // Should move up from 0
			expectHistory: true,
		},
		{
			name:          "down key moves cursor down",
			key:           tea.KeyMsg{Type: tea.KeyDown},
			expectCursor:  1,
			expectHistory: true,
		},
		{
			name:          "ctrl+h toggles history",
			key:           tea.KeyMsg{Type: tea.KeyCtrlH},
			expectHistory: false,
		},
		{
			name:           "typing text enables filtering",
			key:            tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'t'}},
			expectFiltered: true,
			expectHistory:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testModel := model
			if tt.name == "up key moves cursor up" {
				// Set cursor to 0 first so we can move up
				testModel.list.Select(0)
			}

			updatedModel, _ := testModel.Update(tt.key)

			if updatedModel.historyVisible != tt.expectHistory {
				t.Errorf("expected historyVisible %v, got %v", tt.expectHistory, updatedModel.historyVisible)
			}

			if tt.expectFiltered && !updatedModel.list.FilteringEnabled() {
				t.Error("expected filtering to be enabled")
			}
		})
	}
}

func TestPromptModelView(t *testing.T) {
	km := DefaultKeyMap()
	history := []ctrl.HistoryEntry{
		{Prompt: "test", Command: "ls"},
	}

	tests := []struct {
		name           string
		historyVisible bool
		hasHistory     bool
	}{
		{
			name:           "view with visible history",
			historyVisible: true,
			hasHistory:     true,
		},
		{
			name:           "view with hidden history",
			historyVisible: false,
			hasHistory:     true,
		},
		{
			name:           "view with no history",
			historyVisible: true,
			hasHistory:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var historyItems []ctrl.HistoryEntry
			if tt.hasHistory {
				historyItems = history
			}

			model := newPromptModel(km, historyItems)
			model.historyVisible = tt.historyVisible

			view := model.View()
			if view == "" {
				t.Error("view should not be empty")
			}

			// Check if history is shown when it should be
			if tt.hasHistory && tt.historyVisible {
				if !strings.Contains(view, "test") {
					t.Error("view should contain history items when visible")
				}
			}
		})
	}
}

func TestPromptModelShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	history := []ctrl.HistoryEntry{
		{Prompt: "test", Command: "ls"},
	}
	model := newPromptModel(km, history)

	tests := []struct {
		name           string
		historyVisible bool
		hasSelection   bool
		expectKeys     []string
	}{
		{
			name:           "with visible history and selection",
			historyVisible: true,
			hasSelection:   true,
			expectKeys:     []string{"enter", "esc", "↑/ctrl+k", "↓/ctrl+j", "ctrl+h"},
		},
		{
			name:           "with hidden history",
			historyVisible: false,
			hasSelection:   false,
			expectKeys:     []string{"enter", "esc", "ctrl+h"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.historyVisible = tt.historyVisible
			if tt.hasSelection {
				model.list.Select(0)
			}

			help := model.ShortHelp()

			// Check that essential keys are present
			hasSubmit := false
			hasCancel := false
			hasToggleHistory := false

			for _, binding := range help {
				helpKey := binding.Help().Key
				if helpKey == "enter" {
					hasSubmit = true
				}
				if helpKey == "esc" {
					hasCancel = true
				}
				if helpKey == "ctrl+h" {
					hasToggleHistory = true
				}
			}

			if !hasSubmit {
				t.Error("submit key should always be present")
			}
			if !hasCancel {
				t.Error("cancel key should always be present")
			}
			if !hasToggleHistory {
				t.Error("toggle history key should always be present")
			}
		})
	}
}

func TestPromptModelSelected(t *testing.T) {
	km := DefaultKeyMap()
	history := []ctrl.HistoryEntry{
		{Prompt: "history prompt", Command: "history command"},
	}
	model := newPromptModel(km, history)

	tests := []struct {
		name          string
		setupModel    func(*promptModel)
		expectPrompt  string
		expectCommand string
		expectIsNew   bool
	}{
		{
			name: "new prompt entered",
			setupModel: func(m *promptModel) {
				m.textInput.SetValue("new prompt")
				m.historyVisible = false
			},
			expectPrompt:  "new prompt",
			expectCommand: "",
			expectIsNew:   true,
		},
		{
			name: "history item selected",
			setupModel: func(m *promptModel) {
				m.historyVisible = true
				m.list.Select(0)
			},
			expectPrompt:  "history prompt",
			expectCommand: "history command",
			expectIsNew:   false,
		},
		{
			name: "no selection with hidden history",
			setupModel: func(m *promptModel) {
				m.historyVisible = false
				m.textInput.SetValue("")
			},
			expectPrompt:  "",
			expectCommand: "",
			expectIsNew:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testModel := model
			if tt.setupModel != nil {
				tt.setupModel(&testModel)
			}

			selected := testModel.Selected()

			if selected.Prompt != tt.expectPrompt {
				t.Errorf("expected prompt %q, got %q", tt.expectPrompt, selected.Prompt)
			}

			if selected.Command != tt.expectCommand {
				t.Errorf("expected command %q, got %q", tt.expectCommand, selected.Command)
			}

			if selected.IsNew() != tt.expectIsNew {
				t.Errorf("expected IsNew() %v, got %v", tt.expectIsNew, selected.IsNew())
			}
		})
	}
}

func TestPromptModelFilterItems(t *testing.T) {
	km := DefaultKeyMap()
	history := []ctrl.HistoryEntry{
		{Prompt: "list files", Command: "ls"},
		{Prompt: "find files", Command: "find"},
	}
	model := newPromptModel(km, history)

	tests := []struct {
		name           string
		query          string
		expectFiltered bool
	}{
		{
			name:           "empty query disables filtering",
			query:          "",
			expectFiltered: false,
		},
		{
			name:           "single character enables filtering",
			query:          "l",
			expectFiltered: true,
		},
		{
			name:           "multi character query",
			query:          "list",
			expectFiltered: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.filterItems(tt.query)

			if model.list.FilteringEnabled() != tt.expectFiltered {
				t.Errorf("expected filtering enabled %v, got %v", tt.expectFiltered, model.list.FilteringEnabled())
			}
		})
	}
}

func TestInputPromptMethods(t *testing.T) {
	tests := []struct {
		name        string
		prompt      inputPrompt
		expectEmpty bool
		expectNew   bool
	}{
		{
			name:        "empty prompt",
			prompt:      inputPrompt{},
			expectEmpty: true,
			expectNew:   false,
		},
		{
			name:        "new prompt only",
			prompt:      inputPrompt{Prompt: "test"},
			expectEmpty: false,
			expectNew:   true,
		},
		{
			name:        "existing command",
			prompt:      inputPrompt{Prompt: "test", Command: "ls"},
			expectEmpty: false,
			expectNew:   false,
		},
		{
			name:        "command only",
			prompt:      inputPrompt{Command: "ls"},
			expectEmpty: false,
			expectNew:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.prompt.Empty() != tt.expectEmpty {
				t.Errorf("expected Empty() %v, got %v", tt.expectEmpty, tt.prompt.Empty())
			}

			if tt.prompt.IsNew() != tt.expectNew {
				t.Errorf("expected IsNew() %v, got %v", tt.expectNew, tt.prompt.IsNew())
			}
		})
	}
}

func TestHistoryEntryMethods(t *testing.T) {
	entry := historyEntry{
		ctrl.HistoryEntry{
			Prompt:  "test prompt",
			Command: "test command",
		},
	}

	if entry.FilterValue() != "test prompt test command" {
		t.Errorf("expected filter value 'test prompt test command', got %q", entry.FilterValue())
	}

	if entry.Title() != "test prompt" {
		t.Errorf("expected title 'test prompt', got %q", entry.Title())
	}

	if entry.Description() != "test command" {
		t.Errorf("expected description 'test command', got %q", entry.Description())
	}
}

func TestPromptHandleKey(t *testing.T) {
	km := DefaultKeyMap()
	model := newPromptModel(km, []ctrl.HistoryEntry{
		{Prompt: "test", Command: "ls"},
	})

	// Test that unknown keys are passed to text input
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, _ := model.handleKey(keyMsg)

	// The text input should have received the key
	if updatedModel.textInput.Value() != "a" {
		t.Errorf("expected text input value 'a', got %q", updatedModel.textInput.Value())
	}
}
