package profile

import (
	"os"
	"path/filepath"
	"testing"
)

func TestManagerCreation(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
	if len(mgr.List()) != 0 {
		t.Error("New manager should have no profiles")
	}
}

func TestAddProfile(t *testing.T) {
	mgr := NewManager()
	p := &Profile{
		ID:   "test-ssh-1",
		Type: TypeSSH,
		Name: "Test SSH",
		Options: &SSHProfileOptions{
			Host: "example.com",
			Port: 22,
			User: "admin",
			Auth:  "agent",
		},
	}

	err := mgr.Add(p)
	if err != nil {
		t.Fatalf("Failed to add profile: %v", err)
	}

	if len(mgr.List()) != 1 {
		t.Error("Should have 1 profile")
	}
}

func TestAddProfileNoID(t *testing.T) {
	mgr := NewManager()
	p := &Profile{Name: "No ID"}
	err := mgr.Add(p)
	if err == nil {
		t.Error("Should fail adding profile without ID")
	}
}

func TestGetProfile(t *testing.T) {
	mgr := NewManager()
	mgr.Add(&Profile{ID: "p1", Type: TypeSSH, Name: "Profile 1", Options: &SSHProfileOptions{Host: "h1"}})
	mgr.Add(&Profile{ID: "p2", Type: TypeLocal, Name: "Profile 2", Options: &LocalProfileOptions{Command: "/bin/bash"}})

	p, ok := mgr.Get("p1")
	if !ok {
		t.Fatal("Should find profile p1")
	}
	if p.Name != "Profile 1" {
		t.Errorf("Expected name 'Profile 1', got %q", p.Name)
	}

	_, ok = mgr.Get("nonexistent")
	if ok {
		t.Error("Should not find nonexistent profile")
	}
}

func TestRemoveProfile(t *testing.T) {
	mgr := NewManager()
	mgr.Add(&Profile{ID: "p1", Type: TypeSSH, Name: "Profile 1", Options: &SSHProfileOptions{Host: "h1"}})
	mgr.Remove("p1")

	if len(mgr.List()) != 0 {
		t.Error("Should have 0 profiles after removal")
	}
}

func TestListByType(t *testing.T) {
	mgr := NewManager()
	mgr.Add(&Profile{ID: "ssh1", Type: TypeSSH, Name: "SSH 1", Options: &SSHProfileOptions{Host: "h1"}})
	mgr.Add(&Profile{ID: "ssh2", Type: TypeSSH, Name: "SSH 2", Options: &SSHProfileOptions{Host: "h2"}})
	mgr.Add(&Profile{ID: "local1", Type: TypeLocal, Name: "Local 1", Options: &LocalProfileOptions{Command: "bash"}})
	mgr.Add(&Profile{ID: "serial1", Type: TypeSerial, Name: "Serial 1", Options: &SerialProfileOptions{Port: "COM1"}})

	sshProfiles := mgr.ListByType(TypeSSH)
	if len(sshProfiles) != 2 {
		t.Errorf("Expected 2 SSH profiles, got %d", len(sshProfiles))
	}

	localProfiles := mgr.ListByType(TypeLocal)
	if len(localProfiles) != 1 {
		t.Errorf("Expected 1 local profile, got %d", len(localProfiles))
	}

	telnetProfiles := mgr.ListByType(TypeTelnet)
	if len(telnetProfiles) != 0 {
		t.Errorf("Expected 0 telnet profiles, got %d", len(telnetProfiles))
	}
}

func TestListByGroup(t *testing.T) {
	mgr := NewManager()
	mgr.Add(&Profile{ID: "p1", Type: TypeSSH, Name: "P1", Group: "Production", Options: &SSHProfileOptions{Host: "h1"}})
	mgr.Add(&Profile{ID: "p2", Type: TypeSSH, Name: "P2", Group: "Production", Options: &SSHProfileOptions{Host: "h2"}})
	mgr.Add(&Profile{ID: "p3", Type: TypeSSH, Name: "P3", Group: "Staging", Options: &SSHProfileOptions{Host: "h3"}})

	prod := mgr.ListByGroup("Production")
	if len(prod) != 2 {
		t.Errorf("Expected 2 production profiles, got %d", len(prod))
	}
}

