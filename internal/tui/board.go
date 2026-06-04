package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iam-marlonjr/terminal-list/internal/store"
	"github.com/iam-marlonjr/terminal-list/internal/task"
)

type boardEntryKind int

const (
	entryProject boardEntryKind = iota
	entryGroup
	entryTask
)

type boardEntry struct {
	kind    boardEntryKind
	label   string
	project *store.Project
	task    *task.Task
}

type boardState struct {
	entries     []boardEntry
	selected    int
	menuScroll  int
	detailVP    viewport.Model
	detailReady bool
}

func buildBoardEntries(doc *store.Document) []boardEntry {
	var entries []boardEntry
	for _, p := range doc.Projects {
		entries = append(entries, boardEntry{kind: entryProject, label: p.Name, project: p})
		for _, g := range p.Groups {
			if g.Label != "" {
				entries = append(entries, boardEntry{kind: entryGroup, label: g.Label, project: p})
			}
			for _, t := range g.Tasks {
				entries = append(entries, boardEntry{kind: entryTask, task: t, project: p})
			}
		}
	}
	return entries
}

func firstTaskIndex(entries []boardEntry) int {
	for i, e := range entries {
		if e.kind == entryTask {
			return i
		}
	}
	return -1
}

func (m *Model) refreshBoardEntries() {
	m.board.entries = buildBoardEntries(m.doc)
	if m.board.selected >= len(m.board.entries) {
		m.board.selected = 0
	}
	if len(m.board.entries) > 0 && m.board.selected < 0 {
		m.board.selected = 0
	}
	if idx := firstTaskIndex(m.board.entries); idx >= 0 && m.board.selected == 0 {
		if m.board.entries[0].kind != entryTask {
			m.board.selected = idx
		}
	}
	_, menuH := m.boardHeights()
	m.ensureMenuScroll(menuH)
	m.syncBoardDetail()
}

func (m *Model) syncBoardDetail() {
	content := m.renderBoardDetail()
	if !m.board.detailReady {
		menuW, _ := m.menuDetailWidths()
		_, detailH := m.boardHeights()
		m.board.detailVP = viewport.New(menuW, detailH)
		m.board.detailVP.KeyMap = detailViewportKeys()
		m.board.detailReady = true
	}
	_, detailH := m.boardHeights()
	_, detailW := m.menuDetailWidths()
	m.board.detailVP.Width = detailW
	m.board.detailVP.Height = detailH
	m.board.detailVP.SetContent(content)
}

func detailViewportKeys() viewport.KeyMap {
	km := viewport.DefaultKeyMap()
	km.PageDown.SetEnabled(true)
	km.PageUp.SetEnabled(true)
	km.HalfPageDown.SetEnabled(true)
	km.HalfPageUp.SetEnabled(true)
	km.Down.SetEnabled(true)
	km.Up.SetEnabled(true)
	return km
}

func (m *Model) boardSelectedEntry() *boardEntry {
	if m.board.selected < 0 || m.board.selected >= len(m.board.entries) {
		return nil
	}
	e := m.board.entries[m.board.selected]
	return &e
}

func (m *Model) boardSelectedTask() (*task.Task, *store.Project) {
	e := m.boardSelectedEntry()
	if e == nil || e.kind != entryTask {
		return nil, nil
	}
	return e.task, e.project
}

func (m Model) renderBoardMenu() string {
	menuW, _ := m.menuDetailWidths()
	_, menuH := m.boardHeights()
	var lines []string
	accent := colorAccent
	if e := m.boardSelectedEntry(); e != nil && e.project != nil {
		accent = string(accentForProject(e.project.Color))
	}
	selStyle := selectedStyleForAccent(lipgloss.Color(accent))

	start := m.board.menuScroll
	if start < 0 {
		start = 0
	}
	end := start + menuH
	if end > len(m.board.entries) {
		end = len(m.board.entries)
	}

	for i := start; i < end; i++ {
		e := m.board.entries[i]
		var line string
		switch e.kind {
		case entryProject:
			line = sectionStyle.Render(fmt.Sprintf("~ %s ~", strings.ToLower(e.label)))
		case entryGroup:
			line = sectionStyle.Render("  " + strings.ToLower(e.label))
		case entryTask:
			title := strings.ToLower(e.task.Title)
			if len(title) > menuW-4 {
				title = title[:menuW-7] + "..."
			}
			if e.task.Done() {
				line = doneItemStyle.Render("  " + title)
			} else {
				line = menuItemStyle.Render("  " + title)
			}
			if i == m.board.selected {
				line = selStyle.Render("  " + title)
			}
		}
		if e.kind != entryTask && i == m.board.selected {
			line = selStyle.Render(strings.TrimSpace(line))
		}
		lines = append(lines, line)
	}
	if len(lines) == 0 {
		lines = append(lines, mutedStyle.Render("no tasks yet — press a to add"))
	}
	return strings.Join(lines, "\n")
}

