package player

import (
	"tel42edgetts/internal/agi"
)

type Player struct {
	session *agi.Session
}

func New(session *agi.Session) *Player {
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

	switch reply.Res {
	case 0:
		result.Timeout = true
		p.session.Logf(1, "EdgeTTS AGI IVR: No user input (timeout)")
	default:
		if reply.Res > 0 {
			result.UserInput = reply.Dat
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

	if reply.Res > 0 {
		// interrupted playback
		result.UserInput = string(rune(reply.Res))
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

	if waitReply.Res > 0 {
		result.UserInput = string(rune(waitReply.Res))
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
