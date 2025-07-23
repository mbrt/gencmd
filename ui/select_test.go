package ui

import (
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
)

func TestNewSelectModel(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	if len(model.list.Items()) != 0 {
		t.Error("expected empty list initially")
	}

	if model.list.ShowTitle() {
		t.Error("expected title to be hidden")
	}

	if model.list.FilteringEnabled() {
		t.Error("expected filtering to be disabled")
	}

	if model.list.ShowHelp() {
		t.Error("expected help to be hidden")
	}

	if model.list.ShowStatusBar() {
		t.Error("expected status bar to be hidden")
	}
}

func TestSelectModelUpdate_WindowSizeMsg(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	windowMsg := tea.WindowSizeMsg{Width: 100, Height: 50}
	updatedModel, _ := model.Update(windowMsg)

	// We can't directly check the list width/height due to private fields,
	// but we can verify the update doesn't break anything
	if len(updatedModel.list.Items()) != 0 {
		t.Error("window resize should not affect items")
	}
}

func TestSelectModelUpdate_KeyMsg(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)
	model.SetItems([]string{"first", "second", "third"})

	tests := []struct {
		name string
		key  tea.KeyMsg
	}{
		{
			name: "up key",
			key:  tea.KeyMsg{Type: tea.KeyUp},
		},
		{
			name: "down key",
			key:  tea.KeyMsg{Type: tea.KeyDown},
		},
		{
			name: "ctrl+k",
			key:  tea.KeyMsg{Type: tea.KeyCtrlK},
		},
		{
			name: "ctrl+j",
			key:  tea.KeyMsg{Type: tea.KeyCtrlJ},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			updatedModel, cmd := model.Update(tt.key)

			// Should not return any commands for navigation
			if cmd != nil {
				t.Error("navigation keys should not return commands")
			}

			// Model should still have the same items
			if len(updatedModel.list.Items()) != 3 {
				t.Error("navigation should not affect items")
			}
		})
	}
}

func TestSelectModelUpdate_OtherKeyMsg(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	// Test that other keys are ignored
	keyMsg := tea.KeyMsg{Type: tea.KeyRunes, Runes: []rune{'a'}}
	updatedModel, cmd := model.Update(keyMsg)

	if cmd != nil {
		t.Error("unknown keys should not return commands")
	}

	if len(updatedModel.list.Items()) != 0 {
		t.Error("unknown keys should not affect model")
	}
}

func TestSelectModelView(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	tests := []struct {
		name  string
		items []string
	}{
		{
			name:  "empty list",
			items: []string{},
		},
		{
			name:  "single item",
			items: []string{"ls -l"},
		},
		{
			name:  "multiple items",
			items: []string{"ls -l", "ls -la", "ls -lh"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.SetItems(tt.items)
			view := model.View()

			if view == "" {
				t.Error("view should not be empty")
			}

			// Should contain the header message
			if !strings.Contains(view, "Select completion") {
				t.Error("view should contain selection message")
			}

			// If items exist, they should be in the view
			for _, item := range tt.items {
				if !strings.Contains(view, item) {
					t.Errorf("view should contain item %q", item)
				}
			}
		})
	}
}

func TestSelectModelShortHelp(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	help := model.ShortHelp()

	expectedKeys := []string{"enter", "esc", "↑/ctrl+k", "↓/ctrl+j"}

	if len(help) != len(expectedKeys) {
		t.Errorf("expected %d help bindings, got %d", len(expectedKeys), len(help))
	}

	// Check that all expected keys are present
	foundKeys := make(map[string]bool)
	for _, binding := range help {
		foundKeys[binding.Help().Key] = true
	}

	for _, expectedKey := range expectedKeys {
		if !foundKeys[expectedKey] {
			t.Errorf("expected to find key %q in help", expectedKey)
		}
	}
}

func TestSelectModelSetItems(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	items := []string{"cmd1", "cmd2", "cmd3"}
	model.SetItems(items)

	if len(model.list.Items()) != len(items) {
		t.Errorf("expected %d items, got %d", len(items), len(model.list.Items()))
	}

	// Verify each item was added correctly
	for i, expectedItem := range items {
		listItem := model.list.Items()[i]
		shortItem, ok := listItem.(shortItem)
		if !ok {
			t.Errorf("expected shortItem, got %T", listItem)
			continue
		}
		if string(shortItem) != expectedItem {
			t.Errorf("expected item %q, got %q", expectedItem, string(shortItem))
		}
	}
}

func TestSelectModelSelected(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	tests := []struct {
		name     string
		items    []string
		selected int
		expected string
	}{
		{
			name:     "empty list",
			items:    []string{},
			selected: 0,
			expected: "",
		},
		{
			name:     "first item selected",
			items:    []string{"cmd1", "cmd2", "cmd3"},
			selected: 0,
			expected: "cmd1",
		},
		{
			name:     "middle item selected",
			items:    []string{"cmd1", "cmd2", "cmd3"},
			selected: 1,
			expected: "cmd2",
		},
		{
			name:     "last item selected",
			items:    []string{"cmd1", "cmd2", "cmd3"},
			selected: 2,
			expected: "cmd3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model.SetItems(tt.items)
			if len(tt.items) > 0 {
				model.list.Select(tt.selected)
			}

			selected := model.Selected()
			if selected != tt.expected {
				t.Errorf("expected selected %q, got %q", tt.expected, selected)
			}
		})
	}
}

func TestShortItem(t *testing.T) {
	item := shortItem("test command")

	// Test FilterValue method
	if item.FilterValue() != "" {
		t.Errorf("expected empty filter value, got %q", item.FilterValue())
	}
}

func TestShortItemDelegate(t *testing.T) {
	delegate := shortItemDelegate{}

	// Test basic properties
	if delegate.Height() != 1 {
		t.Errorf("expected height 1, got %d", delegate.Height())
	}

	if delegate.Spacing() != 0 {
		t.Errorf("expected spacing 0, got %d", delegate.Spacing())
	}

	// Test Update method (should return nil)
	cmd := delegate.Update(nil, nil)
	if cmd != nil {
		t.Error("Update should return nil")
	}
}

func TestSelectModelInit(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	cmd := model.Init()
	if cmd != nil {
		t.Error("Init should return nil")
	}
}

func TestSelectModelSetItemsEmpty(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	// Test setting empty items
	model.SetItems([]string{})

	if len(model.list.Items()) != 0 {
		t.Error("expected empty items list")
	}

	selected := model.Selected()
	if selected != "" {
		t.Error("expected empty selection")
	}
}

func TestSelectModelSetItemsOverwrite(t *testing.T) {
	km := DefaultKeyMap()
	model := newSelectModel(km)

	// Set initial items
	model.SetItems([]string{"cmd1", "cmd2"})
	if len(model.list.Items()) != 2 {
		t.Error("expected 2 items initially")
	}

	// Overwrite with new items
	model.SetItems([]string{"new1", "new2", "new3"})
	if len(model.list.Items()) != 3 {
		t.Error("expected 3 items after overwrite")
	}

	// Verify the new items are correct
	firstItem := model.list.Items()[0].(shortItem)
	if string(firstItem) != "new1" {
		t.Errorf("expected first item 'new1', got %q", string(firstItem))
	}
}
