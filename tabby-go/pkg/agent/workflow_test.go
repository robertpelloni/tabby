package agent

import (
	"context"
	"testing"
	"time"
)

func TestWorkflowManager_Lifecycle(t *testing.T) {
	mgr := NewWorkflowManager(nil)
	ctx := context.Background()

	wf := mgr.StartWorkflow(ctx, "Test Feature")
	if wf.ID == "" {
		t.Error("Workflow ID should not be empty")
	}

	// Wait for clarification phase
	timeout := time.After(10 * time.Second)
	tick := time.NewTicker(50 * time.Millisecond)
	defer tick.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for clarification phase")
		case <-tick.C:
			current, _ := mgr.GetWorkflow(wf.ID)
			if current.CurrentPhase == PhaseClarification {
				goto Clarify
			}
		}
	}

Clarify:
	err := mgr.SubmitResponse(wf.ID, "Yes please")
	if err != nil {
		t.Fatalf("Failed to submit response: %v", err)
	}

	// Wait for completion
	for {
		select {
		case <-timeout:
			t.Fatal("Timed out waiting for workflow completion")
		case <-tick.C:
			current, _ := mgr.GetWorkflow(wf.ID)
			if current.Status == StatusCompleted {
				if current.PhaseData["clarification_response"] != "Yes please" {
					t.Errorf("Expected response 'Yes please', got '%v'", current.PhaseData["clarification_response"])
				}
				return
			}
		}
	}
}
