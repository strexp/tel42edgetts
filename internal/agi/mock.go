package agi

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"time"
)

// MockReply implements the reply interfaces for mock sessions.
type MockReply struct {
	Res int
	Dat string
}

func (r MockReply) GetRes() int    { return r.Res }
func (r MockReply) GetDat() string { return r.Dat }

// Verify MockSession implements SessionInterface and player session interfaces
var _ SessionInterface = (*MockSession)(nil)

// MockSession implements a mock AGI session for CLI testing.
type MockSession struct {
	variables map[string]string
	verbose   bool
	reader    *bufio.Reader
}

// NewMock creates a new mock AGI session.
func NewMock() *MockSession {
	return &MockSession{
		variables: make(map[string]string),
		verbose:   true,
		reader:    bufio.NewReader(os.Stdin),
	}
}

// Initialize initializes the mock session (no-op for mock).
func (s *MockSession) Initialize() error {
	fmt.Println("[MOCK AGI] Session initialized")
	return nil
}

// GetVariable retrieves a mock variable.
func (s *MockSession) GetVariable(name string) string {
	if v, ok := s.variables[name]; ok {
		return v
	}
	return ""
}

// SetVariable sets a mock variable.
func (s *MockSession) SetVariable(name, value string) (int, error) {
	s.variables[name] = value
	if s.verbose {
		fmt.Printf("[MOCK AGI] SetVariable: %s = %s\n", name, value)
	}
	return 1, nil
}

// Logf logs a formatted message.
func (s *MockSession) Logf(level int, format string, args ...interface{}) {
	if s.verbose {
		fmt.Printf("[MOCK AGI Log L%d] %s\n", level, fmt.Sprintf(format, args...))
	}
}

// SetStatus sets the TTS_STATUS variable.
func (s *MockSession) SetStatus(status string) {
	s.SetVariable("TTS_STATUS", status)
}

// SetUserInput sets the TTS_USERINPUT variable.
func (s *MockSession) SetUserInput(input string) {
	s.SetVariable("TTS_USERINPUT", input)
}

// DisableMusic is a no-op for mock.
func (s *MockSession) DisableMusic() {
	if s.verbose {
		fmt.Println("[MOCK AGI] Music on hold disabled")
	}
}

// ExecPlayback mocks playback by printing and waiting for user input in IVR mode.
func (s *MockSession) ExecPlayback(filePath string) error {
	fmt.Printf("[MOCK AGI] Playing: %s\n", filePath)
	return nil
}

// Verbose prints a verbose message.
func (s *MockSession) Verbose(message string, level int) (int, error) {
	if s.verbose {
		fmt.Printf("[MOCK AGI Verbose L%d] %s\n", level, message)
	}
	return 1, nil
}

// GetData mocks the GetData AGI command for hash mode.
func (s *MockSession) GetData(filePath string, timeoutMs, maxDigits int) (DataReply, error) {
	fmt.Printf("[MOCK AGI] Playing (hash mode): %s\n", filePath)
	fmt.Printf("[MOCK AGI] Waiting for input (max %d digits, timeout %dms)...\n", maxDigits, timeoutMs)
	fmt.Print("Enter digits (end with # or press Enter for timeout): ")

	input, _ := s.reader.ReadString('\n')
	input = strings.TrimSpace(input)

	if input == "" {
		fmt.Println("[MOCK AGI] Timeout - no input received")
		return MockReply{Res: 0, Dat: ""}, nil
	}

	fmt.Printf("[MOCK AGI] Input received: %s\n", input)
	return MockReply{Res: len(input), Dat: input}, nil
}

// ControlStreamFile mocks the ControlStreamFile AGI command for single mode.
func (s *MockSession) ControlStreamFile(filePath, escapeDigits string) (ControlReply, error) {
	fmt.Printf("[MOCK AGI] Playing (single mode): %s\n", filePath)
	fmt.Printf("[MOCK AGI] Escape digits: %s\n", escapeDigits)
	fmt.Print("Press a digit to interrupt (or Enter to let it play): ")

	// Set a short timeout for input
	ch := make(chan string, 1)
	go func() {
		input, _ := s.reader.ReadString('\n')
		ch <- strings.TrimSpace(input)
	}()

	select {
	case input := <-ch:
		if input != "" {
			char := input[0]
			if strings.ContainsRune(escapeDigits, rune(char)) {
				fmt.Printf("[MOCK AGI] Interrupted by digit: %c\n", char)
				return MockReply{Res: int(char), Dat: string(char)}, nil
			}
		}
	case <-time.After(2 * time.Second):
		fmt.Println("[MOCK AGI] No interruption, playback completed")
	}

	return MockReply{Res: 0, Dat: ""}, nil
}

// WaitForDigit mocks the WaitForDigit AGI command.
func (s *MockSession) WaitForDigit(timeoutMs int) (DigitReply, error) {
	fmt.Printf("[MOCK AGI] Waiting for digit (timeout %dms)...\n", timeoutMs)
	fmt.Print("Press a digit (or Enter for timeout): ")

	ch := make(chan string, 1)
	go func() {
		input, _ := s.reader.ReadString('\n')
		ch <- strings.TrimSpace(input)
	}()

	select {
	case input := <-ch:
		if input != "" {
			char := input[0]
			fmt.Printf("[MOCK AGI] Digit received: %c\n", char)
			return MockReply{Res: int(char), Dat: string(char)}, nil
		}
	case <-time.After(time.Duration(timeoutMs) * time.Millisecond):
		fmt.Println("[MOCK AGI] Timeout - no digit received")
	}

	return MockReply{Res: 0, Dat: ""}, nil
}

// SetMusic is a no-op for mock.
func (s *MockSession) SetMusic(onOff string) (int, error) {
	if s.verbose {
		fmt.Printf("[MOCK AGI] SetMusic: %s\n", onOff)
	}
	return 1, nil
}
