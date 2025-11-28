package main

import (
	"database/sql"
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
		l: L,
	}
}

func setupGlobals(L *lua.LState, uiState *UIState, db *sql.DB, lastResult *UIActionResult) {
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
}

func (le *LuaEvaluator) Eval(input string) {
	if err := le.l.DoString(input); err != nil {
		log.Printf("ERROR: %v", err)
	}
}
