package agi

// DataReply defines the interface for GetData results.
type DataReply interface {
	GetRes() int
	GetDat() string
}

// ControlReply defines the interface for ControlStreamFile results.
type ControlReply interface {
	GetRes() int
}

// DigitReply defines the interface for WaitForDigit results.
type DigitReply interface {
	GetRes() int
}

// SessionInterface defines the common interface for AGI sessions.
// Both Session and MockSession implement this interface.
type SessionInterface interface {
	Initialize() error
	GetVariable(name string) string
	SetStatus(status string)
	SetUserInput(input string)
	DisableMusic()
	Logf(level int, format string, args ...interface{})
	ExecPlayback(filePath string) error
	GetData(filePath string, timeoutMs, maxDigits int) (DataReply, error)
	ControlStreamFile(filePath, escapeDigits string) (ControlReply, error)
	WaitForDigit(timeoutMs int) (DigitReply, error)
}
