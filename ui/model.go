package ui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mbrt/gencmd/ctrl"
)

var (
	titleStyle        = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	itemStyle         = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
	helpStyle         = lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2)
	promptStyle       = lipgloss.NewStyle().PaddingTop(1)
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
	wait       waitModel
	help       help.Model
	state      state
	selected   string
	err        error
}

func New(c *ctrl.Controller) Model {
	km := DefaultKeyMap()
	h := help.New()
	return Model{
		controller: c,
		KeyMap:     km,
		prompt:     newPromptModel(km, c.LoadHistory()),
		wait:       newWaitModel(km),
		help:       h,
		state:      statePrompting,
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		m.prompt.Init(),
		m.wait.Init(),
	)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		// Leave space for help and title
		msg.Height -= 4
		// Forward window size message to models
		m, cmd = m.updateModels(msg, false)
		cmds = append(cmds, cmd)

	case tea.KeyMsg:
		m, cmd = m.handleKey(msg)
		cmds = append(cmds, cmd)
		m, cmd = m.updateModels(msg, false)
		cmds = append(cmds, cmd)

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
		m, cmd = m.updateModels(msg, false)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) View() string {
	if m.err != nil {
		return fmt.Sprintf("\nError: %v\n\n", m.err)
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render("gencmd"))
	b.WriteString("\n")

	switch m.state {
	case statePrompting:
		b.WriteString(m.prompt.View())
	case stateGenerating:
		b.WriteString(m.wait.View())
	}

	// Show help text
	b.WriteString(helpStyle.Render(m.help.View(m)))

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

func (m Model) updateModels(msg tea.Msg, onlyActive bool) (Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd

	if !onlyActive || m.state == statePrompting {
		m.prompt, cmd = m.prompt.Update(msg)
		cmds = append(cmds, cmd)
	}
	if !onlyActive || m.state == stateGenerating {
		m.wait, cmd = m.wait.Update(msg)
		cmds = append(cmds, cmd)
	}
	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
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
	}

	return m, nil
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
