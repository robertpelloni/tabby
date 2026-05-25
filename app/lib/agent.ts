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

    goBackend.on('agent.taskUpdated', (task: any) => {
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('agent:taskUpdated', task)
        })
    })
}
