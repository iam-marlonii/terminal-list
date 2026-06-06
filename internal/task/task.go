package task

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

// Status represents whether a task is open or finished.
type Status string

const (
	StatusTodo Status = "todo"
	StatusDone Status = "done"
)

// Task is a single actionable item.
type Task struct {
	ID          string
	Title       string
	Status      Status
	Created     time.Time
	Completed   *time.Time
	Due         *time.Time
	Description string
}

// New builds an open task with a fresh ID and creation timestamp.
func New(title string) Task {
	return Task{
		ID:      NewID(),
		Title:   title,
		Status:  StatusTodo,
		Created: time.Now().UTC(),
	}
}

// Toggle flips the task between open and done, maintaining the completion time.
func (t *Task) Toggle() {
	if t.Status == StatusDone {
		t.Status = StatusTodo
		t.Completed = nil
		return
	}
	t.Status = StatusDone
	now := time.Now().UTC()
	t.Completed = &now
}

// Done reports whether the task is complete.
func (t Task) Done() bool { return t.Status == StatusDone }

// NewID returns a short random hex identifier.
func NewID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
