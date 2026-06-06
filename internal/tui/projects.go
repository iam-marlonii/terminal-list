package tui

import (
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iam-marlonjr/terminal-list/internal/store"
)

const projectCardW = 30

func (m Model) viewProjects() string {
	if m.mode == modeView {
		return m.viewProjectDetail()
	}
	return m.viewProjectsDashboard()
}

func (m Model) viewProjectsDashboard() string {
	projects := m.doc.Projects
	start, end, pages := pageBounds(len(projects), m.projPage)

	var cards []string
	for i := start; i < end; i++ {
		cards = append(cards, m.renderProjectCard(projects[i], i == m.projSel))
	}
	grid := m.gridCards(cards)
	if len(projects) == 0 {
		grid = mutedStyle.Render("no projects yet")
	}

	header := lipgloss.JoinVertical(lipgloss.Left,
		m.centerLine(subtitleStyle.Render("projects · dashboard")),
		m.ruleLine(),
		grid,
	)
	return m.stackTopBottom(header, m.renderPager(m.projPage, pages, true))
}

func (m Model) renderProjectCard(p *store.Project, selected bool) string {
	lines := []string{
		titleStyle.Render(strings.ToLower(p.Name)),
		metaLine("status", projectStatusText(p)),
	}
	if p.Due != nil {
		lines = append(lines, metaLine("due", fmtDate(*p.Due)))
	}
	if p.Area != "" {
		lines = append(lines, metaLine("area", p.Area))
	}
	lines = append(lines, metaLine("progress", progressBar(p.ProgressPercent(), 6)))
	body := strings.Join(lines, "\n")

	style := cardStyle.Width(projectCardW)
	if selected {
		style = style.BorderForeground(lipgloss.Color(colorAccent))
	}
	return style.Render(body)
}

// gridCards arranges cards into rows that fit the content width.
func (m Model) gridCards(cards []string) string {
	if len(cards) == 0 {
		return ""
	}
	perRow := m.layout.contentW / (projectCardW + 3)
	if perRow < 1 {
		perRow = 1
	}
	var rows []string
	for i := 0; i < len(cards); i += perRow {
		end := i + perRow
		if end > len(cards) {
			end = len(cards)
		}
		row := lipgloss.JoinHorizontal(lipgloss.Top, spaceJoin(cards[i:end])...)
		rows = append(rows, row)
	}
	return strings.Join(rows, "\n")
}

func spaceJoin(items []string) []string {
	out := make([]string, 0, len(items)*2)
	for i, it := range items {
		if i > 0 {
			out = append(out, "  ")
		}
		out = append(out, it)
	}
	return out
}

func (m Model) viewProjectDetail() string {
	p := m.currentProject()
	if p == nil {
		return m.centerLine(subtitleStyle.Render("projects · view")) + "\n" + m.ruleLine() +
			"\n" + mutedStyle.Render("no project selected")
	}
	if m.editing {
		return lipgloss.JoinVertical(lipgloss.Left,
			m.centerLine(subtitleStyle.Render("projects · edit")),
			m.ruleLine(),
			m.renderEditForm(),
		)
	}

	meta := projectMetaLines(p, true, true)
	tasks := p.Tasks()
	var taskLines []string
	taskLines = append(taskLines, ruleStyle.Render(strings.Repeat("─", m.layout.contentW)))
	if len(tasks) == 0 {
		taskLines = append(taskLines, mutedStyle.Render("no related tasks"))
	}
	for i, t := range tasks {
		taskLines = append(taskLines, taskCheckLine(t, i == m.projTaskSel, m.layout.contentW))
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		m.centerLine(subtitleStyle.Render("projects · view")),
		m.ruleLine(),
		strings.Join(meta, "\n"),
		strings.Join(taskLines, "\n"),
	)
}

func (m Model) currentProject() *store.Project {
	if m.projSel < 0 || m.projSel >= len(m.doc.Projects) {
		return nil
	}
	return m.doc.Projects[m.projSel]
}

func (m Model) updateProjects(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.mode == modeView {
		return m.updateProjectDetail(msg)
	}
	return m.updateProjectsDashboard(msg)
}

func (m Model) updateProjectsDashboard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	n := len(m.doc.Projects)
	switch msg.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		m.projMove(-1, n)
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		m.projMove(1, n)
		return m, nil
	case tea.KeyEnter:
		if n > 0 {
			m.mode = modeView
			m.projTaskSel = 0
		}
		return m, nil
	}
	return m, nil
}

func (m Model) updateProjectDetail(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	p := m.currentProject()
	switch msg.Type {
	case tea.KeyEsc:
		m.mode = modeDashboard
		return m, nil
	case tea.KeyUp, tea.KeyCtrlP:
		if p != nil {
			m.projTaskMove(-1, len(p.Tasks()))
		}
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		if p != nil {
			m.projTaskMove(1, len(p.Tasks()))
		}
		return m, nil
	case tea.KeySpace:
		if p != nil {
			if tasks := p.Tasks(); m.projTaskSel >= 0 && m.projTaskSel < len(tasks) {
				tasks[m.projTaskSel].Toggle()
				m.save()
			}
		}
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'e':
				if p != nil {
					m.beginProjectEdit(p)
					return m, textinputBlink()
				}
			case 'd':
				if p != nil {
					if tasks := p.Tasks(); m.projTaskSel >= 0 && m.projTaskSel < len(tasks) {
						title := tasks[m.projTaskSel].Title
						m.doc.Remove(tasks[m.projTaskSel])
						m.save()
						m.status = "deleted: " + title
						if m.projTaskSel > 0 && m.projTaskSel >= len(p.Tasks()) {
							m.projTaskSel--
						}
					}
				}
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *Model) projMove(delta, total int) {
	if total == 0 {
		return
	}
	m.projSel += delta
	if m.projSel < 0 {
		m.projSel = total - 1
	}
	if m.projSel >= total {
		m.projSel = 0
	}
	m.projPage = m.projSel / pageSize
}

func (m *Model) projTaskMove(delta, total int) {
	if total == 0 {
		return
	}
	m.projTaskSel += delta
	if m.projTaskSel < 0 {
		m.projTaskSel = total - 1
	}
	if m.projTaskSel >= total {
		m.projTaskSel = 0
	}
}