func TestUpdateProfile(t *testing.T) {
	mgr := NewManager()
	mgr.Add(&Profile{ID: "p1", Type: TypeSSH, Name: "Original", Options: &SSHProfileOptions{Host: "h1"}})

	err := mgr.Update("p1", &Profile{
		Type: TypeSSH,
		Name: "Updated",
		Options: &SSHProfileOptions{Host: "h1-updated"},
	})
	if err != nil {
		t.Fatalf("Failed to update: %v", err)
	}

	p, _ := mgr.Get("p1")
	if p.Name != "Updated" {
		t.Errorf("Expected name 'Updated', got %q", p.Name)
	}
}

func TestUpdateNonexistent(t *testing.T) {
	mgr := NewManager()
	err := mgr.Update("nonexistent", &Profile{Name: "Test"})
	if err == nil {
		t.Error("Should fail updating non-existent profile")
	}
}

func TestImportSSHConfig(t *testing.T) {
	dir := t.TempDir()
	configPath := filepath.Join(dir, "config")

	config := `Host web1
    HostName web1.example.com
    User admin
    Port 2222
    IdentityFile ~/.ssh/id_rsa
    ForwardAgent yes
    ServerAliveInterval 30

Host db1
    HostName db1.example.com
    User postgres
    Port 5432
    ProxyCommand ssh -W %h:%p jump.example.com

Host jump
    HostName jump.example.com
    User jumpuser
`
	os.WriteFile(configPath, []byte(config), 0644)

	mgr := NewManager()
	imported, err := mgr.ImportSSHConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to import: %v", err)
	}

	if len(imported) != 3 {
		t.Fatalf("Expected 3 profiles, got %d", len(imported))
	}

	// Verify web1
	p1, ok := mgr.Get("ssh-import-web1")
	if !ok {
		t.Fatal("Should find web1 profile")
	}
	opts1 := p1.Options.(*SSHProfileOptions)
	if opts1.Host != "web1.example.com" {
		t.Errorf("Expected host 'web1.example.com', got %q", opts1.Host)
	}
	if opts1.Port != 2222 {
		t.Errorf("Expected port 2222, got %d", opts1.Port)
	}
	if opts1.User != "admin" {
		t.Errorf("Expected user 'admin', got %q", opts1.User)
	}
	if !opts1.AgentForward {
		t.Error("AgentForward should be true")
	}
	if opts1.KeepaliveInterval != 30 {
		t.Errorf("Expected keepalive 30, got %d", opts1.KeepaliveInterval)
	}

	// Verify db1
	p2, _ := mgr.Get("ssh-import-db1")
	opts2 := p2.Options.(*SSHProfileOptions)
	if opts2.ProxyCommand != "ssh -W %h:%p jump.example.com" {
		t.Errorf("Unexpected proxy command: %q", opts2.ProxyCommand)
	}

	// Verify jump
	p3, _ := mgr.Get("ssh-import-jump")
	opts3 := p3.Options.(*SSHProfileOptions)
	if opts3.Host != "jump.example.com" {
		t.Errorf("Expected host 'jump.example.com', got %q", opts3.Host)
	}
}

func TestImportNonexistentConfig(t *testing.T) {
	mgr := NewManager()
	_, err := mgr.ImportSSHConfig("/nonexistent/config")
	if err == nil {
		t.Error("Should fail importing nonexistent file")
	}
}

func TestGetDefaultSSHConfigPath(t *testing.T) {
	path := GetDefaultSSHConfigPath()
	if path == "" {
		t.Error("Path should not be empty")
	}
	if !filepath.IsAbs(path) {
		t.Errorf("Path should be absolute: %q", path)
	}
}
