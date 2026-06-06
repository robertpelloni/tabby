import { Injectable } from '@angular/core'
import { Subject, Observable } from 'rxjs'
import { Logger, LogService } from './log.service'
import { VDOMNode } from '../components/widgets/widgetVDOM.component'

export interface AgentTask {
    id: string
    description: string
    status: 'pending' | 'running' | 'completed' | 'failed'
    progress: number
    result?: string
    error?: string
}

export interface AgentWidget {
    id: string
    type: string
    title: string
    vdom?: VDOMNode
}

export interface AgentWorkflow {
    id: string
    description: string
    currentPhase: string
    status: 'pending' | 'running' | 'completed' | 'failed'
    phaseData: any
}

@Injectable({ providedIn: 'root' })
export class AgentService {
    private logger: Logger
    private taskUpdated = new Subject<AgentTask>()
    private widgetCreated = new Subject<AgentWidget>()
    private widgetUpdated = new Subject<AgentWidget>()
    private workflowUpdated = new Subject<AgentWorkflow>()
    private tasks: Map<string, AgentTask> = new Map()
    private widgets: Map<string, AgentWidget> = new Map()
    private workflows: Map<string, AgentWorkflow> = new Map()

    readonly taskUpdated$: Observable<AgentTask> = this.taskUpdated.asObservable()
    readonly widgetCreated$: Observable<AgentWidget> = this.widgetCreated.asObservable()
    readonly widgetUpdated$: Observable<AgentWidget> = this.widgetUpdated.asObservable()
    readonly workflowUpdated$: Observable<AgentWorkflow> = this.workflowUpdated.asObservable()

    constructor (
        log: LogService,
    ) {
        this.logger = log.create('agent')
        const electron = window['require']('electron')
        if (electron && electron.ipcRenderer) {
            const ipc = electron.ipcRenderer
            ipc.on('agent:taskUpdated', (_event, task: AgentTask) => {
                this.tasks.set(task.id, task)
                this.taskUpdated.next(task)
            })
            ipc.on('agent:widgetCreated', (_event, widget: AgentWidget) => {
                this.widgets.set(widget.id, widget)
                this.widgetCreated.next(widget)
            })
            ipc.on('agent:widgetUpdated', (_event, widget: AgentWidget) => {
                this.widgets.set(widget.id, widget)
                this.widgetUpdated.next(widget)
            })
            ipc.on('agent:workflowUpdated', (_event, wf: AgentWorkflow) => {
                this.workflows.set(wf.id, wf)
                this.workflowUpdated.next(wf)
            })
        }
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

    async startWorkflow (description: string): Promise<AgentWorkflow | null> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('agent:startWorkflow', { description })
            this.logger.info(`Started workflow ${result.id}: ${description}`)
            return result
        } catch (e) {
            this.logger.error('Failed to start workflow', e)
            return null
        }
    }

    async submitWorkflowResponse (id: string, response: string): Promise<void> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            await ipc.invoke('agent:submitWorkflowResponse', { id, response })
        } catch (e) {
            this.logger.error(`Failed to submit response for workflow ${id}`, e)
        }
    }


    async listTasks (): Promise<AgentTask[]> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('agent:listTasks')
            for (const task of result) {
                this.tasks.set(task.id, task)
            }
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

    getWidgets (): AgentWidget[] {
        return Array.from(this.widgets.values())
    }

    getWorkflows (): AgentWorkflow[] {
        return Array.from(this.workflows.values())
    }
}
