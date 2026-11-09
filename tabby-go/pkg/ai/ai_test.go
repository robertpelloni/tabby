package ai

import (
	"strings"
	"testing"
)

func TestNewManager(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Expected NewManager to return non-nil")
	}
}

func TestGenerateCommand(t *testing.T) {
	mgr := NewManager()

	tests := []struct {
		prompt   string
		expected string
	}{
		{"extract tar file", "tar -xvf archive.tar.gz"},
		{"find port 8080", "lsof -i :<port>"},
		{"kill process on port 3000", "kill $(lsof -t -i :<port>)"},
		{"check disk space", "df -h"},
		{"undo commit", "git reset --soft HEAD~1"},
	}

	for _, tt := range tests {
		res, err := mgr.GenerateCommand(GenerateCommandParams{Prompt: tt.prompt})
		if err != nil {
			t.Fatalf("GenerateCommand failed for prompt %q: %v", tt.prompt, err)
		}
		if res.Command != tt.expected {
			t.Errorf("For prompt %q expected command %q, got %q", tt.prompt, tt.expected, res.Command)
		}
	}
}

func TestExplainError(t *testing.T) {
	mgr := NewManager()

	tests := []struct {
		cmd      string
		errOut   string
		contains string
	}{
		{"foo", "bash: foo: command not found", "Command Not Found"},
		{"cat /root/secret", "permission denied", "Permission Denied"},
		{"sh script.sh", "syntax error near unexpected token", "Syntax Error"},
		{"cd /invalid", "no such file or directory", "Missing File or Directory"},
		{"listen :80", "address already in use", "Port/Address Already in Use"},
		{"mycmd", "unrecognized crash code 42", "Unknown Error"},
	}

	for _, tt := range tests {
		res, err := mgr.ExplainError(ExplainErrorParams{Command: tt.cmd, ErrorOutput: tt.errOut})
		if err != nil {
			t.Fatalf("ExplainError failed for cmd %q: %v", tt.cmd, err)
		}
		if !strings.Contains(res.Explanation, tt.contains) {
			t.Errorf("Expected explanation to contain %q, got: %s", tt.contains, res.Explanation)
		}
	}
}

func TestChat(t *testing.T) {
	mgr := NewManager()

	res, err := mgr.Chat(ChatParams{Message: "hello there"})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if !strings.Contains(res.Response, "Hello!") {
		t.Errorf("Expected greeting, got: %s", res.Response)
	}

	res, err = mgr.Chat(ChatParams{Message: "can you help me?"})
	if err != nil {
		t.Fatalf("Chat failed: %v", err)
	}
	if !strings.Contains(res.Response, "help") {
		t.Errorf("Expected help info, got: %s", res.Response)
	}
}
