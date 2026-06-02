package tui

import (
	"fmt"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
)

func (m Model) viewDashboard() string {
	var b strings.Builder
	open := m.doc.OpenCountTotal()
	doneToday := m.doc.DoneTodayCount()
	fmt.Fprintf(&b, "Open tasks: %d    Completed today: %d\n\n", open, doneToday)

	b.WriteString(titleStyle.Render("Projects") + "\n")
	for _, p := range m.doc.Projects {
		status := string(p.Status)
		if status == "" {
			status = "active"
		}
		fmt.Fprintf(&b, "  • %s  %s  (%d open)\n", p.Name, mutedStyle.Render("["+status+"]"), p.OpenCount())
	}

	b.WriteString("\n" + titleStyle.Render("Next up") + "\n")
	up := m.doc.AllOpenTasks(8)
	if len(up) == 0 {
		b.WriteString(mutedStyle.Render("  No open tasks. Press 3 for Tasks or a to add.\n"))
	} else {
		for _, item := range up {
			fmt.Fprintf(&b, "  [ ] %s  %s\n", item.Task.Title, mutedStyle.Render("("+item.Project.Name+")"))
		}
	}
	return b.String()
}

func (m Model) updateDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if msg.Type == tea.KeyEnter {
		if len(m.doc.Projects) > 0 {
			m.selectedProject = m.doc.Projects[0]
		}
		m.page = PageTasks
		m.taskList = newTaskList(m.selectedProject, listWidth(m), listHeight(m))
	}
	return m, nil
}
