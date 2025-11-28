package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"slices"
	"strings"
)

// App struct
type App struct {
	ctx            context.Context
	db             *sql.DB
	histFilePath   string
	commandHistory []string

	uiState          UIState
	lastContentIndex int
}

type UIState struct {
	CurrentTab  int    `json:"current_tab"`
	Tabs        []Pane `json:"tabs"`
	FocusedPane []int  `json:"focused_pane"`
}

type Pane struct {
	Layout  string `json:"layout"`
	Panes   []Pane `json:"panes"`
	Content int    `json:"content"`
}

// NewApp creates a new App application struct
func NewApp(db *sql.DB, histFilePath string) *App {
	return &App{
		db:             db,
		histFilePath:   histFilePath,
		commandHistory: []string{},
		uiState: UIState{
			FocusedPane: []int{0},
			CurrentTab:  0,
			Tabs: []Pane{
				{
					Layout: "vsplit",
					Panes: []Pane{
						{
							Layout:  "single",
							Content: 0,
						},
					},
				},
			},
		},
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
		if *action.CommandSubmitted == "createpane" {
			return a.EvalUIAction(UIAction{ActionType: UIActionCreatePane})
		}

		if *action.CommandSubmitted == "deletepane" {
			return a.EvalUIAction(UIAction{ActionType: UIActionDeletePane})
		}

		if *action.CommandSubmitted == "focuspaneprev" {
			return a.EvalUIAction(UIAction{ActionType: UIActionFocusPanePrev})
		}

		if *action.CommandSubmitted == "focuspanenext" {
			return a.EvalUIAction(UIAction{ActionType: UIActionFocusPaneNext})
		}

		if *action.CommandSubmitted == "createtab" {
			return a.EvalUIAction(UIAction{ActionType: UIActionCreateTab})
		}

		if *action.CommandSubmitted == "focustabnext" {
			return a.EvalUIAction(UIAction{ActionType: UIActionFocusTabNext})
		}

		return a.evalCommandSubmitted(*action.CommandSubmitted)

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
			UIState:    &a.uiState,
		}

	case UIActionCreatePane:
		newFocusedPane := a.createPane(&a.uiState.Tabs[a.uiState.CurrentTab], a.uiState.FocusedPane)
		a.uiState.FocusedPane = newFocusedPane

		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    &a.uiState,
		}

	case UIActionDeletePane:
		if len(a.uiState.Tabs[a.uiState.CurrentTab].Panes) <= 1 {
			return UIActionResult{
				ResultType: "ui_state_updated",
				UIState:    &a.uiState,
			}
		}

		newFocusedPane := a.deletePane(&a.uiState.Tabs[a.uiState.CurrentTab], a.uiState.FocusedPane)
		a.uiState.FocusedPane = newFocusedPane

		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    &a.uiState,
		}

	case UIActionFocusPanePrev:
		newFocusedPane := a.focusPanePrev(&a.uiState.Tabs[a.uiState.CurrentTab], a.uiState.FocusedPane)
		a.uiState.FocusedPane = newFocusedPane

		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    &a.uiState,
		}

	case UIActionFocusPaneNext:
		newFocusedPane := a.focusPaneNext(&a.uiState.Tabs[a.uiState.CurrentTab], a.uiState.FocusedPane)
		a.uiState.FocusedPane = newFocusedPane

		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    &a.uiState,
		}

	case UIActionCreateTab:
		a.createTab()

		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    &a.uiState,
		}

	case UIActionFocusTabNext:
		a.focusTabNext()

		return UIActionResult{
			ResultType: "ui_state_updated",
			UIState:    &a.uiState,
		}
	}

	return UIActionResult{
		ResultType: "error",
		Error:      "invalid command",
	}
}

