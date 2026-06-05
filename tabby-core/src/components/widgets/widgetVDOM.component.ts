import { Component, Input } from '@angular/core'

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
            <div *ngIf="!isString(child)" [ngClass]="child.props?.className" [ngStyle]="child.props?.style">
                <widget-vdom [node]="child"></widget-vdom>
            </div>
        </ng-container>
    `,
})
export class WidgetVDOMComponent {
    @Input() node: VDOMNode|null = null

    isString (v: any): v is string {
        return typeof v === 'string'
    }
}
