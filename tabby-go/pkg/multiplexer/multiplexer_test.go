package multiplexer

import (
	"testing"
)

func TestKeyString(t *testing.T) {
	key := MultiplexerKey{
		Host: "example.com",
		Port: 22,
		User: "root",
	}
	result := KeyString(key)
	expected := "example.com:22:root:::0::0"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestKeyStringWithProxy(t *testing.T) {
	key := MultiplexerKey{
		Host:           "example.com",
		Port:           22,
		User:           "root",
		SocksProxyHost: "proxy.local",
		SocksProxyPort: 1080,
	}
	result := KeyString(key)
	expected := "example.com:22:root::proxy.local:1080::0"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestKeyStringWithJumpHost(t *testing.T) {
	key := MultiplexerKey{
		Host:        "target.com",
		Port:        22,
		User:        "admin",
		JumpHostKey: "jump-host-profile-id",
	}
	result := KeyString(key)
	expected := "target.com:22:admin:::0::0$jump-host-profile-id"
	if result != expected {
		t.Errorf("Expected %q, got %q", expected, result)
	}
}

func TestRegisterAndGet(t *testing.T) {
	mgr := NewManager()
	key := MultiplexerKey{Host: "example.com", Port: 22, User: "root"}

	// Initially no connection
	if connID := mgr.Get(key); connID != "" {
		t.Error("Should return empty for unregistered key")
	}

	// Register
	mgr.Register(key, "conn-1")

	// Get should return the connection
	if connID := mgr.Get(key); connID != "conn-1" {
		t.Errorf("Expected 'conn-1', got %q", connID)
	}

	// Ref count should be 2 (register + get)
	active := mgr.ListActive()
	keyStr := KeyString(key)
	if active[keyStr] != 2 {
		t.Errorf("Expected ref count 2, got %d", active[keyStr])
	}
}

func TestRelease(t *testing.T) {
	mgr := NewManager()
	key := MultiplexerKey{Host: "example.com", Port: 22, User: "root"}
	mgr.Register(key, "conn-1")

	// Release once (ref goes from 1 to 0)
	released := mgr.Release(key)
	if !released {
		t.Error("Should be fully released when ref count hits 0")
	}

	// Should no longer exist
	if connID := mgr.Get(key); connID != "" {
		t.Error("Connection should be removed after full release")
	}
}

func TestReleaseMultipleRefs(t *testing.T) {
	mgr := NewManager()
	key := MultiplexerKey{Host: "example.com", Port: 22, User: "root"}
	mgr.Register(key, "conn-1")
	mgr.Get(key) // ref count = 2

	// First release should not fully release
	released := mgr.Release(key)
	if released {
		t.Error("Should not be fully released with ref count 1")
	}

	// Second release should fully release
	released = mgr.Release(key)
	if !released {
		t.Error("Should be fully released when ref count hits 0")
	}
}

func TestRemove(t *testing.T) {
	mgr := NewManager()
	key := MultiplexerKey{Host: "example.com", Port: 22, User: "root"}
	mgr.Register(key, "conn-1")

	// Remove by connection ID
	mgr.Remove("conn-1")

	// Should no longer exist
	active := mgr.ListActive()
	if len(active) != 0 {
		t.Error("Should have no active connections after remove")
	}
}

func TestRemoveNonExistent(t *testing.T) {
	mgr := NewManager()
	mgr.Remove("nonexistent") // Should not panic
}

func TestDifferentKeys(t *testing.T) {
	mgr := NewManager()
	key1 := MultiplexerKey{Host: "a.com", Port: 22, User: "root"}
	key2 := MultiplexerKey{Host: "b.com", Port: 22, User: "root"}

	mgr.Register(key1, "conn-1")
	mgr.Register(key2, "conn-2")

	if connID := mgr.Get(key1); connID != "conn-1" {
		t.Errorf("Expected conn-1, got %q", connID)
	}
	if connID := mgr.Get(key2); connID != "conn-2" {
		t.Errorf("Expected conn-2, got %q", connID)
	}

	active := mgr.ListActive()
	if len(active) != 2 {
		t.Errorf("Expected 2 active connections, got %d", len(active))
	}
}

func TestListActive(t *testing.T) {
	mgr := NewManager()
	active := mgr.ListActive()
	if active == nil || len(active) != 0 {
		t.Error("New manager should have no active connections")
	}
}
