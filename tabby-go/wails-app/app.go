package main

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"runtime"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/colorscheme"
	"github.com/robertpelloni/tabby/tabby-go/pkg/profile"
	"github.com/robertpelloni/tabby/tabby-go/pkg/serial"
	"github.com/robertpelloni/tabby/tabby-go/pkg/session"
	"github.com/robertpelloni/tabby/tabby-go/pkg/pty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
	"github.com/robertpelloni/tabby/tabby-go/pkg/sftp"
	"github.com/robertpelloni/tabby/tabby-go/pkg/settings"
	"github.com/robertpelloni/tabby/tabby-go/pkg/telnet"
	wailsRuntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// App struct holds all the application state and backend managers.
type App struct {
	ctx     context.Context
	sshMgr   *ssh.Manager
	sftpMgr  *sftp.Manager
	ptyMgr   *pty.Manager
	serialMgr *serial.Manager
	telnetMgr *telnet.Manager
}

// NewApp creates a new App application struct with all managers initialized.
func NewApp() *App {
	a := &App{}
	a.ptyMgr = pty.NewManager(a.emit)
	a.sshMgr = ssh.NewManager(a.emit)
	a.sftpMgr = sftp.NewManager(a.sshMgr)
	a.serialMgr = serial.NewManager(a.emit)
	a.telnetMgr = telnet.NewManager(a.emit)
	return a
}

// startup is called when the Wails app starts. It stores the context for later use.
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// ==== PTY Methods ====

// PTYSpawn spawns a local pseudo-terminal process.
func (a *App) PTYSpawn(params api.PTYSpawnParams) (*api.PTYSpawnResult, error) {
	return a.ptyMgr.Spawn(params)
}

// PTYWrite writes base64-encoded data to a PTY process's stdin.
func (a *App) PTYWrite(id string, data string) error {
	return a.ptyMgr.Write(id, data)
}

// PTYResize resizes a PTY process's terminal dimensions.
func (a *App) PTYResize(id string, columns, rows int) error {
	return a.ptyMgr.Resize(id, columns, rows)
}

// PTYKill sends a kill signal to a PTY process.
func (a *App) PTYKill(id string, signal string) error {
	return a.ptyMgr.Kill(id, signal)
}

// ==== SSH Methods ====

// SSHConnect establishes an SSH connection to a remote server.
func (a *App) SSHConnect(params api.SSHConnectParams) (*api.SSHConnectionResult, error) {
	return a.sshMgr.Connect(params)
}

// SSHStartShell starts an interactive shell session over an existing SSH connection.
func (a *App) SSHStartShell(params api.SSHSessionParams) (*api.SSHSessionResult, error) {
	return a.sshMgr.StartShell(params)
}

// SSHWrite writes base64-encoded data to an SSH shell session's stdin.
func (a *App) SSHWrite(params api.SSHWriteParams) error {
	return a.sshMgr.Write(params)
}

// SSHResize resizes an SSH shell session.
func (a *App) SSHResize(params api.SSHResizeParams) error {
	return a.sshMgr.Resize(params)
}

// SSHClose closes an SSH connection or session.
func (a *App) SSHClose(params api.SSHCloseParams) error {
	return a.sshMgr.Close(params)
}

// SSHAddForward adds a port forward to an active SSH connection.
func (a *App) SSHAddForward(params api.PortForwardParams) (*api.PortForwardResult, error) {
	return a.sshMgr.AddForward(params)
}

// SSHRemoveForward removes a port forward.
func (a *App) SSHRemoveForward(params api.PortForwardRemoveParams) error {
	return a.sshMgr.RemoveForward(params)
}

// SSHListForwards lists active port forwards for a connection.
func (a *App) SSHListForwards(connectionID string) []api.PortForwardInfo {
	return a.sshMgr.ListForwards(connectionID)
}

// ==== SFTP Methods ====

// SFTPOpen opens an SFTP session over an existing SSH connection.
func (a *App) SFTPOpen(params api.SFTPOpenParams) (*api.SFTPOpenResult, error) {
	return a.sftpMgr.Open(params)
}

// SFTPList lists files in a directory on the remote server.
func (a *App) SFTPList(params api.SFTPListParams) ([]api.SFTPFile, error) {
	return a.sftpMgr.List(params)
}

// SFTPDownload downloads a file from the remote server.
func (a *App) SFTPDownload(params api.SFTPDownloadParams) (*api.SFTPTransferResult, error) {
	return a.sftpMgr.Download(params)
}

// SFTPUpload uploads a file to the remote server.
func (a *App) SFTPUpload(params api.SFTPUploadParams) (*api.SFTPTransferResult, error) {
	return a.sftpMgr.Upload(params)
}

// SFTPChmod changes file permissions on the remote server.
func (a *App) SFTPChmod(sessionID, filePath string, mode uint32) error {
	return a.sftpMgr.Chmod(sessionID, filePath, mode)
}

