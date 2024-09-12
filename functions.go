package main

import (
	"fmt"
	"os/exec"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

func runCommand(name string, args ...string) tea.Cmd {
	return func() tea.Msg {
		cmd := exec.Command(name, args...)
		err := cmd.Run()
		if err != nil {
			return fmt.Errorf("error running command: %v", err)
		}
		return getDependencies()
	}
}

type loadingTickMsg struct{}

func updateLoading() tea.Cmd {
	return tea.Tick(time.Millisecond*500, func(t time.Time) tea.Msg {
		return loadingTickMsg{}
	})
}

func getDependencies() tea.Msg {
	cmd := exec.Command("go", "list", "-m", "all")
	output, err := cmd.Output()
	if err != nil {
		return fmt.Errorf("error running go list: %v", err)
	}

	lines := strings.Split(string(output), "\n")
	deps := make([]dependency, 0)
	for _, line := range lines[1:] {
		parts := strings.Fields(line)
		if len(parts) >= 2 {
			deps = append(deps, dependency{name: parts[0], version: parts[1]})
		} else if len(parts) == 1 {
			deps = append(deps, dependency{name: "Unknown", version: parts[0]})
		}
	}
	return dependencyMsg(deps)
}
