# efin-ui

A terminal/GUI-based interactive viewer for HTTP request/response data stored in a SQLite database. Provides a keyboard-driven (Vim-style) interface for browsing, searching, and analyzing recorded HTTP traffic.

## Features

- Browse HTTP requests/responses stored in SQLite
- Run arbitrary SQL queries against the database
- View request/response pairs side-by-side
- Export requests as Python or Lua scripts
- Configurable keybindings and themes via Lua settings file
- Multiple built-in themes

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Language | Go |
| GUI Framework | Fyne v2 |
| Database | SQLite (modernc.org/sqlite) |
| Scripting/Config | Lua (gopher-lua) |
| CLI | Cobra |
| Build | Nix Flakes + Go modules |

## Installation

```sh
go build ./...
```

Or with Nix:

```sh
nix build
```

## Usage

```sh
efin-ui -D <path-to-database.db>
```

### Flags

| Flag | Description |
|------|-------------|
| `-D` | Path to SQLite database (required) |
| `-s` | Path to custom Lua settings file |
| `-H` | Path to command history file |

## Modes

efin-ui is modal, similar to Vim:

| Mode | How to Enter | Behavior |
|------|-------------|----------|
| Normal | `Esc` | Navigation (hjkl), pane and tab control |
| Command | `:` (default) | Type and execute Lua expressions |
| Search | `/` (default) | Text search across visible content |
| Help | `?` (default) | Display keybinding reference |

## Querying Data

In command mode, use the `query()` function to run SQL against the database:

```lua
query("SELECT * FROM requests")
query("SELECT * FROM requests WHERE method = 'POST'")
query("SELECT r.*, s.status_code FROM requests r JOIN responses s ON r.request_id = s.response_id")
```

Results are displayed as an interactive table. Press Enter on a row to open the full request/response viewer.

## Keybindings

All keybindings are defined in Lua. Default bindings (Normal mode):

| Key | Action |
|-----|--------|
| `h/j/k/l` | Move left/down/up/right |
| `:`  | Enter command mode |
| `/`  | Enter search mode |
| `?`  | Toggle help |
| `n/N` | Next/previous search result |

## Configuration

efin-ui is configured via a Lua settings file. By default it loads `efin-settings.lua` from the same directory. Override with `-s`:

```sh
efin-ui -D data.db -s ~/.config/efin/settings.lua
```

### Themes

Six built-in themes: `default`, `red`, `green`, `blue`, `synthwave`, `neon_sunset`.

Switch theme at runtime from command mode:

```lua
set_theme("synthwave")
```

### Custom Keybindings

Keybindings are defined per mode in the settings file:

```lua
settings.key_bindings["normal"]["g"] = function()
    move_up()
end
```

## Lua API

These functions are available in command mode and in your settings file:

| Function | Description |
|----------|-------------|
| `query(sql)` | Run SQL query, display results in current pane |
| `export_request(format)` | Export selected request as `"python"` or `"lua"` script |
| `set_theme(name)` | Switch to the named theme |
| `pane_create()` | Create a new pane |
| `pane_close()` | Close the current pane |
| `pane_hsplit()` | Split pane horizontally |
| `pane_vsplit()` | Split pane vertically |
| `tab_create()` | Create a new tab |
| `tab_close()` | Close the current tab |
| `tab_next()` | Switch to next tab |
| `tab_prev()` | Switch to previous tab |
| `move_up/down/left/right()` | Move focus between panes |
| `search(term)` | Search for text in current view |
| `search_next()` | Jump to next search match |
| `search_prev()` | Jump to previous search match |

## Database Schema

efin-ui expects the following SQLite schema:

```sql
CREATE TABLE IF NOT EXISTS requests (
    request_id INTEGER PRIMARY KEY AUTOINCREMENT,
    method TEXT NOT NULL,
    url TEXT NOT NULL,
    body TEXT,
    timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
);
CREATE TABLE IF NOT EXISTS responses (
    response_id INTEGER PRIMARY KEY AUTOINCREMENT,
    status_code INTEGER NOT NULL,
    body TEXT,
    content_length INTEGER
);
CREATE TABLE IF NOT EXISTS headers (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER,
    response_id INTEGER,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (request_id) REFERENCES requests(request_id),
    FOREIGN KEY (response_id) REFERENCES responses(response_id)
);
CREATE TABLE IF NOT EXISTS cookies (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    request_id INTEGER,
    response_id INTEGER,
    name TEXT NOT NULL,
    value TEXT NOT NULL,
    FOREIGN KEY (request_id) REFERENCES requests(request_id),
    FOREIGN KEY (response_id) REFERENCES responses(response_id)
);
CREATE INDEX IF NOT EXISTS idx_requests_url ON requests (url);
CREATE INDEX IF NOT EXISTS idx_responses_status_code ON responses (status_code);
CREATE INDEX IF NOT EXISTS idx_headers_name ON headers (name);
CREATE INDEX IF NOT EXISTS idx_headers_value ON headers (value);
CREATE INDEX IF NOT EXISTS idx_cookies_name ON cookies (name);
CREATE INDEX IF NOT EXISTS idx_cookies_value ON cookies (value);

CREATE INDEX IF NOT EXISTS idx_cookies_request_id ON cookies(request_id);
CREATE INDEX IF NOT EXISTS idx_cookies_response_id ON cookies(response_id);
CREATE INDEX IF NOT EXISTS idx_headers_request_id ON headers(request_id);
CREATE INDEX IF NOT EXISTS idx_headers_response_id ON headers(response_id);
```

## Architecture

Single-package Go application with an MVC-inspired structure:

- **Data layer** (`query.go`) — SQL execution, `Request`/`Response` structs
- **Business logic** (`app.go`) — mode management, keybindings, Lua VM, callbacks
- **UI layer** — Fyne widgets: table, panes, request/response viewer, search

### Key Files

| File | Purpose |
|------|---------|
| `app.go` | Core `App` struct, Lua bindings, mode management |
| `query.go` | Data structs and database interface |
| `multiSplit.go` | N×M pane layout manager |
| `table.go` | Searchable table widget |
| `requestResponseViewer.go` | Side-by-side request/response display |
| `commandEntry.go` | Command input with history |
| `themes.go` | Theme color schemes |
| `efin-settings.lua` | Default Lua configuration |
| `templates/files/` | Embedded export script templates |
