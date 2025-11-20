package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

// App struct
type App struct {
	ctx          context.Context
	db           *sql.DB
	histFilePath string
}

// NewApp creates a new App application struct
func NewApp(db *sql.DB, histFilePath string) *App {
	return &App{
		db:           db,
		histFilePath: histFilePath,
	}
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// SuggestCommand returns a list of command suggestions for the
// partial command given
func (a *App) SuggestCommand(cmd string) []string {
	return []string{
		"query SELECT COUNT(*) FROM responses",
		"query SELECT COUNT(*) FROM requests",
		"query SELECT name, value FROM headers WHERE response_id = 1",
		"query SELECT name, value FROM headers WHERE request_id = 1",
		"query SELECT * FROM responses LIMIT 100",
		"query SELECT * FROM requests LIMIT 100",
	}
}

// EvalCommand evaluates and returns the result of the command given
func (a *App) EvalCommand(cmd string) CommandResult {
	fields := strings.Fields(cmd)
	if len(fields) > 0 && fields[0] == "query" {
		rrTable, err := a.runQuery(strings.Join(fields[1:], " "))
		if err != nil {
			return CommandResult{
				ResultType: "error",
				Error:      err.Error(),
			}
		}

		return CommandResult{
			ResultType:           "request_response_table",
			RequestResponseTable: rrTable,
		}
	}

	return CommandResult{
		ResultType: "error",
		Error:      "invalid command",
	}
}

type CommandResult struct {
	ResultType string `json:"result_type"`
	Error      string `json:"error"`

	RequestResponseTable  RequestResponseTable  `json:"request_response_table"`
	RequestResponseDetail RequestResponseDetail `json:"request_response_detail"`
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
func (a *App) RowAction(row map[string]string) CommandResult {
	requestId, ok := row["request_id"]
	if !ok {
		return CommandResult{
			ResultType: "empty",
		}
	}

	req, err := getRequest(a.db, requestId)
	if err != nil {
		return CommandResult{
			ResultType: "error",
			Error:      err.Error(),
		}
	}

	resp, err := getResponse(a.db, requestId)
	if err != nil {
		return CommandResult{
			ResultType: "error",
			Error:      err.Error(),
		}
	}

	return CommandResult{
		ResultType: "request_response_detail",
		RequestResponseDetail: RequestResponseDetail{
			Request:  req,
			Response: resp,
		},
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
