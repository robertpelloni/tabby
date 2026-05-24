import { ipcMain } from 'electron'
import { goBackend } from './goBackend'

export function initSync() {
    ipcMain.handle('sync:push', async (_event, params) => {
        return goBackend.request('sync.push', params)
    })

    ipcMain.handle('sync:pull', async (_event) => {
        return goBackend.request('sync.pull')
    })
}