func (m Model) renderBoardDetail() string {
	task, project := m.boardSelectedTask()
	if task == nil {
		open := m.doc.OpenCountTotal()
		done := m.doc.DoneTodayCount()
		return strings.Join([]string{
			titleStyle.Render("board"),
			"",
			mutedStyle.Render(fmt.Sprintf("%d open · %d done today", open, done)),
			"",
			mutedStyle.Render("select a task with ↑/↓"),
		}, "\n")
	}

	var b strings.Builder
	b.WriteString(titleStyle.Render(strings.ToLower(task.Title)))
	b.WriteString("\n")
	if project != nil {
		b.WriteString(mutedStyle.Render(strings.ToLower(project.Name)))
		if project.Status != "" {
			b.WriteString(mutedStyle.Render(" · " + string(project.Status)))
		}
		b.WriteString("\n")
	}
	b.WriteString("\n")
	state := "open"
	if task.Done() {
		state = "done"
	}
	b.WriteString(mutedStyle.Render("status: " + state))
	b.WriteString("\n")
	b.WriteString(mutedStyle.Render("created: " + task.Created.Format("2 jan 2006")))
	if task.Completed != nil {
		b.WriteString("\n")
		b.WriteString(mutedStyle.Render("completed: " + task.Completed.Format("2 jan 2006")))
	}
	b.WriteString("\n\n")
	if task.Done() {
		b.WriteString(mutedStyle.Render("marked complete."))
	} else {
		b.WriteString(mutedStyle.Render("a task in your local markdown store."))
	}
	b.WriteString("\n\n")
	btn := buttonStyle.Render("toggle")
	b.WriteString(btn + " " + footerActionStyle.Render("enter"))
	return b.String()
}

func (m Model) viewBoard() string {
	if m.adding {
		return m.viewBoardAdd()
	}
	menu := m.renderBoardMenu()
	detail := m.board.detailVP.View()
	menuW, detailW := m.menuDetailWidths()
	menuH, detailH := m.boardHeights()

	menuBox := lipgloss.NewStyle().Width(menuW).Height(menuH).Render(menu)
	if m.layout.stacked {
		return lipgloss.JoinVertical(lipgloss.Left, menuBox, detail)
	}
	detailBox := lipgloss.NewStyle().Width(detailW).Height(detailH).Render(detail)
	return lipgloss.JoinHorizontal(lipgloss.Top, menuBox, " ", detailBox)
}

func (m Model) viewBoardAdd() string {
	menu := m.renderBoardMenu()
	prompt := inputPromptStyle.Render("new task: ") + m.addInput.View()
	return lipgloss.JoinVertical(lipgloss.Left, menu, "", prompt)
}

func (m Model) updateBoard(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.adding {
		return m.updateBoardAdd(msg)
	}

	// Detail pane scroll when not moving selection
	switch msg.Type {
	case tea.KeyPgDown, tea.KeyPgUp, tea.KeyCtrlD, tea.KeyCtrlU:
		if m.board.detailReady {
			var cmd tea.Cmd
			m.board.detailVP, cmd = m.board.detailVP.Update(msg)
			return m, cmd
		}
	}

	switch msg.Type {
	case tea.KeyUp, tea.KeyCtrlP:
		m.boardMoveSelection(-1)
		return m, nil
	case tea.KeyDown, tea.KeyCtrlN:
		m.boardMoveSelection(1)
		return m, nil
	case tea.KeyEnter, tea.KeySpace:
		if t, _ := m.boardSelectedTask(); t != nil {
			t.Toggle()
			m.save()
			m.refreshBoardEntries()
		}
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case 'a':
				m.adding = true
				m.addInput.Focus()
				m.addInput.SetValue("")
				return m, nil
			case 'd':
				if t, _ := m.boardSelectedTask(); t != nil {
					title := t.Title
					m.doc.Remove(t)
					m.save()
					m.status = "deleted: " + title
					m.refreshBoardEntries()
				}
				return m, nil
			}
		}
	}
	return m, nil
}

func (m *Model) ensureMenuScroll(menuH int) {
	if m.board.selected < m.board.menuScroll {
		m.board.menuScroll = m.board.selected
	}
	if m.board.selected >= m.board.menuScroll+menuH {
		m.board.menuScroll = m.board.selected - menuH + 1
	}
}

func (m *Model) boardMoveSelection(delta int) {
	if len(m.board.entries) == 0 {
		return
	}
	n := len(m.board.entries)
	i := m.board.selected
	for step := 0; step < n; step++ {
		i += delta
		if i < 0 {
			i = n - 1
		}
		if i >= n {
			i = 0
		}
		if m.board.entries[i].kind == entryTask {
			m.board.selected = i
			_, menuH := m.boardHeights()
			m.ensureMenuScroll(menuH)
			m.syncBoardDetail()
			m.board.detailVP.GotoTop()
			return
		}
	}
}

func (m Model) updateBoardAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		title := strings.TrimSpace(m.addInput.Value())
		if title != "" {
			if _, p := m.boardSelectedTask(); p != nil {
				m.doc.AddToProject(p, title)
			} else if e := m.boardSelectedEntry(); e != nil && e.project != nil {
				m.doc.AddToProject(e.project, title)
			} else if len(m.doc.Projects) > 0 {
				m.doc.AddToProject(m.doc.Projects[0], title)
			} else {
				t := task.New(title)
				m.doc.AddToInbox(t)
			}
			m.save()
			m.refreshBoardEntries()
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
