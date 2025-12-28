package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
	"sync"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/widget"
	lua "github.com/yuin/gopher-lua"
)

type App struct {
	db *sql.DB

	settingsScript string
	l              *lua.LState

	histFilePath string
	history      []string

	mode Mode

	keyBindings map[Mode][]string

	search string

	// UI
	fyneApp fyne.App
	window  fyne.Window

	content *fyne.Container

	modeLabel    *ModeLabel
	commandEntry *CommandEntry

	tabs            []*MultiSplit
	currentTabIndex int

	focusedObject fyne.CanvasObject
}

func NewApp(db *sql.DB, histFilePath, settingsScript string) *App {
	a := &App{
		db: db,

		l:              lua.NewState(),
		settingsScript: settingsScript,
		histFilePath:   histFilePath,

		mode: ModeNormal,
	}

	modeLabel := NewModeLabel()

	commandEntry := NewCommandEntry()
	commandEntry.OnCommand = func(code string) {
		a.executeCode(code)
		a.SetMode(ModeNormal)
	}

	commandEntry.OnSearch = func(search string) {
		a.Search(search)
		a.SetMode(ModeNormal)
	}

	tabs := []*MultiSplit{}

	content := container.NewBorder(
		modeLabel,
		commandEntry,
		nil,
		nil,
		nil,
	)

	fyneApp := app.New()
	window := fyneApp.NewWindow("Efin")
	window.SetContent(content)

	a.fyneApp = fyneApp
	a.window = window
	a.content = content
	a.modeLabel = modeLabel
	a.commandEntry = commandEntry
	a.tabs = tabs

	return a
}

func (a *App) Run() {
	a.loadHistory()
	a.commandEntry.SetHistory(a.history)

	a.initializeLuaState()
	a.loadKeyBindingsDefinitions()

	a.SetMode(a.mode)

	a.TabCreate()

	a.window.SetFullScreen(true)
	a.window.Show()
	a.window.SetFullScreen(false)
	a.fyneApp.Run()
}

func (a *App) SetMode(mode Mode) {
	a.mode = mode

	a.modeLabel.SetMode(mode)
	a.commandEntry.SetMode(mode)

	a.applyKeyBindings(mode)

	if mode == ModeNormal {
		a.window.Canvas().Focus(nil)
	} else {
		a.window.Canvas().Focus(a.commandEntry)
	}
}

func (a *App) PaneCreate() {
	a.tabs[a.currentTabIndex].PaneCreate(
		container.NewCenter(widget.NewLabel("Empty")),
	)
}

func (a *App) PaneDelete() {
	a.tabs[a.currentTabIndex].PaneDelete()
}

func (a *App) PaneLineAdd() {
	a.tabs[a.currentTabIndex].PaneLineAdd(
		container.NewCenter(widget.NewLabel("Empty")),
	)
}

func (a *App) PaneFocusRight() {
	a.tabs[a.currentTabIndex].PaneFocusRight()
}

func (a *App) PaneFocusLeft() {
	a.tabs[a.currentTabIndex].PaneFocusLeft()
}

func (a *App) PaneFocusDown() {
	a.tabs[a.currentTabIndex].PaneFocusDown()
}

func (a *App) PaneFocusUp() {
	a.tabs[a.currentTabIndex].PaneFocusUp()
}

func (a *App) TabCreate() {
	tab := NewMultiSplit()
	tab.OnFocusMove = func(o fyne.CanvasObject) {
		a.focusedObject = o
	}

	if a.search != "" {
		tab.Search(a.search, false)
	}

	a.tabs = append(a.tabs, tab)
	a.tabSwitch(len(a.tabs) - 1)

	a.PaneCreate()
}

func (a *App) TabDelete() {
	a.tabs = slices.Delete(a.tabs, a.currentTabIndex, a.currentTabIndex+1)

	if len(a.tabs) == 0 {
		a.TabCreate()
		return
	}

	a.tabSwitch(max(0, a.currentTabIndex-1))
}

func (a *App) TabNext() {
	if a.currentTabIndex == len(a.tabs)-1 {
		return
	}

	a.tabSwitch(a.currentTabIndex + 1)
}

