package config

import (
	"flag"
	"fmt"
	"os"
	"slices"
	"strconv"
	"strings"
)

type Settings struct {
	// TTS
	Text   string
	Lang   string
	Voice  string
	Format string

	// cache
	CacheDir string
	Cache    bool

	// IVR
	IVR        bool
	IVRMode    string
	IVRTimeout int

	// CLI test mode
	CLIMode bool
}

// default values
const (
	DefaultLang       = "en-US"
	DefaultVoice      = "en-US-AvaMultilingualNeural"
	DefaultFormat     = "wav16"
	DefaultCacheDir   = "/tmp"
	DefaultIVRMode    = "single"
	DefaultIVRTimeout = 5000
)

// valid formats from asterisk
var ValidFormats = []string{"mp3", "wav", "wav16"}

func Parse() (*Settings, error) {
	s := &Settings{}

	// get default env
	envCache := getEnvBool("TTS_CACHE", true)
	envIVR := getEnvBool("TTS_IVR", false)
	envIVRMode := getEnvString("TTS_IVR_MODE", DefaultIVRMode)
	envIVRTimeout := getEnvInt("TTS_IVR_TIMEOUT", DefaultIVRTimeout)

	// flags
	flag.StringVar(&s.Lang, "lang", DefaultLang, "Language for TTS (e.g. en-US)")
	flag.StringVar(&s.Voice, "voice", DefaultVoice, "Voice name (e.g. Xiaoyi)")
	flag.StringVar(&s.Format, "format", DefaultFormat, "Output audio format (mp3, wav, wav16)")
	flag.StringVar(&s.CacheDir, "dir", DefaultCacheDir, "Directory to store cached audio files")
	flag.BoolVar(&s.Cache, "cache", envCache, "Enable caching of audio files (env: TTS_CACHE)")
	flag.BoolVar(&s.IVR, "ivr", envIVR, "Enable IVR mode - user input can interrupt playback (env: TTS_IVR)")
	flag.StringVar(&s.IVRMode, "ivr-mode", envIVRMode, "IVR input mode: 'single' for single digit, 'hash' for digits ending with # (env: TTS_IVR_MODE)")
	flag.IntVar(&s.IVRTimeout, "ivr-timeout", envIVRTimeout, "Timeout in milliseconds for waiting user input after playback (env: TTS_IVR_TIMEOUT)")
	flag.BoolVar(&s.CLIMode, "cli", false, "CLI test mode - mock AGI session for testing without Asterisk")

	versionFlag := flag.Bool("version", false, "Print version and exit")

	flag.Usage = func() {
		printUsage()
	}

	flag.Parse()

	if *versionFlag {
		return nil, &VersionError{}
	}

	// positional arguments
	args := flag.Args()
	if len(args) > 0 {
		s.Text = args[0]
	}

	s.Format = strings.ToLower(strings.TrimSpace(s.Format))
	if !isValidFormat(s.Format) {
		s.Format = DefaultFormat
	}

	return s, nil
}

// AGI session variables.
func (s *Settings) ApplyAGIVariables(getVar func(name string) string) {
	if v := getVar("TTS_TEXT"); v != "" {
		s.Text = v
	}
	if v := getVar("TTS_LANG"); v != "" {
		s.Lang = v
	}
	if v := getVar("TTS_VOICE"); v != "" {
		s.Voice = v
	}
	if v := getVar("TTS_FORMAT"); v != "" {
		v = strings.ToLower(strings.TrimSpace(v))
		if isValidFormat(v) {
			s.Format = v
		}
	}
	if v := getVar("TTS_CACHE_DIR"); v != "" {
		s.CacheDir = v
	}
	if v := getVar("TTS_CACHE"); v != "" {
		s.Cache = parseBool(v)
	}
	if v := getVar("TTS_IVR"); v != "" {
		s.IVR = parseBool(v)
	}
	if v := getVar("TTS_IVR_MODE"); v != "" {
		v = strings.ToLower(strings.TrimSpace(v))
		if v == "single" || v == "hash" {
			s.IVRMode = v
		}
	}
	if v := getVar("TTS_IVR_TIMEOUT"); v != "" {
		s.IVRTimeout = parseTimeout(v)
	}
}

func (s *Settings) Validate() error {
	s.Text = strings.TrimSpace(s.Text)
	if s.Text == "" {
		return fmt.Errorf("no text provided to speak")
	}
	return nil
}

type VersionError struct{}

func (e *VersionError) Error() string { return "version requested" }

// helpers

func printUsage() {
	fmt.Fprintf(os.Stderr, "edge-tts-agi - Text-to-Speech for Asterisk using Edge TTS\n\n")
	fmt.Fprintf(os.Stderr, "Usage: %s [OPTIONS] [TEXT]\n", os.Args[0])
	fmt.Fprintf(os.Stderr, "Synthesize text-to-speech using Edge TTS for Asterisk.\n\n")
	fmt.Fprintf(os.Stderr, "Positional Arguments:\n")
	fmt.Fprintf(os.Stderr, "  TEXT         The text to synthesize (if AGI variable TTS_TEXT is not set)\n\n")
	fmt.Fprintf(os.Stderr, "Options:\n")
	flag.PrintDefaults()
	fmt.Fprintf(os.Stderr, "\nAsterisk Dialplan Variables (Take Precedence over flags):\n")
	fmt.Fprintf(os.Stderr, "  TTS_TEXT, TTS_LANG, TTS_VOICE, TTS_FORMAT, TTS_CACHE_DIR, TTS_CACHE\n")
	fmt.Fprintf(os.Stderr, "  TTS_IVR, TTS_IVR_MODE, TTS_IVR_TIMEOUT, TTS_USERINPUT (output)\n")
	fmt.Fprintf(os.Stderr, "\nCLI Test Mode:\n")
	fmt.Fprintf(os.Stderr, "  Use -cli flag to run without Asterisk for testing.\n")
	fmt.Fprintf(os.Stderr, "  In CLI mode, user input is read from stdin to simulate IVR.\n")
	fmt.Fprintf(os.Stderr, "\n")
}

func isValidFormat(format string) bool {
	return slices.Contains(ValidFormats, format)
}

func getEnvBool(name string, defaultVal bool) bool {
	if v := os.Getenv(name); v != "" {
		return parseBool(v)
	}
	return defaultVal
}

func getEnvString(name string, defaultVal string) string {
	if v := os.Getenv(name); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(name string, defaultVal int) int {
	if v := os.Getenv(name); v != "" {
		if i, err := strconv.Atoi(v); err == nil && i >= 0 {
			return i
		}
	}
	return defaultVal
}

func parseBool(s string) bool {
	s = strings.ToLower(strings.TrimSpace(s))
	return s == "true" || s == "1"
}

func parseTimeout(s string) int {
	s = strings.TrimSpace(s)
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	if err != nil || result < 0 {
		return DefaultIVRTimeout
	}
	return result
}
