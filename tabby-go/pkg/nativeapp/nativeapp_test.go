package nativeapp

import (
	"testing"
)

func TestNewNativeApp(t *testing.T) {
	app := NewNativeApp()
	if app == nil {
		t.Fatal("Expected NewNativeApp() to return non-nil instance")
	}
	if app.sshMgr == nil {
		t.Error("Expected sshMgr to be initialized")
	}
	if app.sftpMgr == nil {
		t.Error("Expected sftpMgr to be initialized")
	}
}

func TestSendNotification(t *testing.T) {
	app := NewNativeApp()
	// Test notifications do not panic
	app.sendNotification("ssh.data", nil)
	app.sendNotification("ssh.exit", nil)
	app.sendNotification("ssh.serviceMessage", nil)
	app.sendNotification("ssh.banner", nil)
	app.sendNotification("ssh.hostKeyPrompt", nil)
	app.sendNotification("ssh.keyboardInteractive", nil)
	app.sendNotification("unknown.event", nil)
}
