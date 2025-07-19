package ui

import (
	"fmt"
	"io"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	msgStyle          = lipgloss.NewStyle().Padding(1, 0, 1, 2)
	shortItemStyle    = lipgloss.NewStyle().PaddingLeft(4)
	selectedItemStyle = lipgloss.NewStyle().PaddingLeft(2).Foreground(lipgloss.Color("170"))
)

func newSelectModel(km KeyMap) selectModel {
	// Create the list
	l := list.New(nil, shortItemDelegate{}, 80, 24)
	l.SetShowTitle(false)
	l.SetFilteringEnabled(false)
	l.SetShowHelp(false)
	l.SetShowStatusBar(false)

	return selectModel{
		keyMap: km,
		list:   l,
	}
}

type selectModel struct {
	keyMap KeyMap
	list   list.Model
}

func (m selectModel) Init() tea.Cmd {
	return nil
}

func (m selectModel) Update(msg tea.Msg) (selectModel, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.list.SetWidth(msg.Width)
		m.list.SetHeight(msg.Height - 2) // Leave space for title
		return m, nil

	case tea.KeyMsg:
		switch {
		case key.Matches(msg, m.keyMap.Up):
			m.list.CursorUp()
			return m, nil
		case key.Matches(msg, m.keyMap.Down):
			m.list.CursorDown()
			return m, nil
		}
	}

	return m, nil
}

func (m selectModel) View() string {
	var b strings.Builder
	b.WriteString(msgStyle.Render("Select completion"))
	b.WriteString("\n")
	b.WriteString(m.list.View())
	return b.String()
}

func (m selectModel) ShortHelp() []key.Binding {
	return []key.Binding{
		m.keyMap.Up,
		m.keyMap.Down,
		m.keyMap.Submit,
		m.keyMap.Cancel,
	}
}

func (m *selectModel) SetItems(items []string) {
	listItems := make([]list.Item, len(items))
	for i, item := range items {
		listItems[i] = shortItem(item)
	}
	m.list.SetItems(listItems)
}

func (m selectModel) Selected() string {
	if len(m.list.Items()) == 0 {
		return ""
	}
	selected := m.list.SelectedItem()
	if selected == nil {
		return ""
	}
	return string(selected.(shortItem))
}

type shortItem string

func (i shortItem) FilterValue() string { return "" }

type shortItemDelegate struct{}

func (d shortItemDelegate) Height() int                             { return 1 }
func (d shortItemDelegate) Spacing() int                            { return 0 }
func (d shortItemDelegate) Update(_ tea.Msg, _ *list.Model) tea.Cmd { return nil }
func (d shortItemDelegate) Render(w io.Writer, m list.Model, index int, listshortItem list.Item) {
	i, ok := listshortItem.(shortItem)
	if !ok {
		return
	}

	str := string(i)

	fn := shortItemStyle.Render
	if index == m.Index() {
		fn = func(s ...string) string {
			return selectedItemStyle.Render("> " + strings.Join(s, " "))
		}
	}

	fmt.Fprint(w, fn(str))
}
