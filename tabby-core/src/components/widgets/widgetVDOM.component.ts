import { Component, Input } from '@angular/core'
import { AgentService } from '../../services/agent.service'

export interface VDOMNode {
    tag: string
    props?: any
    children?: (VDOMNode | string)[]
}

@Component({
    selector: 'widget-vdom',
    template: `
        <ng-container *ngFor="let child of node?.children">
            <ng-container *ngIf="isString(child)">
                {{ child }}
            </ng-container>
            <ng-container *ngIf="!isString(child)">
                <div [ngSwitch]="child.tag" [ngClass]="child.props?.className" [ngStyle]="child.props?.style">
                    <h1 *ngSwitchCase="'h1'"><widget-vdom [node]="child"></widget-vdom></h1>
                    <h2 *ngSwitchCase="'h2'"><widget-vdom [node]="child"></widget-vdom></h2>
                    <h3 *ngSwitchCase="'h3'"><widget-vdom [node]="child"></widget-vdom></h3>
                    <p *ngSwitchCase="'p'"><widget-vdom [node]="child"></widget-vdom></p>
                    <button *ngSwitchCase="'button'"
                            class="btn btn-sm btn-primary"
                            (click)="onButtonClick(child)">
                        <widget-vdom [node]="child"></widget-vdom>
                    </button>
                    <div *ngSwitchDefault><widget-vdom [node]="child"></widget-vdom></div>
                </div>
            </ng-container>
        </ng-container>
    `,
})
export class WidgetVDOMComponent {
    @Input() node: VDOMNode|null = null

    constructor (private agent: AgentService) { }

    isString (v: any): v is string {
        return typeof v === 'string'
    }

    onButtonClick (node: VDOMNode) {
        if (node.props?.workflowId && node.props?.action) {
            this.agent.submitWorkflowResponse(node.props.workflowId, node.props.action)
        }
    }
}
