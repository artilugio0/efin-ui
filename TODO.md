# TODO

## Critical Bugs (crashes / panics)

- [x] **`RequestResponseViewer.MoveUp/Down()` are empty stubs** (`requestResponseViewer.go:132`) — users cannot scroll through request/response bodies
- [x] **Nil pointer dereference in `executeCode()`** (`app.go`) — when Lua `LoadString` fails twice, `f` is nil but still passed to `CallByParam` → panic
- [x] **Panic on empty command history** (`commandEntry.go:HistoryPrev`) — `history[historyPos]` accessed without checking if history is empty
- [x] **Stale index panic in `Table.Submit()`** (`table.go`) — no bounds check on `rows[selectedRow]`
- [x] **Unsafe array access in `MultiSplit` pane navigation** (`multiSplit.go`) — `PaneFocusUp/Down/Left/Right()` access `objectsGrid` without checking if it's empty

## Concurrency / Resource Issues

- [x] **Data race in `RunQuery`** (`app.go:~389`) — two goroutines write shared `req`/`resp` variables; detectable with `go test -race`
- [x] **Timer goroutine race in toast cleanup** (`toast.go:~78`) — `time.AfterFunc` fires on a separate goroutine with no lock protecting the slice mutation
- [x] **Lua state never closed** (`app.go`) — `lua.NewState()` in `NewApp()` has no matching `l.Close()`
- [x] **No query timeout** (`query.go:20`) — `context.TODO()` means a slow or hung query blocks indefinitely with no way to cancel

## Code Quality

- [ ] **220 lines of duplicated color parsing** (`app.go:662-891`) — 18 near-identical blocks parsing theme color fields; could be a ~20-line loop
- [ ] **Dead code: unused `messageSend` closure** (`app.go:~636`) — closure is defined but never used; the `message_send` Lua global is set twice with the same value

## UX

- [ ] **Fixed 200px column widths in table** (`table.go:~66`) — columns should resize to content or be user-adjustable
- [ ] **Search is `strings.Contains` only** (`table.go:~118`, `linesList.go:~132`) — no regex support, no field-scoped queries (e.g. `status:404`, `method:POST`), no boolean operators
- [ ] **Toast duration hardcoded at 3s** (`toast.go:~58`) — not configurable
- [ ] **Line wrap width hardcoded at 90 chars** (`linesList.go:~104`) — magic number fallback, no way to adjust

## Missing Features

- [ ] **Only 2 export formats** (Python, Lua) — add cURL at minimum; consider JavaScript (fetch), Go, shell script
- [ ] **No content-type-aware rendering** — JSON and XML bodies shown as raw text; pretty-printing and syntax highlighting would improve readability significantly
- [ ] **No column sorting in query results** — cannot click a column header to sort; must write `ORDER BY` SQL manually
- [ ] **No batch operations** — cannot select multiple rows to export, copy, or delete; no tagging or grouping
- [ ] **No response time / size metadata** — content-length and response timing not shown in the viewer
- [ ] **Lua API not discoverable in-app** — help dialog only shows keybindings; no listing of available Lua functions, no autocomplete, no `help <fn>` introspection
