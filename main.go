package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/RocketChat/rocketchat-tui/ui"
	tea "github.com/charmbracelet/bubbletea"
)

func createModel(debug bool) (*ui.Model, *os.File) {
	var loggerFile *os.File
	var err error

	if debug {
		loggerFile, err = tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("Error setting up logger")
		}
	}
	return ui.IntialModelState(), loggerFile
}

func main() {

	debug := flag.Bool(
		"debug",
		false,
		"passing this flag will allow writing debug output to debug.log",
	)
	flag.Parse()

	model, logger := createModel(*debug)
	if logger != nil {
		defer logger.Close()
	}

	tui := tea.NewProgram(
		model,
		tea.WithAltScreen(),
	)
	if err := tui.Start(); err != nil {
		log.Fatal(err)
		os.Exit(1)
	}
}
