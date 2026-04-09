// Package middleware provides terminal data processing middleware
// for Tabby's Go backend.
//
// It mirrors the TypeScript middleware chain from tabby-terminal:
// - UTF8Splitter: Ensures output is chunked at UTF-8 character boundaries
// - InputProcessor: Transforms keyboard input (backspace mapping)
// - LoginScriptProcessor: Matches output patterns and sends responses
// - OSCProcessor: Parses OSC escape sequences (working directory, etc.)
// - StreamProcessor: Handles newline conversion, echo, hex mode
package middleware

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"
	"unicode/utf8"
)

// ---- UTF8 Splitter ----

// UTF8Splitter ensures data is chunked at valid UTF-8 character boundaries.
// Incomplete multibyte sequences are buffered until the next write.
type UTF8Splitter struct {
	pending []byte
}

// Write processes data and returns complete UTF-8 chunks.
// Incomplete sequences at the end are buffered.
func (u *UTF8Splitter) Write(data []byte) []byte {
	if len(u.pending) > 0 {
		data = append(u.pending, data...)
		u.pending = nil
	}

	// Find the last valid UTF-8 boundary
	validEnd := len(data)
	for validEnd > 0 {
		if utf8.Valid(data[:validEnd]) {
			break
		}
		// Back up to find the start of the incomplete sequence
		validEnd--
		// Check if this could be the start of a multibyte sequence
		if data[validEnd] >= 0xC0 {
			// This is the start of an incomplete multibyte char
			u.pending = make([]byte, len(data)-validEnd)
			copy(u.pending, data[validEnd:])
			data = data[:validEnd]
			break
		}
		// If it's a continuation byte, keep backing up
		if data[validEnd] >= 0x80 && data[validEnd] < 0xC0 {
			continue
		}
		// It's a valid ASCII byte, so everything before it is valid
		validEnd++
		break
	}

	return data
}

// Flush returns any remaining buffered data.
func (u *UTF8Splitter) Flush() []byte {
	data := u.pending
	u.pending = nil
	return data
}

// ---- Input Processor ----

// BackspaceMode defines how the backspace key is interpreted
type BackspaceMode string

const (
	BackspaceCtrlH    BackspaceMode = "ctrl-h"
	BackspaceCtrlQ    BackspaceMode = "ctrl-?"
	BackspaceDelete   BackspaceMode = "delete"
	BackspaceBackspace BackspaceMode = "backspace"
)

// InputProcessingOptions configures input transformation
type InputProcessingOptions struct {
	Backspace BackspaceMode `json:"backspace"`
}

// InputProcessor transforms keyboard input before sending to the session.
// Currently handles backspace key mapping.
type InputProcessor struct {
	Options InputProcessingOptions
}

// NewInputProcessor creates a new input processor
func NewInputProcessor(opts InputProcessingOptions) *InputProcessor {
	return &InputProcessor{Options: opts}
}

// Process transforms input data according to the configured options.
// Returns the (possibly modified) data.
func (p *InputProcessor) Process(data []byte) []byte {
	if len(data) == 1 && data[0] == 0x7f {
		switch p.Options.Backspace {
		case BackspaceCtrlH:
			return []byte{0x08}
		case BackspaceCtrlQ:
			return []byte{0x7f}
		case BackspaceDelete:
			return []byte{0x1b, '[', '3', '~'}
		default:
			return []byte{0x7f}
		}
	}
	return data
}

// ---- Login Script Processor ----

// LoginScript defines a single expect/send pair for login automation
type LoginScript struct {
	Expect   string `json:"expect"`
	Send     string `json:"send"`
	IsRegex  bool   `json:"isRegex,omitempty"`
	Optional bool   `json:"optional,omitempty"`
}

// LoginScriptProcessor matches output patterns and sends responses
type LoginScriptProcessor struct {
	scripts    []LoginScript
	sendFunc   func(data []byte)
	logFunc    func(msg string)
}

