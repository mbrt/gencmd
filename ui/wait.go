package ui

import (
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var spinnerStyle = lipgloss.NewStyle().MarginLeft(2).Foreground(lipgloss.Color("205"))

func newWaitModel(km KeyMap) waitModel {
	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return waitModel{
		spinner: s,
		keyMap:  km,
	}
}

type waitModel struct {
	spinner spinner.Model
	keyMap  KeyMap
}

func (m waitModel) Init() tea.Cmd {
	return m.spinner.Tick
}

func (m waitModel) Update(msg tea.Msg) (waitModel, tea.Cmd) {
	var cmd tea.Cmd
	m.spinner, cmd = m.spinner.Update(msg)
	return m, cmd
}

func (m waitModel) View() string {
	var b strings.Builder
	b.WriteString("\n")
	b.WriteString(m.spinner.View())
	b.WriteString(" Generating commands...\n")
	return b.String()
}

func (m waitModel) ShortHelp() []key.Binding {
	return []key.Binding{m.keyMap.Cancel}
}
