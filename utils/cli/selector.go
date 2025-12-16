package cli

import (
	"fmt"
	"os"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	selectedStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("229")).
			Background(lipgloss.Color("57")).
			Bold(true).
			PaddingLeft(1).
			PaddingRight(1)

	normalStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("250")).
			PaddingLeft(1).
			PaddingRight(1)

	headerStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("39")).
			Bold(true).
			MarginBottom(1)
)

type model struct {
	prompt     string
	items      []string
	cursor     int
	selected   int
	defaultIdx int
	done       bool
}

func (m model) Init() tea.Cmd {
	return nil
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "esc":
			m.selected = m.defaultIdx
			m.done = true
			return m, tea.Quit
		case "enter":
			m.selected = m.cursor
			m.done = true
			return m, tea.Quit
		case "up", "k":
			if m.cursor > 0 {
				m.cursor--
			}
		case "down", "j":
			if m.cursor < len(m.items)-1 {
				m.cursor++
			}
		}
	}
	return m, nil
}

func (m model) View() string {
	if m.done {
		return ""
	}

	s := headerStyle.Render(m.prompt)
	s += "\n"

	for i, item := range m.items {
		cursor := "  "
		if m.cursor == i {
			cursor = "▶ "
		}

		if m.cursor == i {
			s += cursor + selectedStyle.Render(item) + "\n"
		} else {
			s += cursor + normalStyle.Render(item) + "\n"
		}
	}

	return s
}

func CliSelect(prompt string, items []string, defaultIdx int) string {
	if len(items) == 0 {
		return ""
	}

	if defaultIdx < 0 || defaultIdx >= len(items) {
		defaultIdx = 0
	}

	m := model{
		items:      items,
		cursor:     defaultIdx,
		selected:   defaultIdx,
		defaultIdx: defaultIdx,
		done:       false,
		prompt:     prompt,
	}

	p := tea.NewProgram(m)

	finalModel, err := p.Run()
	if err != nil {
		fmt.Printf("Error running program: %v\n", err)
		os.Exit(1)
	}

	if f, ok := finalModel.(model); ok {
		if f.selected >= 0 && f.selected < len(items) {
			return items[f.selected]
		}
	}

	return items[defaultIdx]
}
