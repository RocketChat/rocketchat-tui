package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/RocketChat/rocketchat-tui/ui"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/joho/godotenv"
)

var (
	debug bool
	prod  bool
	url   string
)

// To load environment variables from .env file
func initialiseEnvVariable() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	return nil
}

// It will check for flag arguments.
// If debug flag is set true all logs will be written in debug.log file.
// If prod flag is set true TUI will use Production server url for all rest and realtime function calls.
// If connection url is provided while starting TUI as flag value use that URL.
// Server Url is set in the model state and intial model state of the TUI will be returned for Tea to start TUI
func createModel() (*ui.Model, *os.File) {
	var loggerFile *os.File
	var err error
	var sUrl string

	if debug {
		loggerFile, err = tea.LogToFile("debug.log", "debug")
		if err != nil {
			fmt.Println("Error setting up logger")
		}
	}
	if prod {
		sUrl = os.Getenv("PROD_SERVER_URL")
	} else if url != "" {
		sUrl = url
	} else {
		sUrl = os.Getenv("DEV_SERVER_URL")
	}
	return ui.IntialModelState(sUrl), loggerFile
}

// It intialises environment variables to pick them from .env file.
// It defines all the flags and parse their values.
// Initial model state with all required methods is used to start the TUI..
func main() {
	err := initialiseEnvVariable()
	if err != nil {
		log.Println("Environment variable file not found", err)
		panic(err)
	}

	flag.BoolVar(&debug,
		"debug",
		false,
		"passing this flag will allow writing debug output to debug.log",
	)
	flag.BoolVar(&prod,
		"prod",
		false,
		"passing this flag will use production server url for connecting",
	)
	flag.StringVar(&url,
		"url",
		"",
		"user can pass sever url in it default is loacalhost",
	)
	flag.Parse()

	model, logger := createModel()
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
