package main

import (
	"fmt"
	"os"

	"tel42edgetts/internal/agi"
	"tel42edgetts/internal/config"
)

var (
	Version = "dev"
)

func main() {
	settings, err := config.Parse()
	if err != nil {
		handleConfigError(err)
	}

	// CLI test mode: use mock AGI session
	if settings.CLIMode {
		runCLIMode(settings)
		os.Exit(0)
	}

	// Normal AGI mode
	session := agi.New()
	if err := session.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize AGI: %v\n", err)
		os.Exit(1)
	}

	settings.ApplyAGIVariables(session.GetVariable)

	if err := settings.Validate(); err != nil {
		handleValidationError(session, err)
	}

	app := NewApp(session, settings)
	if err := app.Run(); err != nil {
		handleRunError(session, err)
	}

	os.Exit(0)
}

// runCLIMode runs the application in CLI test mode with mock AGI.
func runCLIMode(settings *config.Settings) {
	if err := settings.Validate(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	session := agi.NewMock()
	if err := session.Initialize(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to initialize mock AGI: %v\n", err)
		os.Exit(1)
	}

	app := NewCLIApp(session, settings)
	if err := app.Run(); err != nil {
		session.Logf(1, "EdgeTTS CLI Error: %v", err)
		session.SetStatus("ERROR")
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	os.Exit(0)
}

func handleConfigError(err error) {
	if _, ok := err.(*config.VersionError); ok {
		fmt.Printf("tel42edgetts version %s\n", Version)
		os.Exit(0)
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

// handleValidationError handles validation errors for both AGI and CLI modes.
func handleValidationError(session agi.SessionInterface, err error) {
	session.Logf(1, "EdgeTTS AGI Error: %v", err)
	session.SetStatus("ERROR")
	os.Exit(0)
}

// handleRunError handles run errors for both AGI and CLI modes.
func handleRunError(session agi.SessionInterface, err error) {
	session.Logf(1, "EdgeTTS AGI Error: %v", err)
	session.SetStatus("ERROR")
	os.Exit(0)
}
