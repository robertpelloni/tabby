import { Injectable } from '@angular/core'
import { Logger, LogService } from './log.service'


export interface SyncData {
    workflows: any[]
    profiles: any[]
    envVars: any[]
}

@Injectable({ providedIn: 'root' })
export class SyncService {
    private logger: Logger

    constructor (

        log: LogService,
    ) {
        this.logger = log.create('sync')
    }

    async push (data: SyncData): Promise<boolean> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('sync.push', { data })
            this.logger.info('Pushed sync data', result)
            return result.success
        } catch (e) {
            this.logger.error('Failed to push sync data', e)
            return false
        }
    }

    async pull (): Promise<SyncData | null> {
        try {
            const ipc = window['require']('electron').ipcRenderer
            const result = await ipc.invoke('sync.pull')
            this.logger.info('Pulled sync data', result)
            return result.data
        } catch (e) {
            this.logger.error('Failed to pull sync data', e)
            return null
        }
    }
}
