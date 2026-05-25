// Package sftp implements SFTP file transfer over SSH connections.
//
// It provides file listing, upload, download, delete, rename, and
// directory creation — everything needed for the SFTP file manager UI.
package sftp

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	"github.com/robertpelloni/tabby/tabby-go/pkg/api"
	"github.com/robertpelloni/tabby/tabby-go/pkg/ssh"
	sftpPkg "github.com/pkg/sftp"
)

// Manager manages SFTP sessions
type Manager struct {
	sshMgr   *ssh.Manager
	sessions map[string]*Session
	mu       sync.RWMutex
}

// Session represents an active SFTP session
type Session struct {
	ID           string
	ConnectionID string
	Client       *sftpPkg.Client
}

// NewManager creates a new SFTP session manager
func NewManager(sshMgr *ssh.Manager) *Manager {
	return &Manager{
		sshMgr:   sshMgr,
		sessions: make(map[string]*Session),
	}
}

// Open opens an SFTP session over an existing SSH connection
func (m *Manager) Open(params api.SFTPOpenParams) (*api.SFTPOpenResult, error) {
	sshClient, err := m.sshMgr.GetConnection(params.ConnectionID)
	if err != nil {
		return nil, err
	}

	sftpClient, err := sftpPkg.NewClient(sshClient)
	if err != nil {
		return nil, fmt.Errorf("failed to open SFTP session: %w", err)
	}

	sessionID := fmt.Sprintf("sftp-%d", time.Now().UnixMilli())
	sess := &Session{
		ID:           sessionID,
		ConnectionID: params.ConnectionID,
		Client:       sftpClient,
	}

	m.mu.Lock()
	m.sessions[sessionID] = sess
	m.mu.Unlock()

	return &api.SFTPOpenResult{SessionID: sessionID}, nil
}

// List lists files in a directory
func (m *Manager) List(params api.SFTPListParams) ([]api.SFTPFile, error) {
	m.mu.RLock()
	sess, ok := m.sessions[params.SessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", params.SessionID)
	}

	entries, err := sess.Client.ReadDir(params.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to list directory: %w", err)
	}

	files := make([]api.SFTPFile, 0, len(entries))
	for _, entry := range entries {
		files = append(files, api.SFTPFile{
			Name:    entry.Name(),
			Size:    entry.Size(),
			Mode:    uint32(entry.Mode()),
			ModTime: entry.ModTime().Format(time.RFC3339),
			IsDir:   entry.IsDir(),
		})
	}

	return files, nil
}

// Download downloads a file from the remote server
func (m *Manager) Download(params api.SFTPDownloadParams) (*api.SFTPTransferResult, error) {
	m.mu.RLock()
	sess, ok := m.sessions[params.SessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", params.SessionID)
	}

	remoteFile, err := sess.Client.Open(params.RemotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to open remote file: %w", err)
	}
	defer remoteFile.Close()

	localFile, err := os.Create(params.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create local file: %w", err)
	}
	defer localFile.Close()

	n, err := io.Copy(localFile, remoteFile)
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}

	return &api.SFTPTransferResult{BytesTransferred: n}, nil
}

// Upload uploads a file to the remote server
func (m *Manager) Upload(params api.SFTPUploadParams) (*api.SFTPTransferResult, error) {
	m.mu.RLock()
	sess, ok := m.sessions[params.SessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", params.SessionID)
	}

	localFile, err := os.Open(params.LocalPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open local file: %w", err)
	}
	defer localFile.Close()

	// Ensure remote directory exists
	remoteDir := path.Dir(params.RemotePath)
	sess.Client.MkdirAll(remoteDir)

	remoteFile, err := sess.Client.Create(params.RemotePath)
	if err != nil {
		return nil, fmt.Errorf("failed to create remote file: %w", err)
	}
	defer remoteFile.Close()

	n, err := io.Copy(remoteFile, localFile)
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}

	return &api.SFTPTransferResult{BytesTransferred: n}, nil
}

// Delete removes a file or directory
func (m *Manager) Delete(sessionID, remotePath string) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Remove(remotePath)
}

// Rename renames a file or directory
func (m *Manager) Rename(sessionID, oldPath, newPath string) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Rename(oldPath, newPath)
}

