package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

// viewDashboard is kept for summary stats embedded in board detail; board is the main view.
func (m Model) viewDashboard() string {
	var b strings.Builder
	open := m.doc.OpenCountTotal()
	doneToday := m.doc.DoneTodayCount()
	fmt.Fprintf(&b, "%s\n\n", titleStyle.Render("overview"))
	fmt.Fprintf(&b, "%s %d    %s %d\n\n",
		mutedStyle.Render("open:"), open,
		mutedStyle.Render("done today:"), doneToday)

	b.WriteString(titleStyle.Render("projects") + "\n")
	for _, p := range m.doc.Projects {
		status := string(p.Status)
		if status == "" {
			status = "active"
		}
		fmt.Fprintf(&b, "  %s  %s  (%d open)\n", strings.ToLower(p.Name), mutedStyle.Render("["+status+"]"), p.OpenCount())
	}

	b.WriteString("\n" + titleStyle.Render("next up") + "\n")
	up := m.doc.AllOpenTasks(8)
	if len(up) == 0 {
		b.WriteString(mutedStyle.Render("  no open tasks — press 1 for board\n"))
	} else {
		for _, item := range up {
			fmt.Fprintf(&b, "  [ ] %s  %s\n", strings.ToLower(item.Task.Title), mutedStyle.Render("("+item.Project.Name+")"))
		}
	}
	return b.String()
}

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		m.page = PageBoard
		m.refreshBoardEntries()
	}
	return m, nil
}
