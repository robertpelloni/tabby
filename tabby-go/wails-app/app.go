package main

import (
	"context"
	"fmt"

	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
	"github.com/robertpelloni/tabby/tabby-go/pkg/sftp"
	"github.com/robertpelloni/tabby/tabby-go/pkg/pty"
	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	a.sshMgr = ssh.NewManager(a.sendNotification)
	a.sftpMgr = sftp.NewManager(a.sshMgr)
	a.ptyMgr = pty.NewManager(a.sendNotification)
	return a
}

// startup is called when the app starts. The context is saved
// so we can call the runtime methods
func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

// Greet returns a greeting for the given name
func (a *App) Greet(name string) string {
	return fmt.Sprintf("Hello %s, It's show time!", name)
}

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

// sendNotification sends a notification to the frontend
func (a *App) sendNotification(method string, params interface{}) {
	if a.ctx != nil {
		runtime.EventsEmit(a.ctx, method, params)
	}
}
