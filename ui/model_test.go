package ui

import (
	"errors"
	"testing"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/stretchr/testify/assert"

	"github.com/mbrt/gencmd/ctrl"
)

// TestCompleteUIWorkflow tests the complete user workflow from start to finish
func TestCompleteUIWorkflow(t *testing.T) {
	tests := []struct {
		name           string
		setupHistory   []ctrl.HistoryEntry
		commands       []string
		userActions    []userAction
		expectedResult workflowResult
	}{
		{
			name: "new prompt with single command",
			setupHistory: []ctrl.HistoryEntry{
				{Prompt: "old prompt", Command: "old command"},
			},
			commands: []string{"ls -l"},
			userActions: []userAction{
				{action: typeText, value: "list files"},
				{action: pressKey, key: tea.KeyEnter},
			},
			expectedResult: workflowResult{
				selectedCommand: "ls -l",
				finalState:      stateSelected,
			},
		},
		{
			name:     "new prompt with multiple commands",
			commands: []string{"ls -l", "ls -la", "find . -name '*.txt'"},
			userActions: []userAction{
				{action: typeText, value: "list files"},
				{action: pressKey, key: tea.KeyEnter},
				{action: pressKey, key: tea.KeyDown},
				{action: pressKey, key: tea.KeyEnter},
			},
			expectedResult: workflowResult{
				selectedCommand: "ls -la",
				finalState:      stateSelected,
			},
		},
		{
			name: "select from history",
			setupHistory: []ctrl.HistoryEntry{
				{Prompt: "list files", Command: "ls -l"},
				{Prompt: "find files", Command: "find . -name '*.txt'"},
			},
			userActions: []userAction{
				{action: pressKey, key: tea.KeyDown},
				{action: pressKey, key: tea.KeyEnter},
			},
			expectedResult: workflowResult{
				selectedCommand: "find . -name '*.txt'",
				finalState:      stateSelected,
			},
		},
		{
			name: "user cancellation",
			userActions: []userAction{
				{action: typeText, value: "test prompt"},
				{action: pressKey, key: tea.KeyEsc},
			},
			expectedResult: workflowResult{
				finalState: stateSelected,
				error:      "cancelled",
			},
		},
		{
			name:     "empty command generation",
			commands: []string{}, // No commands generated
			userActions: []userAction{
				{action: typeText, value: "invalid prompt"},
				{action: pressKey, key: tea.KeyEnter},
			},
			expectedResult: workflowResult{
				finalState: stateSelected,
				error:      "no commands",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup controller with test data
			controller := &FakeController{
				history:  tt.setupHistory,
				commands: tt.commands,
			}

			model := New(controller)

			// Execute user actions
			for _, action := range tt.userActions {
				model = executeAction(t, model, action)
			}

			// Verify final state
			assert.Equal(t, tt.expectedResult.finalState, model.state)

			if tt.expectedResult.error == "" {
				assert.NoError(t, model.err)
			} else {
				assert.ErrorContains(t, model.err, tt.expectedResult.error)
			}

			assert.Equal(t, tt.expectedResult.selectedCommand, model.selected)
		})
	}
}

// TestUIStateTransitions tests that the UI transitions between states correctly
func TestUIStateTransitions(t *testing.T) {
	controller := &FakeController{
		history:  []ctrl.HistoryEntry{{Prompt: "test", Command: "echo test"}},
		commands: []string{"ls -l", "ls -la"},
	}

	model := New(controller)

	// Initial state should be prompting
	assert.Equal(t, statePrompting, model.state)

	// Type a new prompt
	model = typeTextIntoModel(model, "new prompt")

	// Press enter to generate commands
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Should transition to selecting state (multiple commands)
	assert.Equal(t, stateSelecting, model.state)

	// Select a command
	model = updateModel(model, tea.KeyMsg{Type: tea.KeyEnter})

	// Should transition to selected state
	assert.Equal(t, stateSelected, model.state)
}

// TestUIErrorHandling tests error scenarios
func TestUIErrorHandling(t *testing.T) {
	tests := []struct {
		name          string
		controllerErr error
		expectedErr   error
	}{
		{
			name:          "controller generation error",
			controllerErr: errors.New("API error"),
			expectedErr:   errors.New("API error"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			controller := &FakeController{
				history:     []ctrl.HistoryEntry{},
				commands:    []string{},
				generateErr: tt.controllerErr,
			}

			model := New(controller)

			// Type prompt and submit
			model = typeTextIntoModel(model, "test prompt")
			model = updateModel(model, tea.KeyMsg{Type: tea.KeyEnter})

			// Should be in selected state with error
			assert.Equal(t, stateSelected, model.state)
			assert.ErrorContains(t, model.err, tt.expectedErr.Error())
		})
	}
}

