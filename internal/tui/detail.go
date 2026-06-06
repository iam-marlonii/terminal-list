package tui

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/lipgloss"

	"github.com/iam-marlonjr/terminal-list/internal/store"
	"github.com/iam-marlonjr/terminal-list/internal/task"
)

// fmtDate renders a date the way the mockups show it ("4 jun 2026").
func fmtDate(t time.Time) string {
	return strings.ToLower(t.Format("2 Jan 2006"))
}

// taskStatusText maps a task to its display status.
func taskStatusText(t *task.Task) string {
	if t.Done() {
		return "done"
	}
	return "active"
}

func metaLine(label, value string) string {
	return mutedStyle.Render(label+": ") + mutedStyle.Render(value)
}

// renderTaskDetail draws the shared task detail block (title, meta, description).
func renderTaskDetail(t *task.Task, project *store.Project, width int) string {
	var lines []string
	box := "[ ]"
	if t.Done() {
		box = "[x]"
	}
	lines = append(lines, titleStyle.Render(box+" "+strings.ToLower(t.Title)))
	lines = append(lines, metaLine("status", taskStatusText(t)))
	lines = append(lines, metaLine("created", fmtDate(t.Created)))
	if t.Due != nil {
		lines = append(lines, metaLine("due", fmtDate(*t.Due)))
	}
	if project != nil {
		if project.Area != "" {
			lines = append(lines, metaLine("area", project.Area))
		}
		lines = append(lines, metaLine("project", strings.ToLower(project.Name)))
	}
	lines = append(lines, ruleStyle.Render(strings.Repeat("─", width)))
	desc := t.Description
	if desc == "" {
		desc = mutedStyle.Render("no description")
	} else {
		desc = lipgloss.NewStyle().Width(width).Render(strings.ToLower(desc))
	}
	lines = append(lines, desc)
	return strings.Join(lines, "\n")
}

// progressBar renders a fixed-width bar of "|" scaled to percent.
func progressBar(percent float64, width int) string {
	if width < 1 {
		width = 1
	}
	filled := int(percent/100*float64(width) + 0.5)
	if filled > width {
		filled = width
	}
	return strings.Repeat("|", filled)
}

func projectStatusText(p *store.Project) string {
	if p.Status == "" {
		return string(store.StatusInProgress)
	}
	return string(p.Status)
}

func taskCheckLine(t *task.Task, selected bool, width int) string {
	box := "[ ]"
	if t.Done() {
		box = "[x]"
	}
	text := box + " " + strings.ToLower(t.Title)
	if selected {
		return selectedItemStyle.Width(width).Render(text)
	}
	if t.Done() {
		return doneItemStyle.Render(text)
	}
	return menuItemStyle.Render(text)
}

// projectMetaLines is shared by the projects view and the import review card.
func projectMetaLines(p *store.Project, includeProgress, includeRelated bool) []string {
	var lines []string
	lines = append(lines, titleStyle.Render(strings.ToLower(p.Name)))
	lines = append(lines, metaLine("status", projectStatusText(p)))
	if includeProgress {
		lines = append(lines, metaLine("progress", fmt.Sprintf("%.2f%%", p.ProgressPercent())))
	}
	lines = append(lines, metaLine("created", fmtDate(p.Created)))
	if p.Due != nil {
		lines = append(lines, metaLine("due", fmtDate(*p.Due)))
	}
	if p.Area != "" {
		lines = append(lines, metaLine("area", p.Area))
	}
	if includeRelated {
		lines = append(lines, metaLine("related tasks", fmt.Sprintf("%d", p.TaskCount())))
	}
	return lines
}
