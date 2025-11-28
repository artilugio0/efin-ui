package main

import (
	"log"

	lua "github.com/yuin/gopher-lua"
)

type LuaEvaluator struct {
	l       *lua.LState
	uiState *UIState
}

func NewLuaEvaluator(uiState *UIState) *LuaEvaluator {
	L := lua.NewState()
	L.OpenLibs()

	setupGlobals(L, uiState)

	return &LuaEvaluator{
		l: L,
	}
}

func setupGlobals(L *lua.LState, uiState *UIState) {
	L.SetGlobal("pane_create", L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneCreate()
		return 0
	}))

	L.SetGlobal("pane_delete", L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneDelete()
		return 0
	}))

	L.SetGlobal("pane_focus_next", L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneFocusNext()
		return 0
	}))

	L.SetGlobal("pane_focus_prev", L.NewFunction(func(ls *lua.LState) int {
		uiState.PaneFocusPrev()
		return 0
	}))

	L.SetGlobal("focused_pane_set_content", L.NewFunction(func(ls *lua.LState) int {
		newContent := L.ToInt(1)
		uiState.FocusedPaneSetContent(newContent)
		return 0
	}))

	L.SetGlobal("tab_create", L.NewFunction(func(ls *lua.LState) int {
		uiState.TabCreate()
		return 0
	}))

	L.SetGlobal("tab_focus_next", L.NewFunction(func(ls *lua.LState) int {
		uiState.TabFocusNext()
		return 0
	}))

	L.SetGlobal("tab_focus_prev", L.NewFunction(func(ls *lua.LState) int {
		uiState.TabFocusPrev()
		return 0
	}))
}

func (le *LuaEvaluator) Eval(input string) {
	if err := le.l.DoString(input); err != nil {
		log.Printf("ERROR: %v", err)
	}
}
