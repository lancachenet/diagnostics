package main

import "github.com/charmbracelet/bubbles/list"

type CDN struct {
	Name string
	File string
}

type Lookup struct {
	Resolver    string
	Hostname    string
	Address     []string
	ContainerID string
	Time        string
}

type Item struct {
	title    string
	selected bool
}

type Model struct {
	Title          string
	Height         int
	Items          []Item
	Style          Style
	List           list.Model
	Selected       string
	MultiSelected  []string
	Filter         string
	Error          string
	MultiSelection bool
	Quitting       bool
}

type Style struct {
	Model *Model
}
