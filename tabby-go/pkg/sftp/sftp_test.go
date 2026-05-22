package sftp

import (
	"testing"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
)

// TestManagerCreation verifies the SFTP manager is created correctly
func TestManagerCreation(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

// TestOpenNonexistentConnection verifies opening SFTP on non-existent connection fails
func TestOpenNonexistentConnection(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.Open(api.SFTPOpenParams{ConnectionID: "nonexistent"})
	if err == nil {
		t.Error("Should fail opening SFTP on non-existent connection")
	}
}

// TestListNonexistentSession verifies listing with non-existent session fails
func TestListNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.List(api.SFTPListParams{SessionID: "nonexistent", Path: "/"})
	if err == nil {
		t.Error("Should fail listing non-existent session")
	}
}

// TestDownloadNonexistentSession verifies download with non-existent session fails
func TestDownloadNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.Download(api.SFTPDownloadParams{
		SessionID:  "nonexistent",
		RemotePath: "/tmp/test",
		LocalPath:  "/tmp/test",
	})
	if err == nil {
		t.Error("Should fail downloading from non-existent session")
	}
}

// TestUploadNonexistentSession verifies upload with non-existent session fails
func TestUploadNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.Upload(api.SFTPUploadParams{
		SessionID:  "nonexistent",
		LocalPath:  "/tmp/test",
		RemotePath: "/tmp/test",
	})
	if err == nil {
		t.Error("Should fail uploading to non-existent session")
	}
}

// TestDeleteNonexistentSession verifies delete with non-existent session fails
func TestDeleteNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Delete("nonexistent", "/tmp/test")
	if err == nil {
		t.Error("Should fail deleting from non-existent session")
	}
}

// TestRenameNonexistentSession verifies rename with non-existent session fails
func TestRenameNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Rename("nonexistent", "/old", "/new")
	if err == nil {
		t.Error("Should fail renaming in non-existent session")
	}
}

// TestMkdirNonexistentSession verifies mkdir with non-existent session fails
func TestMkdirNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Mkdir("nonexistent", "/tmp/newdir")
	if err == nil {
		t.Error("Should fail mkdir in non-existent session")
	}
}

// TestMkdirAllNonexistentSession verifies mkdirAll with non-existent session fails
func TestMkdirAllNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.MkdirAll("nonexistent", "/tmp/a/b/c")
	if err == nil {
		t.Error("Should fail mkdirAll in non-existent session")
	}
}

// TestStatNonexistentSession verifies stat with non-existent session fails
func TestStatNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.Stat("nonexistent", "/tmp/test")
	if err == nil {
		t.Error("Should fail stat in non-existent session")
	}
}

// TestLstatNonexistentSession verifies lstat with non-existent session fails
func TestLstatNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.Lstat("nonexistent", "/tmp/test")
	if err == nil {
		t.Error("Should fail lstat in non-existent session")
	}
}

// TestChmodNonexistentSession verifies chmod with non-existent session fails
func TestChmodNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Chmod("nonexistent", "/tmp/test", 0755)
	if err == nil {
		t.Error("Should fail chmod in non-existent session")
	}
}

// TestReadlinkNonexistentSession verifies readlink with non-existent session fails
func TestReadlinkNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.Readlink("nonexistent", "/tmp/link")
	if err == nil {
		t.Error("Should fail readlink in non-existent session")
	}
}

// TestSymlinkNonexistentSession verifies symlink with non-existent session fails
func TestSymlinkNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Symlink("nonexistent", "/tmp/target", "/tmp/link")
	if err == nil {
		t.Error("Should fail symlink in non-existent session")
	}
}

// TestRmdirNonexistentSession verifies rmdir with non-existent session fails
func TestRmdirNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Rmdir("nonexistent", "/tmp/dir")
	if err == nil {
		t.Error("Should fail rmdir in non-existent session")
	}
}

// TestReadDirNonexistentSession verifies readDir with non-existent session fails
func TestReadDirNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	_, err := mgr.ReadDir("nonexistent", "/tmp")
	if err == nil {
		t.Error("Should fail readDir in non-existent session")
	}
}

// TestCloseNonexistentSession verifies closing non-existent session fails
func TestCloseNonexistentSession(t *testing.T) {
	sshMgr := ssh.NewManager(func(method string, params interface{}) {})
	mgr := NewManager(sshMgr, nil)

	err := mgr.Close("nonexistent")
	if err == nil {
		t.Error("Should fail closing non-existent session")
	}
}