func (a *App) TabPrev() {
	if a.currentTabIndex == 0 {
		return
	}

	a.tabSwitch(a.currentTabIndex - 1)
}

func (a *App) tabSwitch(i int) {
	a.currentTabIndex = i
	a.content.Objects[0] = a.tabs[i]
	a.content.Refresh()
}

func (a *App) MoveUp() {
	if a.focusedObject == nil {
		return
	}

	if m, ok := a.focusedObject.(Mover); ok {
		m.MoveUp()
	}
}

func (a *App) MoveDown() {
	if a.focusedObject == nil {
		return
	}

	if m, ok := a.focusedObject.(Mover); ok {
		m.MoveDown()
	}
}

func (a *App) MoveLeft() {
	if a.focusedObject == nil {
		return
	}

	if m, ok := a.focusedObject.(Mover); ok {
		m.MoveLeft()
	}
}

func (a *App) MoveRight() {
	if a.focusedObject == nil {
		return
	}

	if m, ok := a.focusedObject.(Mover); ok {
		m.MoveRight()
	}
}

func (a *App) Submit() {
	if a.focusedObject == nil {
		return
	}

	if s, ok := a.focusedObject.(Submitter); ok {
		s.Submit()
	}
}

func (a *App) Search(search string) {
	a.search = search

	for _, tab := range a.tabs {
		tab.Search(search, false)
	}
}

func (a *App) SearchPrev() {
	if a.focusedObject == nil {
		return
	}

	if s, ok := a.focusedObject.(Searcher); ok {
		s.SearchPrev()
	}
}

func (a *App) SearchNext() {
	if a.focusedObject == nil {
		return
	}

	if s, ok := a.focusedObject.(Searcher); ok {
		s.SearchNext()
	}
}

func (a *App) SearchClear() {
	a.search = ""

	for _, tab := range a.tabs {
		tab.SearchClear()
	}
}

func (a *App) RunQuery(query string) error {
	result, err := runQuery(a.db, query)
	if err != nil {
		return err
	}

	resultsTable := NewTable(result)
	resultsTable.OnSubmit = func(row []string) {
		var wg sync.WaitGroup
		wg.Add(1)
		var req *Request
		go func() {
			defer wg.Done()
			var err error
			req, err = getRequest(a.db, row[0])
			if err != nil {
				log.Printf("ERROR: %v", err)
				return
			}
		}()

		wg.Add(1)
		var resp *Response
		go func() {
			defer wg.Done()
			var err error
			resp, err = getResponse(a.db, row[0])
			if err != nil {
				log.Printf("ERROR: %v", err)
				return
			}
		}()
		wg.Wait()

		reqResViewer := NewRequestResponseViewer()
		reqResViewer.SetData(req, resp)
		a.tabs[a.currentTabIndex].PaneCreate(reqResViewer)
	}

	a.tabs[a.currentTabIndex].SetCurrentPane(resultsTable)

	return nil
}

