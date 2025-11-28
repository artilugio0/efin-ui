package main

import (
	"os"
	"strings"
)

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
