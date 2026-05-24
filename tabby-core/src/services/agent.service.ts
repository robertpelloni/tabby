import { Injectable } from '@angular/core'
import { Logger, LogService } from './log.service'

export interface AgentTask {
    id: string
    description: string
    status: 'pending' | 'running' | 'completed' | 'failed'
    progress: number
    result?: string
    error?: string
}

@Injectable({ providedIn: 'root' })
export class AgentService {
    private logger: Logger

    constructor (
        log: LogService,
    ) {
        this.logger = log.create('agent')
    }

    async runTask (description: string): Promise<AgentTask | null> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('agent:runTask', { description })
            this.logger.info(`Started task ${result.id}: ${description}`)
            return result
        } catch (e) {
            this.logger.error('Failed to run agent task', e)
            return null
        }
    }

    async listTasks (): Promise<AgentTask[]> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('agent:listTasks')
            return result
        } catch (e) {
            this.logger.error('Failed to list agent tasks', e)
            return []
        }
    }

    async getTaskStatus (id: string): Promise<AgentTask | null> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('agent:getTaskStatus', { id })
            return result
        } catch (e) {
            this.logger.error(`Failed to get status for task ${id}`, e)
            return null
        }
    }
}
