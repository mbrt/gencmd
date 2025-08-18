package ui

import (
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/mbrt/gencmd/ctrl"
)

func newPromptModel(km KeyMap, c Controller) promptModel {
	history := c.LoadHistory()
	items := make([]list.Item, len(history))
	for i, entry := range history {
		items[i] = historyEntry{entry}
	}

	// Create the list
	l := list.New(items, list.NewDefaultDelegate(), 80, 24)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(true)
	l.SetShowHelp(false)

	// Create text input
	ti := textinput.New()
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	res := promptModel{
		controller:     c,
		keyMap:         km,
		list:           l,
		textInput:      ti,
		historyVisible: true,
	}
	res.updateDefaultText()
	return res
}

type promptModel struct {
	controller     Controller
	keyMap         KeyMap
	list           list.Model
	textInput      textinput.Model
	historyVisible bool
}

func (m promptModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m promptModel) Update(msg tea.Msg) (promptModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		// Leave space for input
		m.list.SetHeight(msg.Height - 2)
		m.textInput.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		cmd := m.handleKey(msg)
		return m, cmd

	default:
		return m, nil
	}
}

func (m promptModel) View() string {
	var b strings.Builder

	if len(m.list.Items()) > 0 && m.historyVisible {
		b.WriteString(m.list.View())
		b.WriteString("\n")
	}

	// Show the text input
	b.WriteString(promptStyle.Render(m.textInput.View()))
	b.WriteString("\n")

	return b.String()
}

func (m promptModel) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		m.keyMap.Submit,
		m.keyMap.Cancel,
	}
	if m.list.SelectedItem() != nil && m.historyVisible {
		bindings = append(bindings, m.keyMap.Up, m.keyMap.Down)
	}
	if len(m.list.Items()) > 0 {
		bindings = append(bindings, m.keyMap.ToggleHistory)
	}
	return bindings
}

func (m promptModel) Selected() inputPrompt {
	if selectedItem := m.list.SelectedItem(); selectedItem != nil && m.historyVisible {
		he := selectedItem.(historyEntry)
		return inputPrompt{Prompt: he.Prompt, Command: he.Command}
	}
	// User typed a new command
	return inputPrompt{Prompt: m.textInput.Value()}
}

func (m *promptModel) handleKey(msg tea.KeyMsg) tea.Cmd {
	switch {
	case key.Matches(msg, m.keyMap.Up):
		m.list.CursorUp()
		return nil
	case key.Matches(msg, m.keyMap.Down):
		m.list.CursorDown()
		return nil
	case key.Matches(msg, m.keyMap.ToggleHistory):
		m.historyVisible = !m.historyVisible
		m.updateDefaultText()
	case key.Matches(msg, m.keyMap.DeleteHistory):
		return m.handleDeleteHistory()
	}
	// Handle text input updates.
	// Store the old value, update and compare for changes.
	oldValue := m.textInput.Value()
	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	newValue := m.textInput.Value()
	if newValue != oldValue {
		m.filterItems(newValue)
	}
	return cmd
}

func (m *promptModel) filterItems(query string) {
	if len(query) == 0 {
		m.list.SetFilteringEnabled(false)
		return
	}
	if len(query) == 1 {
		m.list.SetFilteringEnabled(true)
	}
	m.list.SetFilterText(query)
	m.updateDefaultText()
}

func (m *promptModel) updateDefaultText() {
	if len(m.list.Items()) > 0 && m.historyVisible {
		m.textInput.Placeholder = "Search history or type a new prompt"
	} else {
		m.textInput.Placeholder = "Type a prompt"
	}
}

type inputPrompt struct {
	Prompt  string
	Command string
}

func (p inputPrompt) Empty() bool {
	return p.Prompt == "" && p.Command == ""
}

func (p inputPrompt) IsNew() bool {
	return p.Prompt != "" && p.Command == ""
}

type historyEntry struct {
	ctrl.HistoryEntry
}

// Implement list.Item interface
func (h historyEntry) FilterValue() string {
	return h.Prompt + " " + h.Command
}

// Implement list.DefaultItem interface
func (h historyEntry) Title() string {
	return h.Prompt
}

func (h historyEntry) Description() string {
	return h.Command
}

func (m *promptModel) handleDeleteHistory() tea.Cmd {
	// Only allow deletion if history is visible and an item is selected
	if !m.historyVisible || len(m.list.Items()) == 0 {
		return nil
	}

	selectedItem := m.list.SelectedItem()
	if selectedItem == nil {
		return nil
	}

	var cmds []tea.Cmd

	// Get the selected history entry
	he := selectedItem.(historyEntry).HistoryEntry

	// Remove the item from the list
	items := m.list.Items()
	selectedIndex := m.list.GlobalIndex()

	// Create new items slice without the selected item
	newItems := slices.Delete(items, selectedIndex, selectedIndex+1)
	// Update the list with new items
	cmds = append(cmds, tea.Sequence(
		m.list.SetItems(newItems),
		func() tea.Msg {
			m.filterItems(m.textInput.Value())
			return nil
		},
	))

	// Adjust cursor position if necessary
	if selectedIndex >= len(newItems) && len(newItems) > 0 {
		m.list.Select(len(newItems) - 1)
	}

	// Delete from the controller (async)
	cmds = append(cmds, func() tea.Msg {
		err := m.controller.DeleteHistory(he)
		if err != nil {
			return errMsg(err)
		}
		return nil
	})
	return tea.Batch(cmds...)
}
