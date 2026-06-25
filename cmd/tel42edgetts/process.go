package main

import (
	"fmt"
	"os"
	"time"

	"tel42edgetts/internal/audio"
	"tel42edgetts/internal/tts"
)

func Process(text, lang, voiceName, format, cacheDir string, cacheEnabled bool) (string, error) {
	selectedVoice := tts.ResolveVoice(lang, voiceName)

	hash := audio.GenerateHash(text, lang, selectedVoice, format)
	if !cacheEnabled {
		hash = fmt.Sprintf("%s_%d", hash, time.Now().UnixNano())
	}
	
	pathNoExt, fullPath := audio.GetPaths(cacheDir, hash, format)

	if cacheEnabled && audio.Exists(fullPath) {
		return pathNoExt, nil
	}

	mp3Data, err := tts.Download(text, selectedVoice)
	if err != nil {
		return "", fmt.Errorf("edge TTS download failed: %w", err)
	}

	if format == "mp3" {
		if err := os.WriteFile(fullPath, mp3Data, 0644); err != nil {
			return "", fmt.Errorf("failed to write mp3 cache file: %w", err)
		}
	} else {
		targetRate := 8000
		if format == "wav16" {
			targetRate = 16000
		}
		if err := audio.ConvertToWav(fullPath, mp3Data, targetRate); err != nil {
			return "", fmt.Errorf("failed to convert to wav: %w", err)
		}
	}

	return pathNoExt, nil
}
