import { Component } from '@angular/core'
import { AgentService, AgentWorkflow } from '../../services/agent.service'

@Component({
    selector: 'workflow-progress',
    template: `
        <div class="workflow-container card mb-2" *ngFor="let wf of agent.getWorkflows()">
            <div class="card-header d-flex align-items-center">
                <div class="me-auto small">{{ wf.description }}</div>
                <div class="badge" [class.bg-primary]="wf.status === 'running'" [class.bg-success]="wf.status === 'completed'">
                    {{ wf.currentPhase }}
                </div>
            </div>
            <div class="card-body">
                <div class="phases d-flex justify-content-between mb-3">
                    <div *ngFor="let phase of phases"
                         class="phase-dot"
                         [class.active]="wf.currentPhase === phase"
                         [class.completed]="isCompleted(wf, phase)"
                         [ngbTooltip]="phase">
                    </div>
                </div>

                <div *ngIf="wf.currentPhase === 'clarification'">
                    <p class="x-small text-muted mb-2">Agent needs clarification:</p>
                    <div class="input-group">
                        <input type="text" class="form-control form-control-sm" #responseInput placeholder="Your response...">
                        <button class="btn btn-primary btn-sm" (click)="agent.submitWorkflowResponse(wf.id, responseInput.value)">Send</button>
                    </div>
                </div>

                <div class="progress mt-2" style="height: 4px;">
                    <div class="progress-bar progress-bar-striped progress-bar-animated"
                         [style.width]="getProgress(wf) + '%'">
                    </div>
                </div>
            </div>
        </div>
    `,
    styles: [`
        .workflow-container {
            background: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
        .phase-dot {
            width: 8px;
            height: 8px;
            border-radius: 50%;
            background: #444;
        }
        .phase-dot.active {
            background: #007bff;
            box-shadow: 0 0 5px #007bff;
        }
        .phase-dot.completed {
            background: #28a745;
        }
        .x-small { font-size: 0.7rem; }
    `]
})
export class WorkflowProgressComponent {
    phases = ['discovery', 'exploration', 'clarification', 'design', 'implementation', 'review', 'summary']

    constructor (public agent: AgentService) { }

    isCompleted (wf: AgentWorkflow, phase: string): boolean {
        const currentIdx = this.phases.indexOf(wf.currentPhase)
        const phaseIdx = this.phases.indexOf(phase)
        return phaseIdx < currentIdx || wf.status === 'completed'
    }

    getProgress (wf: AgentWorkflow): number {
        const currentIdx = this.phases.indexOf(wf.currentPhase)
        return ((currentIdx + 1) / this.phases.length) * 100
    }
}