// NewLoginScriptProcessor creates a new login script processor
func NewLoginScriptProcessor(scripts []LoginScript, sendFunc func([]byte), logFunc func(string)) *LoginScriptProcessor {
	// Unescape sequences in scripts
	processed := make([]LoginScript, len(scripts))
	for i, s := range scripts {
		expect := s.Expect
		if !s.IsRegex {
			expect = unescape(s.Expect)
		}
		processed[i] = LoginScript{
			Expect:   expect,
			Send:     unescape(s.Send),
			IsRegex:  s.IsRegex,
			Optional: s.Optional,
		}
	}
	return &LoginScriptProcessor{
		scripts:  processed,
		sendFunc: sendFunc,
		logFunc:  logFunc,
	}
}

// ProcessOutput examines session output and triggers matching scripts
func (l *LoginScriptProcessor) ProcessOutput(data []byte) {
	dataStr := string(data)

	for i := 0; i < len(l.scripts); i++ {
		script := l.scripts[i]
		if script.Expect == "" {
			continue
		}

		matched := false
		if script.IsRegex {
			re, err := regexp.Compile(script.Expect)
			if err == nil {
				matched = re.MatchString(dataStr)
			}
		} else {
			matched = strings.Contains(dataStr, script.Expect)
		}

		if matched {
			if l.logFunc != nil {
				l.logFunc(fmt.Sprintf("Executing login script: send %q", script.Send))
			}
			l.sendFunc([]byte(script.Send + "\n"))
			// Remove this script from the list
			l.scripts = append(l.scripts[:i], l.scripts[i+1:]...)
			i--
		} else if script.Optional {
			// Skip optional scripts that don't match
			l.scripts = append(l.scripts[:i], l.scripts[i+1:]...)
			i--
		} else {
			// Stop at first non-matching non-optional script
			break
		}
	}
}

// ExecuteUnconditionalScripts sends all scripts with empty expect patterns
func (l *LoginScriptProcessor) ExecuteUnconditionalScripts() {
	for i := 0; i < len(l.scripts); i++ {
		if l.scripts[i].Expect == "" {
			if l.logFunc != nil {
				l.logFunc(fmt.Sprintf("Executing unconditional script: %s", l.scripts[i].Send))
			}
			l.sendFunc([]byte(l.scripts[i].Send + "\n"))
			l.scripts = append(l.scripts[:i], l.scripts[i+1:]...)
			i--
		} else {
			break
		}
	}
}

// ---- OSC Processor ----

// OSCProcessor parses OSC (Operating System Command) escape sequences
// and extracts information like reported working directory.
type OSCProcessor struct {
	onCWD       func(cwd string)
	homeDir     string
}

// NewOSCProcessor creates a new OSC processor
func NewOSCProcessor(homeDir string, onCWD func(string)) *OSCProcessor {
	return &OSCProcessor{
		onCWD:   onCWD,
		homeDir: homeDir,
	}
}

// ProcessOutput examines data for OSC sequences and extracts information.
// Returns the data unchanged (OSC sequences are not stripped).
func (o *OSCProcessor) ProcessOutput(data []byte) []byte {
	oscPrefix := []byte{0x1b, ']'}
	bellSuffix := []byte{0x07}
	stSuffix := []byte{0x1b, '\\'}

	start := 0
	for {
		idx := bytes.Index(data[start:], oscPrefix)
		if idx == -1 {
			break
		}
		idx += start

		// Find the end of the OSC sequence
		params := data[idx+2:]
		bellIdx := bytes.Index(params, bellSuffix)
		stIdx := bytes.Index(params, stSuffix)

		endIdx := -1
		var suffixLen int
		if bellIdx >= 0 && (stIdx < 0 || bellIdx < stIdx) {
			endIdx = bellIdx
			suffixLen = 1
		} else if stIdx >= 0 {
			endIdx = stIdx
			suffixLen = 2
		}

		if endIdx == -1 {
			break // Incomplete sequence
		}

		oscString := string(params[:endIdx])
		start = idx + 2 + endIdx + suffixLen

		// Parse the OSC sequence
		parts := strings.SplitN(oscString, ";", 2)
		if len(parts) < 1 {
			continue
		}

		oscCode := 0
		fmt.Sscanf(parts[0], "%d", &oscCode)

		if oscCode == 1337 && len(parts) > 1 {
			paramString := parts[1]
			if strings.HasPrefix(paramString, "CurrentDir=") {
				cwd := strings.TrimPrefix(paramString, "CurrentDir=")
				if strings.HasPrefix(cwd, "~") {
					cwd = o.homeDir + cwd[1:]
				}
				if o.onCWD != nil {
					o.onCWD(cwd)
				}
			}
		}
	}

	return data
}

