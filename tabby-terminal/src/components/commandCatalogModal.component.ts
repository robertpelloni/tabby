import { Component } from '@angular/core'
import { NgbActiveModal } from '@ng-bootstrap/ng-bootstrap'

@Component({
    selector: 'command-catalog-modal',
    template: `
        <div class="modal-header">
            <h5 class="modal-title">Command Catalog & Workflows</h5>
            <button type="button" class="close" (click)="modal.dismiss()">
                <span aria-hidden="true">&times;</span>
            </button>
        </div>
        <div class="modal-body">
            <input type="text" class="form-control mb-3" placeholder="Search saved commands..." [(ngModel)]="searchQuery" (ngModelChange)="filterCommands()" autofocus>
            <div class="list-group">
                <button type="button" class="list-group-item list-group-item-action" *ngFor="let cmd of filteredCommands" (click)="selectCommand(cmd)">
                    <strong>{{ cmd.title }}</strong><br>
                    <small class="text-muted" style="font-family: monospace;" [innerHTML]="formatCommand(cmd.command)"></small>
                </button>
            </div>
            <div *ngIf="filteredCommands.length === 0" class="text-center text-muted mt-3">
                No commands match your search.
            </div>
        </div>
        <div class="modal-footer">
            <button type="button" class="btn btn-secondary" (click)="modal.dismiss()">Cancel</button>
        </div>
    `
})
export class CommandCatalogModalComponent {
    searchQuery = ''
    commands = [
        { title: 'Docker Build & Push', command: 'docker build -t {{image_name}} . && docker push {{image_name}}' },
        { title: 'Find text in files', command: 'grep -rn "{{search_text}}" .' },
        { title: 'Tail Docker logs', command: 'docker logs -f --tail 100 {{container_id}}' },
        { title: 'Git Rebase Interactive', command: 'git rebase -i HEAD~{{num_commits}}' },
        { title: 'Kubernetes get pods', command: 'kubectl get pods -n {{namespace}}' },
    ]
    filteredCommands = [...this.commands]

    constructor (public modal: NgbActiveModal) {}

    filterCommands() {
        if (!this.searchQuery) {
            this.filteredCommands = [...this.commands]
            return
        }
        const query = this.searchQuery.toLowerCase()
        this.filteredCommands = this.commands.filter(cmd =>
            cmd.title.toLowerCase().includes(query) || cmd.command.toLowerCase().includes(query)
        )
    }

    selectCommand(cmd: any) {
        let finalCommand = cmd.command;
        const matches = [...finalCommand.matchAll(/{{([^}]+)}}/g)];

        for (const match of matches) {
            const paramName = match[1];
            const val = prompt(`Enter value for ${paramName}:`);
            if (val === null) {
                // User cancelled the prompt
                return;
            }
            finalCommand = finalCommand.replace(match[0], val);
        }

        this.modal.close(finalCommand);
    }

    formatCommand(cmd: string): string {
        return cmd.replace(/{{([^}]+)}}/g, '<span style="color: #ffb86c;">{{$1}}</span>')
    }
}
