package ai

import "fmt"

type Manager struct {}

func NewManager() *Manager {
	return &Manager{}
}

type GenerateCommandParams struct {
	Prompt string `json:"prompt"`
}

type GenerateCommandResult struct {
	Command string `json:"command"`
}

func (m *Manager) GenerateCommand(params GenerateCommandParams) (*GenerateCommandResult, error) {
	// Stub implementation for now until real LLM integration is ready.
	// We'll just echo back a placeholder.
	return &GenerateCommandResult{
		Command: fmt.Sprintf("echo 'AI generated command for: %s'", params.Prompt),
	}, nil
}

type ExplainErrorParams struct {
	Command string `json:"command"`
	ErrorOutput string `json:"errorOutput"`
}

type ExplainErrorResult struct {
	Explanation string `json:"explanation"`
}

func (m *Manager) ExplainError(params ExplainErrorParams) (*ExplainErrorResult, error) {
	return &ExplainErrorResult{
		Explanation: fmt.Sprintf("AI Explanation: It looks like the command `%s` failed because of: %s", params.Command, params.ErrorOutput),
	}, nil
}
