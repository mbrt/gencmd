package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewWaitModel(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	// Test that the model is initialized properly
	// Note: KeyMap contains non-comparable types, so we check functionality instead

	// Test that spinner is initialized
	view := model.View()
	if view == "" {
		t.Error("view should not be empty")
	}

	// Should contain the waiting message
	if !strings.Contains(view, "Generating commands...") {
		t.Error("view should contain generating message")
	}
}

func TestWaitModelInit(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	cmd := model.Init()
	if cmd == nil {
		t.Error("Init should return spinner tick command")
	}
}

func TestWaitModelUpdate(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	tests := []struct {
		name string
		msg  tea.Msg
	}{
		{
			name: "window size message",
			msg:  tea.WindowSizeMsg{Width: 100, Height: 50},
		},
		{
			name: "key message",
			msg:  tea.KeyMsg{Type: tea.KeyEsc},
		},
		{
			name: "generic message",
			msg:  "test message",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, cmd := model.Update(tt.msg)

			// The model should be updated (spinner might tick)
			// Note: KeyMap contains non-comparable types, so we verify functionality

			// Some messages might return spinner tick commands
			// We don't check the specific command since it depends on internal spinner state
			_ = cmd
		})
	}
}

func TestWaitModelView(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	view := model.View()

	// Check basic structure
	if view == "" {
		t.Error("view should not be empty")
	}

	// Should start with newline
	if view[0] != '\n' {
		t.Error("view should start with newline")
	}

	// Should contain the generating message
	if !strings.Contains(view, "Generating commands...") {
		t.Error("view should contain 'Generating commands...'")
	}

	// Should end with newline
	if view[len(view)-1] != '\n' {
		t.Error("view should end with newline")
	}
}

func TestWaitModelShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	help := model.ShortHelp()

	// Should only have cancel key available during waiting
	if len(help) != 1 {
		t.Errorf("expected 1 help binding, got %d", len(help))
	}

	binding := help[0]
	if binding.Help().Key != "esc" {
		t.Errorf("expected cancel key 'esc', got %q", binding.Help().Key)
	}

	if binding.Help().Desc != "cancel" {
		t.Errorf("expected cancel description 'cancel', got %q", binding.Help().Desc)
	}
}

func TestWaitModelSpinnerUpdate(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	// Simulate spinner tick - this is internal to the spinner
	// We can't directly test the spinner animation, but we can ensure
	// that updates don't break the model
	for i := 0; i < 5; i++ {
		updatedModel, _ := model.Update(nil)
		view := updatedModel.View()

		if view == "" {
			t.Errorf("view should not be empty after update %d", i)
		}

		if !strings.Contains(view, "Generating commands...") {
			t.Errorf("view should still contain message after update %d", i)
		}

		model = updatedModel
	}
}

func TestWaitModelConsistentView(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	// Get initial view
	view1 := model.View()

	// Update the model (this might advance spinner)
	updatedModel, _ := model.Update(nil)
	view2 := updatedModel.View()

	// Both views should contain the same text (spinner might be different)
	if !strings.Contains(view1, "Generating commands...") {
		t.Error("initial view should contain generating message")
	}

	if !strings.Contains(view2, "Generating commands...") {
		t.Error("updated view should contain generating message")
	}

	// Both should have same structure (newlines at start/end)
	if view1[0] != '\n' || view2[0] != '\n' {
		t.Error("views should start with newline")
	}

	if view1[len(view1)-1] != '\n' || view2[len(view2)-1] != '\n' {
		t.Error("views should end with newline")
	}
}

func TestWaitModelKeyMapFunctionality(t *testing.T) {
	km := DefaultKeyMap()
	model := newWaitModel(km)

	// Verify that help functionality works consistently after updates
	initialHelp := model.ShortHelp()

	// Update with various messages
	messages := []tea.Msg{
		tea.WindowSizeMsg{Width: 80, Height: 24},
		tea.KeyMsg{Type: tea.KeyEsc},
		"random message",
	}

	for _, msg := range messages {
		updatedModel, _ := model.Update(msg)

		updatedHelp := updatedModel.ShortHelp()
		if len(updatedHelp) != len(initialHelp) {
			t.Error("help should remain consistent after updates")
		}

		model = updatedModel
	}
}
