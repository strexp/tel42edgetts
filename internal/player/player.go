package player

import (
	"fmt"
	"io"

	"tel42edgetts/internal/agi"
)

// Player handles audio playback with AGI.
type Player struct {
	session agi.SessionInterface
}

// New creates a new Player with the given session.
func New(session agi.SessionInterface) *Player {
	return &Player{session: session}
}

type Mode string

const (
	ModeSingle Mode = "single"
	ModeHash   Mode = "hash"
)

type Result struct {
	UserInput   string
	Interrupted bool
	Timeout     bool
	Error       error
}

func (p *Player) Play(filePath string, ivrMode Mode, timeoutMs int) Result {
	if ivrMode == "" {
		return p.playback(filePath)
	}

	switch ivrMode {
	case ModeHash:
		return p.playHashMode(filePath, timeoutMs)
	case ModeSingle:
		return p.playSingleMode(filePath, timeoutMs)
	default:
		return p.playback(filePath)
	}
}

func (p *Player) playback(filePath string) Result {
	if err := p.session.ExecPlayback(filePath); err != nil {
		p.session.Logf(1, "EdgeTTS AGI Playback Error: %v", err)
		return Result{Error: err}
	}
	return Result{}
}

func (p *Player) playHashMode(filePath string, timeoutMs int) Result {
	result := Result{}

	reply, err := p.session.GetData(filePath, timeoutMs, 20)
	if err != nil {
		p.session.Logf(1, "EdgeTTS AGI IVR GetData Error: %v", err)
		result.Error = err
		return result
	}

	switch reply.GetRes() {
	case 0:
		result.Timeout = true
		p.session.Logf(1, "EdgeTTS AGI IVR: No user input (timeout)")
	default:
		if reply.GetRes() > 0 {
			result.UserInput = reply.GetDat()
			p.session.Logf(1, "EdgeTTS AGI IVR: User input received (hash mode): %s", result.UserInput)
		}
	}

	return result
}

func (p *Player) playSingleMode(filePath string, timeoutMs int) Result {
	result := Result{}

	escapeDigits := "0123456789*#"
	reply, err := p.session.ControlStreamFile(filePath, escapeDigits)
	if err != nil {
		p.session.Logf(1, "EdgeTTS AGI IVR ControlStreamFile Error: %v", err)
		result.Error = err
		return result
	}

	if reply.GetRes() > 0 {
		// interrupted playback
		result.UserInput = string(rune(reply.GetRes()))
		result.Interrupted = true
		p.session.Logf(1, "EdgeTTS AGI IVR: User input received (single mode, interrupted): %s", result.UserInput)
		return result
	}

	// completed without interruption, wait for input
	p.session.Logf(1, "EdgeTTS AGI IVR: Playback completed, waiting for user input...")

	waitReply, waitErr := p.session.WaitForDigit(timeoutMs)
	if waitErr != nil {
		p.session.Logf(1, "EdgeTTS AGI IVR WaitForDigit Error: %v", waitErr)
		return result
	}

	if waitReply.GetRes() > 0 {
		result.UserInput = string(rune(waitReply.GetRes()))
		p.session.Logf(1, "EdgeTTS AGI IVR: User input received (single mode, after playback): %s", result.UserInput)
	} else {
		result.Timeout = true
		p.session.Logf(1, "EdgeTTS AGI IVR: No user input after playback (timeout)")
	}

	return result
}

func (p *Player) LogSummary(text, lang, voice, format string, cacheEnabled, ivrEnabled bool, ivrMode Mode, timeoutMs int) {
	p.session.Logf(1, "EdgeTTS AGI: Synthesizing '%s' [Lang: %s, Voice: %s, Format: %s, Cache: %v, IVR: %v, Mode: %s, Timeout: %dms]",
		text, lang, voice, format, cacheEnabled, ivrEnabled, ivrMode, timeoutMs)
}

func (p *Player) LogResult(result Result) {
	if result.Error != nil {
		p.session.Logf(1, "EdgeTTS AGI Playback failed: %v", result.Error)
	}
}

// CLIPlayer extends Player with CLI-specific output methods.
type CLIPlayer struct {
	*Player
	stdout io.Writer
}

// NewCLI creates a new CLI player.
func NewCLI(session agi.SessionInterface, stdout io.Writer) *CLIPlayer {
	return &CLIPlayer{
		Player: New(session),
		stdout: stdout,
	}
}

// PrintSummary prints a CLI summary of the TTS request.
func (p *CLIPlayer) PrintSummary(text, lang, voice, format string, cacheEnabled, ivrEnabled bool, ivrMode Mode, timeoutMs int) {
	fmt.Fprintln(p.stdout)
	fmt.Fprintln(p.stdout, "=== EdgeTTS CLI Test Mode ===")
	fmt.Fprintln(p.stdout)
	fmt.Fprintf(p.stdout, "Text:       %s\n", text)
	fmt.Fprintf(p.stdout, "Language:   %s\n", lang)
	fmt.Fprintf(p.stdout, "Voice:      %s\n", voice)
	fmt.Fprintf(p.stdout, "Format:     %s\n", format)
	fmt.Fprintf(p.stdout, "Cache:      %v\n", cacheEnabled)
	fmt.Fprintf(p.stdout, "IVR:        %v\n", ivrEnabled)
	fmt.Fprintf(p.stdout, "IVR Mode:   %s\n", ivrMode)
	fmt.Fprintf(p.stdout, "Timeout:    %dms\n", timeoutMs)
	fmt.Fprintln(p.stdout)
}

// PrintResult prints the playback result for CLI mode.
func (p *CLIPlayer) PrintResult(result Result) {
	fmt.Fprintln(p.stdout)
	fmt.Fprintln(p.stdout, "=== Playback Result ===")
	if result.Error != nil {
		fmt.Fprintf(p.stdout, "Error:      %v\n", result.Error)
	} else {
		fmt.Fprintf(p.stdout, "User Input: %s\n", result.UserInput)
		fmt.Fprintf(p.stdout, "Timeout:    %v\n", result.Timeout)
		fmt.Fprintf(p.stdout, "Interrupted: %v\n", result.Interrupted)
	}
	fmt.Fprintln(p.stdout)
}