func (a *App) initializeLuaState() {
	setModeFunc := a.l.NewFunction(func(ls *lua.LState) int {
		modeStr := a.l.ToString(1)
		mode := ModeNormal

		switch strings.ToLower(modeStr) {
		case "normal":
			mode = ModeNormal
		case "command":
			mode = ModeCommand
		case "search":
			mode = ModeSearch
		}

		a.SetMode(mode)

		return 0
	})
	a.l.SetGlobal("set_mode", setModeFunc)

	setModeNormalFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SetMode(ModeNormal)
		return 0
	})
	a.l.SetGlobal("set_mode_normal", setModeNormalFunc)

	setModeCommandFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SetMode(ModeCommand)
		return 0
	})
	a.l.SetGlobal("set_mode_command", setModeCommandFunc)

	setModeSearchFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SetMode(ModeSearch)
		return 0
	})
	a.l.SetGlobal("set_mode_search", setModeSearchFunc)

	queryFunc := a.l.NewFunction(func(ls *lua.LState) int {
		queryStr := a.l.ToString(1)

		if err := a.RunQuery(queryStr); err != nil {
			a.l.RaiseError("could not run query: %v", err)
		}

		return 0
	})
	a.l.SetGlobal("query", queryFunc)

	paneDeleteFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneDelete()
		return 0
	})
	a.l.SetGlobal("pane_delete", paneDeleteFunc)

	paneCreateFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneCreate()
		return 0
	})
	a.l.SetGlobal("pane_create", paneCreateFunc)

	paneLineAddFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneLineAdd()
		return 0
	})
	a.l.SetGlobal("pane_line_add", paneLineAddFunc)

	paneFocusUpFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneFocusUp()
		return 0
	})
	a.l.SetGlobal("pane_focus_up", paneFocusUpFunc)

	paneFocusDownFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneFocusDown()
		return 0
	})
	a.l.SetGlobal("pane_focus_down", paneFocusDownFunc)

	paneFocusLeftFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneFocusLeft()
		return 0
	})
	a.l.SetGlobal("pane_focus_left", paneFocusLeftFunc)

	paneFocusRightFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.PaneFocusRight()
		return 0
	})
	a.l.SetGlobal("pane_focus_right", paneFocusRightFunc)

	commandHistoryPrevFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.commandEntry.HistoryPrev()
		return 0
	})
	a.l.SetGlobal("command_history_prev", commandHistoryPrevFunc)

	commandHistoryNextFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.commandEntry.HistoryNext()
		return 0
	})
	a.l.SetGlobal("command_history_next", commandHistoryNextFunc)

	tabCreateFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.TabCreate()
		return 0
	})
	a.l.SetGlobal("tab_create", tabCreateFunc)

	tabDeleteFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.TabDelete()
		return 0
	})
	a.l.SetGlobal("tab_delete", tabDeleteFunc)

	tabNextFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.TabNext()
		return 0
	})
	a.l.SetGlobal("tab_next", tabNextFunc)

	tabPrevFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.TabPrev()
		return 0
	})
	a.l.SetGlobal("tab_prev", tabPrevFunc)

	moveLeftFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.MoveLeft()
		return 0
	})
	a.l.SetGlobal("move_left", moveLeftFunc)

	moveRightFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.MoveRight()
		return 0
	})
	a.l.SetGlobal("move_right", moveRightFunc)

	moveUpFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.MoveUp()
		return 0
	})
	a.l.SetGlobal("move_up", moveUpFunc)

	moveDownFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.MoveDown()
		return 0
	})
	a.l.SetGlobal("move_down", moveDownFunc)

	submitFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.Submit()
		return 0
	})
	a.l.SetGlobal("submit", submitFunc)

	searchResultPrevFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SearchPrev()
		return 0
	})
	a.l.SetGlobal("search_result_prev", searchResultPrevFunc)

	searchResultNextFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SearchNext()
		return 0
	})
	a.l.SetGlobal("search_result_next", searchResultNextFunc)

	searchClearFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SearchClear()
		return 0
	})
	a.l.SetGlobal("search_clear", searchClearFunc)

	settingsTable := a.l.NewTable()
	keyBindingsTable := a.l.NewTable()
	normalModeTable := a.l.NewTable()
	commandModeTable := a.l.NewTable()
	searchModeTable := a.l.NewTable()

	a.l.SetField(settingsTable, "key_bindings", keyBindingsTable)
	a.l.SetField(keyBindingsTable, "normal", normalModeTable)
	a.l.SetField(keyBindingsTable, "command", commandModeTable)
	a.l.SetField(keyBindingsTable, "search", searchModeTable)

	a.l.SetField(normalModeTable, "ctrl i", setModeCommandFunc)
	a.l.SetField(normalModeTable, "ctrl /", setModeSearchFunc)
	a.l.SetField(normalModeTable, "escape", searchClearFunc)

	a.l.SetField(commandModeTable, "escape", setModeNormalFunc)
	a.l.SetField(commandModeTable, "ctrl j", commandHistoryNextFunc)
	a.l.SetField(commandModeTable, "ctrl k", commandHistoryPrevFunc)

	a.l.SetField(searchModeTable, "escape", setModeNormalFunc)

	a.l.SetField(normalModeTable, "ctrl d", paneDeleteFunc)
	a.l.SetField(normalModeTable, "ctrl n", paneCreateFunc)

	a.l.SetField(normalModeTable, "ctrl l", paneFocusRightFunc)
	a.l.SetField(normalModeTable, "ctrl h", paneFocusLeftFunc)
	a.l.SetField(normalModeTable, "ctrl k", paneFocusUpFunc)
	a.l.SetField(normalModeTable, "ctrl j", paneFocusDownFunc)

	a.l.SetField(normalModeTable, "ctrl shift n", tabCreateFunc)
	a.l.SetField(normalModeTable, "ctrl shift d", tabDeleteFunc)
	a.l.SetField(normalModeTable, "ctrl shift h", tabPrevFunc)
	a.l.SetField(normalModeTable, "ctrl shift l", tabNextFunc)

	a.l.SetField(normalModeTable, "n", searchResultNextFunc)
	a.l.SetField(normalModeTable, "p", searchResultPrevFunc)

	a.l.SetField(normalModeTable, "h", moveLeftFunc)
	a.l.SetField(normalModeTable, "l", moveRightFunc)
	a.l.SetField(normalModeTable, "k", moveUpFunc)
	a.l.SetField(normalModeTable, "j", moveDownFunc)

	a.l.SetField(normalModeTable, "enter", submitFunc)
	a.l.SetField(normalModeTable, "return", submitFunc)

	a.l.SetGlobal("settings", settingsTable)

	if err := a.l.DoString(a.settingsScript); err != nil {
		log.Printf("ERROR: evaluation of settings script failed: %v", err)
	}
}

