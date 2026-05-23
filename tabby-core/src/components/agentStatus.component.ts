import { Component } from '@angular/core'
import { AgentService, AgentTask } from '../services/agent.service'

@Component({
    selector: 'agent-status',
    templateUrl: './agentStatus.component.pug',
    styles: [require('./transfersMenu.component.scss')],
})
export class AgentStatusComponent {
    tasks: AgentTask[] = []

    constructor (
        public agent: AgentService,
    ) { }

    async ngOnInit () {
        setInterval(async () => {
            this.tasks = await this.agent.listTasks()
        }, 1000)
    }

    get runningTasks () {
        return this.tasks.filter(t => t.status === 'running' || t.status === 'pending')
    }
}
