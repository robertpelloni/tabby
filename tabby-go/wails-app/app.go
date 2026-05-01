package main

import (
	"context"
	"os"
	"os/exec"
	"runtime"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/pty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
	"github.com/robertpelloni/tabby/tabby-go/pkg/sftp"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct
type App struct {
	ctx     context.Context
	sshMgr  *ssh.Manager
	sftpMgr *sftp.Manager
	ptyMgr  *pty.Manager
}

// NewApp creates a new App application struct
func NewApp() *App {
	a := &App{}
	a.ptyMgr = pty.NewManager(a.emit)
	a.sshMgr = ssh.NewManager(a.emit)
	a.sftpMgr = sftp.NewManager(a.sshMgr)
	return a
}

// startup is called when the app starts
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ---- PTY Methods ----

// PTYSpawn spawns a local PTY
func (a *App) PTYSpawn(params api.PTYSpawnParams) (*api.PTYSpawnResult, error) {
	return a.ptyMgr.Spawn(params)
}

// PTYWrite writes data to a PTY
func (a *App) PTYWrite(id string, data string) error {
	return a.ptyMgr.Write(id, data)
}

// PTYResize resizes a PTY
func (a *App) PTYResize(id string, columns, rows int) error {
	return a.ptyMgr.Resize(id, columns, rows)
}

// PTYKill kills a PTY
func (a *App) PTYKill(id string, signal string) error {
	return a.ptyMgr.Kill(id, signal)
}

// ---- SSH Methods ----

// SSHConnect connects to an SSH server
func (a *App) SSHConnect(params api.SSHConnectParams) (*api.SSHConnectionResult, error) {
	return a.sshMgr.Connect(params)
}

// SSHStartShell starts a shell session
func (a *App) SSHStartShell(params api.SSHSessionParams) (*api.SSHSessionResult, error) {
	return a.sshMgr.StartShell(params)
}

// SSHWrite writes data to a session
func (a *App) SSHWrite(params api.SSHWriteParams) error {
	return a.sshMgr.Write(params)
}

// ---- System Methods ----

// GetDefaultShell returns the default shell for the current OS
func (a *App) GetDefaultShell() string {
	if runtime.GOOS == "windows" {
		// Prefer PowerShell, fall back to cmd
		if path, err := exec.LookPath("powershell.exe"); err == nil {
			return path
		}
		if path, err := exec.LookPath("pwsh.exe"); err == nil {
			return path
		}
		return "cmd.exe"
	}
	// Unix: check SHELL env, then fall back
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}

// GetAvailableShells returns a list of available shells
func (a *App) GetAvailableShells() []string {
	var shells []string
	if runtime.GOOS == "windows" {
		candidates := []string{
			"powershell.exe", "pwsh.exe",
			"cmd.exe",
			"bash.exe", "wsl.exe",
		}
		for _, c := range candidates {
			if path, err := exec.LookPath(c); err == nil {
				shells = append(shells, path)
			}
		}
	} else {
		candidates := []string{
			"/bin/bash", "/bin/zsh", "/bin/fish",
			"/bin/sh", "/usr/local/bin/fish",
			"/opt/homebrew/bin/fish",
		}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				shells = append(shells, c)
			}
		}
		// Also check SHELL env
		if shell := os.Getenv("SHELL"); shell != "" {
			found := false
			for _, s := range shells {
				if s == shell {
					found = true
					break
				}
			}
			if !found {
				shells = append([]string{shell}, shells...)
			}
		}
	}
	return shells
}

// GetHomeDir returns the user's home directory
func (a *App) GetHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// GetHostname returns the machine hostname
func (a *App) GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

// GetUsername returns the current username
func (a *App) GetUsername() string {
	return os.Getenv("USER")
}

// GetPlatform returns OS platform info
func (a *App) GetPlatform() map[string]string {
	return map[string]string{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": "1.0.0",
	}
}

// OpenInBrowser opens a URL in the default browser
func (a *App) OpenInBrowser(url string) {
	if a.ctx != nil {
		wailsRuntime.BrowserOpenURL(a.ctx, url)
	}
}

// SelectDirectory opens a directory picker dialog
func (a *App) SelectDirectory(title string) string {
	if a.ctx != nil {
		path, _ := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{
			Title: title,
		})
		return path
	}
	return ""
}

// SetWindowTitle changes the window title
func (a *App) SetWindowTitle(title string) {
	if a.ctx != nil {
		wailsRuntime.WindowSetTitle(a.ctx, title)
	}
}

// ---- Internal ----

// emit sends a notification to the frontend via Wails events
func (a *App) emit(method string, params interface{}) {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, method, params)
	}
}

