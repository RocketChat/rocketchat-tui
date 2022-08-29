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

func initialiseEnvVariable() error {
	err := godotenv.Load(".env")
	if err != nil {
		return err
	}
	return nil
}

func createModel(debug bool, prod bool) (*ui.Model, *os.File) {
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
	} else {
		sUrl = os.Getenv("DEV_SERVER_URL")
	}
	return ui.IntialModelState(sUrl), loggerFile
}

func main() {
	err := initialiseEnvVariable()
	if err != nil {
		log.Println("Environment variable file not found", err)
		panic(err)
	}
	
	debug := flag.Bool(
		"debug",
		false,
		"passing this flag will allow writing debug output to debug.log",
	)

	prod := flag.Bool(
		"prod",
		false,
		"passing this flag will use production server url for connecting",
	)
	flag.Parse()

	model, logger := createModel(*debug, *prod)
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
