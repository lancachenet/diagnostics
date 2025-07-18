package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/huh"
)

func newModel(title string, options []string, multi bool) Model {
	items := toItems(options)

	var modelItems []Item
	for _, li := range items {
		if it, ok := li.(Item); ok {
			modelItems = append(modelItems, it)
		}
	}

	m := Model{
		Title:          title,
		Height:         len(options),
		Items:          modelItems,
		MultiSelection: multi,
	}
	m.Style = Style{Model: &m}

	l := list.New(items, m.Style, 0, 0)
	l.SetShowTitle(false)
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.DisableQuitKeybindings()
	l.SetShowPagination(false)
	l.SetShowHelp(false)
	m.List = l

	return m
}

func (m *Model) Init() tea.Cmd {
	return nil
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {

	case tea.KeyMsg:
		switch {

		case key.Matches(msg, key.NewBinding(key.WithKeys("ctrl+c", "esc"))):
			if m.Filter != "" {
				m.Filter = ""
				filtered := make([]list.Item, 0, len(m.Items))
				for _, item := range m.Items {
					filtered = append(filtered, item)
				}
				m.List.SetItems(filtered)
				m.List.Select(0)
				return m, nil
			}
			m.Quitting = true
			return m, tea.Quit

		case msg.String() == " ":
			if m.MultiSelection {
				if i, ok := m.List.SelectedItem().(Item); ok {
					m.toggleSelection(i.title)
					m.Filter = ""
					m.applyFilter("")
				}
			}

		case msg.String() == "left":
			m.clearSelections()

		case msg.String() == "right":
			m.selectAll()

		case msg.String() == "enter":
			if m.MultiSelection {
				m.MultiSelected = m.getSelections()
				if len(m.MultiSelected) == 0 {
					m.Error = "At least one CDN is required"
					return m, nil
				}
			} else {
				if i, ok := m.List.Items()[m.List.Index()].(Item); ok {
					m.Selected = i.title
				}
			}
			m.Quitting = true
			return m, tea.Quit

		case msg.String() == "backspace":
			if len(m.Filter) > 0 {
				m.Filter = m.Filter[:len(m.Filter)-1]
				m.applyFilter(m.Filter)
			}

		case len(msg.String()) == 1:
			m.Filter += msg.String()
			m.applyFilter(m.Filter)

		case msg.String() == "up", msg.String() == "down":
		}

	case tea.WindowSizeMsg:
		maxVisible := len(m.List.Items())
		availableHeight := msg.Height

		if availableHeight > maxVisible {
			availableHeight = maxVisible
		}

		m.List.SetSize(msg.Width, availableHeight)
	}

	var cmd tea.Cmd
	m.List, cmd = m.List.Update(msg)
	return m, cmd
}

func (m *Model) View() string {
	theme := huh.ThemeCharm()
	title := theme.Focused.Base.Render() + theme.Focused.Title.Render(m.Title)
	help := styledHelp(m.MultiSelection)

	if m.Filter != "" {
		title = title + " " + theme.Focused.TextInput.Prompt.Render(m.Filter)
	}
	if m.Error != "" {
		help = theme.Focused.ErrorMessage.Render(m.Error) + "\n" + help
	}

	return fmt.Sprintf(
		"%s\n%s\n\n%s",
		title,
		m.List.View(),
		help,
	)
}

func (m *Model) toggleSelection(name string) {
	for i := range m.Items {
		if m.Items[i].title == name {
			if m.MultiSelection {
				m.Items[i].selected = !m.Items[i].selected
			} else {
				m.clearSelections()
				m.Items[i].selected = true
			}
			break
		}
	}
}

func (m *Model) selectAll() {
	for i := range m.Items {
		m.Items[i].selected = true
	}
	m.applyFilter(m.Filter)
}

func (m *Model) clearSelections() {
	for i := range m.Items {
		m.Items[i].selected = false
	}
	m.applyFilter(m.Filter)
}

func (m *Model) getSelections() []string {
	var selected []string
	for _, i := range m.Items {
		if i.selected {
			selected = append(selected, i.title)
		}
	}
	return selected
}

func (m *Model) applyFilter(filter string) {
	filter = strings.ToLower(filter)
	var items []list.Item
	for _, i := range m.Items {
		if strings.Contains(strings.ToLower(i.title), filter) {
			items = append(items, i)
		}
	}
	m.List.SetItems(items)
}

func styledHelp(multi bool) string {
	theme := huh.ThemeCharm().Help

	segment := func(key, desc string) string {
		return theme.ShortKey.Render(key) + " " + theme.ShortDesc.Render(desc)
	}

	sep := theme.ShortDesc.Render(" • ")

	segments := []string{
		segment("↑", "up"),
		segment("↓", "down"),
	}

	if multi {
		segments = append(segments,
			segment("spacebar", "select"),
		)
	}

	segments = append(segments,
		segment("enter", "submit"),
	)

	if multi {
		segments = append(segments,
			segment("→", "select all"),
			segment("←", "select none"),
		)
	}

	segments = append(segments,
		segment("esc", "exit"),
		theme.ShortDesc.Render("type to filter"),
	)

	return strings.Join(segments, sep)
}