// ---- Stream Processor ----

// NewlineMode defines how newlines are transformed
type NewlineMode string

const (
	NewlineNone        NewlineMode = ""
	NewlineCR          NewlineMode = "cr"
	NewlineLF          NewlineMode = "lf"
	NewlineCRLF        NewlineMode = "crlf"
	NewlineImplicitCR  NewlineMode = "implicit_cr"
	NewlineImplicitLF  NewlineMode = "implicit_lf"
)

// StreamProcessingOptions configures stream processing
type StreamProcessingOptions struct {
	InputMode     string      `json:"inputMode"`     // null, "local-echo", "readline", "readline-hex"
	InputNewlines NewlineMode `json:"inputNewlines"`
	OutputMode    string      `json:"outputMode"`    // null, "hex"
	OutputNewlines NewlineMode `json:"outputNewlines"`
}

// StreamProcessor handles newline conversion, echo modes, and hex display
type StreamProcessor struct {
	Options StreamProcessingOptions
}

// NewStreamProcessor creates a new stream processor
func NewStreamProcessor(opts StreamProcessingOptions) *StreamProcessor {
	return &StreamProcessor{Options: opts}
}

// ProcessOutput transforms output data (newline conversion, hex mode)
func (s *StreamProcessor) ProcessOutput(data []byte) []byte {
	data = s.replaceNewlines(data, s.Options.OutputNewlines)
	// Hex mode would require hex dump formatting — skip for Go backend
	// since xterm.js handles display
	return data
}

// ProcessInput transforms input data (newline conversion)
func (s *StreamProcessor) ProcessInput(data []byte) []byte {
	return s.replaceNewlines(data, s.Options.InputNewlines)
}

// replaceNewlines transforms newline characters in data
func (s *StreamProcessor) replaceNewlines(data []byte, mode NewlineMode) []byte {
	if mode == "" {
		return data
	}

	switch mode {
	case NewlineImplicitCR:
		// Replace bare \n with \r\n (but not if already \r\n)
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\n"), []byte("\r\n"))
		return data

	case NewlineImplicitLF:
		// Replace bare \r with \r\n (but not if already \r\n)
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\r"), []byte("\r\n"))
		return data

	default:
		// Normalize to \n first, then replace
		data = bytes.ReplaceAll(data, []byte("\r\n"), []byte("\n"))
		data = bytes.ReplaceAll(data, []byte("\r"), []byte("\n"))
		replacement := map[NewlineMode][]byte{
			NewlineCR:   {0x0d},
			NewlineLF:   {0x0a},
			NewlineCRLF: {0x0d, 0x0a},
		}[mode]
		if replacement != nil {
			data = bytes.ReplaceAll(data, []byte("\n"), replacement)
		}
		return data
	}
}

// ---- Helpers ----

// unescape processes escape sequences in login script strings
func unescape(s string) string {
	s = regexp.MustCompile(`\\x([0-9a-fA-F]{2})`).ReplaceAllStringFunc(s, func(match string) string {
		var b byte
		fmt.Sscanf(match[2:], "%02x", &b)
		return string(b)
	})
	s = regexp.MustCompile(`\\u([0-9a-fA-F]{4})`).ReplaceAllStringFunc(s, func(match string) string {
		var r rune
		fmt.Sscanf(match[2:], "%04x", &r)
		return string(r)
	})

	// Map of escape sequence -> replacement
	escapeMap := map[byte]string{
		'a': "\x07", 'b': "\x08", 'e': "\x1b", 'f': "\x0c",
		'n': "\x0a", 'r': "\x0d", 't': "\x09", 'v': "\x0b",
	}

	result := make([]byte, 0, len(s))
	i := 0
	for i < len(s) {
		if s[i] == '\\' && i+1 < len(s) {
			next := s[i+1]
			if repl, ok := escapeMap[next]; ok {
				result = append(result, repl...)
			} else {
				result = append(result, next)
			}
			i += 2
		} else {
			result = append(result, s[i])
			i++
		}
	}
	return string(result)
}