// SFTPReadlink reads a symbolic link on the remote server.
func (a *App) SFTPReadlink(sessionID, linkPath string) (string, error) {
	return a.sftpMgr.Readlink(sessionID, linkPath)
}

// SFTPSymlink creates a symbolic link on the remote server.
func (a *App) SFTPSymlink(sessionID, oldPath, newPath string) error {
	return a.sftpMgr.Symlink(sessionID, oldPath, newPath)
}

// SFTPDelete deletes a file on the remote server.
func (a *App) SFTPDelete(sessionID, filePath string) error {
	return a.sftpMgr.Delete(sessionID, filePath)
}

// SFTPRename renames a file on the remote server.
func (a *App) SFTPRename(sessionID, oldPath, newPath string) error {
	return a.sftpMgr.Rename(sessionID, oldPath, newPath)
}

// SFTPMkdir creates a directory on the remote server.
func (a *App) SFTPMkdir(sessionID, dirPath string) error {
	return a.sftpMgr.Mkdir(sessionID, dirPath)
}

// SFTPStat gets file info on the remote server.
func (a *App) SFTPStat(sessionID, filePath string) (*api.SFTPFile, error) {
	return a.sftpMgr.Stat(sessionID, filePath)
}

// SFTPClose closes an SFTP session.
func (a *App) SFTPClose(sessionID string) error {
	return a.sftpMgr.Close(sessionID)
}

// SFTPRmdir removes a directory on the remote server.
func (a *App) SFTPRmdir(sessionID, dirPath string) error {
	return a.sftpMgr.Rmdir(sessionID, dirPath)
}

// SFTPReadDir reads a directory listing on the remote server.
func (a *App) SFTPReadDir(sessionID, dirPath string) ([]api.SFTPFile, error) {
	return a.sftpMgr.ReadDir(sessionID, dirPath)
}

// SFTPMkdirAll creates a directory and all parent directories.
func (a *App) SFTPMkdirAll(sessionID, dirPath string) error {
	return a.sftpMgr.MkdirAll(sessionID, dirPath)
}

// ==== System Methods ====

// GetDefaultShell returns the default shell for the current OS.
func (a *App) GetDefaultShell() string {
	if runtime.GOOS == "windows" {
		if path, err := exec.LookPath("powershell.exe"); err == nil {
			return path
		}
		if path, err := exec.LookPath("pwsh.exe"); err == nil {
			return path
		}
		return "cmd.exe"
	}
	if shell := os.Getenv("SHELL"); shell != "" {
		return shell
	}
	return "/bin/bash"
}

// GetAvailableShells returns a list of all installed shells on the system.
func (a *App) GetAvailableShells() []string {
	var shells []string
	if runtime.GOOS == "windows" {
		candidates := []string{"powershell.exe", "pwsh.exe", "cmd.exe", "bash.exe", "wsl.exe"}
		for _, c := range candidates {
			if path, err := exec.LookPath(c); err == nil {
				shells = append(shells, path)
			}
		}
	} else {
		candidates := []string{"/bin/bash", "/bin/zsh", "/bin/fish", "/bin/sh", "/usr/local/bin/fish", "/opt/homebrew/bin/fish"}
		for _, c := range candidates {
			if _, err := os.Stat(c); err == nil {
				shells = append(shells, c)
			}
		}
		if shell := os.Getenv("SHELL"); shell != "" {
			found := false
			for _, s := range shells {
				if s == shell { found = true; break }
			}
			if !found { shells = append([]string{shell}, shells...) }
		}
	}
	return shells
}

// GetHomeDir returns the user's home directory.
func (a *App) GetHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}

// GetHostname returns the machine hostname.
func (a *App) GetHostname() string {
	hostname, _ := os.Hostname()
	return hostname
}

// GetUsername returns the current OS username.
func (a *App) GetUsername() string {
	if runtime.GOOS == "windows" {
		return os.Getenv("USERNAME")
	}
	return os.Getenv("USER")
}

// GetPlatform returns OS platform info as a map.
func (a *App) GetPlatform() map[string]string {
	return map[string]string{
		"os":      runtime.GOOS,
		"arch":    runtime.GOARCH,
		"version": "1.0.0",
	}
}

// OpenInBrowser opens a URL in the default web browser.
func (a *App) OpenInBrowser(url string) {
	if a.ctx != nil {
		wailsRuntime.BrowserOpenURL(a.ctx, url)
	}
}

// SelectDirectory opens a native directory picker dialog.
func (a *App) SelectDirectory(title string) string {
	if a.ctx != nil {
		path, _ := wailsRuntime.OpenDirectoryDialog(a.ctx, wailsRuntime.OpenDialogOptions{Title: title})
		return path
	}
	return ""
}

// SetWindowTitle changes the application window title.
func (a *App) SetWindowTitle(title string) {
	if a.ctx != nil {
		wailsRuntime.WindowSetTitle(a.ctx, title)
	}
}

