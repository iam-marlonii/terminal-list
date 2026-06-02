// Package config resolves paths and runtime options.
package config

import (
	"os"
	"path/filepath"
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

// UseAltScreen reports whether the TUI should use the alternate screen buffer.
func UseAltScreen() bool {
	return os.Getenv("TODO_ALT_SCREEN") == "1"
}
