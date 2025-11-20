package main

import (
	"database/sql"
	"embed"
	"fmt"
	"os"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"

	_ "modernc.org/sqlite"
)

//go:embed all:frontend/dist
var assets embed.FS

func main() {
	dbFile := "data.db"
	if len(os.Args) > 1 {
		dbFile = os.Args[1]
	}

	histFilePath := "efin.history"
	if len(os.Args) > 2 {
		histFilePath = os.Args[2]
	}

	db, err := sql.Open("sqlite", dbFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: could not open DB file: %v", err)
		os.Exit(1)
	}
	defer db.Close()

	// Create an instance of the app structure
	app := NewApp(db, histFilePath)

	// Create application with options
	err = wails.Run(&options.App{
		Title:  "efin-ui",
		Width:  1024,
		Height: 768,
		AssetServer: &assetserver.Options{
			Assets: assets,
		},
		BackgroundColour: &options.RGBA{R: 27, G: 38, B: 54, A: 1},
		OnStartup:        app.startup,
		Bind: []interface{}{
			app,
		},
	})

	if err != nil {
		println("Error:", err.Error())
	}
}