// ==== Color Schemes ====

// GetColorSchemes returns all built-in color schemes.
func (a *App) GetColorSchemes() []colorscheme.ColorScheme {
	return colorscheme.BuiltInSchemes
}

// GetColorSchemeNames returns just the names of all built-in color schemes.
func (a *App) GetColorSchemeNames() []string {
	return colorscheme.GetSchemeNames()
}

// GetColorScheme returns a single color scheme by name.
func (a *App) GetColorScheme(name string) *colorscheme.ColorScheme {
	return colorscheme.GetBuiltInScheme(name)
}

// ==== Settings ====

// GetSettings returns the current user settings loaded from disk.
func (a *App) GetSettings() (settings.Settings, error) {
	return settings.LoadSettings()
}

// SaveSettings persists user settings to disk.
func (a *App) SaveSettings(s settings.Settings) error {
	return settings.SaveSettings(s)
}

// ResetSettings resets all settings to defaults and saves them.
func (a *App) ResetSettings() error {
	return settings.ResetSettings()
}

// GetDefaultShellPreferred returns the user's configured shell from settings,
// or auto-detects if none is configured.
func (a *App) GetDefaultShellPreferred() (string, error) {
	s, err := a.GetSettings()
	if err != nil {
		return "", err
	}
	if s.Shell != "" {
		return s.Shell, nil
	}
	return a.GetDefaultShell(), nil
}

// ==== Connection Profiles ====

// GetProfiles returns all saved connection profiles.
func (a *App) GetProfiles() ([]profile.ConnectionProfile, error) {
	return profile.LoadProfiles()
}

// SaveProfiles persists connection profiles to disk.
func (a *App) SaveProfiles(profiles []profile.ConnectionProfile) error {
	return profile.SaveProfiles(profiles)
}

// ==== Session Persistence ====

// SaveSessionState persists which tabs were open so they can be restored on restart.
func (a *App) SaveSessionState(tabs []session.TabState) error {
	return session.SaveSession(tabs)
}

// LoadSessionState reads the last saved session state from disk.
func (a *App) LoadSessionState() (*session.SessionState, error) {
	return session.LoadSession()
}

// ClearSessionState removes the saved session file.
func (a *App) ClearSessionState() error {
	return session.ClearSession()
}

// ==== Telnet Methods ====

// TelnetConnect establishes a Telnet connection.
func (a *App) TelnetConnect(host string, port int) (*telnet.TelnetConnectResult, error) {
	return a.telnetMgr.Connect(telnet.TelnetConnectParams{Host: host, Port: port})
}

// TelnetWrite sends base64-encoded data to a Telnet connection.
func (a *App) TelnetWrite(connectionID string, data string) error {
	return a.telnetMgr.Write(connectionID, data)
}

// TelnetResize resizes the terminal for NAWS support.
func (a *App) TelnetResize(connectionID string, width, height int) error {
	return a.telnetMgr.Resize(connectionID, width, height)
}

// TelnetClose closes a Telnet connection.
func (a *App) TelnetClose(connectionID string) error {
	return a.telnetMgr.Close(connectionID)
}

// ==== Serial Port Methods ====

// SerialOpen opens a serial port connection.
func (a *App) SerialOpen(params api.SerialOpenParams) (*api.SerialOpenResult, error) {
	return a.serialMgr.Open(params)
}

// SerialWrite writes base64-encoded data to a serial port.
func (a *App) SerialWrite(id string, data string) error {
	return a.serialMgr.Write(id, data)
}

// SerialClose closes a serial port connection.
func (a *App) SerialClose(id string) error {
	return a.serialMgr.Close(id)
}

// SerialListPorts returns a list of available serial ports on the system.
func (a *App) SerialListPorts() ([]api.SerialPortInfo, error) {
	return a.serialMgr.ListPorts()
}

// ==== SSH Config Import ====

// ImportSSHConfig reads ~/.ssh/config and returns parsed host entries.
func (a *App) ImportSSHConfig() ([]profile.ConnectionProfile, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("cannot determine home directory: %w", err)
	}
	configPath := home + "/.ssh/config"
	// Also check Windows OpenSSH path
	if runtime.GOOS == "windows" {
		alt := home + "/.ssh/config"
		if _, err := os.Stat(alt); err != nil {
			configPath = home + "/.ssh/config"
		}
	}
	profiles, err := profile.ImportSSHConfigAsProfiles(configPath)
	if err != nil {
		return nil, fmt.Errorf("failed to import SSH config: %w", err)
	}
	return profiles, nil
}

// ==== Internal ====

// emit sends a named event with parameters to the frontend via Wails events.
func (a *App) emit(method string, params interface{}) {
	if a.ctx != nil {
		wailsRuntime.EventsEmit(a.ctx, method, params)
	}
}
