// Package config resolves paths and runtime options.
package config

import (
	"os"
	"path/filepath"

	"github.com/mattn/go-isatty"
)

// TasksFile returns the markdown task database path.
// Priority: TODO_FILE → $XDG_CONFIG_HOME/terminal-list/tasks.md → ~/.config/terminal-list/tasks.md → ./tasks.md
func TasksFile() string {
	if v := os.Getenv("TODO_FILE"); v != "" {
		return v
	}
	base := os.Getenv("XDG_CONFIG_HOME")
	if base == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "tasks.md"
		}
		base = filepath.Join(home, ".config")
	}
	return filepath.Join(base, "terminal-list", "tasks.md")
}

// ImportDir returns the directory where staged import files live.
// It is a sibling of the tasks file, so TODO_FILE overrides are respected.
func ImportDir() string {
	return filepath.Join(filepath.Dir(TasksFile()), "imports")
}

// TrashDir returns the directory where consumed import files are moved.
func TrashDir() string {
	return filepath.Join(filepath.Dir(TasksFile()), "trash")
}

// UseAltScreen reports whether the TUI should use the alternate screen buffer.
// Defaults to on for real TTYs unless TODO_ALT_SCREEN=0.
func UseAltScreen() bool {
	switch os.Getenv("TODO_ALT_SCREEN") {
	case "1":
		return true
	case "0":
		return false
	}
	return isatty.IsTerminal(os.Stdout.Fd())
}
