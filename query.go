package main

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/url"
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

	Body      []byte `json:"body"`
	Host      string `json:"host"`
	Method    string `json:"method"`
	Timestamp string `json:"timestamp"`
	URL       string `json:"url"`

	Headers []Header `json:"headers"`
}

func (r *Request) Raw() []byte {
	if r == nil {
		return nil
	}

	var buf bytes.Buffer

	// Request line: METHOD PATH HTTP/1.1
	path := r.URL
	if !strings.HasPrefix(path, "/") {
		// In case someone stored full URL in URL field
		if u, err := url.Parse(r.URL); err == nil {
			path = u.RequestURI()
		} else {
			path = "/"
		}
	}

	fmt.Fprintf(&buf, "%s %s HTTP/1.1\r\n", r.Method, path)

	// Special headers
	if r.Host != "" {
		fmt.Fprintf(&buf, "Host: %s\r\n", r.Host)
	}

	// Other headers
	for _, h := range r.Headers {
		// Skip Host if already written
		if strings.ToLower(h.Name) == "host" {
			continue
		}
		fmt.Fprintf(&buf, "%s: %s\r\n", h.Name, h.Value)
	}

	// End of headers
	buf.WriteString("\r\n")

	// Body
	if len(r.Body) > 0 {
		buf.Write(r.Body)
	}

	return buf.Bytes()
}

type Response struct {
	ID string `json:"id"`

	StatusCode int    `json:"status_code"`
	Body       []byte `json:"body"`

	Headers []Header `json:"headers"`
}

func (resp *Response) Raw() []byte {
	if resp == nil {
		return nil
	}

	var buf bytes.Buffer

	// Status line
	reason := http.StatusText(resp.StatusCode)
	if reason == "" {
		reason = "Unknown Status"
	}
	fmt.Fprintf(&buf, "HTTP/1.1 %d %s\r\n", resp.StatusCode, reason)

	// Headers
	for _, h := range resp.Headers {
		fmt.Fprintf(&buf, "%s: %s\r\n", h.Name, h.Value)
	}

	// End of headers
	buf.WriteString("\r\n")

	// Body
	if len(resp.Body) > 0 {
		buf.Write(resp.Body)
	}

	return buf.Bytes()
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
