package main

import (
	"fmt"
	"io"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/huh/v2"
)

func (s Style) Height() int                               { return 1 }
func (s Style) Spacing() int                              { return 0 }
func (s Style) Update(msg tea.Msg, m *list.Model) tea.Cmd { return nil }

func (s Style) Render(w io.Writer, m list.Model, index int, listItem list.Item) {
	it, ok := listItem.(Item)
	if !ok {
		return
	}

	theme := huh.ThemeCharm(false)
	line := theme.Focused.Base.Render()
	isMulti := s.Model != nil && s.Model.MultiSelection

	if isMulti {
		if index == m.Index() {
			line += theme.Focused.MultiSelectSelector.Render()
		} else {
			line += theme.Blurred.Base.Render()
		}

		if it.selected {
			line += theme.Focused.SelectedPrefix.Render() + theme.Focused.SelectedOption.Render(it.title)
		} else {
			line += theme.Focused.UnselectedPrefix.Render() + theme.Focused.UnselectedOption.Render(it.title)
		}
	} else {
		if index == m.Index() {
			line += theme.Focused.SelectSelector.Render()
			line += theme.Focused.SelectedOption.Render(it.title)
		} else {
			line += "  "
			line += theme.Focused.UnselectedOption.Render(it.title)
		}
	}

	_, err := fmt.Fprint(w, line)
	if err != nil {
		_, _ = fmt.Fprint(w, fmt.Errorf("error: %w", err))
	}
}
