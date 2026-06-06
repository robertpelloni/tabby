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

    ipcMain.handle('agent:createWidget', async (_event, params) => {
        return goBackend.request('agent.createWidget', params)
    })

    ipcMain.handle('agent:updateWidgetVDOM', async (_event, params) => {
        return goBackend.request('agent.updateWidgetVDOM', params)
    })

    ipcMain.handle('agent:startWorkflow', async (_event, params) => {
        return goBackend.request('agent.startWorkflow', params)
    })

    ipcMain.handle('agent:submitWorkflowResponse', async (_event, params) => {
        return goBackend.request('agent.submitWorkflowResponse', params)
    })

    goBackend.on('agent.taskUpdated', (task: any) => {
        broadcast('agent:taskUpdated', task)
    })

    goBackend.on('agent.widgetCreated', (widget: any) => {
        broadcast('agent:widgetCreated', widget)
    })

    goBackend.on('agent.widgetUpdated', (widget: any) => {
        broadcast('agent:widgetUpdated', widget)
    })

    goBackend.on('agent.workflowUpdated', (wf: any) => {
        broadcast('agent:workflowUpdated', wf)
    })
}

function broadcast(channel: string, payload: any) {
    const { BrowserWindow } = require('electron')
    BrowserWindow.getAllWindows().forEach(w => {
        w.webContents.send(channel, payload)
    })
}
