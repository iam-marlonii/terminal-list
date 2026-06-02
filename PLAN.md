# Terminal To-Do List Application

> **Note:** The repo currently ships a local-first Markdown task board (see
> [README.md](README.md)). This document describes the longer-term Supabase +
> notification-daemon architecture.

## Architecture Overview

The application consists of two components:

1. **Main TUI Application** (Server-side): Runs on `todo.marlonausbyjr.com`, accessible via SSH
2. **Notification Daemon** (Client-side): Runs on the user's local machine, shows desktop notifications

### Communication Flow

```
┌─────────────────┐           ┌──────────────────┐         ┌─────────────┐
│  Local Machine  │   SSH     │  Remote Server   │         │   Supabase  │
│                 │──────────>│    (TUI App)     │         │   Database  │
│                 │           │                  │<───────>│             │
│  Notification   │<──────────│  Notification    │         │             │
│  Daemon         │ WebSocket │    Service       │         │             │
└─────────────────┘           └──────────────────┘         └─────────────┘
```

## Project Structure

```
terminal-list/
├── cmd/
│   ├── todo/              # Main TUI application
│   │   └── main.go
│   └── notifyd/           # Notification daemon
│       └── main.go
├── internal/
│   ├── db/                # Database layer (Supabase client)
│   │   └── client.go
│   ├── models/            # Data models
│   │   └── todo.go
│   ├── tui/               # TUI components
│   │   ├── app.go         # Main Bubble Tea app
│   │   ├── list.go        # Todo list view
│   │   ├── form.go        # Add/edit form
│   │   └── styles.go      # Lipgloss styles
│   ├── notifications/     # Notification service
│   │   └── server.go      # WebSocket/HTTP notification endpoint
│   └── config/            # Configuration management
│       └── config.go
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## Implementation Details

### 1. Main TUI Application (`cmd/todo`)

**Dependencies:**

- `github.com/charmbracelet/bubbletea` - TUI framework
- `github.com/charmbracelet/bubbles` - UI components (list, textinput, etc.)
- `github.com/charmbracelet/lipgloss` - Styling
- `github.com/supabase-community/supabase-go` - Supabase client
- `github.com/gorilla/websocket` - WebSocket server for notifications

**Features:**

- List todos (with status: pending, in-progress, completed)
- Add new todos
- Edit existing todos
- Delete todos
- Full CRUD flow wired to Supabase (create/read/update/delete)
- Mark todos as complete
- Set due dates and reminders
- Filter/search todos
- Keyboard navigation (vim-like keys)

**Database Schema (Supabase):**

```sql
CREATE TABLE todos (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  status TEXT NOT NULL DEFAULT 'pending', -- pending, in_progress, completed
  due_date TIMESTAMPTZ,
  reminder_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT NOW(),
  updated_at TIMESTAMPTZ DEFAULT NOW()
);

CREATE INDEX idx_todos_user_status ON todos(user_id, status);
CREATE INDEX idx_todos_reminder ON todos(reminder_at) WHERE reminder_at IS NOT NULL;
```

### 2. Notification Daemon (`cmd/notifyd`)

**Dependencies:**

- `github.com/getlantern/systray` or `fyne.io/systray` - System tray icon
- `github.com/gen2brain/beeep` - Cross-platform desktop notifications
- `github.com/gorilla/websocket` - WebSocket client

**Features:**

- Runs as background service/daemon
- Connects to server via WebSocket
- Shows desktop notifications for:
  - Reminder alerts (when `reminder_at` is reached)
  - New todos assigned
  - Status changes
- System tray icon with menu:
  - Show/hide notifications
  - Quit
- Auto-start on system boot (optional)

**Configuration:**

- Server URL/endpoint
- Authentication token
- Notification preferences

### 3. Notification Service (`internal/notifications`)

**Implementation Options:**

**Option A: WebSocket Server (Recommended)**

- Run alongside TUI app
- Clients connect and subscribe to user-specific channels
- Push notifications when events occur (reminders, updates)

**Option B: Supabase Realtime**

- Use Supabase's built-in realtime subscriptions
- Daemon subscribes to `todos` table changes
- Filter by user_id and trigger notifications

**Option C: HTTP Polling**

- Simple HTTP endpoint on server
- Daemon polls for pending notifications
- Less efficient but simpler to implement

**Recommended: Option B (Supabase Realtime)** - leverages existing infrastructure

### 4. Configuration

**Environment Variables:**

- `SUPABASE_URL` - Supabase project URL
- `SUPABASE_KEY` - Supabase anon/service key
- `NOTIFICATION_SERVER_URL` - WebSocket server URL (if not using Realtime)
- `USER_ID` - User identifier (could be SSH username or config file)

**Config File (`~/.config/todo/config.yaml`):**

```yaml
supabase:
  url: "http://localhost:54321"  # Local Supabase
  key: "your-anon-key"
user:
  id: "marlonausbyjr"
notifications:
  enabled: true
  server_url: "ws://todo.marlonausbyjr.com:8080/notifications"
```

## Key Files to Create

1. **`cmd/todo/main.go`** - Entry point for TUI app
2. **`internal/tui/app.go`** - Main Bubble Tea model and update loop
3. **`internal/tui/list.go`** - Todo list component using bubbles/list
4. **`internal/db/client.go`** - Supabase client wrapper
5. **`internal/models/todo.go`** - Todo struct and methods
6. **`cmd/notifyd/main.go`** - Notification daemon entry point
7. **`internal/notifications/realtime.go`** - Supabase Realtime subscription handler
8. **`go.mod`** - Go module dependencies
9. **`Makefile`** - Build and deployment scripts

## Deployment Considerations

**SSH Access Setup:**

- Configure SSH server to run the TUI app as a shell or command
- Use `ForceCommand` in `sshd_config` or shell wrapper script
- Example: `ssh todo.marlonausbyjr.com` → launches TUI app directly

**Local Supabase:**

- Run Supabase locally using Docker: `supabase start`
- Default local URL: `http://localhost:54321`
- Access via `http://localhost:54323` (Studio)

**Notification Daemon:**

- Build platform-specific binaries
- Provide installation scripts for:
  - macOS: LaunchAgent plist
  - Linux: systemd service
  - Windows: Task Scheduler

## Development Phases

1. **Phase 1: Core TUI App**
   - Set up Bubble Tea structure
   - Implement basic CRUD operations
   - Connect to Supabase

2. **Phase 2: Notification System**
   - Set up Supabase Realtime subscriptions
   - Create notification daemon
   - Implement desktop notifications

3. **Phase 3: Polish & Deployment**
   - SSH integration
   - Configuration management
   - Build scripts and documentation

## Todo List

- [ ] Initialize Go module, create project structure, and set up dependencies (bubbletea, supabase-go, beeep, etc.)
- [ ] Create Supabase database schema with todos table, indexes, and RLS policies
- [ ] Implement database client wrapper for Supabase operations (CRUD, queries)
- [ ] Wire full CRUD flows end-to-end in TUI against Supabase
- [ ] Build main TUI application with Bubble Tea: list view, add/edit forms, keyboard navigation
- [ ] Implement Supabase Realtime subscription handler for todo changes and reminders
- [ ] Create notification daemon with systray icon and desktop notification support
- [ ] Add configuration file support and environment variable handling
- [ ] Set up SSH server configuration to launch TUI app on connection
- [ ] Create Makefile with build targets for both applications and cross-platform binaries
