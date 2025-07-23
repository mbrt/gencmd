package ui

import (
	"errors"
	"strings"
	"testing"

	tea "github.com/charmbracelet/bubbletea"

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
				noError:         true,
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
				noError:         true,
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
				noError:         true,
			},
		},
		{
			name: "user cancellation",
			userActions: []userAction{
				{action: typeText, value: "test prompt"},
				{action: pressKey, key: tea.KeyEsc},
			},
			expectedResult: workflowResult{
				selectedCommand: "",
				finalState:      stateSelected,
				expectError:     ErrUserCancel,
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
				selectedCommand: "",
				finalState:      stateSelected,
				hasError:        true,
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
			if model.state != tt.expectedResult.finalState {
				t.Errorf("expected final state %v, got %v", tt.expectedResult.finalState, model.state)
			}

			if tt.expectedResult.noError && model.err != nil {
				t.Errorf("expected no error, got %v", model.err)
			}

			if tt.expectedResult.expectError != nil && model.err != tt.expectedResult.expectError {
				t.Errorf("expected error %v, got %v", tt.expectedResult.expectError, model.err)
			}

			if tt.expectedResult.hasError && model.err == nil {
				t.Error("expected an error but got none")
			}

			if model.selected != tt.expectedResult.selectedCommand {
				t.Errorf("expected selected command %q, got %q", tt.expectedResult.selectedCommand, model.selected)
			}
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
	if model.state != statePrompting {
		t.Errorf("expected initial state statePrompting, got %v", model.state)
	}

	// Type a new prompt
	model = typeTextIntoModel(model, "new prompt")

	// Press enter to generate commands
	updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(Model)

	// Should transition to generating state
	if model.state != stateGenerating {
		t.Errorf("expected state stateGenerating after enter, got %v", model.state)
	}

	// Simulate command generation completion
	generateMsg := generateMsg{Prompt: "new prompt", Commands: []string{"cmd1", "cmd2"}}
	updatedModel, _ = model.Update(generateMsg)
	model = updatedModel.(Model)

	// Should transition to selecting state (multiple commands)
	if model.state != stateSelecting {
		t.Errorf("expected state stateSelecting after generation, got %v", model.state)
	}

	// Select a command
	updatedModel, _ = model.Update(tea.KeyMsg{Type: tea.KeyEnter})
	model = updatedModel.(Model)

	// Should transition to selected state
	if model.state != stateSelected {
		t.Errorf("expected state stateSelected after selection, got %v", model.state)
	}
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
			updatedModel, _ := model.Update(tea.KeyMsg{Type: tea.KeyEnter})
			model = updatedModel.(Model)

			// Should be in generating state
			if model.state != stateGenerating {
				t.Errorf("expected generating state, got %v", model.state)
			}

			// Execute the generate command (this will trigger the error)
			cmd := model.runGenerate("test prompt")
			result := cmd()

			// Update model with error message
			updatedModel, _ = model.Update(result)
			model = updatedModel.(Model)

			// Should be in selected state with error
			if model.state != stateSelected {
				t.Errorf("expected selected state after error, got %v", model.state)
			}

			if model.err == nil {
				t.Error("expected error but got none")
			}

			if !strings.Contains(model.err.Error(), tt.expectedErr.Error()) {
				t.Errorf("expected error containing %q, got %q", tt.expectedErr.Error(), model.err.Error())
			}
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
			if view == "" {
				t.Error("view should not be empty")
			}

			// All views should contain title unless there's an error
			if testModel.err == nil && !strings.Contains(view, "gencmd") {
				t.Error("view should contain title when no error")
			}

			// Error views should show error message
			if testModel.err != nil && !strings.Contains(view, "Error:") {
				t.Error("error view should contain 'Error:'")
			}
		})
	}
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
	noError         bool
	expectError     error
	hasError        bool
}

func executeAction(t *testing.T, model Model, action userAction) Model {
	switch action.action {
	case typeText:
		return typeTextIntoModel(model, action.value)
	case pressKey:
		updatedModel, cmd := model.Update(tea.KeyMsg{Type: action.key})
		model = updatedModel.(Model)

		// If a command was returned, execute it and feed the result back
		if cmd != nil {
			result := cmd()
			if result != nil {
				updatedModel, _ := model.Update(result)
				model = updatedModel.(Model)
			}
		}

		return model
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
