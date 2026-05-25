import { Component, Input } from '@angular/core'
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap'
import { SyncService } from '../services/sync.service'

@Component({
    templateUrl: './commandCatalogModal.component.pug',
    styles: [require('./selectorModal.component.scss')],
})
export class CommandCatalogModalComponent {
    @Input() workflows: any[] = []
    filter = ''

    constructor (
        public modal: NgbActiveModal,
        private sync: SyncService,
    ) { }

    async ngOnInit () {
        const data = await this.sync.pull()
        if (data) {
            this.workflows = data.workflows
        }
    }

    getFilteredWorkflows () {
        if (!this.filter) {
            return this.workflows
        }
        const f = this.filter.toLowerCase()
        return this.workflows.filter(w =>
            w.name.toLowerCase().includes(f) ||
            w.command.toLowerCase().includes(f) ||
            (w.tags && w.tags.some(t => t.toLowerCase().includes(f)))
        )
    }

    select (workflow: any) {
        this.modal.close(workflow)
    }
}
