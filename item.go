package main

import "github.com/charmbracelet/bubbles/list"

func (i Item) Title() string       { return i.title }
func (i Item) FilterValue() string { return i.title }

func toItems(items []string) []list.Item {
	var listItems []list.Item
	for _, item := range items {
		listItems = append(listItems, Item{title: item})
	}
	return listItems
}