// Mkdir creates a directory
func (m *Manager) Mkdir(sessionID, dirPath string) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Mkdir(dirPath)
}

// Stat returns file information
func (m *Manager) Stat(sessionID, filePath string) (*api.SFTPFile, error) {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	info, err := sess.Client.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &api.SFTPFile{
		Name:    path.Base(filePath),
		Size:    info.Size(),
		Mode:    uint32(info.Mode()),
		ModTime: info.ModTime().Format(time.RFC3339),
		IsDir:   info.IsDir(),
	}, nil
}

// Close closes an SFTP session
func (m *Manager) Close(sessionID string) error {
	m.mu.Lock()
	sess, ok := m.sessions[sessionID]
	if ok {
		delete(m.sessions, sessionID)
	}
	m.mu.Unlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Close()
}

// Chmod changes file permissions
func (m *Manager) Chmod(sessionID, filePath string, mode uint32) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Chmod(filePath, os.FileMode(mode))
}

// Readlink reads the target of a symbolic link
func (m *Manager) Readlink(sessionID, linkPath string) (string, error) {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return "", fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.ReadLink(linkPath)
}

// Symlink creates a symbolic link
func (m *Manager) Symlink(sessionID, oldPath, newPath string) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Symlink(oldPath, newPath)
}

// Rmdir removes a directory
func (m *Manager) Rmdir(sessionID, dirPath string) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.RemoveDirectory(dirPath)
}

// Lstat returns file info without following symlinks
func (m *Manager) Lstat(sessionID, filePath string) (*api.SFTPFile, error) {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	info, err := sess.Client.Lstat(filePath)
	if err != nil {
		return nil, err
	}

	return &api.SFTPFile{
		Name:      path.Base(filePath),
		FullPath:  filePath,
		Size:      info.Size(),
		Mode:      uint32(info.Mode()),
		ModTime:   info.ModTime().Format(time.RFC3339),
		IsDir:     info.IsDir(),
		IsSymlink: info.Mode()&os.ModeSymlink != 0,
	}, nil
}

// ReadDir reads a directory, returning full file info with symlink detection
func (m *Manager) ReadDir(sessionID, dirPath string) ([]api.SFTPFile, error) {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("session not found: %s", sessionID)
	}

	entries, err := sess.Client.ReadDir(dirPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read directory: %w", err)
	}

	files := make([]api.SFTPFile, 0, len(entries))
	for _, entry := range entries {
		fullPath := path.Join(dirPath, entry.Name())
		isSymlink := entry.Mode()&os.ModeSymlink != 0

		// Resolve symlink target if it is a symlink
		var symlinkTarget string
		if isSymlink {
			if target, err := sess.Client.ReadLink(fullPath); err == nil {
				symlinkTarget = target
			}
		}

		file := api.SFTPFile{
			Name:      entry.Name(),
			FullPath:  fullPath,
			Size:      entry.Size(),
			Mode:      uint32(entry.Mode()),
			ModTime:   entry.ModTime().Format(time.RFC3339),
			IsDir:     entry.IsDir(),
			IsSymlink: isSymlink,
		}

		// If symlink points to a directory, mark it as directory too
		if isSymlink && symlinkTarget != "" {
			if targetInfo, err := sess.Client.Stat(fullPath); err == nil {
				file.IsDir = targetInfo.IsDir()
			}
		}

		_ = symlinkTarget

		files = append(files, file)
	}

	return files, nil
}

// MkdirAll creates a directory and all parent directories
func (m *Manager) MkdirAll(sessionID, dirPath string) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.MkdirAll(dirPath)
}

// Truncate truncates a file to the specified size
func (m *Manager) Truncate(sessionID, filePath string, size int64) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Truncate(filePath, size)
}

// Chtimes changes the access and modification times of a file
func (m *Manager) Chtimes(sessionID, filePath string, atime, mtime time.Time) error {
	m.mu.RLock()
	sess, ok := m.sessions[sessionID]
	m.mu.RUnlock()

	if !ok {
		return fmt.Errorf("session not found: %s", sessionID)
	}

	return sess.Client.Chtimes(filePath, atime, mtime)
}
