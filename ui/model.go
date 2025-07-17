package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/sahilm/fuzzy"
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

type Model struct {
	history      []HistoryEntry
	matches      fuzzy.Matches
	textInput    textinput.Model
	cursor       int
	selected     string
	isNewCommand bool
	cancelled    bool
	err          error
	width        int
	height       int
}

func New(history []HistoryEntry) Model {
	ti := textinput.New()
	ti.Placeholder = "Search history or type a new command"
	ti.Focus()
	ti.CharLimit = 156
	ti.Width = 80 // Default width, will be updated when window size is known

	return Model{
		history:   history,
		textInput: ti,
		err:       nil,
		width:     80, // Default width
		height:    24, // Default height
	}
}

func (m Model) Init() tea.Cmd {
	return textinput.Blink
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.textInput.Width = msg.Width - 4 // Leave some padding
		return m, nil

	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEnter:
			if m.textInput.Value() == "" {
				return m, tea.Quit
			}
			if len(m.matches) > 0 && m.cursor < len(m.matches) {
				m.selected = m.history[m.matches[m.cursor].Index].Command
				m.isNewCommand = false
			} else {
				m.selected = m.textInput.Value()
				m.isNewCommand = true
			}
			return m, tea.Quit

		case tea.KeyCtrlC, tea.KeyEsc:
			m.cancelled = true
			return m, tea.Quit

		case tea.KeyUp, tea.KeyCtrlK:
			if len(m.matches) == 0 {
				break
			}
			m.cursor--
			if m.cursor < 0 {
				m.cursor = len(m.matches) - 1
			}

		case tea.KeyDown, tea.KeyCtrlJ:
			if len(m.matches) == 0 {
				break
			}
			m.cursor++
			if m.cursor >= len(m.matches) {
				m.cursor = 0
			}

		default:
			m.textInput, cmd = m.textInput.Update(msg)
			m.updateMatches()
			return m, cmd
		}

	case errMsg:
		m.err = msg
		return m, nil
	}

	m.textInput, cmd = m.textInput.Update(msg)
	m.updateMatches()
	return m, cmd
}

func (m *Model) updateMatches() {
	var historyStrings []string
	for _, h := range m.history {
		historyStrings = append(historyStrings, fmt.Sprintf("%s -> %s", h.Prompt, h.Command))
	}
	m.matches = fuzzy.Find(m.textInput.Value(), historyStrings)
	if m.cursor >= len(m.matches) {
		m.cursor = 0
	}
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("gencmd"))
	b.WriteString("\n\n")

	var entries []HistoryEntry
	if m.textInput.Value() == "" {
		// If no input, show all history
		entries = m.history
	} else {
		for _, match := range m.matches {
			entries = append(entries, m.history[match.Index])
		}
	}

	// Display the matches
	for i := 0; i < len(entries); i++ {
		entry := entries[i]
		line := fmt.Sprintf("%s -> %s", entry.Prompt, entry.Command)
		if m.cursor == i {
			b.WriteString(selectedItemStyle.Render("> " + line))
		} else {
			b.WriteString(itemStyle.Render(line))
		}
		b.WriteString("\n")
	}

	b.WriteString(promptStyle.Render(m.textInput.View()))
	b.WriteString("\n")

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