// TestUIViewRendering tests that views render correctly in different states
func TestUIViewRendering(t *testing.T) {
	controller := NewFakeController()
	model := New(controller)

	states := []struct {
		name  string
		state state
		setup func(*Model)
	}{
		{
			name:  "prompting state",
			state: statePrompting,
			setup: func(m *Model) { m.state = statePrompting },
		},
		{
			name:  "generating state",
			state: stateGenerating,
			setup: func(m *Model) { m.state = stateGenerating },
		},
		{
			name:  "selecting state",
			state: stateSelecting,
			setup: func(m *Model) {
				m.state = stateSelecting
				m.selectCmp.SetItems([]string{"cmd1", "cmd2"})
			},
		},
		{
			name:  "error state",
			state: stateSelected,
			setup: func(m *Model) {
				m.state = stateSelected
				m.err = errors.New("test error")
			},
		},
	}

	for _, tt := range states {
		t.Run(tt.name, func(t *testing.T) {
			testModel := model
			tt.setup(&testModel)

			view := testModel.View()
			assert.NotEmpty(t, view)

			// All views should contain title unless there's an error
			if testModel.err == nil {
				assert.Contains(t, view, "gencmd")
			}

			// Error views should show error message
			if testModel.err != nil {
				assert.Contains(t, view, "Error:")
			}
		})
	}
}

// TestDeleteHistoryUI tests the UI functionality for deleting history entries
func TestDeleteHistoryUI(t *testing.T) {
	t.Run("delete selected history entry", func(t *testing.T) {
		controller := &FakeController{
			history: []ctrl.HistoryEntry{
				{Prompt: "p1", Command: "c1"},
				{Prompt: "p2", Command: "c2"},
				{Prompt: "p3", Command: "c3"},
			},
		}

		model := New(controller)

		// Navigate to second entry (index 1)
		model = updateModel(model, tea.KeyMsg{Type: tea.KeyDown})
		// Delete
		model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlD})

		// Verify the result
		remainingHistory := controller.LoadHistory()
		assert.Equal(t, []ctrl.HistoryEntry{
			{Prompt: "p1", Command: "c1"},
			{Prompt: "p3", Command: "c3"},
		}, remainingHistory)
	})

	t.Run("delete with history not visible - should be no-op", func(t *testing.T) {
		controller := &FakeController{
			history: []ctrl.HistoryEntry{
				{Prompt: "p1", Command: "c1"},
				{Prompt: "p2", Command: "c2"},
			},
		}

		model := New(controller)

		// Hide history
		model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlN})
		// Try to simulate delete - should have no effect
		model = updateModel(model, tea.KeyMsg{Type: tea.KeyCtrlD})

		// Verify history is unchanged
		remainingHistory := controller.LoadHistory()
		assert.Equal(t, 2, len(remainingHistory))
	})
}

// Test helper types and functions

type actionType int

const (
	typeText actionType = iota
	pressKey
)

type userAction struct {
	action actionType
	value  string
	key    tea.KeyType
}

type workflowResult struct {
	selectedCommand string
	finalState      state
	error           string
}

func executeAction(t *testing.T, model Model, action userAction) Model {
	switch action.action {
	case typeText:
		return typeTextIntoModel(model, action.value)
	case pressKey:
		return updateModel(model, tea.KeyMsg{Type: action.key})
	default:
		t.Fatalf("unknown action type: %v", action.action)
		return model
	}
}

func typeTextIntoModel(model Model, text string) Model {
	for _, char := range text {
		updatedModel, _ := model.Update(tea.KeyMsg{
			Type:  tea.KeyRunes,
			Runes: []rune{char},
		})
		model = updatedModel.(Model)
	}
	return model
}

func updateModel(m Model, msg tea.Msg) Model {
	if msg == nil || msg == tea.Quit() {
		return m
	}
	// Handle batch messages
	if msg, ok := msg.(tea.BatchMsg); ok {
		for _, cmd := range msg {
			m = updateModel(m, cmd())
		}
		return m
	}
	// Update the model
	um, cmd := m.Update(msg)
	m = um.(Model)
	if cmd == nil {
		return m
	}
	return updateModel(m, cmd())
}
