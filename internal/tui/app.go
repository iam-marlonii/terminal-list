// Package tui implements the multi-page terminal UI.
package tui

import (
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
	PageBoard Page = iota
	PageProjects
	PageTasks
	PageImport
	PagePreview
)

// Model is the root Bubble Tea model.
type Model struct {
	doc      *store.Document
	path     string
	page     Page
	layout   layoutState
	status   string
	quitting bool

	// navigation context
	selectedProject *store.Project

	// sub-models
	board       boardState
	projectList list.Model
	taskList    list.Model
	addInput    textinput.Model
	adding      bool
	previewMD   string

	// import page
	importFiles       []string
	importSel         int
	generating        bool
	pendingImportFile string
}

// New constructs the app for the given document and path.
func New(doc *store.Document, path string) Model {
	m := Model{
		doc:  doc,
		path: path,
		page: PageBoard,
	}
	m.refreshBoardEntries()
	m.projectList = newProjectList(doc, 24, 12)
	m.taskList = newTaskList(nil, 24, 12)
	m.refreshImports()
	m.addInput = textinput.New()
	m.addInput.Placeholder = "new task…"
	m.addInput.CharLimit = 200
	m.addInput.Width = 50
	m.addInput.PromptStyle = inputPromptStyle
	m.addInput.TextStyle = menuItemStyle
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.recalcLayout(msg)
		m.resizeLists()
		if m.page == PageBoard {
			m.syncBoardDetail()
		}
		return m, nil
	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}
		if m.page == PagePreview {
			return m.updatePreview(msg)
		}
		if m.adding && m.page != PageBoard {
			return m.updateAdd(msg)
		}
		return m.updateKey(msg)
	case genDoneMsg:
		imported, err := store.ParseMarkdown([]byte(msg.md))
		if err != nil {
			m.generating = false
			m.status = "import error: " + err.Error()
			return m, nil
		}
		normalizeImported(imported)
		m.previewMD = msg.md
		m.pendingImportFile = msg.file
		m.generating = false
		m.page = PagePreview
		return m, nil
	case genErrMsg:
		m.generating = false
		m.status = "import error: " + msg.err.Error()
		return m, nil
	}
	return m, nil
}

func (m *Model) resizeLists() {
	h := m.layout.contentH
	if h < 4 {
		h = 4
	}
	w := m.layout.contentW
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
			case 'q', 'Q':
				m.save()
				m.quitting = true
				return m, tea.Quit
			case '1':
				m.page = PageBoard
				m.refreshBoardEntries()
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
				m.refreshImports()
				return m, nil
			}
		}
	}

	switch m.page {
	case PageBoard:
		return m.updateBoard(msg)
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
	w := m.layout.contentW
	if w < 20 {
		return 20
	}
	return w
}

func listHeight(m Model) int {
	h := m.layout.contentH
	if h < 4 {
		return 4
	}
	return h
}

func (m Model) View() string {
	if m.quitting {
		return ""
	}
	if m.layout.termW == 0 {
		m.layout.termW = 80
		m.layout.termH = 24
		m.recalcLayout(tea.WindowSizeMsg{Width: 80, Height: 24})
	}

	var body string
	switch m.page {
	case PageBoard:
		body = m.viewBoard()
	case PageProjects:
		body = m.viewProjects()
	case PageTasks:
		body = m.viewTasks()
	case PageImport:
		body = m.viewImport()
	case PagePreview:
		body = m.viewPreview()
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		"",
		baseStyle.Width(m.layout.contentW).Render(body),
		"",
		m.renderFooter(),
	)
	return m.placeWindow(inner)
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
