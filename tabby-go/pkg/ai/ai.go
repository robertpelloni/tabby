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
	errOut := strings.ToLower(params.ErrorOutput)
	var explanation string

	if strings.Contains(errOut, "command not found") {
		explanation = fmt.Sprintf(`### Command Not Found

The command "%s" was not found in your system's PATH.

**Possible solutions:**
1. Check for typos in the command.
2. Install the missing package via your package manager (e.g., "apt install %s", "brew install %s").`, params.Command, params.Command, params.Command)
	} else if strings.Contains(errOut, "permission denied") {
		explanation = fmt.Sprintf(`### Permission Denied

You do not have the required permissions to execute "%s" or access the target file/directory.

**Possible solutions:**
1. Prepend "sudo" to run the command as an administrator.
2. Check the file permissions using "ls -l" and modify them with "chmod" if necessary.`, params.Command)
	} else if strings.Contains(errOut, "syntax error") || strings.Contains(errOut, "parse error") {
		explanation = fmt.Sprintf(`### Syntax Error

There appears to be a syntax error in your command or the script you are trying to execute.

**Possible solutions:**
1. Check for missing quotes, parentheses, or brackets.
2. Verify the correct usage of flags and parameters for "%s".`, params.Command)
	} else if strings.Contains(errOut, "no such file or directory") {
		explanation = fmt.Sprintf(`### Missing File or Directory

The command "%s" attempted to access a file or directory that does not exist.

**Possible solutions:**
1. Double-check the path for typos.
2. Ensure you are in the correct working directory ("pwd").`, params.Command)
	} else if strings.Contains(errOut, "already in use") || strings.Contains(errOut, "address already in use") {
		explanation = fmt.Sprintf(`### Port/Address Already in Use

The command "%s" failed because a network port it needs is already bound by another process.

**Possible solutions:**
1. Identify the process using the port with "lsof -i :<port>" or "netstat -tulpn | grep <port>".
2. Stop the conflicting process (e.g., "kill -9 <PID>").`, params.Command)
	} else {
		explanation = fmt.Sprintf(`### Unknown Error

**Command:** "%s"

**Error Output:**
"""text
%s
"""

I couldn't identify the exact cause of this error. Please review the output above or search online for the specific error message.`, params.Command, params.ErrorOutput)
	}

	return &ExplainErrorResult{
		Explanation: explanation,
	}, nil
}
