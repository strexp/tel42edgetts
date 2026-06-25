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

func handleConfigError(err error) {
	if _, ok := err.(*config.VersionError); ok {
		fmt.Printf("tel42edgetts version %s\n", Version)
		os.Exit(0)
	}
	fmt.Fprintf(os.Stderr, "Error: %v\n", err)
	os.Exit(1)
}

func handleValidationError(session *agi.Session, err error) {
	session.Logf(1, "EdgeTTS AGI Error: %v", err)
	session.SetStatus("ERROR")
	os.Exit(0)
}

func handleRunError(session *agi.Session, err error) {
	session.Logf(1, "EdgeTTS AGI Error: %v", err)
	session.SetStatus("ERROR")
	os.Exit(0)
}
