package main

import (
	"database/sql"
	"fmt"
	"log"
	"strconv"

	lua "github.com/yuin/gopher-lua"
)

type LuaEvaluator struct {
	l       *lua.LState
	uiState *UIState
}

func NewLuaEvaluator(uiState *UIState, db *sql.DB, lastResult *UIActionResult) *LuaEvaluator {
	L := lua.NewState()
	L.OpenLibs()

	setupGlobals(L, uiState, db, lastResult)

	return &LuaEvaluator{
		l:       L,
		uiState: uiState,
	}
}

func setupGlobals(L *lua.LState, uiState *UIState, db *sql.DB, lastResult *UIActionResult) {
	paneCreateFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneCreate()
		return 0
	})
	L.SetGlobal("pane_create", paneCreateFunc)

	paneDeleteFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneDelete()
		return 0
	})
	L.SetGlobal("pane_delete", paneDeleteFunc)

	paneFocusNextFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneFocusNext()
		return 0
	})
	L.SetGlobal("pane_focus_next", paneFocusNextFunc)

	paneFocusPrevFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneFocusPrev()
		return 0
	})
	L.SetGlobal("pane_focus_prev", paneFocusPrevFunc)

	L.SetGlobal("focused_pane_set_content", L.NewFunction(func(ls *lua.LState) int {
		newContent := L.ToInt(1)
		uiState.FocusedPaneSetContent(newContent)
		return 0
	}))

	tabCreateFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.TabCreate()
		return 0
	})
	L.SetGlobal("tab_create", tabCreateFunc)

	tabFocusNextFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.TabFocusNext()
		return 0
	})
	L.SetGlobal("tab_focus_next", tabFocusNextFunc)

	tabFocusPrevFunc := L.NewFunction(func(ls *lua.LState) int {
		uiState.TabFocusPrev()
		return 0
	})
	L.SetGlobal("tab_focus_prev", tabFocusPrevFunc)

	L.SetGlobal("query", L.NewFunction(func(ls *lua.LState) int {
		queryStr := L.ToString(1)
		queryResult, err := runQuery(db, queryStr)
		if err != nil {
			*lastResult = UIActionResult{
				ResultType: "error",
				Error:      err.Error(),
			}

			return 0
		}

		*lastResult = UIActionResult{
			ResultType:           "request_response_table",
			RequestResponseTable: queryResult,
			UIState:              uiState,
		}

		uiState.IncreaseLastContentIndex()
		uiState.FocusedPaneSetContentToLast()

		return 0
	}))

	L.SetGlobal("request_show", L.NewFunction(func(ls *lua.LState) int {
		requestId := L.ToInt(1)
		req, err := getRequest(db, strconv.Itoa(requestId))
		if err != nil {
			*lastResult = UIActionResult{
				ResultType: "error",
				Error:      err.Error(),
			}

			return 0
		}

		resp, err := getResponse(db, strconv.Itoa(requestId))
		if err != nil {
			*lastResult = UIActionResult{
				ResultType: "error",
				Error:      err.Error(),
			}

			return 0
		}

		*lastResult = UIActionResult{
			ResultType: "request_response_detail",
			RequestResponseDetail: RequestResponseDetail{
				Request:  req,
				Response: resp,
			},
			UIState: uiState,
		}

		uiState.IncreaseLastContentIndex()
		uiState.FocusedPaneSetContentToLast()

		return 0
	}))

	settingsTable := L.NewTable()
	keyBindingsTable := L.NewTable()
	normalModeTable := L.NewTable()
	L.SetField(settingsTable, "key_bindings", keyBindingsTable)
	L.SetField(keyBindingsTable, "normal", normalModeTable)

	L.SetField(normalModeTable, "ctrl T", tabCreateFunc)
	L.SetField(normalModeTable, "ctrl L", tabFocusNextFunc)
	L.SetField(normalModeTable, "ctrl H", tabFocusPrevFunc)

	L.SetField(normalModeTable, "ctrl p", paneCreateFunc)
	L.SetField(normalModeTable, "ctrl l", paneFocusNextFunc)
	L.SetField(normalModeTable, "ctrl h", paneFocusPrevFunc)

	L.SetGlobal("settings", settingsTable)

	updateUIStateKeyBindings(L, uiState)
}

func updateUIStateKeyBindings(L *lua.LState, uiState *UIState) {
	settingsTable, ok := L.GetGlobal("settings").(*lua.LTable)
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
		return
	}

	uiState.KeyBindings = keyBindings
}

func (le *LuaEvaluator) Eval(input string) {
	if err := le.l.DoString(input); err != nil {
		log.Printf("ERROR: %v", err)
	}

	updateUIStateKeyBindings(le.l, le.uiState)
}