func (a *App) executeCode(code string) {
	var f *lua.LFunction
	f, err := a.l.LoadString("return " + code)
	if err != nil {
		f, err = a.l.LoadString(code)
		if err != nil {
			log.Printf("ERROR: %v", err)
		}
	}

	if err := a.l.CallByParam(lua.P{
		Fn:      f,
		NRet:    0,
		Protect: true,
	}); err != nil {
		log.Printf("ERROR: %v", err)
	}

	if err == nil {
		a.historyAppend(code)
		a.commandEntry.SetHistory(a.history)
	}

	a.loadKeyBindingsDefinitions()
}

func (a *App) historyAppend(cmd string) error {
	f, err := os.OpenFile(a.histFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(cmd + "\n")); err != nil {
		return err
	}

	a.history = append(a.history, cmd)

	return nil
}

func (a *App) loadHistory() error {
	fbytes, err := os.ReadFile(a.histFilePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}

	a.history = []string{}
	content := string(fbytes)
	for line := range strings.Lines(content) {
		a.history = append(a.history, strings.TrimRight(line, "\n"))
	}

	return nil
}

func (a *App) loadKeyBindingsDefinitions() {
	settingsTable, ok := a.l.GetGlobal("settings").(*lua.LTable)
	if !ok {
		log.Printf("invalid settings table found")
		return
	}

	keyBindingsTable, ok := settingsTable.RawGet(lua.LString("key_bindings")).(*lua.LTable)
	if !ok {
		log.Printf("invalid key_bindings table found")
		return
	}

	var err error
	keyBindings := map[Mode][]string{}

	keyBindingsTable.ForEach(func(k lua.LValue, v lua.LValue) {
		if err != nil {
			return
		}

		modeLS, ok := k.(lua.LString)
		if !ok {
			err = fmt.Errorf("invalid mode found in key_bindings table: %+v", k)
			log.Printf(err.Error())
			return
		}
		var mode Mode
		switch modeLS.String() {
		case "normal":
			mode = ModeNormal
		case "command":
			mode = ModeCommand
		case "search":
			mode = ModeSearch
		default:
			err = fmt.Errorf("invalid mode name found in key_bindings table: %+v", k)
			return
		}

		modeKeyBindingsTable, ok := v.(*lua.LTable)
		if !ok {
			err = fmt.Errorf("invalid value found in key_bindings table for %s mode", mode)
			log.Printf(err.Error())
			return
		}

		keyBindings[mode] = []string{}
		modeKeyBindingsTable.ForEach(func(kbK lua.LValue, kbV lua.LValue) {
			if err != nil {
				return
			}

			kbStr, ok := kbK.(lua.LString)
			if !ok {
				err = fmt.Errorf("invalid key binding found in %s mode key_bindings", mode)
				log.Printf(err.Error())
				return
			}

			keyBindings[mode] = append(keyBindings[mode], kbStr.String())
		})
	})

	if err != nil {
		log.Printf("could not load keybindings: %v", err)
		return
	}

	a.keyBindings = keyBindings
}

