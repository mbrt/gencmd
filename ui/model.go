package ui

import (
	"errors"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/help"
	"github.com/charmbracelet/bubbles/key"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/mbrt/gencmd/ctrl"
)

var UserCancelErr = errors.New("user cancelled")

var (
	titleStyle  = lipgloss.NewStyle().Background(lipgloss.Color("62")).Foreground(lipgloss.Color("230")).Padding(0, 1)
	itemStyle   = lipgloss.NewStyle().PaddingLeft(4)
	helpStyle   = lipgloss.NewStyle().PaddingTop(1).PaddingLeft(2)
	promptStyle = lipgloss.NewStyle().PaddingTop(1)
)

type state int

const (
	statePrompting state = iota
	stateGenerating
	stateSelecting
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
	if finalModel.err != nil {
		return finalModel.err
	}
	if finalModel.selected != "" {
		fmt.Println(finalModel.selected)
	}
	return nil
}

type Model struct {
	KeyMap KeyMap

	controller *ctrl.Controller
	prompt     promptModel
	wait       waitModel
	selectCmp  selectModel
	help       help.Model
	state      state
	promptText string
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
		selectCmp:  newSelectModel(km),
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
		cmd = m.handleCompletion(msg.Prompt, msg.Commands)
		return m, cmd

	case errMsg:
		cmd = m.quitWithError(msg)
		return m, cmd

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
	case stateSelecting:
		b.WriteString(m.selectCmp.View())
	}

	// Show help text
	b.WriteString(helpStyle.Render(m.help.View(m)))

	return b.String()
}

func (m Model) ShortHelp() []key.Binding {
	switch m.state {
	case statePrompting:
		return m.prompt.ShortHelp()
	case stateGenerating:
		return m.wait.ShortHelp()
	case stateSelecting:
		return m.selectCmp.ShortHelp()
	default:
		return []key.Binding{m.KeyMap.Cancel}
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
	if !onlyActive || m.state == stateSelecting {
		m.selectCmp, cmd = m.selectCmp.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

func (m Model) handleKey(msg tea.KeyMsg) (Model, tea.Cmd) {
	switch {
	case key.Matches(msg, m.KeyMap.Cancel):
		cmd := m.quitWithError(UserCancelErr)
		return m, cmd

	case key.Matches(msg, m.KeyMap.Submit):
		switch m.state {

		case statePrompting:
			// User submitted a prompt
			selected := m.prompt.Selected()
			if selected.IsNew() {
				cmd := m.runGenerate(selected.Prompt)
				return m, cmd
			}
			// User selected an existing command
			m.selected = selected.Command
			m.state = stateSelected
			return m, tea.Quit

		case stateSelecting:
			// User selected a command from the list
			selected := m.selectCmp.Selected()
			cmd := m.selectCommand(m.promptText, selected)
			return m, cmd
		}
	}

	return m, nil
}

func (m *Model) runGenerate(prompt string) tea.Cmd {
	m.promptText = prompt
	m.state = stateGenerating
	return func() tea.Msg {
		cmds, err := m.controller.GenerateCommands(prompt)
		if err != nil {
			return errMsg(err)
		}
		return generateMsg{Prompt: prompt, Commands: cmds}
	}
}

func (m *Model) handleCompletion(prompt string, commands []string) tea.Cmd {
	if len(commands) == 0 {
		return m.quitWithError(fmt.Errorf("no commands generated"))
	}

	// If we have multiple commands, show a selection list
	if len(commands) > 1 {
		m.selectCmp.SetItems(commands)
		m.state = stateSelecting
		return nil
	}

	// If there's only one command, select it directly
	return m.selectCommand(prompt, commands[0])
}

func (m *Model) selectCommand(prompt string, command string) tea.Cmd {
	if command == "" {
		return m.quitWithError(fmt.Errorf("no command selected"))
	}
	m.selected = command
	m.controller.UpdateHistory(prompt, command)
	m.state = stateSelected
	return tea.Quit
}

func (m *Model) quitWithError(err error) tea.Cmd {
	m.err = err
	m.state = stateSelected
	m.selected = ""
	return tea.Quit

}

type errMsg error

type generateMsg struct {
	Prompt   string
	Commands []string
}
