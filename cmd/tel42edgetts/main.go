package main

import (
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/zaf/agi"
)

var Version = "dev" // Populated at build time

func getAgiVar(session *agi.Session, name string) string {
	if reply, err := session.GetVariable(name); err == nil && reply.Res == 1 {
		return reply.Dat
	}
	return ""
}

func main() {
	defaultCache := true
	if envCache := os.Getenv("TTS_CACHE"); strings.ToLower(envCache) == "false" || envCache == "0" {
		defaultCache = false
	}

	cmdLang := flag.String("lang", "en-US", "Language for TTS (e.g. en-US)")
	cmdVoice := flag.String("voice", "en-US-AvaMultilingualNeural", "Voice name (e.g. Xiaoyi)")
	cmdFormat := flag.String("format", "wav16", "Output audio format (mp3, wav, wav16)")
	cmdCacheDir := flag.String("dir", "/tmp", "Directory to store cached audio files")
	cmdCache := flag.Bool("cache", defaultCache, "Enable caching of audio files (env: TTS_CACHE)")
	cmdVersion := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "edge-tts-agi version %s\n\n", Version)
		fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [TEXT]\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Synthesize text-to-speech using Edge TTS for Asterisk.\n\n")
		fmt.Fprintf(os.Stderr, "Positional Arguments:\n")
		fmt.Fprintf(os.Stderr, "  TEXT         The text to synthesize (if AGI variable TTS_TEXT is not set)\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		flag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nAsterisk Dialplan Variables (Take Precedence over flags):\n")
		fmt.Fprintf(os.Stderr, "  TTS_TEXT, TTS_LANG, TTS_VOICE, TTS_FORMAT, TTS_CACHE_DIR, TTS_CACHE\n")
		fmt.Fprintf(os.Stderr, "\n")
	}

	flag.Parse()

	if *cmdVersion {
		fmt.Printf("tel42edgetts version %s\n", Version)
		os.Exit(0)
	}

	args := flag.Args()

	session := agi.New()
	if err := session.Init(nil); err != nil {
		os.Exit(1)
	}

	text := getAgiVar(session, "TTS_TEXT")
	if text == "" && len(args) > 0 {
		text = args[0]
	}
	text = strings.TrimSpace(text)
	if text == "" {
		session.Verbose("EdgeTTS AGI Error: No text provided to speak.", 1)
		session.SetVariable("TTS_STATUS", "ERROR")
		os.Exit(0)
	}

	lang := getAgiVar(session, "TTS_LANG")
	if lang == "" {
		lang = *cmdLang
	}

	voice := getAgiVar(session, "TTS_VOICE")
	if voice == "" {
		voice = *cmdVoice
	}

	format := getAgiVar(session, "TTS_FORMAT")
	if format == "" {
		format = *cmdFormat
	}
	format = strings.ToLower(strings.TrimSpace(format))
	if format != "mp3" && format != "wav" && format != "wav16" {
		format = "wav16"
	}

	cacheDir := getAgiVar(session, "TTS_CACHE_DIR")
	if cacheDir == "" {
		cacheDir = *cmdCacheDir
	}

	cacheEnabled := *cmdCache
	agiCache := getAgiVar(session, "TTS_CACHE")
	if agiCache != "" {
		agiCache = strings.ToLower(strings.TrimSpace(agiCache))
		switch agiCache {
		case "false", "0":
			cacheEnabled = false
		case "true", "1":
			cacheEnabled = true
		}
	}

	session.Verbose(fmt.Sprintf("EdgeTTS AGI: Synthesizing '%s' [Lang: %s, Voice: %s, Format: %s, Cache: %v]", text, lang, voice, format, cacheEnabled), 1)

	filePath, err := Process(text, lang, voice, format, cacheDir, cacheEnabled)
	if err != nil {
		session.Verbose(fmt.Sprintf("EdgeTTS AGI Processing Error: %v", err), 1)
		session.SetVariable("TTS_STATUS", "ERROR")
		os.Exit(0)
	}

	_, _ = session.SetVariable("TTS_STATUS", "SUCCESS")

	_, err = session.Exec("Playback", filePath)
	if err != nil {
		session.Verbose(fmt.Sprintf("EdgeTTS AGI Playback Error: %v", err), 1)
	}

	if !cacheEnabled {
		_ = os.Remove(fmt.Sprintf("%s.%s", filePath, format))
	}
}
