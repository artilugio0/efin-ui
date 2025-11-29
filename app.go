package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"slices"
	"strings"
)

// App struct
type App struct {
	ctx            context.Context
	db             *sql.DB
	histFilePath   string
	commandHistory []string

	uiState      *UIState
	luaEvaluator *LuaEvaluator
	lastResult   *UIActionResult
}

// NewApp creates a new App application struct
func NewApp(db *sql.DB, histFilePath string) *App {
	uiState := &UIState{
		FocusedPane: []int{0},
		CurrentTab:  0,
		Tabs: []*Pane{
			{
				Layout: "vsplit",
				Panes: []*Pane{
					{
						Layout:  "single",
						Content: 0,
					},
				},
			},
		},
	}

	lastResult := UIActionResult{}

	return &App{
		luaEvaluator:   NewLuaEvaluator(uiState, db, &lastResult),
		db:             db,
		histFilePath:   histFilePath,
		commandHistory: []string{},
		uiState:        uiState,
		lastResult:     &lastResult,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
	if err := a.loadHistory(); err != nil {
		log.Printf("could not read command history: %v", err)
	}
}

// SuggestCommand returns a list of command suggestions for the
// partial command given
func (a *App) SuggestCommand(cmd string) []string {
	suggestions := []string{}
	exists := map[string]bool{}

	for _, histCmd := range slices.Backward(a.commandHistory) {
		lc := strings.ToLower(histCmd)
		if strings.HasPrefix(lc, strings.ToLower(cmd)) && !exists[lc] {
			suggestions = append(suggestions, histCmd)
			exists[lc] = true
		}
	}

	return suggestions
}

// EvalCommand evaluates and returns the result of the command given
func (a *App) EvalUIAction(action UIAction) UIActionResult {
	switch action.ActionType {
	case UIActionCommandSubmitted:
		if strings.HasPrefix(*action.CommandSubmitted, "lua") {
			if err := a.updateHistory(*action.CommandSubmitted); err != nil {
				log.Printf("could not update history: %v", err)
			}

			*a.lastResult = UIActionResult{
				ResultType: "ui_state_updated",
			}

			a.luaEvaluator.Eval(strings.Join(strings.Fields(*action.CommandSubmitted)[1:], " "))

			a.lastResult.UIState = a.uiState
			return *a.lastResult
		}

		return a.evalCommandSubmitted(*action.CommandSubmitted)

	case UIActionKeyBinding:
		log.Printf("attempting key binding: %s", *action.KeyBinding)
		*a.lastResult = UIActionResult{
			ResultType: "ui_state_updated",
		}

		evalString := fmt.Sprintf("settings.key_bindings.normal['%s']()", *action.KeyBinding)
		a.luaEvaluator.Eval(evalString)

		a.lastResult.UIState = a.uiState
		return *a.lastResult

	case UIActionRowSubmitted:
		return a.evalRowSubmitted(*action.RowSubmitted)

	case UIActionCommandSuggestionRequested:
		suggestions := a.SuggestCommand(*action.CommandSuggestionRequested)
		return UIActionResult{
			ResultType:        "command_suggestion",
			CommandSuggestion: suggestions,
		}

	case UIActionUIStateRequested:
		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    a.uiState,
		}

	}

	log.Printf("invalid action: %+v", action)
	return UIActionResult{
		ResultType: "error",
		Error:      "invalid action",
	}
}

// EvalCommand evaluates and returns the result of the command given
func (a *App) evalCommandSubmitted(cmd string) UIActionResult {
	return UIActionResult{
		ResultType: "error",
		Error:      "invalid command",
	}
}

type UIActionResult struct {
	ResultType string `json:"result_type"`
	Error      string `json:"error"`

	RequestResponseTable  RequestResponseTable  `json:"request_response_table"`
	RequestResponseDetail RequestResponseDetail `json:"request_response_detail"`
	CommandSuggestion     []string              `json:"command_suggestion"`

	UIState *UIState `json:"ui_state"`
}

const (
	UIActionCommandSubmitted           string = "command_submitted"
	UIActionRowSubmitted               string = "row_submitted"
	UIActionCommandSuggestionRequested string = "command_suggestion_requested"
	UIActionKeyBinding                 string = "key_binding"
	UIActionUIStateRequested           string = "ui_state_requested"
)

type UIAction struct {
	ActionType string `json:"action_type"`

	CommandSubmitted           *string            `json:"command_submitted"`
	RowSubmitted               *map[string]string `json:"row_submitted"`
	CommandSuggestionRequested *string            `json:"command_suggestion_requested"`
	KeyBinding                 *string            `json:"key_binding"`
}

type RequestResponseDetail struct {
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
}

// RowAction
func (a *App) evalRowSubmitted(row map[string]string) UIActionResult {
	requestId, ok := row["request_id"]
	if !ok {
		return UIActionResult{
			ResultType: "empty",
		}
	}

	req, err := getRequest(a.db, requestId)
	if err != nil {
		return UIActionResult{
			ResultType: "error",
			Error:      err.Error(),
		}
	}

	resp, err := getResponse(a.db, requestId)
	if err != nil {
		return UIActionResult{
			ResultType: "error",
			Error:      err.Error(),
		}
	}

	a.uiState.IncreaseLastContentIndex()
	a.uiState.FocusedPaneSetContentToLast()

	return UIActionResult{
		ResultType: "request_response_detail",
		RequestResponseDetail: RequestResponseDetail{
			Request:  req,
			Response: resp,
		},
		UIState: a.uiState,
	}
}
