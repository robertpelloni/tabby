import { Component, Input, ViewChild, ElementRef } from '@angular/core'
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap'

@Component({
    template: `
        <div class="modal-header">
            <h4 class="modal-title">{{prompt}}</h4>
        </div>
        <div class="modal-body">
            <input
                #input
                [type]="password ? 'password' : 'text'"
                class="form-control"
                [(ngModel)]="value"
                (keyup.enter)="submit()"
                (keyup.esc)="close()"
                [placeholder]="placeholder || ''"
                autofocus
            >
        </div>
        <div class="modal-footer">
            <button class="btn btn-primary" (click)="submit()">OK</button>
            <button class="btn btn-secondary" (click)="close()">Cancel</button>
        </div>
    `,
})
export class PromptModalComponent {
    @Input() prompt: string
    @Input() value = ''
    @Input() password = false
    @Input() placeholder = ''

    @ViewChild('input') input: ElementRef

    constructor (
        private modal: NgbActiveModal,
    ) { }

    ngAfterViewInit () {
        setTimeout(() => {
            this.input.nativeElement.focus()
        })
    }

    submit () {
        this.modal.close(this.value)
    }

    close () {
        this.modal.dismiss()
    }
}
