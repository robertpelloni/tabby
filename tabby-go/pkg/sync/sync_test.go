package sync

import (
	"path/filepath"
	"testing"
)

func TestSyncPushAndPull(t *testing.T) {
	tmpDir := t.TempDir()
	syncFile := filepath.Join(tmpDir, "cloud_sync.json")

	mgr := &Manager{
		syncPath: syncFile,
	}

	// 1. Initial pull on non-existent file returns empty structure
	pullRes, err := mgr.Pull()
	if err != nil {
		t.Fatalf("unexpected error pulling empty sync: %v", err)
	}
	if len(pullRes.Data.Workflows) != 0 || len(pullRes.Data.Profiles) != 0 {
		t.Fatalf("expected empty sync data initially")
	}

	// 2. Push test sync data
	testData := SyncData{
		Workflows: []Workflow{
			{ID: "wf-1", Name: "Deploy", Command: "docker compose up -d", Tags: []string{"prod", "docker"}},
		},
		Profiles: []Profile{
			{ID: "prof-1", Name: "Prod SSH", Type: "ssh"},
		},
		EnvVars: []EnvVar{
			{Key: "ENVIRONMENT", Value: "production"},
		},
	}

	pushRes, err := mgr.Push(PushParams{Data: testData})
	if err != nil {
		t.Fatalf("unexpected error pushing sync: %v", err)
	}
	if !pushRes.Success {
		t.Fatalf("expected push success true")
	}

	// 3. Pull again and verify persisted data
	pullRes2, err := mgr.Pull()
	if err != nil {
		t.Fatalf("unexpected error pulling sync: %v", err)
	}
	if len(pullRes2.Data.Workflows) != 1 || pullRes2.Data.Workflows[0].Name != "Deploy" {
		t.Fatalf("workflows not restored correctly")
	}
	if len(pullRes2.Data.Profiles) != 1 || pullRes2.Data.Profiles[0].Name != "Prod SSH" {
		t.Fatalf("profiles not restored correctly")
	}
	if len(pullRes2.Data.EnvVars) != 1 || pullRes2.Data.EnvVars[0].Value != "production" {
		t.Fatalf("env vars not restored correctly")
	}
}
