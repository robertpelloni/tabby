import { ipcMain } from 'electron'
import { goBackend } from './goBackend'

export function initAI () {
    ipcMain.handle('ai:generateCommand', async (_event, params) => {
        return await goBackend.request('ai.generateCommand', params)
    })

    ipcMain.handle('ai:explainError', async (_event, params) => {
        return await goBackend.request('ai.explainError', params)
    })
}
