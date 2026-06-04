package tui

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

type headerTab struct {
	key  string
	name string
	page Page
}

func headerTabs() []headerTab {
	return []headerTab{
		{"1", "board", PageBoard},
		{"2", "projects", PageProjects},
		{"3", "tasks", PageTasks},
		{"4", "import", PageImport},
	}
}

func (m Model) renderHeader() string {
	div := headerDividerStyle.Render("│")
	var segments []string
	segments = append(segments, brandStyle.Render("todo"))
	for _, t := range headerTabs() {
		active := m.page == t.page || (m.page == PagePreview && t.page == PageImport)
		key := headerKeyStyle.Render(t.key)
		label := t.name
		var seg string
		if active {
			seg = headerSegActiveStyle.Render(key + " " + label)
		} else {
			seg = headerSegStyle.Render(key + " " + label)
		}
		segments = append(segments, seg)
	}
	open := m.doc.OpenCountTotal()
	cart := fmt.Sprintf("%d open", open)
	segments = append(segments, headerSegStyle.Render(cart))
	line := lipgloss.JoinHorizontal(lipgloss.Top, segments[0], div)
	for i := 1; i < len(segments); i++ {
		line = lipgloss.JoinHorizontal(lipgloss.Top, line, div, segments[i])
	}
	return baseStyle.Width(m.layout.containerW - 2).Render(line)
}
