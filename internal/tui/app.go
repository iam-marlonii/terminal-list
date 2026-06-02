// Package tui implements the multi-page terminal UI.
package tui

import (
	"strings"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iam-marlonjr/terminal-list/internal/config"
	"github.com/iam-marlonjr/terminal-list/internal/store"
)

// Page is a top-level screen.
type Page int

const (
	PageDashboard Page = iota
	PageProjects
	PageTasks
	PageImport
	PagePreview
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	tabActive   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212")).Underline(true)
	tabInactive = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	helpStyle   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	statusStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("214"))
	mutedStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("245"))
)

// Model is the root Bubble Tea model.
type Model struct {
	doc      *store.Document
	path     string
	page     Page
	width    int
	height   int
	status   string
	quitting bool

	// navigation context
	selectedProject *store.Project

	// sub-models
	projectList list.Model
	taskList    list.Model
	importInput textinput.Model
	addInput    textinput.Model
	adding      bool
	previewMD   string
}

// New constructs the app for the given document and path.
func New(doc *store.Document, path string) Model {
	m := Model{
		doc:  doc,
		path: path,
		page: PageDashboard,
	}
	m.projectList = newProjectList(doc, 24, 12)
	m.taskList = newTaskList(nil, 24, 12)
	m.importInput = textinput.New()
	m.importInput.Placeholder = "path/to/document.pdf"
	m.importInput.CharLimit = 512
	m.importInput.Width = 60
	m.addInput = textinput.New()
	m.addInput.Placeholder = "New task…"
	m.addInput.CharLimit = 200
	m.addInput.Width = 50
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.resizeLists()
		return m, nil
	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}
		if m.page == PagePreview {
			return m.updatePreview(msg)
		}
		if m.adding {
			return m.updateAdd(msg)
		}
		return m.updateKey(msg)
	}
	return m, nil
}

func (m *Model) resizeLists() {
	h := m.height - 8
	if h < 4 {
		h = 4
	}
	w := m.width - 4
	if w < 20 {
		w = 20
	}
	m.projectList.SetSize(w, h)
	m.taskList.SetSize(w, h)
}

func (m *Model) save() {
	if err := m.doc.Save(m.path); err != nil {
		m.status = "save error: " + err.Error()
	} else {
		m.status = ""
	}
}

func (m Model) updateKey(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyCtrlC:
		m.save()
		m.quitting = true
		return m, tea.Quit
	case tea.KeyRunes:
		if len(msg.Runes) == 1 {
			switch msg.Runes[0] {
			case '1':
				m.page = PageDashboard
				return m, nil
			case '2':
				m.page = PageProjects
				m.projectList = newProjectList(m.doc, listWidth(m), listHeight(m))
				return m, nil
			case '3':
				if m.selectedProject == nil && len(m.doc.Projects) > 0 {
					m.selectedProject = m.doc.Projects[0]
				}
				m.page = PageTasks
				m.taskList = newTaskList(m.selectedProject, listWidth(m), listHeight(m))
				return m, nil
			case '4':
				m.page = PageImport
				return m, nil
			}
		}
	}

	switch m.page {
	case PageDashboard:
		return m.updateDashboard(msg)
	case PageProjects:
		return m.updateProjects(msg)
	case PageTasks:
		return m.updateTasks(msg)
	case PageImport:
		return m.updateImport(msg)
	}
	return m, nil
}

func listWidth(m Model) int {
	w := m.width - 2
	if w < 20 {
		return 20
	}
	return w
}

func listHeight(m Model) int {
	return contentHeight(m)
}

// contentHeight is the vertical space for the main panel (between header and footer).
func contentHeight(m Model) int {
	h := m.height - 4
	if m.status != "" {
		h--
	}
	if h < 4 {
		return 4
	}
	return h
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	var b strings.Builder
	b.WriteString(m.renderTabs())
	b.WriteString("\n\n")

	var body string
	switch m.page {
	case PageDashboard:
		body = m.viewDashboard()
	case PageProjects:
		body = m.viewProjects()
	case PageTasks:
		body = m.viewTasks()
	case PageImport:
		body = m.viewImport()
	case PagePreview:
		body = m.viewPreview()
	}
	b.WriteString(body)

	if m.status != "" {
		b.WriteString("\n" + statusStyle.Render(m.status))
	}
	b.WriteString("\n" + m.renderFooter())
	return b.String()
}

func (m Model) renderTabs() string {
	tabs := []struct {
		key  string
		name string
		page Page
	}{
		{"1", "Dashboard", PageDashboard},
		{"2", "Projects", PageProjects},
		{"3", "Tasks", PageTasks},
		{"4", "Import", PageImport},
	}
	var parts []string
	for _, t := range tabs {
		style := tabInactive
		if m.page == t.page || (m.page == PagePreview && t.page == PageImport) {
			style = tabActive
		}
		parts = append(parts, style.Render(t.key+":"+t.name))
	}
	line := strings.Join(parts, "  ")
	return titleStyle.Render("todo") + "  " + line + mutedStyle.Render("  |  "+m.path)
}

func (m Model) renderFooter() string {
	if m.page == PagePreview {
		return helpStyle.Render("y: merge import  •  n: cancel  •  q: quit")
	}
	if m.adding {
		return helpStyle.Render("enter: save  •  esc: cancel")
	}
	base := "1-4: pages  •  q: quit & save"
	switch m.page {
	case PageProjects:
		return helpStyle.Render(base + "  •  enter: open tasks  •  j/k: move")
	case PageTasks:
		return helpStyle.Render(base + "  •  space: toggle  •  a: add  •  d: delete  •  esc: projects")
	case PageImport:
		return helpStyle.Render(base + "  •  enter: run import (CLI)  •  type path")
	default:
		return helpStyle.Render(base + "  •  enter: go to tasks")
	}
}

// Run starts the TUI program.
func Run(doc *store.Document, path string) error {
	opts := []tea.ProgramOption{}
	if config.UseAltScreen() {
		opts = append(opts, tea.WithAltScreen())
	}
	p := tea.NewProgram(New(doc, path), opts...)
	_, err := p.Run()
	return err
}

// SetPreview switches to preview page (called from import flow via message).
func (m *Model) SetPreview(md string) {
	m.previewMD = md
	m.page = PagePreview
}
