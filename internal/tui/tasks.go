package tui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/list"
	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/iam-marlonjr/terminal-list/internal/store"
	"github.com/iam-marlonjr/terminal-list/internal/task"
)

type taskItem struct {
	task    *task.Task
	project *store.Project
}

func (i taskItem) FilterValue() string { return i.task.Title }
func (i taskItem) Title() string {
	box := "[ ]"
	if i.task.Done() {
		box = "[x]"
	}
	return fmt.Sprintf("%s %s", box, i.task.Title)
}
func (i taskItem) Description() string { return "" }

func newTaskList(project *store.Project, width, height int) list.Model {
	items := make([]list.Item, 0)
	if project != nil {
		items = projectGroupsAsItems(project)
	}
	delegate := list.NewDefaultDelegate()
	l := list.New(items, delegate, width, height)
	title := "Tasks"
	if project != nil {
		title = "Tasks — " + project.Name
	}
	l.Title = title
	l.SetShowStatusBar(false)
	l.SetFilteringEnabled(false)
	l.Styles.Title = titleStyle
	return l
}

func projectGroupsAsItems(project *store.Project) []list.Item {
	var items []list.Item
	for _, g := range project.Groups {
		for _, t := range g.Tasks {
			items = append(items, taskItem{task: t, project: project})
		}
	}
	return items
}

func (m Model) viewTasks() string {
	if m.selectedProject == nil {
		return mutedStyle.Render("No project selected. Press 2 to pick one.\n")
	}
	if m.adding {
		return m.taskList.View() + "\n\n" + m.addInput.View()
	}
	return m.taskList.View()
}

func (m Model) updateTasks(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEsc:
		m.page = PageProjects
		m.projectList = newProjectList(m.doc, listWidth(m), listHeight(m))
		return m, nil
	case tea.KeySpace:
		if it, ok := m.taskList.SelectedItem().(taskItem); ok {
			it.task.Toggle()
			m.save()
			m.status = "toggled: " + it.task.Title
			m.taskList = newTaskList(m.selectedProject, listWidth(m), listHeight(m))
			m.taskList.Select(m.taskList.Index())
		}
		return m, nil
	case tea.KeyRunes:
		if len(msg.Runes) == 1 && msg.Runes[0] == 'a' {
			m.adding = true
			m.addInput.Focus()
			m.addInput.SetValue("")
			return m, textinput.Blink
		}
		if len(msg.Runes) == 1 && msg.Runes[0] == 'd' {
			if it, ok := m.taskList.SelectedItem().(taskItem); ok {
				title := it.task.Title
				m.doc.Remove(it.task)
				m.save()
				m.status = "deleted: " + title
				m.taskList = newTaskList(m.selectedProject, listWidth(m), listHeight(m))
			}
			return m, nil
		}
	}
	var cmd tea.Cmd
	m.taskList, cmd = m.taskList.Update(msg)
	return m, cmd
}

func (m Model) updateAdd(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.Type {
	case tea.KeyEnter:
		title := m.addInput.Value()
		if title != "" && m.selectedProject != nil {
			m.doc.AddToProject(m.selectedProject, title)
			m.save()
			m.status = "added: " + title
			m.taskList = newTaskList(m.selectedProject, listWidth(m), listHeight(m))
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
