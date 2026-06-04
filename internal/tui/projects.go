package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iam-marlonjr/terminal-list/internal/store"
)

type projectItem struct {
	project *store.Project
}

func (i projectItem) FilterValue() string { return i.project.Name }
func (i projectItem) Title() string       { return i.project.Name }
func (i projectItem) Description() string {
	return fmt.Sprintf("%s · %d open", i.project.Status, i.project.OpenCount())
}

func newProjectList(doc *store.Document, width, height int) list.Model {
	items := make([]list.Item, 0, len(doc.Projects))
	for _, p := range doc.Projects {
		items = append(items, projectItem{project: p})
	}
	delegate := list.NewDefaultDelegate()
	delegate.Styles.SelectedTitle = selectedItemStyle
	delegate.Styles.SelectedDesc = selectedItemStyle
	delegate.Styles.NormalTitle = menuItemStyle
	delegate.Styles.NormalDesc = mutedStyle
	l := list.New(items, delegate, width, height)
	l.Title = "projects"
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	return l
}

func (m Model) viewProjects() string {
	return m.projectList.View()
}

func (m Model) updateProjects(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		if it, ok := m.projectList.SelectedItem().(projectItem); ok {
			m.selectedProject = it.project
			m.page = PageTasks
			m.taskList = newTaskList(m.selectedProject, listWidth(m), listHeight(m))
		}
		return m, nil
	}
	var cmd tea.Cmd
	m.projectList, cmd = m.projectList.Update(msg)
	return m, cmd
}
