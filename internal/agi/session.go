package agi

import (
	"fmt"

	"github.com/zaf/agi"
)

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
