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
        this.tasks = await this.agent.listTasks()
        this.agent.taskUpdated$.subscribe(task => {
            const index = this.tasks.findIndex(t => t.id === task.id)
            if (index !== -1) {
                this.tasks[index] = task
            } else {
                this.tasks.push(task)
            }
        })
    }

    get runningTasks () {
        return this.tasks.filter(t => t.status === 'running' || t.status === 'pending')
    }
}
