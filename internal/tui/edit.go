package tui

import (
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iam-marlonjr/terminal-list/internal/store"
)

type editKind int

const (
	editTaskKind editKind = iota
	editProjectKind
)

type editField struct {
	label string
	value string
}

func (m *Model) beginTaskEdit(r taskRef) {
	m.editing = true
	m.editKind = editTaskKind
	m.editTask = r.task
	m.editProject = nil
	due := ""
	if r.task.Due != nil {
		due = r.task.Due.Format("2006-01-02")
	}
	m.editFields = []editField{
		{"title", r.task.Title},
		{"due", due},
		{"description", r.task.Description},
	}
	m.editIndex = 0
	m.focusEditField()
}

func (m *Model) beginProjectEdit(p *store.Project) {
	m.editing = true
	m.editKind = editProjectKind
	m.editProject = p
	m.editTask = nil
	due := ""
	if p.Due != nil {
		due = p.Due.Format("2006-01-02")
	}
	m.editFields = []editField{
		{"name", p.Name},
		{"status", projectStatusText(p)},
		{"due", due},
		{"area", p.Area},
	}
	m.editIndex = 0
	m.focusEditField()
}

func (m *Model) focusEditField() {
	m.editInput.SetValue(m.editFields[m.editIndex].value)
	m.editInput.CursorEnd()
	m.editInput.Focus()
}

// commitField stores the current input back into the focused field.
func (m *Model) commitField() {
	m.editFields[m.editIndex].value = m.editInput.Value()
}

func (m Model) renderEditForm() string {
	var lines []string
	for i, f := range m.editFields {
		if i == m.editIndex {
			lines = append(lines, inputPromptStyle.Render(f.label+": ")+m.editInput.View())
		} else {
			lines = append(lines, mutedStyle.Render(f.label+": ")+menuItemStyle.Render(f.value))
		}
	}
	help := mutedStyle.Render("↑/↓ move between fields · enter to save · esc to cancel")
	return lipgloss.JoinVertical(lipgloss.Left, strings.Join(lines, "\n"), "", help)
}

func (m Model) updateEdit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.endEdit()
		m.status = "edit cancelled"
		return m, nil
	case tea.KeyEnter:
		m.commitField()
		m.applyEdit()
		m.save()
		m.endEdit()
		m.status = "saved"
		return m, nil
	case tea.KeyUp, tea.KeyShiftTab:
		m.commitField()
		if m.editIndex > 0 {
			m.editIndex--
		} else {
			m.editIndex = len(m.editFields) - 1
		}
		m.focusEditField()
		return m, nil
	case tea.KeyDown, tea.KeyTab:
		m.commitField()
		if m.editIndex < len(m.editFields)-1 {
			m.editIndex++
		} else {
			m.editIndex = 0
		}
		m.focusEditField()
		return m, nil
	}
	var cmd tea.Cmd
	m.editInput, cmd = m.editInput.Update(msg)
	return m, cmd
}

func (m *Model) endEdit() {
	m.editing = false
	m.editInput.Blur()
	m.editInput.SetValue("")
	m.editTask = nil
	m.editProject = nil
	m.editFields = nil
	m.editIndex = 0
}

func (m *Model) applyEdit() {
	field := func(label string) string {
		for _, f := range m.editFields {
			if f.label == label {
				return strings.TrimSpace(f.value)
			}
		}
		return ""
	}
	switch m.editKind {
	case editTaskKind:
		if m.editTask == nil {
			return
		}
		if title := field("title"); title != "" {
			m.editTask.Title = title
		}
		m.editTask.Due = parseDueInput(field("due"))
		m.editTask.Description = field("description")
	case editProjectKind:
		if m.editProject == nil {
			return
		}
		if name := field("name"); name != "" {
			m.editProject.Name = name
		}
		if status := field("status"); status != "" {
			m.editProject.Status = store.ProjectStatus(status)
		}
		m.editProject.Due = parseDueInput(field("due"))
		m.editProject.Area = field("area")
	}
}

// parseDueInput accepts "2006-01-02" or "2 Jan 2006" (case-insensitive); empty clears.
func parseDueInput(s string) *time.Time {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	for _, layout := range []string{"2006-01-02", "2 Jan 2006", "2 January 2006"} {
		if ts, err := time.Parse(layout, normalizeMonthCase(s)); err == nil {
			return &ts
		}
	}
	return nil
}

// normalizeMonthCase title-cases tokens so "jun" parses against the "Jan" layout.
func normalizeMonthCase(s string) string {
	parts := strings.Fields(s)
	for i, p := range parts {
		if len(p) > 0 {
			parts[i] = strings.ToUpper(p[:1]) + strings.ToLower(p[1:])
		}
	}
	return strings.Join(parts, " ")
}
