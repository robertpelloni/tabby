package middleware

import (
	"bytes"
	"testing"
)

// ---- UTF8 Splitter Tests ----

func TestUTF8SplitterASCII(t *testing.T) {
	u := &UTF8Splitter{}
	result := u.Write([]byte("Hello World"))
	if string(result) != "Hello World" {
		t.Errorf("Expected 'Hello World', got %q", string(result))
	}
}

func TestUTF8SplitterMultibyte(t *testing.T) {
	u := &UTF8Splitter{}
	input := []byte("Hello 世界")
	result := u.Write(input)
	if string(result) != string(input) {
		t.Errorf("Expected %q, got %q", string(input), string(result))
	}
}

func TestUTF8SplitterIncomplete(t *testing.T) {
	u := &UTF8Splitter{}

	// Send first byte of a 3-byte UTF-8 character (世 = E4 B8 96)
	result := u.Write([]byte{0xe4})
	if len(result) != 0 {
		t.Errorf("Expected empty result for incomplete sequence, got %v", result)
	}

	// Send second byte
	result = u.Write([]byte{0xb8})
	if len(result) != 0 {
		t.Errorf("Expected empty result for incomplete sequence, got %v", result)
	}

	// Send final byte - should emit the complete character
	result = u.Write([]byte{0x96})
	if len(result) != 3 || !bytes.Equal(result, []byte{0xe4, 0xb8, 0x96}) {
		t.Errorf("Expected complete UTF-8 character, got %v", result)
	}
}

func TestUTF8SplitterFlush(t *testing.T) {
	u := &UTF8Splitter{}
	u.Write([]byte{0xe4, 0xb8}) // Incomplete
	result := u.Flush()
	if len(result) != 2 {
		t.Errorf("Flush should return pending bytes, got %d bytes", len(result))
	}
}

func TestUTF8SplitterMixed(t *testing.T) {
	u := &UTF8Splitter{}
	// ASCII + incomplete multibyte + more ASCII
	result1 := u.Write([]byte("A"))
	if string(result1) != "A" {
		t.Errorf("Expected 'A', got %q", string(result1))
	}
	result2 := u.Write([]byte{0xe4}) // Start of 世
	if len(result2) != 0 {
		t.Errorf("Expected empty for incomplete, got %v", result2)
	}
	result3 := u.Write([]byte{0xb8, 0x96, "B"[0]}) // Complete 世 + B
	if len(result3) != 4 {
		t.Errorf("Expected 4 bytes (世+B), got %d", len(result3))
	}
}

// ---- Input Processor Tests ----

func TestInputProcessorNoChange(t *testing.T) {
	p := NewInputProcessor(InputProcessingOptions{Backspace: BackspaceCtrlH})
	result := p.Process([]byte("hello"))
	if string(result) != "hello" {
		t.Errorf("Expected 'hello', got %q", string(result))
	}
}

func TestInputProcessorBackspaceCtrlH(t *testing.T) {
	p := NewInputProcessor(InputProcessingOptions{Backspace: BackspaceCtrlH})
	result := p.Process([]byte{0x7f})
	if len(result) != 1 || result[0] != 0x08 {
		t.Errorf("Expected 0x08 for ctrl-h backspace, got %v", result)
	}
}

func TestInputProcessorBackspaceDelete(t *testing.T) {
	p := NewInputProcessor(InputProcessingOptions{Backspace: BackspaceDelete})
	result := p.Process([]byte{0x7f})
	expected := []byte{0x1b, '[', '3', '~'}
	if !bytes.Equal(result, expected) {
		t.Errorf("Expected delete sequence, got %v", result)
	}
}

func TestInputProcessorBackspaceDefault(t *testing.T) {
	p := NewInputProcessor(InputProcessingOptions{Backspace: BackspaceBackspace})
	result := p.Process([]byte{0x7f})
	if len(result) != 1 || result[0] != 0x7f {
		t.Errorf("Expected 0x7f for default backspace, got %v", result)
	}
}

func TestInputProcessorMultiByte(t *testing.T) {
	p := NewInputProcessor(InputProcessingOptions{Backspace: BackspaceCtrlH})
	result := p.Process([]byte("ab"))
	if string(result) != "ab" {
		t.Errorf("Multi-byte input should pass through, got %q", string(result))
	}
}

// ---- Login Script Processor Tests ----

func TestLoginScriptSimpleMatch(t *testing.T) {
	var sent [][]byte
	p := NewLoginScriptProcessor([]LoginScript{
		{Expect: "Password:", Send: "secret"},
	}, func(d []byte) { sent = append(sent, d) }, nil)

	p.ExecuteUnconditionalScripts()
	if len(sent) != 0 {
		t.Error("No unconditional scripts, nothing should be sent")
	}

	p.ProcessOutput([]byte("Please enter Password:"))
	if len(sent) != 1 || string(sent[0]) != "secret\n" {
		t.Errorf("Expected 'secret\\n', got %q", sent)
	}
}

