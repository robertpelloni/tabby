package notification

import (
	"sync"
	"testing"
	"time"
)

func TestManagerCreation(t *testing.T) {
	mgr := NewManager()
	if mgr == nil {
		t.Fatal("Manager should not be nil")
	}
}

func TestInfo(t *testing.T) {
	mgr := NewManager()
	mgr.Info("Title", "Message")

	unread := mgr.GetUnread()
	if len(unread) != 1 {
		t.Fatalf("Expected 1 unread, got %d", len(unread))
	}
	if unread[0].Title != "Title" || unread[0].Message != "Message" {
		t.Errorf("Notification content mismatch: %+v", unread[0])
	}
	if unread[0].Level != LevelInfo {
		t.Errorf("Expected LevelInfo, got %d", unread[0].Level)
	}
}

func TestWarning(t *testing.T) {
	mgr := NewManager()
	mgr.Warning("Warn", "Something happened")

	unread := mgr.GetUnread()
	if unread[0].Level != LevelWarning {
		t.Errorf("Expected LevelWarning, got %d", unread[0].Level)
	}
}

func TestError(t *testing.T) {
	mgr := NewManager()
	mgr.Error("Error", "Something broke")

	unread := mgr.GetUnread()
	if unread[0].Level != LevelError {
		t.Errorf("Expected LevelError, got %d", unread[0].Level)
	}
}

func TestMarkRead(t *testing.T) {
	mgr := NewManager()
	mgr.Info("Test", "test")

	unread := mgr.GetUnread()
	if len(unread) != 1 {
		t.Fatal("Should have 1 unread")
	}

	mgr.MarkRead(unread[0].ID)

	if len(mgr.GetUnread()) != 0 {
		t.Error("Should have 0 unread after marking read")
	}
}

func TestGetAll(t *testing.T) {
	mgr := NewManager()
	mgr.Info("A", "a")
	mgr.Info("B", "b")
	mgr.MarkRead(mgr.GetUnread()[0].ID)

	all := mgr.GetAll()
	if len(all) != 2 {
		t.Errorf("Expected 2 total, got %d", len(all))
	}

	unread := mgr.GetUnread()
	if len(unread) != 1 {
		t.Errorf("Expected 1 unread, got %d", len(unread))
	}
}

func TestClear(t *testing.T) {
	mgr := NewManager()
	mgr.Info("A", "a")
	mgr.Info("B", "b")
	mgr.Clear()

	if len(mgr.GetAll()) != 0 {
		t.Error("All notifications should be cleared")
	}
}

func TestOnChange(t *testing.T) {
	mgr := NewManager()

	var mu sync.Mutex
	var count int
	mgr.OnChange(func(n []Notification) {
		mu.Lock()
		count++
		mu.Unlock()
	})

	mgr.Info("Test", "test")
	time.Sleep(10 * time.Millisecond)

	mu.Lock()
	if count != 1 {
		t.Errorf("Expected 1 change callback, got %d", count)
	}
	mu.Unlock()
}

func TestMaxNotifications(t *testing.T) {
	mgr := NewManager()
	for i := 0; i < 150; i++ {
		mgr.Info("Test", "test")
	}

	all := mgr.GetAll()
	if len(all) != 100 {
		t.Errorf("Expected 100 (max), got %d", len(all))
	}
}

func TestServiceMessage(t *testing.T) {
	mgr := NewManager()
	mgr.ServiceMessage("ssh-1", "Connected")

	unread := mgr.GetUnread()
	if len(unread) != 1 {
		t.Fatal("Should have 1 notification")
	}
	if unread[0].Title != "ssh-1" {
		t.Errorf("Expected 'ssh-1' as title, got %q", unread[0].Title)
	}
}
