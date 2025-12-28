package main

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type QueryResult [][]string

// runQuery runs the given query and returns a table with the results
func runQuery(db *sql.DB, query string) (QueryResult, error) {
	rows, err := db.QueryContext(context.TODO(), query)
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
	query := `
    SELECT 
        resp.response_id,
        resp.status_code,
        resp.body,
        
        -- Response headers as JSON array of objects
        COALESCE((
            SELECT json_group_array(json_object('name', h.name, 'value', h.value))
            FROM headers h 
            WHERE h.response_id = resp.response_id
        ), '[]') AS response_headers
        
    FROM responses resp
    WHERE resp.response_id = ?
    `

	row := db.QueryRow(query, id)

	var resp Response
	var headersJSON string

	err := row.Scan(
		&resp.ID,
		&resp.StatusCode,
		&resp.Body,
		&headersJSON,
	)
	if err != nil {
		return nil, err
	}

	// Unmarshal headers
	if err := json.Unmarshal([]byte(headersJSON), &resp.Headers); err != nil {
		return nil, err
	}

	return &resp, nil
}
