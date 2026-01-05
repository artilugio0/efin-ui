package efinui

import (
	"database/sql"
	"fmt"
	"image/color"
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

	keyBindings map[string][]string

	search string

	// UI
	fyneApp fyne.App
	window  fyne.Window

	content *fyne.Container

	modeLabel    *ModeLabel
	commandEntry *CommandEntry

	tabs            []*MultiSplit
	currentTabIndex int

	toastSet *ToastSet

	helpDialog *HelpDialog

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
		widget.NewLabel("placeholder"),
	)

	toastSet := NewToastSet()

	helpDialog := NewHelpDialog(map[string]string{})

	windowContent := container.NewStack(content, helpDialog, toastSet)

	fyneApp := app.New()
	window := fyneApp.NewWindow("Efin")
	window.SetContent(windowContent)

	a.fyneApp = fyneApp
	a.window = window
	a.content = content
	a.modeLabel = modeLabel
	a.commandEntry = commandEntry
	a.tabs = tabs
	a.toastSet = toastSet
	a.helpDialog = helpDialog

	return a
}

func (a *App) Run() {
	a.loadHistory()
	a.commandEntry.SetHistory(a.history)

	a.initializeLuaState()
	a.loadKeyBindingsDefinitions()

	a.TabCreate()

	a.helpDialog.Hide()

	a.SetMode(a.mode)

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
		if f, ok := a.focusedObject.(fyne.Focusable); ok {
			a.window.Canvas().Focus(f)
		} else {
			a.window.Canvas().Focus(nil)
		}

		a.helpDialog.Hide()
		a.helpDialog.Refresh()
	} else if mode == ModeHelp {
		a.helpDialog.Show()
		a.helpDialog.Refresh()

		if f, ok := a.focusedObject.(fyne.Focusable); ok {
			a.window.Canvas().Focus(f)
		} else {
			a.window.Canvas().Focus(nil)
		}
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

func (a *App) updateHelpDialog() {
	settingsTable, ok := a.l.GetGlobal("settings").(*lua.LTable)
	if !ok {
		return
	}

	keyBindingsTable, ok := settingsTable.RawGet(lua.LString("key_bindings")).(*lua.LTable)
	if !ok {
		return
	}

	kbWidget, ok := a.focusedObject.(KeyBinder)
	if !ok {
		return
	}

	kbWidgetTable, ok := keyBindingsTable.RawGet(lua.LString(kbWidget.WidgetName())).(*lua.LTable)
	if !ok {
		return
	}

	keyBindings := map[string]string{}
	kbWidgetTable.ForEach(func(k lua.LValue, v lua.LValue) {
		if s, ok := k.(lua.LString); ok {
			description := "<no description>"

			if t, ok := v.(*lua.LTable); ok {
				if desc, ok := t.RawGet(lua.LString("desc")).(lua.LString); ok {
					description = string(desc)
				}
			}

			keyBindings[string(s)] = description
		}
	})

	a.helpDialog.SetDescriptions(keyBindings)
}

func (a *App) TabCreate() {
	tab := NewMultiSplit()
	tab.OnFocusMove = func(o fyne.CanvasObject) {
		a.focusedObject = o

		a.applyKeyBindingsToFocusedObject(a.mode)

		if f, ok := a.focusedObject.(fyne.Focusable); ok {
			a.window.Canvas().Focus(f)
		} else {
			a.window.Canvas().Focus(nil)
		}

		a.updateHelpDialog()
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

func (a *App) ToastMessage(message string) {
	a.toastSet.CreateToastMessage(message)
}

func (a *App) ToastError(message string) {
	a.toastSet.CreateToastError(message)
}

func (a *App) ThemeSet(theme CustomTheme) {
	a.fyneApp.Settings().SetTheme(theme)
}

func (a *App) RunQuery(query string) error {
	result, err := runQuery(a.db, query)
	if err != nil {
		return err
	}

	resultsTable := NewTable(result)
	resultsTable.ShowToastMessageFunc = a.ToastMessage

	resultsTable.OnSubmit = func(row []string) {
		var wg sync.WaitGroup
		wg.Add(1)
		var req *Request
		go func() {
			defer wg.Done()
			var err error
			req, err = getRequest(a.db, row[0])
			if err != nil {
				a.ToastError(fmt.Sprintf("ERROR: %v", err))
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
				a.ToastError(fmt.Sprintf("ERROR: %v", err))
				return
			}
		}()
		wg.Wait()

		reqResViewer := NewRequestResponseViewer(req, resp)
		reqResViewer.ShowToastMessageFunc = a.ToastMessage

		a.tabs[a.currentTabIndex].PaneCreate(reqResViewer)
	}

	a.tabs[a.currentTabIndex].SetCurrentPane(resultsTable)

	return nil
}

func (a *App) MessageSend(message Message) {
	if mh, ok := a.focusedObject.(MessageHandler); ok {
		mh.MessageHandle(message)
	}
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
		case "help":
			mode = ModeHelp
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

	setModeHelpFunc := a.l.NewFunction(func(ls *lua.LState) int {
		a.SetMode(ModeHelp)
		return 0
	})
	a.l.SetGlobal("set_mode_help", setModeHelpFunc)

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

	messageSendFunc := a.l.NewFunction(func(ls *lua.LState) int {
		message := a.l.ToString(1)
		a.MessageSend(message)
		return 0
	})
	a.l.SetGlobal("message_send", messageSendFunc)

	messageSend := func(message string) *lua.LFunction {
		return a.l.NewFunction(func(ls *lua.LState) int {
			a.MessageSend(message)
			return 0
		})
	}
	a.l.SetGlobal("message_send", messageSendFunc)

	toastFunc := a.l.NewFunction(func(ls *lua.LState) int {
		message := a.l.ToString(1)

		a.ToastMessage(message)

		return 0
	})
	a.l.SetGlobal("toast", toastFunc)

	toastErrFunc := a.l.NewFunction(func(ls *lua.LState) int {
		message := a.l.ToString(1)

		a.ToastError(message)

		return 0
	})
	a.l.SetGlobal("toast_err", toastErrFunc)

	themeSetFunc := a.l.NewFunction(func(ls *lua.LState) int {
		themeTable := a.l.ToTable(1)
		if themeTable == nil {
			a.l.ArgError(1, "theme table expected")
			return 0
		}

		theme := CustomTheme{}

		activeBackgroundNumber, ok := themeTable.RawGet(lua.LString("active_background")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'active_background'")
		}

		activeBackgroundValue := int(activeBackgroundNumber)
		theme.ActiveBackground = color.RGBA{
			R: uint8((activeBackgroundValue & 0xFF000000) >> 24),
			G: uint8((activeBackgroundValue & 0xFF0000) >> 16),
			B: uint8((activeBackgroundValue & 0xFF00) >> 8),
			A: uint8((activeBackgroundValue & 0xFF)),
		}

		backgroundNumber, ok := themeTable.RawGet(lua.LString("background")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'background'")
		}

		backgroundValue := int(backgroundNumber)
		theme.Background = color.RGBA{
			R: uint8((backgroundValue & 0xFF000000) >> 24),
			G: uint8((backgroundValue & 0xFF0000) >> 16),
			B: uint8((backgroundValue & 0xFF00) >> 8),
			A: uint8((backgroundValue & 0xFF)),
		}

		disabledNumber, ok := themeTable.RawGet(lua.LString("disabled")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'disabled'")
			return 0
		}

		disabledValue := int(disabledNumber)
		theme.Disabled = color.RGBA{
			R: uint8((disabledValue & 0xFF000000) >> 24),
			G: uint8((disabledValue & 0xFF0000) >> 16),
			B: uint8((disabledValue & 0xFF00) >> 8),
			A: uint8((disabledValue & 0xFF)),
		}

		floatingBackgroundNumber, ok := themeTable.RawGet(lua.LString("floating_background")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'floating_background'")
			return 0
		}

		floatingBackgroundValue := int(floatingBackgroundNumber)
		theme.FloatingBackground = color.RGBA{
			R: uint8((floatingBackgroundValue & 0xFF000000) >> 24),
			G: uint8((floatingBackgroundValue & 0xFF0000) >> 16),
			B: uint8((floatingBackgroundValue & 0xFF00) >> 8),
			A: uint8((floatingBackgroundValue & 0xFF)),
		}

		foregroundNumber, ok := themeTable.RawGet(lua.LString("foreground")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'foreground'")
			return 0
		}

		foregroundValue := int(foregroundNumber)
		theme.Foreground = color.RGBA{
			R: uint8((foregroundValue & 0xFF000000) >> 24),
			G: uint8((foregroundValue & 0xFF0000) >> 16),
			B: uint8((foregroundValue & 0xFF00) >> 8),
			A: uint8((foregroundValue & 0xFF)),
		}

		headerBackgroundNumber, ok := themeTable.RawGet(lua.LString("header_background")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'header_background'")
			return 0
		}

		headerBackgroundValue := int(headerBackgroundNumber)
		theme.HeaderBackground = color.RGBA{
			R: uint8((headerBackgroundValue & 0xFF000000) >> 24),
			G: uint8((headerBackgroundValue & 0xFF0000) >> 16),
			B: uint8((headerBackgroundValue & 0xFF00) >> 8),
			A: uint8((headerBackgroundValue & 0xFF)),
		}

		hoverNumber, ok := themeTable.RawGet(lua.LString("hover")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'hover'")
			return 0
		}

		hoverValue := int(hoverNumber)
		theme.Hover = color.RGBA{
			R: uint8((hoverValue & 0xFF000000) >> 24),
			G: uint8((hoverValue & 0xFF0000) >> 16),
			B: uint8((hoverValue & 0xFF00) >> 8),
			A: uint8((hoverValue & 0xFF)),
		}

		inputBackgroundNumber, ok := themeTable.RawGet(lua.LString("input_background")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'input_background'")
			return 0
		}

		inputBackgroundValue := int(inputBackgroundNumber)
		theme.InputBackground = color.RGBA{
			R: uint8((inputBackgroundValue & 0xFF000000) >> 24),
			G: uint8((inputBackgroundValue & 0xFF0000) >> 16),
			B: uint8((inputBackgroundValue & 0xFF00) >> 8),
			A: uint8((inputBackgroundValue & 0xFF)),
		}

		inputBorderNumber, ok := themeTable.RawGet(lua.LString("input_border")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'input_border'")
			return 0
		}

		inputBorderValue := int(inputBorderNumber)
		theme.InputBorder = color.RGBA{
			R: uint8((inputBorderValue & 0xFF000000) >> 24),
			G: uint8((inputBorderValue & 0xFF0000) >> 16),
			B: uint8((inputBorderValue & 0xFF00) >> 8),
			A: uint8((inputBorderValue & 0xFF)),
		}

		placeHolderNumber, ok := themeTable.RawGet(lua.LString("place_holder")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'place_holder'")
			return 0
		}

		placeHolderValue := int(placeHolderNumber)
		theme.PlaceHolder = color.RGBA{
			R: uint8((placeHolderValue & 0xFF000000) >> 24),
			G: uint8((placeHolderValue & 0xFF0000) >> 16),
			B: uint8((placeHolderValue & 0xFF00) >> 8),
			A: uint8((placeHolderValue & 0xFF)),
		}

		primaryNumber, ok := themeTable.RawGet(lua.LString("primary")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'primary'")
			return 0
		}

		primaryValue := int(primaryNumber)
		theme.Primary = color.RGBA{
			R: uint8((primaryValue & 0xFF000000) >> 24),
			G: uint8((primaryValue & 0xFF0000) >> 16),
			B: uint8((primaryValue & 0xFF00) >> 8),
			A: uint8((primaryValue & 0xFF)),
		}

		scrollBarNumber, ok := themeTable.RawGet(lua.LString("scroll_bar")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'scroll_bar'")
			return 0
		}

		scrollBarValue := int(scrollBarNumber)
		theme.ScrollBar = color.RGBA{
			R: uint8((scrollBarValue & 0xFF000000) >> 24),
			G: uint8((scrollBarValue & 0xFF0000) >> 16),
			B: uint8((scrollBarValue & 0xFF00) >> 8),
			A: uint8((scrollBarValue & 0xFF)),
		}

		scrollBarBackgroundNumber, ok := themeTable.RawGet(lua.LString("scroll_bar_background")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'scroll_bar_background'")
			return 0
		}

		scrollBarBackgroundValue := int(scrollBarBackgroundNumber)
		theme.ScrollBarBackground = color.RGBA{
			R: uint8((scrollBarBackgroundValue & 0xFF000000) >> 24),
			G: uint8((scrollBarBackgroundValue & 0xFF0000) >> 16),
			B: uint8((scrollBarBackgroundValue & 0xFF00) >> 8),
			A: uint8((scrollBarBackgroundValue & 0xFF)),
		}

		selectionNumber, ok := themeTable.RawGet(lua.LString("selection")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'selection'")
			return 0
		}

		selectionValue := int(selectionNumber)
		theme.Selection = color.RGBA{
			R: uint8((selectionValue & 0xFF000000) >> 24),
			G: uint8((selectionValue & 0xFF0000) >> 16),
			B: uint8((selectionValue & 0xFF00) >> 8),
			A: uint8((selectionValue & 0xFF)),
		}

		separatorNumber, ok := themeTable.RawGet(lua.LString("separator")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'separator'")
			return 0
		}

		separatorValue := int(separatorNumber)
		theme.Separator = color.RGBA{
			R: uint8((separatorValue & 0xFF000000) >> 24),
			G: uint8((separatorValue & 0xFF0000) >> 16),
			B: uint8((separatorValue & 0xFF00) >> 8),
			A: uint8((separatorValue & 0xFF)),
		}

		shadowNumber, ok := themeTable.RawGet(lua.LString("shadow")).(lua.LNumber)
		if !ok {
			a.l.ArgError(1, "invalid value in theme table 'shadow'")
			return 0
		}

		shadowValue := int(shadowNumber)
		theme.Shadow = color.RGBA{
			R: uint8((shadowValue & 0xFF000000) >> 24),
			G: uint8((shadowValue & 0xFF0000) >> 16),
			B: uint8((shadowValue & 0xFF00) >> 8),
			A: uint8((shadowValue & 0xFF)),
		}

		a.ThemeSet(theme)

		return 0
	})
	a.l.SetGlobal("theme_set", themeSetFunc)

	settingsTable := a.l.NewTable()
	keyBindingsTable := a.l.NewTable()
	normalModeTable := a.l.NewTable()
	commandModeTable := a.l.NewTable()
	searchModeTable := a.l.NewTable()
	helpModeTable := a.l.NewTable()

	a.l.SetField(settingsTable, "key_bindings", keyBindingsTable)
	a.l.SetField(keyBindingsTable, "normal", normalModeTable)
	a.l.SetField(keyBindingsTable, "command", commandModeTable)
	a.l.SetField(keyBindingsTable, "search", searchModeTable)
	a.l.SetField(keyBindingsTable, "help", helpModeTable)

	a.l.SetField(normalModeTable, "ctrl i", setModeCommandFunc)
	a.l.SetField(normalModeTable, "ctrl /", setModeSearchFunc)
	a.l.SetField(normalModeTable, "ctrl q", setModeHelpFunc)
	a.l.SetField(normalModeTable, "escape", searchClearFunc)

	a.l.SetField(commandModeTable, "escape", setModeNormalFunc)
	a.l.SetField(commandModeTable, "ctrl j", commandHistoryNextFunc)
	a.l.SetField(commandModeTable, "ctrl k", commandHistoryPrevFunc)

	a.l.SetField(searchModeTable, "escape", setModeNormalFunc)

	a.l.SetField(helpModeTable, "escape", setModeNormalFunc)
	a.l.SetField(helpModeTable, "ctrl i", setModeCommandFunc)
	a.l.SetField(helpModeTable, "ctrl /", setModeSearchFunc)
	a.l.SetField(helpModeTable, "ctrl q", setModeHelpFunc)

	a.l.SetField(normalModeTable, "ctrl d", paneDeleteFunc)
	a.l.SetField(normalModeTable, "ctrl n", paneCreateFunc)
	a.l.SetField(helpModeTable, "ctrl d", paneDeleteFunc)
	a.l.SetField(helpModeTable, "ctrl n", paneCreateFunc)

	a.l.SetField(normalModeTable, "ctrl l", paneFocusRightFunc)
	a.l.SetField(normalModeTable, "ctrl h", paneFocusLeftFunc)
	a.l.SetField(normalModeTable, "ctrl k", paneFocusUpFunc)
	a.l.SetField(normalModeTable, "ctrl j", paneFocusDownFunc)
	a.l.SetField(helpModeTable, "ctrl l", paneFocusRightFunc)
	a.l.SetField(helpModeTable, "ctrl h", paneFocusLeftFunc)
	a.l.SetField(helpModeTable, "ctrl k", paneFocusUpFunc)
	a.l.SetField(helpModeTable, "ctrl j", paneFocusDownFunc)

	a.l.SetField(normalModeTable, "ctrl shift n", tabCreateFunc)
	a.l.SetField(normalModeTable, "ctrl shift d", tabDeleteFunc)
	a.l.SetField(normalModeTable, "ctrl shift h", tabPrevFunc)
	a.l.SetField(normalModeTable, "ctrl shift l", tabNextFunc)
	a.l.SetField(helpModeTable, "ctrl shift n", tabCreateFunc)
	a.l.SetField(helpModeTable, "ctrl shift d", tabDeleteFunc)
	a.l.SetField(helpModeTable, "ctrl shift h", tabPrevFunc)
	a.l.SetField(helpModeTable, "ctrl shift l", tabNextFunc)

	a.l.SetField(normalModeTable, "n", searchResultNextFunc)
	a.l.SetField(normalModeTable, "p", searchResultPrevFunc)
	a.l.SetField(helpModeTable, "n", searchResultNextFunc)
	a.l.SetField(helpModeTable, "p", searchResultPrevFunc)

	a.l.SetField(normalModeTable, "h", moveLeftFunc)
	a.l.SetField(normalModeTable, "l", moveRightFunc)
	a.l.SetField(normalModeTable, "k", moveUpFunc)
	a.l.SetField(normalModeTable, "j", moveDownFunc)
	a.l.SetField(helpModeTable, "h", moveLeftFunc)
	a.l.SetField(helpModeTable, "l", moveRightFunc)
	a.l.SetField(helpModeTable, "k", moveUpFunc)
	a.l.SetField(helpModeTable, "j", moveDownFunc)

	a.l.SetField(normalModeTable, "enter", submitFunc)
	a.l.SetField(normalModeTable, "return", submitFunc)
	a.l.SetField(helpModeTable, "enter", submitFunc)
	a.l.SetField(helpModeTable, "return", submitFunc)

	tableTable := a.l.NewTable()
	a.l.SetField(keyBindingsTable, "table", tableTable)
	a.l.SetField(tableTable, "c", toDescCallTable(a.l, "Copy row to clipboard", messageSend(TableMessageCopyRow)))

	requestResponseViewerTable := a.l.NewTable()
	a.l.SetField(keyBindingsTable, "request_response_viewer", requestResponseViewerTable)
	a.l.SetField(requestResponseViewerTable, "c",
		toDescCallTable(a.l, "Copy request to clipboard", messageSend(RequestResponseViewerMessageCopyRequest)))
	a.l.SetField(requestResponseViewerTable, "s",
		toDescCallTable(a.l, "Copy request script to clipboard", messageSend(RequestResponseViewerMessageCopyRequestScript)))
	a.l.SetField(requestResponseViewerTable, "r",
		toDescCallTable(a.l, "Copy request response to clipboad", messageSend(RequestResponseViewerMessageCopyResponse)))

	a.l.SetGlobal("settings", settingsTable)

	if err := a.l.DoString(a.settingsScript); err != nil {
		a.ToastError(fmt.Sprintf("Evaluation of settings script failed: %v", err))
	}
}

func (a *App) executeCode(code string) {
	var f *lua.LFunction
	f, err := a.l.LoadString("return " + code)
	if err != nil {
		f, err = a.l.LoadString(code)
		if err != nil {
			a.ToastError(fmt.Sprintf("ERROR: %v", err))
		}
	}

	if err := a.l.CallByParam(lua.P{
		Fn:      f,
		NRet:    0,
		Protect: true,
	}); err != nil {
		a.ToastError(fmt.Sprintf("ERROR: %v", err))
	}

	if err == nil {
		a.historyAppend(code)
		a.commandEntry.SetHistory(a.history)
	}

	a.loadKeyBindingsDefinitions()
	a.updateHelpDialog()
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
	keyBindings := map[string][]string{}

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
		mode := modeLS.String()

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

	kbCall := kbModeTable.RawGet(lua.LString(kb))
	var kbFunc *lua.LFunction
	switch kbCall.Type() {
	case lua.LTFunction:
		kbFunc = kbCall.(*lua.LFunction)

	case lua.LTTable:
		t := kbCall.(*lua.LTable)
		f, ok := t.RawGet(lua.LString("call")).(*lua.LFunction)
		if !ok {
			log.Printf("invalid key binding definition: %s %s", a.mode, kb)
			return
		}
		kbFunc = f

	default:
		// Check if there is a key binding specific for the widget
		kbWidget, ok := a.focusedObject.(KeyBinder)
		if !ok {
			log.Printf("key binding not found: %s %s", a.mode, kb)
			return
		}

		kbModeTable, ok := keyBindingsTable.RawGet(lua.LString(kbWidget.WidgetName())).(*lua.LTable)
		if !ok {
			log.Printf("key binding table not available for mode %s or widget %s", a.mode, kbWidget.WidgetName())
			return
		}

		kbWidgetCall := kbModeTable.RawGet(lua.LString(kb))
		switch kbWidgetCall.Type() {
		case lua.LTFunction:
			kbFunc = kbWidgetCall.(*lua.LFunction)

		case lua.LTTable:
			t := kbWidgetCall.(*lua.LTable)
			f, ok := t.RawGet(lua.LString("call")).(*lua.LFunction)
			if !ok {
				log.Printf("invalid key binding definition: %s %s", a.mode, kb)
				return
			}
			kbFunc = f

		default:
			log.Printf("key binding not found: %s %s", a.mode, kb)
			return
		}
	}

	if err := a.l.CallByParam(lua.P{
		Fn:      kbFunc,
		NRet:    0,
		Protect: true,
	}); err != nil {
		a.ToastError(fmt.Sprintf("ERROR: %v", err))
		return
	}
}

func (a *App) configureCanvasKeyBindings(mode Mode) {
	// Remove current shortcuts
	for _, mode := range []Mode{ModeNormal, ModeCommand, ModeSearch, ModeHelp} {
		for _, kb := range a.keyBindings[mode.String()] {
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
	for _, kb := range a.keyBindings[mode.String()] {
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
	activeKeyBindings := a.keyBindings[mode.String()]

	if len(activeKeyBindings) > 0 {
		a.commandEntry.SetKeyBindings(NewKeyBindings(
			activeKeyBindings,
			a.executeKeyBinding,
		))
	}

	a.applyKeyBindingsToFocusedObject(mode)

	a.configureCanvasKeyBindings(mode)
}

func (a *App) applyKeyBindingsToFocusedObject(mode Mode) {
	activeKeyBindings := a.keyBindings[mode.String()]
	if widg, ok := a.focusedObject.(KeyBinder); ok {
		if kbs, ok := a.keyBindings[widg.WidgetName()]; ok {
			activeKeyBindings = append(activeKeyBindings, kbs...)
		}

		widg.SetKeyBindings(NewKeyBindings(
			activeKeyBindings,
			a.executeKeyBinding,
		))
	}
}

func toDescCallTable(l *lua.LState, desc string, f *lua.LFunction) *lua.LTable {
	descCallableTable := l.NewTable()
	l.SetField(descCallableTable, "desc", lua.LString(desc))
	l.SetField(descCallableTable, "call", f)

	mt := l.GetTypeMetatable("desc_callable")
	if mt == lua.LNil {
		mt = l.NewTypeMetatable("desc_callable")
		l.SetGlobal("desc_callable", mt)
		l.SetField(mt, "__call", l.NewFunction(func(ls *lua.LState) int {
			self := l.ToTable(1)

			callFunc := self.RawGet(lua.LString("call")).(*lua.LFunction)

			if err := l.CallByParam(lua.P{
				Fn:      callFunc,
				NRet:    0,
				Protect: true,
			}); err != nil {
				log.Printf("ERROR: %v", err)
			}

			return 0
		}))
	}

	l.SetMetatable(descCallableTable, mt)

	return descCallableTable
}