func TestLoginScriptNoMatch(t *testing.T) {
	var sent [][]byte
	p := NewLoginScriptProcessor([]LoginScript{
		{Expect: "Password:", Send: "secret"},
	}, func(d []byte) { sent = append(sent, d) }, nil)

	p.ProcessOutput([]byte("Login successful"))
	if len(sent) != 0 {
		t.Error("No match, nothing should be sent")
	}
}

func TestLoginScriptRegex(t *testing.T) {
	var sent [][]byte
	p := NewLoginScriptProcessor([]LoginScript{
		{Expect: "Pass\\w+:", Send: "secret", IsRegex: true},
	}, func(d []byte) { sent = append(sent, d) }, nil)

	p.ProcessOutput([]byte("Enter Passphrase:"))
	if len(sent) != 1 || string(sent[0]) != "secret\n" {
		t.Errorf("Expected regex match, got %q", sent)
	}
}

func TestLoginScriptUnconditional(t *testing.T) {
	var sent [][]byte
	p := NewLoginScriptProcessor([]LoginScript{
		{Expect: "", Send: "unconditional"},
		{Expect: "Password:", Send: "secret"},
	}, func(d []byte) { sent = append(sent, d) }, nil)

	p.ExecuteUnconditionalScripts()
	if len(sent) != 1 || string(sent[0]) != "unconditional\n" {
		t.Errorf("Expected unconditional script, got %q", sent)
	}
}

// ---- OSC Processor Tests ----

func TestOSCProcessorCurrentDir(t *testing.T) {
	var cwd string
	p := NewOSCProcessor("/home/user", func(c string) { cwd = c })

	data := []byte("\x1b]1337;CurrentDir=/home/user/projects\x07some text")
	result := p.ProcessOutput(data)

	if cwd != "/home/user/projects" {
		t.Errorf("Expected '/home/user/projects', got %q", cwd)
	}

	// Data should be returned unchanged
	if !bytes.Equal(result, data) {
		t.Error("OSC processor should return data unchanged")
	}
}

func TestOSCProcessorTildeExpansion(t *testing.T) {
	var cwd string
	p := NewOSCProcessor("/home/user", func(c string) { cwd = c })

	data := []byte("\x1b]1337;CurrentDir=~/projects\x07")
	p.ProcessOutput(data)

	if cwd != "/home/user/projects" {
		t.Errorf("Expected tilde expansion, got %q", cwd)
	}
}

func TestOSCProcessorNoOSC(t *testing.T) {
	var cwd string
	p := NewOSCProcessor("/home/user", func(c string) { cwd = c })

	data := []byte("Hello World")
	result := p.ProcessOutput(data)

	if cwd != "" {
		t.Error("Should not detect CWD without OSC sequence")
	}
	if !bytes.Equal(result, data) {
		t.Error("Data should pass through unchanged")
	}
}

// ---- Stream Processor Tests ----

func TestStreamProcessorNoChange(t *testing.T) {
	p := NewStreamProcessor(StreamProcessingOptions{})
	data := []byte("hello\r\nworld")
	result := p.ProcessOutput(data)
	if !bytes.Equal(result, data) {
		t.Errorf("Expected unchanged data, got %q", result)
	}
}

func TestStreamProcessorOutputNewlines(t *testing.T) {
	p := NewStreamProcessor(StreamProcessingOptions{OutputNewlines: NewlineCRLF})

	tests := []struct {
		input    string
		expected string
	}{
		{"hello\nworld", "hello\r\nworld"},
		{"hello\r\nworld", "hello\r\nworld"},
		{"hello\rworld", "hello\r\nworld"},
	}

	for _, tt := range tests {
		result := p.ProcessOutput([]byte(tt.input))
		if string(result) != tt.expected {
			t.Errorf("Input %q: expected %q, got %q", tt.input, tt.expected, string(result))
		}
	}
}

func TestStreamProcessorInputNewlines(t *testing.T) {
	p := NewStreamProcessor(StreamProcessingOptions{InputNewlines: NewlineLF})
	result := p.ProcessInput([]byte("hello\r\nworld"))
	if string(result) != "hello\nworld" {
		t.Errorf("Expected LF-only, got %q", string(result))
	}
}

func TestStreamProcessorImplicitCR(t *testing.T) {
	p := NewStreamProcessor(StreamProcessingOptions{OutputNewlines: NewlineImplicitCR})
	result := p.ProcessOutput([]byte("hello\nworld"))
	if string(result) != "hello\r\nworld" {
		t.Errorf("Expected CRLF, got %q", string(result))
	}
}

// ---- Unescape Tests ----

func TestUnescapeHex(t *testing.T) {
	result := unescape(`\x41\x42`)
	if result != "AB" {
		t.Errorf("Expected 'AB', got %q", result)
	}
}

func TestUnescapeUnicode(t *testing.T) {
	result := unescape(`\u0041`)
	if result != "A" {
		t.Errorf("Expected 'A', got %q", result)
	}
}

func TestUnescapeSpecialChars(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`\n`, "\n"},
		{`\r`, "\r"},
		{`\t`, "\t"},
		{`\\\\`, `\\`}, // \\\\\\ -> \\
	}
	for _, tt := range tests {
		result := unescape(tt.input)
		if result != tt.expected {
			t.Errorf("unescape(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}
