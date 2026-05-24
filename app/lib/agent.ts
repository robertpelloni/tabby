import { ipcMain } from 'electron'
import { goBackend } from './goBackend'

export function initAgent() {
    ipcMain.handle('agent:runTask', async (_event, params) => {
        return goBackend.request('agent.runTask', params)
    })

    ipcMain.handle('agent:listTasks', async (_event) => {
        return goBackend.request('agent.listTasks')
    })

    ipcMain.handle('agent:getTaskStatus', async (_event, params) => {
        return goBackend.request('agent.getTaskStatus', params)
    })
}
