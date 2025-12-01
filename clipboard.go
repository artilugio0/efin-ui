package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"html/template"
	"io"
	"log"
	"os/exec"
	"runtime"
	"strconv"
	"strings"

	_ "embed"
)

//go:embed files/request.tpl.lua
var testifierScript string

func copyToClipboard(text string) error {
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "darwin":
		cmd = exec.Command("pbcopy")
	case "windows":
		cmd = exec.Command("clip")
	case "linux":
		// Try xclip first
		if _, err := exec.LookPath("xclip"); err == nil {
			cmd = exec.Command("xclip", "-selection", "clipboard")
		} else if _, err := exec.LookPath("xsel"); err == nil {
			cmd = exec.Command("xsel", "--clipboard", "--input")
		} else {
			return fmt.Errorf("no clipboard utility found (xclip or xsel)")
		}
	default:
		return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
	}

	in, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	if err := cmd.Start(); err != nil {
		return err
	}

	if _, err := io.Copy(in, strings.NewReader(text)); err != nil {
		in.Close()
		return err
	}

	if err := in.Close(); err != nil {
		return err
	}

	return cmd.Wait()
}

func CopyRequestToClipboard(db *sql.DB, reqId int) {
	req, err := getRequest(db, strconv.Itoa(reqId))
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	funcs := map[string]any{
		"contains": strings.Contains,
		"contains_bytes": func(s []byte, c string) bool {
			return bytes.Contains(s, []byte(c))
		},
	}

	t, err := template.New("make_request").Funcs(funcs).Parse(testifierScript)
	if err != nil {
		log.Printf("ERROR: %v", err)
	}

	f := &strings.Builder{}

	if err := t.Execute(f, req); err != nil {
		log.Printf("ERROR: %v", err)
	}

	if err := copyToClipboard(f.String()); err != nil {
		log.Printf("ERROR: %v", err)
	}
}
