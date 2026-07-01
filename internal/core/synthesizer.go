package core

import (
	"fmt"
	"os"
	"time"

	"tel42edgetts/internal/audio"
	"tel42edgetts/internal/config"
	"tel42edgetts/internal/tts"
)

// Synthesizer handles TTS audio synthesis and caching.
type Synthesizer struct {
	settings *config.Settings
}

// NewSynthesizer creates a new synthesizer with the given settings.
func NewSynthesizer(settings *config.Settings) *Synthesizer {
	return &Synthesizer{settings: settings}
}

// Result holds the synthesis result.
type Result struct {
	PathNoExt string
	FullPath  string
	Cached    bool
}

// Synthesize generates or retrieves cached audio for the configured text.
// Returns the result and any error that occurred.
func (s *Synthesizer) Synthesize() (*Result, error) {
	voice := tts.ResolveVoice(s.settings.Lang, s.settings.Voice)
	hash := audio.GenerateHash(s.settings.Text, s.settings.Lang, voice, s.settings.Format)

	if !s.settings.Cache {
		hash = fmt.Sprintf("%s_%d", hash, time.Now().UnixNano())
	}

	pathNoExt, fullPath := audio.GetPaths(s.settings.CacheDir, hash, s.settings.Format)

	if s.settings.Cache && audio.Exists(fullPath) {
		return &Result{
			PathNoExt: pathNoExt,
			FullPath:  fullPath,
			Cached:    true,
		}, nil
	}

	mp3Data, err := tts.Download(s.settings.Text, voice)
	if err != nil {
		return nil, fmt.Errorf("TTS download failed: %w", err)
	}

	if err := s.saveAudio(fullPath, mp3Data); err != nil {
		return nil, err
	}

	return &Result{
		PathNoExt: pathNoExt,
		FullPath:  fullPath,
		Cached:    false,
	}, nil
}

func (s *Synthesizer) saveAudio(filePath string, mp3Data []byte) error {
	if s.settings.Format == "mp3" {
		return s.saveMP3(filePath, mp3Data)
	}
	return s.saveWAV(filePath, mp3Data)
}

func (s *Synthesizer) saveMP3(filePath string, mp3Data []byte) error {
	if err := os.WriteFile(filePath, mp3Data, 0644); err != nil {
		return fmt.Errorf("failed to write MP3 file: %w", err)
	}
	return nil
}

func (s *Synthesizer) saveWAV(filePath string, mp3Data []byte) error {
	targetRate := 8000
	if s.settings.Format == "wav16" {
		targetRate = 16000
	}

	if err := audio.ConvertToWav(filePath, mp3Data, targetRate); err != nil {
		return fmt.Errorf("failed to convert to WAV: %w", err)
	}
	return nil
}

// Cleanup removes the audio file if caching is disabled.
func (s *Synthesizer) Cleanup(pathNoExt string) {
	if !s.settings.Cache {
		_ = os.Remove(fmt.Sprintf("%s.%s", pathNoExt, s.settings.Format))
	}
}
