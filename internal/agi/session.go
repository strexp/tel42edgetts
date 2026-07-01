package agi

import (
	"fmt"

	"github.com/zaf/agi"
)

// Reply wraps agi.Reply to implement our interfaces.
type Reply struct {
	Res int
	Dat string
}

func (r Reply) GetRes() int    { return r.Res }
func (r Reply) GetDat() string { return r.Dat }

// Verify Session implements SessionInterface
var _ SessionInterface = (*Session)(nil)

// Session wraps an AGI session with helper methods.
type Session struct {
	*agi.Session
	verbose bool
}

// New creates a new AGI session wrapper.
func New() *Session {
	return &Session{
		Session: agi.New(),
		verbose: true,
	}
}

// Initialize initializes the AGI session.
func (s *Session) Initialize() error {
	return s.Init(nil)
}

// GetVariable retrieves a variable from the AGI session.
// Returns empty string if the variable is not set.
func (s *Session) GetVariable(name string) string {
	if reply, err := s.Session.GetVariable(name); err == nil && reply.Res == 1 {
		return reply.Dat
	}
	return ""
}

// Logf logs a formatted message at the specified verbosity level.
func (s *Session) Logf(level int, format string, args ...interface{}) {
	if s.verbose {
		_, _ = s.Verbose(fmt.Sprintf(format, args...), level)
	}
}

// SetStatus sets the TTS_STATUS variable.
func (s *Session) SetStatus(status string) {
	_, _ = s.SetVariable("TTS_STATUS", status)
}

// SetUserInput sets the TTS_USERINPUT variable.
func (s *Session) SetUserInput(input string) {
	_, _ = s.SetVariable("TTS_USERINPUT", input)
}

// DisableMusic turns off music on hold.
func (s *Session) DisableMusic() {
	_, _ = s.SetMusic("off")
}

// ExecPlayback executes the Playback AGI command.
func (s *Session) ExecPlayback(filePath string) error {
	_, err := s.Exec("Playback", filePath)
	return err
}

// GetData executes the GetData AGI command.
func (s *Session) GetData(filePath string, timeoutMs, maxDigits int) (DataReply, error) {
	reply, err := s.Session.GetData(filePath, timeoutMs, maxDigits)
	if err != nil {
		return nil, err
	}
	return Reply{Res: reply.Res, Dat: reply.Dat}, nil
}

// ControlStreamFile executes the ControlStreamFile AGI command.
func (s *Session) ControlStreamFile(filePath, escapeDigits string) (ControlReply, error) {
	reply, err := s.Session.ControlStreamFile(filePath, escapeDigits)
	if err != nil {
		return nil, err
	}
	return Reply{Res: reply.Res, Dat: reply.Dat}, nil
}

// WaitForDigit executes the WaitForDigit AGI command.
func (s *Session) WaitForDigit(timeoutMs int) (DigitReply, error) {
	reply, err := s.Session.WaitForDigit(timeoutMs)
	if err != nil {
		return nil, err
	}
	return Reply{Res: reply.Res, Dat: reply.Dat}, nil
}