func (a *App) executeKeyBinding(kb string) {
	settingsTable, ok := a.l.GetGlobal("settings").(*lua.LTable)
	if !ok {
		log.Printf("invalid settings table found")
		return
	}

	keyBindingsTable, ok := settingsTable.RawGet(lua.LString("key_bindings")).(*lua.LTable)
	if !ok {
		log.Printf("invalid key_bindings table found")
		return
	}

	kbModeTable, ok := keyBindingsTable.RawGet(lua.LString(a.mode.String())).(*lua.LTable)
	if !ok {
		log.Printf("key binding table not available for mode %s", a.mode)
		return
	}

	kbFunc, ok := kbModeTable.RawGet(lua.LString(kb)).(*lua.LFunction)
	if !ok {
		log.Printf("key binding not found: %s %s", a.mode, kb)
		return
	}

	if err := a.l.CallByParam(lua.P{
		Fn:      kbFunc,
		NRet:    0,
		Protect: true,
	}); err != nil {
		log.Printf("ERROR: key binding failed: %v", err)
		return
	}
}

func (a *App) configureCanvasKeyBindings(mode Mode) {
	// Remove current shortcuts
	for _, mode := range []Mode{ModeNormal, ModeCommand, ModeSearch} {
		for _, kb := range a.keyBindings[mode] {
			var mod fyne.KeyModifier
			keys := strings.Fields(strings.ToLower(kb))
			for _, k := range keys {
				switch k {
				case "ctrl":
					mod |= fyne.KeyModifierControl
				case "shift":
					mod |= fyne.KeyModifierShift
				case "alt":
					mod |= fyne.KeyModifierAlt
				case "super":
					mod |= fyne.KeyModifierSuper
				}
			}

			keyName := keys[len(keys)-1]

			shortcut := &desktop.CustomShortcut{KeyName: fyne.KeyName(strings.ToUpper(keyName)), Modifier: mod}
			a.window.Canvas().RemoveShortcut(shortcut)
		}
	}

	// configure new shortcuts
	singleKeyKeyBindings := map[string]bool{}
OUTER:
	for _, kb := range a.keyBindings[mode] {
		var mod fyne.KeyModifier
		keys := strings.Fields(strings.ToLower(kb))

		if len(keys) == 1 {
			singleKeyKeyBindings[strings.ToLower(keys[0])] = true
		}

		for _, k := range keys[:len(keys)-1] {
			switch k {
			case "ctrl":
				mod |= fyne.KeyModifierControl
			case "shift":
				mod |= fyne.KeyModifierShift
			case "alt":
				mod |= fyne.KeyModifierAlt
			case "super":
				mod |= fyne.KeyModifierSuper
			default:
				log.Printf("WARNING: invalid key binding: '%s'", kb)
				continue OUTER
			}
		}

		keyName := keys[len(keys)-1]

		shortcut := &desktop.CustomShortcut{KeyName: fyne.KeyName(strings.ToUpper(keyName)), Modifier: mod}
		a.window.Canvas().AddShortcut(shortcut, func(shortcut fyne.Shortcut) {
			a.executeKeyBinding(kb)
		})
	}

	a.window.Canvas().SetOnTypedKey(func(k *fyne.KeyEvent) {
		kb := strings.ToLower(string(k.Name))
		if ok := singleKeyKeyBindings[kb]; ok {
			a.executeKeyBinding(kb)
		}
	})
}

func (a *App) applyKeyBindings(mode Mode) {
	if kbs, ok := a.keyBindings[mode]; ok {
		a.commandEntry.SetKeyBindings(NewKeyBindings(
			kbs,
			a.executeKeyBinding,
		))
	}

	a.configureCanvasKeyBindings(mode)
}
