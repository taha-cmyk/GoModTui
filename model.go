package main

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type dependency struct {
	name    string
	version string
}

func (d dependency) Title() string       { return d.name }
func (d dependency) Description() string { return d.version }
func (d dependency) FilterValue() string { return d.name }

type dependencyMsg []dependency

type model struct {
	list          list.Model
	input         textinput.Model
	inputting     bool
	inputMode     inputMode
	selectedIndex int
	confirmDelete bool
	loading       bool
	loadingDots   int
}

func (m model) Init() tea.Cmd {
	return getDependencies
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			return m, tea.Quit
		case "a":
			m.inputting = true
			m.inputMode = addMode
			m.input.SetValue("")
			m.input.Placeholder = "Enter new dependency (format: name@version)"
			return m, textinput.Blink
		case "u":
			if len(m.list.Items()) > 0 {
				m.inputting = true
				m.inputMode = updateMode
				m.selectedIndex = m.list.Index()
				dep := m.list.Items()[m.selectedIndex].(dependency)
				m.input.SetValue(fmt.Sprintf("%s@%s", dep.name, dep.version))
				m.input.Placeholder = "Update dependency (format: name@version)"
				return m, textinput.Blink
			}
		case "d":
			if len(m.list.Items()) > 0 {
				m.inputMode = deleteMode
				m.selectedIndex = m.list.Index()
				m.confirmDelete = true
				return m, nil
			}
		case "t":
			m.loading = true
			return m, tea.Batch(
				runCommand("go", "mod", "tidy"),
				updateLoading(),
			)
		case "y":
			if m.confirmDelete {
				dep := m.list.Items()[m.selectedIndex].(dependency)
				m.list.RemoveItem(m.selectedIndex)
				m.loading = true
				m.confirmDelete = false
				m.inputMode = normalMode
				return m, tea.Batch(
					runCommand("go", "get", fmt.Sprintf("%s@none", dep.name)),
					updateLoading(),
				)
			}
		case "n":
			if m.confirmDelete {
				m.confirmDelete = false
				m.inputMode = normalMode
				return m, nil
			}
		}

		if m.inputting {
			switch msg.String() {
			case "enter":
				parts := strings.Split(m.input.Value(), "@")
				if len(parts) == 2 {
					newDep := dependency{name: parts[0], version: parts[1]}
					switch m.inputMode {
					case addMode:
						m.list.InsertItem(0, newDep)
						m.loading = true
						return m, tea.Batch(
							runCommand("go", "get", m.input.Value()),
							updateLoading(),
						)
					case updateMode:
						m.list.SetItem(m.selectedIndex, newDep)
						m.loading = true
						return m, tea.Batch(
							runCommand("go", "get", m.input.Value()),
							updateLoading(),
						)
					}
				}
				m.input.SetValue("")
				m.inputting = false
				m.inputMode = normalMode
				return m, nil
			case "esc":
				m.inputting = false
				m.inputMode = normalMode
				return m, nil
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	case tea.WindowSizeMsg:
		h, v := docStyle.GetFrameSize()
		m.list.SetSize(msg.Width-h, msg.Height-v)
	case dependencyMsg:
		items := make([]list.Item, len(msg))
		for i, dep := range msg {
			items[i] = dep
		}
		m.list.SetItems(items)
		m.loading = false
	case loadingTickMsg:
		m.loadingDots = (m.loadingDots + 1) % 4
		if m.loading {
			return m, updateLoading()
		}
	case error:
		m.list.Title = fmt.Sprintf("Error: %v", msg)
		m.loading = false
	}

	var cmd tea.Cmd
	m.list, cmd = m.list.Update(msg)
	return m, cmd
}

var docStyle = lipgloss.NewStyle().Margin(1, 2)

func (m model) View() string {
	if m.loading {
		dots := strings.Repeat(".", m.loadingDots)
		return fmt.Sprintf("\n\n   Loading%s\n\n%s", dots, "Press q to quit")
	}
	if m.inputting {
		return fmt.Sprintf(
			"%s\n\n%s\n\n(esc to cancel)",
			m.input.Placeholder,
			m.input.View(),
		) + "\n"
	}
	if m.confirmDelete {
		dep := m.list.Items()[m.selectedIndex].(dependency)
		return fmt.Sprintf(
			"Are you sure you want to delete %s@%s?\n\n(y/n)",
			dep.name, dep.version,
		) + "\n"
	}
	return docStyle.Render(m.list.View() + "\n\nPress 'a' to add, 'u' to update, 'd' to delete, 't' for go mod tidy, 'q' to quit")
}
