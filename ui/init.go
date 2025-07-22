package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/mbrt/gencmd/config"
)

// SelectProvider allows the user to select a provider and input its configuration.
func SelectProvider(providers []config.ProviderDoc) (string, map[string]string, error) {
	if len(providers) == 0 {
		return "", nil, fmt.Errorf("no providers available")
	}

	// Let the user select a provider.
	provider, err := selectProvider(providers)
	if err != nil {
		return "", nil, err
	}

	// Ask for the provider's options.
	options, err := askOptions(provider)
	if err != nil {
		return "", nil, err
	}

	return provider.ID, options, nil
}

func selectProvider(providers []config.ProviderDoc) (config.ProviderDoc, error) {
	m, err := tea.NewProgram(newSelectProviderModel(providers)).Run()
	if err != nil {
		return config.ProviderDoc{}, err
	}
	res := m.(selectProviderModel)
	if res.selected == nil {
		return config.ProviderDoc{}, fmt.Errorf("no provider selected")
	}
	return *res.selected, nil
}

func askOptions(provider config.ProviderDoc) (map[string]string, error) {
	res := make(map[string]string)
	for _, opt := range provider.Options {
		m, err := tea.NewProgram(newAskOptionModel(opt)).Run()
		if err != nil {
			return nil, err
		}
		model := m.(askOptionModel)
		if model.submitted == "" {
			return nil, fmt.Errorf("no value provided for %q", opt.Name)
		}
		res[opt.EnvVar] = model.submitted
	}
	return res, nil
}

func newSelectProviderModel(providers []config.ProviderDoc) selectProviderModel {
	items := make([]list.Item, len(providers))
	for i, p := range providers {
		items[i] = providerItem{p}
	}
	l := list.New(items, list.NewDefaultDelegate(), 80, 15)
	l.Title = "Select a provider"
	l.SetShowStatusBar(false)
	l.SetShowHelp(false)

	return selectProviderModel{
		list: l,
	}
}

type selectProviderModel struct {
	list     list.Model
	selected *config.ProviderDoc
}

func (m selectProviderModel) Init() tea.Cmd {
	return nil
}

func (m selectProviderModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			if sel, ok := m.list.SelectedItem().(providerItem); ok {
				m.selected = &sel.ProviderDoc
			}
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

func (m selectProviderModel) View() string {
	return m.list.View()
}

// providerItem wraps a config.ProviderDoc to implement the list.Item interface.
type providerItem struct {
	config.ProviderDoc
}

func (i providerItem) FilterValue() string { return i.ID }
func (i providerItem) Title() string       { return i.ProviderDoc.Name }
func (i providerItem) Description() string { return i.ProviderDoc.URL }

func newAskOptionModel(option config.ProviderOption) askOptionModel {
	ti := textinput.New()
	ti.Placeholder = option.Description
	ti.Focus()
	ti.CharLimit = 256
	ti.Width = 80

	return askOptionModel{
		textInput: ti,
		option:    option,
	}
}

type askOptionModel struct {
	textInput textinput.Model
	option    config.ProviderOption
	submitted string
}

func (m askOptionModel) Init() tea.Cmd {
	return textinput.Blink
}

func (m askOptionModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyEscape, tea.KeyCtrlC:
			return m, tea.Quit
		case tea.KeyEnter:
			m.submitted = m.textInput.Value()
			return m, tea.Quit
		}
	}

	var cmd tea.Cmd
	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m askOptionModel) View() string {
	return fmt.Sprintf(
		"Enter value for %s:\n\n%s\n\n",
		m.option.Name,
		m.textInput.View(),
	)
}
