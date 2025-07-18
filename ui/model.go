package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mbrt/gencmd/ctrl"
)

type (
	errMsg error
)

var (
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().Padding(1, 0, 0, 2)
	promptStyle       = lipgloss.NewStyle().PaddingTop(1)
)

func RunUI(c *ctrl.Controller) error {
	p := tea.NewProgram(New(c), tea.WithAltScreen())
	// TODO: Add a fallback for when we don't have a terminal
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("running UI: %w", err)
	}
	finalModel := m.(Model)
	if finalModel.cancelled {
		return nil
	}
	c.OutputCommand(finalModel.selected)
	return nil
}

type Model struct {
	controller   *ctrl.Controller
	KeyMap       KeyMap
	list         list.Model
	help         help.Model
	textInput    textinput.Model
	history      []list.Item
	selected     string
	isNewCommand bool
	cancelled    bool
	err          error
}

func New(c *ctrl.Controller) Model {
	// Convert history entries to list items
	history := c.LoadHistory()
	items := make([]list.Item, len(history))
	for i, entry := range history {
		items[i] = historyEntry{entry}
	}

	// Create the list
	l := list.New(items, list.NewDefaultDelegate(), 80, 24)
	l.Title = "gencmd"
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle
	l.SetShowHelp(false)

	// Create help model
	h := help.New()

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Search history or type a new command"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	return Model{
		controller: c,
		KeyMap:     DefaultKeyMap(),
		list:       l,
		help:       h,
		textInput:  ti,
		history:    items,
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Leave space for input and title
		m.textInput.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		return m.handleKey(msg)

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, nil
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("gencmd"))
	b.WriteString("\n\n")

	b.WriteString(m.list.View())
	b.WriteString("\n")

	// Show help text
	b.WriteString(helpStyle.Render(m.help.View(m)))

	// Always show the text input
	b.WriteString(promptStyle.Render(m.textInput.View()))
	b.WriteString("\n")

	return b.String()
}

func (m Model) ShortHelp() []key.Binding {
	bindings := []key.Binding{
		m.KeyMap.Submit,
		m.KeyMap.Cancel,
	}
	if m.list.SelectedItem() != nil {
		bindings = append(bindings, m.KeyMap.Up, m.KeyMap.Down)
	}
	return bindings
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Cancel):
		m.cancelled = true
		return m, tea.Quit

	case key.Matches(msg, m.KeyMap.Submit):
		if selectedItem := m.list.SelectedItem(); selectedItem != nil {
			m.selected = selectedItem.(historyEntry).Command
			return m, tea.Quit
		}
		// User typed a new command
		prompt := m.textInput.Value()
		cmds, err := m.controller.GenerateCommands(prompt)
		if err != nil {
			m.err = err
			return m, nil
		}
		m.selected = cmds[0]
		m.controller.UpdateHistory(prompt, cmds[0])
		return m, tea.Quit

	case key.Matches(msg, m.KeyMap.Up):
		m.list.CursorUp()

	case key.Matches(msg, m.KeyMap.Down):
		m.list.CursorDown()
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

	return m, cmd
}

func (m *Model) filterItems(query string) {
	if len(query) == 0 {
		m.list.SetFilteringEnabled(false)
		return
	}
	if len(query) == 1 {
		m.list.SetFilteringEnabled(true)
	}
	m.list.SetFilterText(query)
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
