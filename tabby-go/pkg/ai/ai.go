package ai

import (
	"fmt"
	"strings"
)

type Manager struct{}

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
	prompt := strings.ToLower(params.Prompt)

	var command string
	if strings.Contains(prompt, "extract tar") || strings.Contains(prompt, "unzip tar") {
		command = "tar -xvf archive.tar.gz"
	} else if strings.Contains(prompt, "find port") || strings.Contains(prompt, "process on port") {
		command = "lsof -i :<port>"
	} else if strings.Contains(prompt, "kill") && strings.Contains(prompt, "port") {
		command = "kill $(lsof -t -i :<port>)"
	} else if strings.Contains(prompt, "disk space") {
		command = "df -h"
	} else if strings.Contains(prompt, "undo") && strings.Contains(prompt, "commit") {
		command = "git reset --soft HEAD~1"
	} else {
		// Fallback for mock
		command = fmt.Sprintf("echo 'Mock AI command for: %s'", params.Prompt)
	}

	return &GenerateCommandResult{
		Command: command,
	}, nil
}

type ExplainErrorParams struct {
	Command     string `json:"command"`
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
