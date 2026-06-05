import { Component } from '@angular/core'
import { AgentService } from '../../services/agent.service'

@Component({
    selector: 'agent-widgets',
    template: `
        <div class="agent-widgets-container" *ngIf="agent.getWidgets().length > 0 || agent.getWorkflows().length > 0">
            <workflow-progress></workflow-progress>

            <div class="widget-card card mb-2" *ngFor="let widget of agent.getWidgets()">
                <div class="card-header d-flex align-items-center">
                    <div class="me-auto">{{ widget.title }}</div>
                    <div class="badge bg-secondary">{{ widget.type }}</div>
                </div>
                <div class="card-body p-0">
                    <widget-vdom *ngIf="widget.type === 'vdom'" [node]="widget.vdom"></widget-vdom>
                    <div *ngIf="widget.type !== 'vdom'" class="p-3">
                        Widget type {{ widget.type }} not yet supported in frontend.
                    </div>
                </div>
            </div>
        </div>
    `,
    styles: [`
        .agent-widgets-container {
            position: absolute;
            right: 10px;
            top: 50px;
            width: 300px;
            z-index: 1000;
            max-height: 80vh;
            overflow-y: auto;
        }
        .widget-card {
            background: rgba(0, 0, 0, 0.8);
            backdrop-filter: blur(10px);
            border: 1px solid rgba(255, 255, 255, 0.1);
        }
    `]
})
export class AgentWidgetsComponent {
    constructor (public agent: AgentService) { }
}
