package agent

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

type WorkflowPhase string

const (
	PhaseDiscovery      WorkflowPhase = "discovery"
	PhaseExploration    WorkflowPhase = "exploration"
	PhaseClarification  WorkflowPhase = "clarification"
	PhaseDesign         WorkflowPhase = "design"
	PhaseImplementation WorkflowPhase = "implementation"
	PhaseReview         WorkflowPhase = "review"
	PhaseSummary        WorkflowPhase = "summary"
)

type Workflow struct {
	ID           string           `json:"id"`
	Description  string           `json:"description"`
	CurrentPhase WorkflowPhase    `json:"currentPhase"`
	Status       TaskStatus       `json:"status"`
	CreatedAt    time.Time        `json:"createdAt"`
	PhaseData    map[string]any   `json:"phaseData"`
	UserInput    chan string      `json:"-"`
}

type WorkflowManager struct {
	workflows map[string]*Workflow
	mu        sync.RWMutex
	notify    NotifyFunc
}

func NewWorkflowManager(notify NotifyFunc) *WorkflowManager {
	return &WorkflowManager{
		workflows: make(map[string]*Workflow),
		notify:    notify,
	}
}

func (wm *WorkflowManager) StartWorkflow(ctx context.Context, description string) *Workflow {
	wf := &Workflow{
		ID:           uuid.New().String(),
		Description:  description,
		CurrentPhase: PhaseDiscovery,
		Status:       StatusRunning,
		CreatedAt:    time.Now(),
		PhaseData:    make(map[string]any),
		UserInput:    make(chan string),
	}

	wm.mu.Lock()
	wm.workflows[wf.ID] = wf
	wm.mu.Unlock()

	go wm.runWorkflow(wf)

	return wf
}

func (wm *WorkflowManager) runWorkflow(wf *Workflow) {
	phases := []WorkflowPhase{
		PhaseDiscovery, PhaseExploration, PhaseClarification,
		PhaseDesign, PhaseImplementation, PhaseReview, PhaseSummary,
	}

	for _, phase := range phases {
		wf.CurrentPhase = phase
		wm.emitUpdate(wf)

		if phase == PhaseClarification {
			// Wait for user input
			select {
			case input := <-wf.UserInput:
				wf.PhaseData["clarification_response"] = input
			case <-time.After(5 * time.Minute):
				wf.Status = StatusFailed
				wf.PhaseData["error"] = "timeout waiting for user clarification"
				wm.emitUpdate(wf)
				return
			}
		} else {
			// Simulate work for other phases
			time.Sleep(1 * time.Second)
		}
	}

	wf.Status = StatusCompleted
	wm.emitUpdate(wf)
}

func (wm *WorkflowManager) SubmitResponse(workflowID string, response string) error {
	wm.mu.RLock()
	wf, ok := wm.workflows[workflowID]
	wm.mu.RUnlock()

	if !ok {
		return fmt.Errorf("workflow not found: %s", workflowID)
	}

	if wf.CurrentPhase != PhaseClarification {
		return fmt.Errorf("workflow not in clarification phase: %s", wf.CurrentPhase)
	}

	wf.UserInput <- response
	return nil
}

func (wm *WorkflowManager) emitUpdate(wf *Workflow) {
	if wm.notify != nil {
		wm.notify("agent.workflowUpdated", wf)
	}
}

func (wm *WorkflowManager) GetWorkflow(id string) (*Workflow, bool) {
	wm.mu.RLock()
	defer wm.mu.RUnlock()
	wf, ok := wm.workflows[id]
	return wf, ok
}