func (a *App) createPane(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		pane.Panes = append(pane.Panes, Pane{
			Layout:  "single",
			Content: a.lastContentIndex,
		})
		return []int{len(pane.Panes) - 1}
	}

	restFocusedPane := a.createPane(&pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (a *App) deletePane(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		newPanes := []Pane{}
		for i, p := range pane.Panes {
			if i != focusedPane[0] {
				newPanes = append(newPanes, p)
			}
		}
		pane.Panes = newPanes
		if len(pane.Panes) == 0 {
			return []int{}
		}
		return []int{0}
	}

	restFocusedPane := a.deletePane(&pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (a *App) focusPaneNext(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		return []int{(focusedPane[0] + 1) % len(pane.Panes)}
	}

	restFocusedPane := a.focusPaneNext(&pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (a *App) focusPanePrev(pane *Pane, focusedPane []int) []int {
	if len(focusedPane) == 1 {
		return []int{(focusedPane[0] - 1) % len(pane.Panes)}
	}

	restFocusedPane := a.focusPanePrev(&pane.Panes[focusedPane[0]], focusedPane[1:])
	newFocusedPane := []int{focusedPane[0]}
	return append(newFocusedPane, restFocusedPane...)
}

func (a *App) updateFocusedPaneContent(pane Pane, focusedPane []int, newContent int) {
	if len(focusedPane) == 1 {
		pane.Panes[focusedPane[0]].Content = newContent
		return
	}

	a.updateFocusedPaneContent(pane.Panes[focusedPane[0]], focusedPane[1:], newContent)
}

func (a *App) createTab() {
	a.uiState.Tabs = append(a.uiState.Tabs, Pane{
		Layout: "vsplit",
		Panes: []Pane{
			{
				Layout:  "single",
				Content: 0,
			},
		},
	})

	a.uiState.CurrentTab = len(a.uiState.Tabs) - 1
	a.uiState.FocusedPane = []int{0}
}

func (a *App) focusTabNext() {
	a.uiState.CurrentTab = (a.uiState.CurrentTab + 1) % len(a.uiState.Tabs)
}

// EvalCommand evaluates and returns the result of the command given
func (a *App) evalCommandSubmitted(cmd string) UIActionResult {
	if err := a.updateHistory(cmd); err != nil {
		log.Printf("could not update history: %v", err)
	}

	fields := strings.Fields(cmd)
	if len(fields) > 0 && fields[0] == "query" {
		rrTable, err := a.runQuery(strings.Join(fields[1:], " "))
		if err != nil {
			return UIActionResult{
				ResultType: "error",
				Error:      err.Error(),
			}
		}

		a.lastContentIndex++
		a.updateFocusedPaneContent(a.uiState.Tabs[a.uiState.CurrentTab], a.uiState.FocusedPane, a.lastContentIndex)

		return UIActionResult{
			ResultType:           "request_response_table",
			RequestResponseTable: rrTable,
			UIState:              &a.uiState,
		}
	}

	return UIActionResult{
		ResultType: "error",
		Error:      "invalid command",
	}
}

// updateHistory adds the given command to the command history
func (a *App) updateHistory(cmd string) error {
	f, err := os.OpenFile(a.histFilePath, os.O_APPEND|os.O_WRONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer f.Close()

	if _, err := f.Write([]byte(cmd + "\n")); err != nil {
		return err
	}

	a.commandHistory = append(a.commandHistory, cmd)

	return nil
}

// readHistory reads command history from file
func (a *App) loadHistory() error {
	fbytes, err := os.ReadFile(a.histFilePath)
	if err != nil {
		return err
	}

	a.commandHistory = []string{}
	content := string(fbytes)
	for line := range strings.Lines(content) {
		a.commandHistory = append(a.commandHistory, strings.TrimRight(line, "\n"))
	}

	return nil
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
	UIActionCreatePane                 string = "create_pane"
	UIActionDeletePane                 string = "delete_pane"
	UIActionFocusPanePrev              string = "focus_pane_prev"
	UIActionFocusPaneNext              string = "focus_pane_next"
	UIActionCreateTab                  string = "create_tab"
	UIActionFocusTabNext               string = "focus_tab_next"
	UIActionUIStateRequested           string = "ui_state_requested"
)

type UIAction struct {
	ActionType string `json:"action_type"`

	CommandSubmitted           *string            `json:"command_submitted"`
	RowSubmitted               *map[string]string `json:"row_submitted"`
	CommandSuggestionRequested *string            `json:"command_suggestion_requested"`
}

type RequestResponseTable [][]string

// runQuery runs the given query and returns a table with the results
func (a *App) runQuery(query string) (RequestResponseTable, error) {
	rows, err := a.db.QueryContext(context.TODO(), query)
	if err != nil {
		log.Printf("error query: %v", err)
		return nil, err
	}
	defer rows.Close()

	columns, err := rows.Columns()
	if err != nil {
		log.Printf("error columns: %v", err)
		return nil, err
	}

	var table [][]string
	table = append(table, columns)
	values := make([]interface{}, len(columns))
	valuePtrs := make([]interface{}, len(columns))
	for i := range values {
		valuePtrs[i] = &values[i]
	}

	for rows.Next() {
		err := rows.Scan(valuePtrs...)
		if err != nil {
			log.Printf("error scan: %v", err)
			return nil, err
		}

		row := make([]string, len(columns))
		for i, val := range values {
			if val != nil {
				row[i] = fmt.Sprintf("%v", val)
			} else {
				row[i] = ""
			}
		}
		table = append(table, row)
	}

	if err := rows.Err(); err != nil {
		log.Printf("error rows err: %v", err)
		return nil, err
	}

	return table, nil
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

	a.lastContentIndex++
	a.updateFocusedPaneContent(a.uiState.Tabs[a.uiState.CurrentTab], a.uiState.FocusedPane, a.lastContentIndex)

	return UIActionResult{
		ResultType: "request_response_detail",
		RequestResponseDetail: RequestResponseDetail{
			Request:  req,
			Response: resp,
		},
		UIState: &a.uiState,
	}
}

type RequestResponseDetail struct {
	Request  *Request  `json:"request"`
	Response *Response `json:"response"`
}

type Request struct {
	ID string `json:"id"`

	Body      string `json:"body"`
	Host      string `json:"host"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`

	Headers []Header `json:"headers"`
}

type Response struct {
	ID string `json:"id"`

	StatusCode int    `json:"status_code"`
	Body       string `json:"body"`

	Headers []Header `json:"headers"`
}

type Header struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

func getRequest(db *sql.DB, id string) (*Request, error) {
	query := `
    SELECT 
        req.timestamp, req.request_id, req.method, req.url, req.body,
        COALESCE((
            SELECT json_group_array(json_object('name', h.name, 'value', h.value))
            FROM headers h 
            WHERE h.request_id = req.request_id
        ), '[]') AS req_headers
    FROM requests req
    WHERE req.request_id = ?
    `

	row := db.QueryRow(query, id)

	var req Request
	var headersJSON string

	err := row.Scan(&req.Timestamp, &req.ID, &req.Method, &req.URL, &req.Body, &headersJSON)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(headersJSON), &req.Headers); err != nil {
		return nil, err
	}

	// Extract Host if needed
	for _, h := range req.Headers {
		if strings.EqualFold(h.Name, "host") {
			req.Host = h.Value
			break
		}
	}

	return &req, nil
}

func getResponse(db *sql.DB, id string) (*Response, error) {
	log.Printf("getting response")
	respQuery := `
		SELECT resp.response_id, resp.status_code, resp.body
		FROM responses resp
		WHERE resp.response_id = ?
	`
	respRow := db.QueryRow(respQuery, id)
	if err := respRow.Err(); err != nil {
		return nil, err
	}
	resp := Response{}
	if err := respRow.Scan(&resp.ID, &resp.StatusCode, &resp.Body); err != nil {
		return nil, err
	}

	log.Printf("getting response headers")
	respHeadersQuery := `
		SELECT h.name, h.value
		FROM headers h
		WHERE h.response_id = ?
	`
	respHeadersRows, err := db.Query(respHeadersQuery, id)
	if err != nil {
		return nil, err
	}
	defer respHeadersRows.Close()

	respHeaders := []Header{}
	for respHeadersRows.Next() {
		h := Header{}
		if err := respHeadersRows.Scan(&h.Name, &h.Value); err != nil {
			return nil, err
		}
		respHeaders = append(respHeaders, h)
	}
	if err := respHeadersRows.Err(); err != nil {
		return nil, err
	}
	resp.Headers = respHeaders

	return &resp, nil
}
