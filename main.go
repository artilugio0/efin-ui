package efinui

import (
	"database/sql"
	"fmt"
	"os"

	_ "modernc.org/sqlite"
)

func main() {
	dbFile := "data.db"
	if len(os.Args) > 1 {
		dbFile = os.Args[1]
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not open DB file: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	histFilePath := ".efin.history"

	settingsScriptPath := "efin-settings.lua"
	if len(os.Args) > 2 {
		settingsScriptPath = os.Args[2]
	}
	settingsScript := ""
	settingsScriptBytes, err := os.ReadFile(settingsScriptPath)
	if err == nil {
		settingsScript = string(settingsScriptBytes)
	}

	app := NewApp(db, histFilePath, settingsScript)

	app.Run()
}
