package main

import (
	"fmt"
	"os"

	"tel42edgetts/internal/agi"
	"tel42edgetts/internal/config"
	"tel42edgetts/internal/core"
	"tel42edgetts/internal/player"
)

// App represents the TTS application.
type App struct {
	session     agi.SessionInterface
	settings    *config.Settings
	player      *player.Player
	synthesizer *core.Synthesizer
}

// NewApp creates a new App instance.
func NewApp(session agi.SessionInterface, settings *config.Settings) *App {
	return &App{
		session:     session,
		settings:    settings,
		player:      player.New(session),
		synthesizer: core.NewSynthesizer(settings),
	}
}

// Run executes the TTS application.
func (a *App) Run() error {
	a.logStart()

	result, err := a.synthesizer.Synthesize()
	if err != nil {
		return err
	}

	a.session.SetStatus("SUCCESS")
	a.session.DisableMusic()

	playResult := a.playAudio(result.PathNoExt)

	if playResult.Error != nil {
		a.session.SetStatus("ERROR")
	} else {
		a.session.SetUserInput(playResult.UserInput)
	}

	a.synthesizer.Cleanup(result.PathNoExt)

	return nil
}

func (a *App) logStart() {
	ivrMode := player.Mode(a.settings.IVRMode)
	a.player.LogSummary(
		a.settings.Text,
		a.settings.Lang,
		a.settings.Voice,
		a.settings.Format,
		a.settings.Cache,
		a.settings.IVR,
		ivrMode,
		a.settings.IVRTimeout,
	)
}

func (a *App) playAudio(filePath string) player.Result {
	ivrMode := player.Mode(a.settings.IVRMode)
	if a.settings.IVR {
		return a.player.Play(filePath, ivrMode, a.settings.IVRTimeout)
	}
	return a.player.Play(filePath, "", 0)
}

// CLIApp extends App with CLI-specific behavior.
type CLIApp struct {
	*App
	cliPlayer *player.CLIPlayer
}

// NewCLIApp creates a new CLI app instance.
func NewCLIApp(session *agi.MockSession, settings *config.Settings) *CLIApp {
	baseApp := &App{
		session:     session,
		settings:    settings,
		player:      player.New(session),
		synthesizer: core.NewSynthesizer(settings),
	}
	return &CLIApp{
		App:       baseApp,
		cliPlayer: player.NewCLI(session, os.Stdout),
	}
}

// Run executes the CLI TTS application.
func (a *CLIApp) Run() error {
	// Print CLI summary
	ivrMode := player.Mode(a.settings.IVRMode)
	a.cliPlayer.PrintSummary(
		a.settings.Text,
		a.settings.Lang,
		a.settings.Voice,
		a.settings.Format,
		a.settings.Cache,
		a.settings.IVR,
		ivrMode,
		a.settings.IVRTimeout,
	)

	// Log to AGI
	a.logStart()

	result, err := a.synthesizer.Synthesize()
	if err != nil {
		return err
	}

	// Print synthesis status
	if result.Cached {
		fmt.Printf("[CLI] Using cached file: %s\n", result.FullPath)
	} else {
		fmt.Printf("[CLI] Downloaded and saved: %s\n", result.FullPath)
	}

	a.session.SetStatus("SUCCESS")
	a.session.DisableMusic()

	playResult := a.playAudio(result.PathNoExt)

	if playResult.Error != nil {
		a.session.SetStatus("ERROR")
	} else {
		a.session.SetUserInput(playResult.UserInput)
	}

	a.cliPlayer.PrintResult(playResult)
	a.synthesizer.Cleanup(result.PathNoExt)

	return nil
}
