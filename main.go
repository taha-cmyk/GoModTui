package main

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
)

type inputMode int

const (
	normalMode inputMode = iota
	addMode
	updateMode
	deleteMode
)

func initialModel() model {
	l := list.New([]list.Item{}, list.NewDefaultDelegate(), 0, 0)
	l.Title = "Go Dependencies"

	ti := textinput.New()
	ti.Placeholder = "Enter dependency..."
	ti.Focus()

	return model{
		list:  l,
		input: ti,
	}
}

func main() {
	p := tea.NewProgram(initialModel())
	if _, err := p.Run(); err != nil {
		fmt.Println("Error running program:", err)
	}
}
