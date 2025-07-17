package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type (
	errMsg error
)

var (
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	promptStyle       = lipgloss.NewStyle().Margin(1, 0, 0, 0)
)

type HistoryEntry struct {
	Prompt  string
	Command string
}

// Implement list.Item interface
func (h HistoryEntry) FilterValue() string {
	return h.Prompt + " " + h.Command
}

// Implement list.DefaultItem interface
func (h HistoryEntry) Title() string {
	return h.Prompt
}

func (h HistoryEntry) Description() string {
	return h.Command
}

type Model struct {
	list         list.Model
	textInput    textinput.Model
	allItems     []list.Item
	selected     string
	isNewCommand bool
	cancelled    bool
	err          error
}

func New(history []HistoryEntry) Model {
	// Convert history entries to list items
	items := make([]list.Item, len(history))
	for i, entry := range history {
		items[i] = entry
	}

	// Create the list
	l := list.New(items, list.NewDefaultDelegate(), 80, 24)
	l.Title = "Command History"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(true)
	l.Styles.Title = titleStyle

	// Create text input
	ti := textinput.New()
	ti.Placeholder = "Search history or type a new command"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80

	return Model{
		list:      l,
		textInput: ti,
		allItems:  items, // Store all items for filtering
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 4) // Leave space for input and title
		m.textInput.Width = msg.Width - 4
		return m, nil

	case tea.KeyMsg:
		// Handle global keys first
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyEnter:
			if selectedItem := m.list.SelectedItem(); selectedItem != nil {
				entry := selectedItem.(HistoryEntry)
				m.selected = entry.Command
				m.isNewCommand = false
				return m, tea.Quit
			}
			// User typed a new command
			m.selected = m.textInput.Value()
			m.isNewCommand = true
			return m, tea.Quit

		case tea.KeyUp, tea.KeyCtrlK:
			m.list.CursorUp()

		case tea.KeyDown, tea.KeyCtrlJ:
			m.list.CursorDown()
		}

		// Store the old value to detect changes
		oldValue := m.textInput.Value()

		// Update text input
		m.textInput, cmd = m.textInput.Update(msg)
		cmds = append(cmds, cmd)

		// Check if the input value changed and update filtering
		newValue := m.textInput.Value()
		if newValue != oldValue {
			m.filterItems(newValue)
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	return m, tea.Batch(cmds...)
}

func (m *Model) filterItems(query string) {
	m.list.SetFilterText(query)
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

	// Always show the text input
	b.WriteString(promptStyle.Render(m.textInput.View()))
	b.WriteString("\n")

	// Show help text
	b.WriteString(lipgloss.NewStyle().Foreground(lipgloss.Color("241")).Render("Enter: confirm • Esc: cancel"))

	return b.String()
}

func (m Model) Selected() (string, bool) {
	if m.cancelled {
		return "", false
	}
	return m.selected, m.isNewCommand
}

func RunUI(history []HistoryEntry) (string, bool, error) {
	p := tea.NewProgram(New(history), tea.WithAltScreen())
	m, err := p.Run()
	if err != nil {
		return "", false, err
	}
	finalModel := m.(Model)
	selected, isNew := finalModel.Selected()
	return selected, isNew, nil
}
