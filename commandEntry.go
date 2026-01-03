package efinui

import (
	"slices"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/widget"
)

type CommandEntry struct {
	widget.Entry

	keyBindings *KeyBindings

	mode      Mode
	OnCommand func(string)
	OnSearch  func(string)

	history    []string
	historyPos int
}

func NewCommandEntry() *CommandEntry {
	e := &CommandEntry{
		mode: ModeNormal,
	}
	e.ExtendBaseWidget(e)

	e.OnSubmitted = func(text string) {
		if text != "" {
			switch e.mode {
			case ModeCommand:
				e.OnCommand(text)

			case ModeSearch:
				e.OnSearch(text)
			}

			e.SetText("")
		}
	}

	return e
}

func (ce *CommandEntry) SetMode(mode Mode) {
	ce.mode = mode

	if mode == ModeNormal {
		ce.Disable()
	} else {
		ce.Enable()
	}

	ce.Refresh()
}

func (ce *CommandEntry) SetKeyBindings(kbs *KeyBindings) {
	ce.keyBindings = kbs
}

func (ce *CommandEntry) TypedKey(ev *fyne.KeyEvent) {
	if ok := ce.keyBindings.OnTypedKey(ev); ok {
		return
	}

	ce.Entry.TypedKey(ev)
}

func (ce *CommandEntry) TypedShortcut(sc fyne.Shortcut) {
	if ok := ce.keyBindings.OnTypedShortcut(sc); ok {
		return
	}

	ce.Entry.TypedShortcut(sc)
}

func (ce *CommandEntry) SetHistory(history []string) {
	ce.history = []string{}

	present := map[string]bool{}
	for _, cmd := range slices.Backward(history) {
		if _, ok := present[cmd]; !ok {
			ce.history = append(ce.history, cmd)
			present[cmd] = true
		}
	}

	slices.Reverse(ce.history)
	ce.historyPos = len(ce.history)
}

func (ce *CommandEntry) HistoryNext() {
	ce.historyPos = min(len(ce.history), ce.historyPos+1)

	if ce.historyPos < len(ce.history) {
		ce.SetText(ce.history[ce.historyPos])
	} else {
		ce.SetText("")
	}
	ce.Refresh()
}

func (ce *CommandEntry) HistoryPrev() {
	ce.historyPos = max(0, ce.historyPos-1)
	ce.SetText(ce.history[ce.historyPos])
	ce.Refresh()
}
