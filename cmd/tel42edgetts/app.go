package main

import (
	"fmt"
	"time"

	"tel42edgetts/internal/agi"
	"tel42edgetts/internal/audio"
	"tel42edgetts/internal/config"
	"tel42edgetts/internal/player"
	"tel42edgetts/internal/tts"
)

type App struct {
	session  *agi.Session
	settings *config.Settings
	player   *player.Player
}

func NewApp(session *agi.Session, settings *config.Settings) *App {
	return &App{
		session:  session,
		settings: settings,
		player:   player.New(session),
	}
}

func (a *App) Run() error {
	a.logStart()

	filePath, err := a.synthesizeAudio()
	if err != nil {
		return err
	}

	a.session.SetStatus("SUCCESS")
	a.session.DisableMusic()

	result := a.playAudio(filePath)

	if result.Error != nil {
		a.session.SetStatus("ERROR")
	} else {
		a.session.SetUserInput(result.UserInput)
	}

	a.cleanup(filePath)

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

func (a *App) synthesizeAudio() (string, error) {
	voice := tts.ResolveVoice(a.settings.Lang, a.settings.Voice)
	hash := audio.GenerateHash(a.settings.Text, a.settings.Lang, voice, a.settings.Format)

	if !a.settings.Cache {
		hash = fmt.Sprintf("%s_%d", hash, time.Now().UnixNano())
	}

	pathNoExt, fullPath := audio.GetPaths(a.settings.CacheDir, hash, a.settings.Format)

	if a.settings.Cache && audio.Exists(fullPath) {
		return pathNoExt, nil
	}

	mp3Data, err := tts.Download(a.settings.Text, voice)
	if err != nil {
		return "", fmt.Errorf("TTS download failed: %w", err)
	}

	if err := a.saveAudio(fullPath, mp3Data); err != nil {
		return "", err
	}

	return pathNoExt, nil
}

func (a *App) saveAudio(filePath string, mp3Data []byte) error {
	if a.settings.Format == "mp3" {
		return a.saveMP3(filePath, mp3Data)
	}
	return a.saveWAV(filePath, mp3Data)
}

func (a *App) saveMP3(filePath string, mp3Data []byte) error {
	if err := writeFile(filePath, mp3Data); err != nil {
		return fmt.Errorf("failed to write MP3 file: %w", err)
	}
	return nil
}

func (a *App) saveWAV(filePath string, mp3Data []byte) error {
	targetRate := 8000
	if a.settings.Format == "wav16" {
		targetRate = 16000
	}

	if err := audio.ConvertToWav(filePath, mp3Data, targetRate); err != nil {
		return fmt.Errorf("failed to convert to WAV: %w", err)
	}
	return nil
}

func (a *App) playAudio(filePath string) player.Result {
	ivrMode := player.Mode(a.settings.IVRMode)
	if a.settings.IVR {
		return a.player.Play(filePath, ivrMode, a.settings.IVRTimeout)
	}
	return a.player.Play(filePath, "", 0)
}

func (a *App) cleanup(filePath string) {
	if !a.settings.Cache {
		_ = removeFile(fmt.Sprintf("%s.%s", filePath, a.settings.Format))
	}
}
