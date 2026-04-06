# TODO

## UX

- [ ] **Fixed 200px column widths in table** (`table.go:~66`) — columns should resize to content or be user-adjustable
- [ ] **Search is `strings.Contains` only** (`table.go:~118`, `linesList.go:~132`) — no regex support, no field-scoped queries (e.g. `status:404`, `method:POST`), no boolean operators
- [ ] **Toast duration hardcoded at 3s** (`toast.go:~58`) — not configurable
- [ ] **Line wrap width hardcoded at 90 chars** (`linesList.go:~104`) — magic number fallback, no way to adjust

## Missing Features

- [ ] **Lua API not discoverable in-app** — help dialog only shows keybindings; no listing of available Lua functions, no autocomplete, no `help <fn>` introspection
