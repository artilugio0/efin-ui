package cmd

import (
	"database/sql"
	"fmt"
	"os"

	efinui "github.com/artilugio0/efin-ui"
	"github.com/spf13/cobra"
)

const (
	DefaultDBFile       string = ""
	DefaultHistoryFile  string = ".efin.history"
	DefaultSettingsFile string = "efin-settings.lua"
)

// Execute runs the root command.
func Execute() {
	if err := NewUICmd("efin-ui").Execute(); err != nil {
		os.Exit(1)
	}
}

func NewUICmd(use string) *cobra.Command {
	var (
		dbFile       string
		settingsFile string
		historyFile  string
	)

	efinUICmd := &cobra.Command{
		Use:   use,
		Short: "Run Efin UI",
		Run: func(cmd *cobra.Command, args []string) {
			db, err := sql.Open("sqlite", dbFile)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Error: could not open DB file: %v", err)
				os.Exit(1)
			}
			defer db.Close()

			settingsScript := ""
			settingsScriptBytes, err := os.ReadFile(settingsFile)
			if err == nil {
				settingsScript = string(settingsScriptBytes)
			}

			app := efinui.NewApp(db, historyFile, settingsScript)
			app.Run()
		},
	}

	efinUICmd.Flags().StringVarP(
		&dbFile,
		"db-file",
		"D",
		DefaultDBFile,
		"Save requests and responses in the specified Sqlite3 db file",
	)

	efinUICmd.Flags().StringVarP(
		&settingsFile,
		"settings",
		"s",
		DefaultSettingsFile,
		"Settings file",
	)

	efinUICmd.Flags().StringVarP(
		&historyFile,
		"history",
		"H",
		DefaultHistoryFile,
		"History file",
	)

	efinUICmd.MarkFlagRequired("db-file")

	return efinUICmd
}
