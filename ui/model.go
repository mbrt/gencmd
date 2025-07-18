package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/spinner"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mbrt/gencmd/ctrl"
)

var (
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().Padding(1, 0, 0, 2)
	promptStyle       = lipgloss.NewStyle().PaddingTop(1)
	spinnerStyle      = lipgloss.NewStyle().Foreground(lipgloss.Color("205"))
)

type state int

const (
	statePrompting state = iota
	stateGenerating
	stateSelected
)

func RunUI(c *ctrl.Controller) error {
	p := tea.NewProgram(New(c), tea.WithAltScreen())
	// TODO: Add a fallback for when we don't have a terminal
	m, err := p.Run()
	if err != nil {
		return fmt.Errorf("running UI: %w", err)
	}
	finalModel := m.(Model)
	c.OutputCommand(finalModel.selected)
	return nil
}

type Model struct {
	KeyMap KeyMap

	controller *ctrl.Controller
	prompt     promptModel
	help       help.Model
	spinner    spinner.Model
	state      state
	selected   string
	err        error
}

func New(c *ctrl.Controller) Model {
	// Create prompt model
	promptM := newPromptModel(DefaultKeyMap(), c.LoadHistory())

	// Create help
	h := help.New()

	// Create spinner
	s := spinner.New()
	s.Spinner = spinner.Dot
	s.Style = spinnerStyle

	return Model{
		controller: c,
		KeyMap:     DefaultKeyMap(),
		prompt:     promptM,
		help:       h,
		state:      statePrompting,
		spinner:    s,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(textinput.Blink, m.spinner.Tick)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		msg.Height -= 4 // Leave space for help and title
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(msg)
		return m, cmd

	case tea.KeyMsg:
		return m.handleKey(msg)

	case generateMsg:
		if len(msg.Commands) == 0 {
			m.err = fmt.Errorf("no commands generated")
			return m, nil
		}
		m.controller.UpdateHistory(msg.Prompt, msg.Commands[0])
		m.selected = msg.Commands[0]
		m.state = stateSelected
		return m, tea.Quit

	case errMsg:
		m.state = stateSelected
		m.err = msg
		return m, nil

	default:
		var cmd tea.Cmd
		m.spinner, cmd = m.spinner.Update(msg)
		return m, cmd
	}
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("gencmd"))
	b.WriteString("\n\n")

	if m.state == stateGenerating {
		b.WriteString(m.spinner.View())
		b.WriteString(" Generating commands...\n")
	} else {
		b.WriteString(m.prompt.View())
		// Show help text
		b.WriteString(helpStyle.Render(m.help.View(m)))
	}

	return b.String()
}

func (m Model) ShortHelp() []key.Binding {
	if m.state == stateGenerating {
		return []key.Binding{m.KeyMap.Cancel}
	} else {
		return m.prompt.ShortHelp()
	}
}

func (m Model) FullHelp() [][]key.Binding {
	return [][]key.Binding{m.ShortHelp()}
}

func (m Model) handleKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Cancel):
		m.selected = ""
		m.state = stateSelected
		return m, tea.Quit

	case key.Matches(msg, m.KeyMap.Submit):
		if m.state == statePrompting {
			selected := m.prompt.Selected()
			if selected.IsNew() {
				return m, m.runGenerate(selected.Prompt)
			}
			// User selected an existing command
			m.selected = selected.Command
			m.state = stateSelected
			return m, tea.Quit
		}
		// TODO:
		return m, nil

	default:
		var cmd tea.Cmd
		m.prompt, cmd = m.prompt.Update(msg)
		return m, cmd
	}
}

func (m *Model) runGenerate(prompt string) tea.Cmd {
	m.state = stateGenerating
	return func() tea.Msg {
		cmds, err := m.controller.GenerateCommands(prompt)
		if err != nil {
			return errMsg(err)
		}
		return generateMsg{Prompt: prompt, Commands: cmds}
	}
}

type errMsg error

type generateMsg struct {
	Prompt   string
	Commands []string
}
