import { ipcMain } from 'electron'
import { goBackend } from './goBackend'

export function initSSH () {
    // ---- SSH ----
    ipcMain.handle('ssh:connect', async (_event, params) => {
        return await goBackend.request('ssh.connect', params)
    })

    ipcMain.handle('ssh:startShell', async (_event, params) => {
        return await goBackend.request('ssh.startShell', params)
    })

    ipcMain.handle('ssh:resize', async (_event, params) => {
        return await goBackend.request('ssh.resize', params)
    })

    ipcMain.handle('ssh:close', async (_event, params) => {
        return await goBackend.request('ssh.close', params)
    })

    ipcMain.handle('ssh:addForward', async (_event, params) => {
        return await goBackend.request('ssh.addForward', params)
    })

    ipcMain.handle('ssh:removeForward', async (_event, params) => {
        return await goBackend.request('ssh.removeForward', params)
    })

    ipcMain.handle('ssh:verifyHostKey', async (_event, connectionId, accept) => {
        return await goBackend.request('ssh.verifyHostKey', { connectionId, accept })
    })

    ipcMain.handle('ssh:keyboardInteractiveResp', async (_event, connectionId, responses) => {
        return await goBackend.request('ssh.keyboardInteractiveResp', { connectionId, responses })
    })

    ipcMain.on('ssh:write', (_event, connectionId, sessionId, data) => {
        goBackend.request('ssh.write', { connectionId, sessionId, data }).catch(() => {})
    })

    // ---- SFTP ----
    ipcMain.handle('sftp:open', async (_event, connectionId) => {
        return await goBackend.request('sftp.open', { connectionId })
    })

    ipcMain.handle('sftp:close', async (_event, params) => {
        return await goBackend.request('sftp.close', params)
    })

    ipcMain.handle('sftp:list', async (_event, params) => {
        return await goBackend.request('sftp.list', params)
    })

    ipcMain.handle('sftp:readlink', async (_event, params) => {
        return await goBackend.request('sftp.readlink', params)
    })

    ipcMain.handle('sftp:stat', async (_event, params) => {
        return await goBackend.request('sftp.stat', params)
    })

    ipcMain.handle('sftp:delete', async (_event, params) => {
        return await goBackend.request('sftp.delete', params)
    })

    ipcMain.handle('sftp:mkdir', async (_event, params) => {
        return await goBackend.request('sftp.mkdir', params)
    })

    ipcMain.handle('sftp:rename', async (_event, params) => {
        return await goBackend.request('sftp.rename', params)
    })

    ipcMain.handle('sftp:chmod', async (_event, params) => {
        return await goBackend.request('sftp.chmod', params)
    })

    ipcMain.handle('sftp:upload', async (_event, params) => {
        return await goBackend.request('sftp.upload', params)
    })

    ipcMain.handle('sftp:download', async (_event, params) => {
        return await goBackend.request('sftp.download', params)
    })

    // Listen to async notifications from Go daemon and route them to Electron windows
    goBackend.on('ssh.data', (params: any) => {
        const { connectionId, sessionId, data } = params
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:data', connectionId, sessionId, data)
        })
    })

    goBackend.on('ssh.exit', (params: any) => {
        const { connectionId, sessionId, exitCode } = params
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:exit', connectionId, sessionId, exitCode)
        })
    })

    goBackend.on('ssh.serviceMessage', (params: any) => {
        const { connectionId, message } = params
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:serviceMessage', connectionId, message)
        })
    })

    goBackend.on('ssh.hostKeyPrompt', (params: any) => {
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:hostKeyPrompt', params)
        })
    })

    goBackend.on('ssh.keyboardInteractive', (params: any) => {
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:keyboardInteractive', params)
        })
    })

    goBackend.on('ssh.banner', (params: any) => {
        const { connectionId, message } = params
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:banner', connectionId, message)
        })
    })

    goBackend.on('ssh.portForwardEvent', (params: any) => {
        const { BrowserWindow } = require('electron')
        BrowserWindow.getAllWindows().forEach(w => {
            w.webContents.send('ssh:portForwardEvent', params)
        })
    })
}
