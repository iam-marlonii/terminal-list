// Package tui implements the multi-page terminal UI.
package tui

import (
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"

	"github.com/iam-marlonjr/terminal-list/internal/config"
	"github.com/iam-marlonjr/terminal-list/internal/store"
	"github.com/iam-marlonjr/terminal-list/internal/task"
)

// Page is a top-level screen.
type Page int

const (
	PageBoard Page = iota
	PageTasks
	PageProjects
	PageImport
)

// viewMode is the sub-state within a page: a list (dashboard) or a detail.
type viewMode int

const (
	modeDashboard viewMode = iota
	modeView               // detail view; for import this is the review screen
)

const pageSize = 10

// taskRef pairs a task with its owning project for flat lists.
type taskRef struct {
	task    *task.Task
	project *store.Project
}

// Model is the root Bubble Tea model.
type Model struct {
	doc      *store.Document
	path     string
	page     Page
	mode     viewMode
	layout   layoutState
	status   string
	quitting bool

	// board
	boardSel  int
	boardPage int

	// tasks
	taskSel  int
	taskPage int

	// projects
	projSel     int
	projPage    int
	projTaskSel int

	// import
	importFiles       []string
	importSel         int
	generating        bool
	pendingImportFile string

	// import review
	reviewDoc     *store.Document
	reviewProject *store.Project
	reviewTaskSel int

	// inline add (import review)
	adding   bool
	addInput textinput.Model

	// field editor (tasks / projects view)
	editing     bool
	editKind    editKind
	editFields  []editField
	editIndex   int
	editInput   textinput.Model
	editTask    *task.Task
	editProject *store.Project
}

// New constructs the app for the given document and path.
func New(doc *store.Document, path string) Model {
	m := Model{
		doc:  doc,
		path: path,
		page: PageBoard,
	}
	m.refreshImports()

	m.addInput = textinput.New()
	m.addInput.Placeholder = "new task…"
	m.addInput.CharLimit = 200
	m.addInput.Width = 50
	m.addInput.PromptStyle = inputPromptStyle
	m.addInput.TextStyle = menuItemStyle

	m.editInput = textinput.New()
	m.editInput.CharLimit = 400
	m.editInput.Width = 50
	m.editInput.PromptStyle = inputPromptStyle
	m.editInput.TextStyle = menuItemStyle

	return m
}

func (m Model) Init() tea.Cmd { return nil }

func textinputBlink() tea.Cmd { return textinput.Blink }

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.recalcLayout(msg)
		return m, nil
	case tea.KeyMsg:
		if m.quitting {
			return m, tea.Quit
		}
		if m.editing {
			return m.updateEdit(msg)
		}
		if m.adding {
			return m.updateReviewAdd(msg)
		}
		return m.updateKey(msg)
	case genDoneMsg:
		return m.onGenDone(msg)
	case genErrMsg:
		m.generating = false
		m.status = "import error: " + msg.err.Error()
		return m, nil
	}
	return m, nil
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
			case 'b', 'B':
				m.page = PageBoard
				m.mode = modeDashboard
				return m, nil
			case 't', 'T':
				m.page = PageTasks
				m.mode = modeDashboard
				return m, nil
			case 'p', 'P':
				m.page = PageProjects
				m.mode = modeDashboard
				return m, nil
			case 'i', 'I':
				m.page = PageImport
				m.mode = modeDashboard
				m.refreshImports()
				return m, nil
			}
		}
	}

	switch m.page {
	case PageBoard:
		return m.updateBoard(msg)
	case PageTasks:
		return m.updateTasks(msg)
	case PageProjects:
		return m.updateProjects(msg)
	case PageImport:
		return m.updateImport(msg)
	}
	return m, nil
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
	case PageTasks:
		body = m.viewTasks()
	case PageProjects:
		body = m.viewProjects()
	case PageImport:
		body = m.viewImport()
	}

	inner := lipgloss.JoinVertical(lipgloss.Left,
		m.renderHeader(),
		"",
		baseStyle.Width(m.layout.contentW).Height(m.layout.contentH).Render(body),
		"",
		m.renderFooter(),
	)
	return m.placeWindow(inner)
}

// allTaskRefs returns every task across all projects in document order.
func (m Model) allTaskRefs() []taskRef {
	var refs []taskRef
	for _, p := range m.doc.Projects {
		for _, g := range p.Groups {
			for _, t := range g.Tasks {
				refs = append(refs, taskRef{task: t, project: p})
			}
		}
	}
	return refs
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
