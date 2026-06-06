package tui

import (
	"github.com/charmbracelet/lipgloss"
)

type headerTab struct {
	key  string
	name string
	page Page
}

func headerTabs() []headerTab {
	return []headerTab{
		{"b", "board", PageBoard},
		{"t", "tasks", PageTasks},
		{"p", "projects", PageProjects},
		{"i", "import", PageImport},
	}
}

func (m Model) renderHeader() string {
	segments := []string{brandStyle.Render("kanban")}
	for _, t := range headerTabs() {
		active := m.page == t.page
		label := tabKeyStyle.Render(t.key) + " " + t.name
		if active {
			segments = append(segments, tabActiveStyle.Render(label))
		} else {
			segments = append(segments, tabStyle.Render(label))
		}
	}
	bar := lipgloss.JoinHorizontal(lipgloss.Center, segments...)
	return baseStyle.Width(m.layout.contentW).
		Align(lipgloss.Center).
		Render(bar)
}
