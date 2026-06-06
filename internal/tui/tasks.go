package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewTasks() string {
	if m.mode == modeView {
		return m.viewTaskDetail()
	}
	return m.viewTasksDashboard()
}

func (m Model) viewTasksDashboard() string {
	refs := m.allTaskRefs()
	start, end, pages := pageBounds(len(refs), m.taskPage)

	var rows []string
	if len(refs) == 0 {
		rows = append(rows, mutedStyle.Render("no tasks yet — press a to add"))
	}
	for i := start; i < end; i++ {
		rows = append(rows, taskCheckLine(refs[i].task, i == m.taskSel, m.layout.contentW))
	}
	list := strings.Join(rows, "\n")

	header := lipgloss.JoinVertical(lipgloss.Left,
		m.centerLine(subtitleStyle.Render("tasks · dashboard")),
		m.ruleLine(),
		list,
	)
	if m.adding {
		prompt := inputPromptStyle.Render("new task: ") + m.addInput.View()
		return lipgloss.JoinVertical(lipgloss.Left, header, "", prompt)
	}
	return m.stackTopBottom(header, m.renderPager(m.taskPage, pages, true))
}

func (m Model) viewTaskDetail() string {
	refs := m.allTaskRefs()
	if m.taskSel < 0 || m.taskSel >= len(refs) {
		return m.centerLine(subtitleStyle.Render("tasks · view")) + "\n" + m.ruleLine() +
			"\n" + mutedStyle.Render("no task selected")
	}
	if m.editing {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.centerLine(subtitleStyle.Render("tasks · edit")),
			m.ruleLine(),
			m.renderEditForm(),
		)
	}
	r := refs[m.taskSel]
	return lipgloss.JoinVertical(lipgloss.Left,
		m.centerLine(subtitleStyle.Render("tasks · view")),
		m.ruleLine(),
		renderTaskDetail(r.task, r.project, m.layout.contentW),
	)
}

func (m Model) updateTasks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeView {
		return m.updateTaskDetail(msg)
	}
	return m.updateTasksDashboard(msg)
}

func (m Model) updateTasksDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.adding {
		return m.updateTaskAdd(msg)
	}
	refs := m.allTaskRefs()
	switch msg.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		m.taskMove(-1, len(refs))
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		m.taskMove(1, len(refs))
		return m, nil
	case tea.KeyLeft:
		if m.taskPage > 0 {
			m.taskPage--
			m.taskSel = m.taskPage * pageSize
		}
		return m, nil
	case tea.KeyRight:
		_, _, pages := pageBounds(len(refs), m.taskPage)
		if m.taskPage < pages-1 {
			m.taskPage++
			m.taskSel = m.taskPage * pageSize
		}
		return m, nil
	case tea.KeySpace:
		if m.taskSel >= 0 && m.taskSel < len(refs) {
			refs[m.taskSel].task.Toggle()
			m.save()
		}
		return m, nil
	case tea.KeyEnter:
		if m.taskSel >= 0 && m.taskSel < len(refs) {
			m.mode = modeView
		}
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'a':
				m.adding = true
				m.addInput.Focus()
				m.addInput.SetValue("")
				return m, textinputBlink()
			case 'd':
				if m.taskSel >= 0 && m.taskSel < len(refs) {
					title := refs[m.taskSel].task.Title
					m.doc.Remove(refs[m.taskSel].task)
					m.save()
					m.status = "deleted: " + title
					if m.taskSel >= len(refs)-1 && m.taskSel > 0 {
						m.taskSel--
					}
					m.taskPage = m.taskSel / pageSize
				}
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) updateTaskDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	refs := m.allTaskRefs()
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeDashboard
		return m, nil
	case tea.KeyUp, tea.KeyCtrlP:
		m.taskMove(-1, len(refs))
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		m.taskMove(1, len(refs))
		return m, nil
	case tea.KeySpace:
		if m.taskSel >= 0 && m.taskSel < len(refs) {
			refs[m.taskSel].task.Toggle()
			m.save()
		}
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'e':
				if m.taskSel >= 0 && m.taskSel < len(refs) {
					m.beginTaskEdit(refs[m.taskSel])
					return m, textinputBlink()
				}
			case 's':
				m.save()
				m.status = "saved"
				return m, nil
			}
		}
	}
	return m, nil
}

func (m Model) updateTaskAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		title := strings.TrimSpace(m.addInput.Value())
		if title != "" {
			target := m.doc.EnsureInbox()
			if refs := m.allTaskRefs(); m.taskSel >= 0 && m.taskSel < len(refs) {
				target = refs[m.taskSel].project
			}
			m.doc.AddToProject(target, title)
			m.save()
			m.status = "added: " + title
		}
		m.adding = false
		m.addInput.Blur()
		m.addInput.SetValue("")
		return m, nil
	case tea.KeyEsc:
		m.adding = false
		m.addInput.Blur()
		m.addInput.SetValue("")
		return m, nil
	}
	var cmd tea.Cmd
	m.addInput, cmd = m.addInput.Update(msg)
	return m, cmd
}

func (m *Model) taskMove(delta, total int) {
	if total == 0 {
		return
	}
	m.taskSel += delta
	if m.taskSel < 0 {
		m.taskSel = total - 1
	}
	if m.taskSel >= total {
		m.taskSel = 0
	}
	m.taskPage = m.taskSel / pageSize
}
