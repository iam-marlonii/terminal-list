package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

func (m Model) viewBoard() string {
	refs := m.allTaskRefs()
	menuW, detailW := m.menuDetailWidths()
	start, end, pages := pageBounds(len(refs), m.boardPage)

	var lines []string
	lines = append(lines, sectionStyle.Render("~ task box ~"))
	if len(refs) == 0 {
		lines = append(lines, mutedStyle.Render("no tasks yet"))
	}
	for i := start; i < end; i++ {
		r := refs[i]
		title := truncate(strings.ToLower(r.task.Title), menuW-2)
		switch {
		case i == m.boardSel:
			lines = append(lines, selectedItemStyle.Width(menuW).Render(title))
		case r.task.Done():
			lines = append(lines, doneItemStyle.Render(title))
		default:
			lines = append(lines, menuItemStyle.Render(title))
		}
	}
	menu := strings.Join(lines, "\n")
	menuCol := lipgloss.JoinVertical(lipgloss.Left, menu, "", m.renderPager(m.boardPage, pages, false))

	detail := mutedStyle.Render("select a task with ↑/↓")
	if m.boardSel >= 0 && m.boardSel < len(refs) {
		r := refs[m.boardSel]
		detail = renderTaskDetail(r.task, r.project, detailW-2)
	}

	menuBox := lipgloss.NewStyle().Width(menuW).Render(menuCol)
	detailBox := lipgloss.NewStyle().Width(detailW).Render(detail)
	return lipgloss.JoinHorizontal(lipgloss.Top, menuBox, " ", detailBox)
}

func (m Model) updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	refs := m.allTaskRefs()
	switch msg.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		m.boardMove(-1, len(refs))
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		m.boardMove(1, len(refs))
		return m, nil
	case tea.KeySpace:
		if m.boardSel >= 0 && m.boardSel < len(refs) {
			refs[m.boardSel].task.Toggle()
			m.save()
		}
		return m, nil
	case tea.KeyEnter:
		if m.boardSel >= 0 && m.boardSel < len(refs) {
			m.page = PageTasks
			m.mode = modeView
			m.taskSel = m.boardSel
			m.taskPage = m.boardSel / pageSize
		}
		return m, nil
	case tea.KeyEsc:
		m.save()
		m.quitting = true
		return m, tea.Quit
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'd' {
			if m.boardSel >= 0 && m.boardSel < len(refs) {
				title := refs[m.boardSel].task.Title
				m.doc.Remove(refs[m.boardSel].task)
				m.save()
				m.status = "deleted: " + title
				if m.boardSel >= len(refs)-1 && m.boardSel > 0 {
					m.boardSel--
				}
				m.boardPage = m.boardSel / pageSize
			}
			return m, nil
		}
	}
	return m, nil
}

func (m *Model) boardMove(delta, total int) {
	if total == 0 {
		return
	}
	m.boardSel += delta
	if m.boardSel < 0 {
		m.boardSel = total - 1
	}
	if m.boardSel >= total {
		m.boardSel = 0
	}
	m.boardPage = m.boardSel / pageSize
}
