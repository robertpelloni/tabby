/* eslint-disable @typescript-eslint/no-unused-vars */
import { Subject, Observable } from 'rxjs'
import { posix as posixPath } from 'path'
import { Injector } from '@angular/core'
import { FileDownload, FileUpload, Logger, LogService, PlatformService, FileTransfer } from 'tabby-core'

export interface SFTPFile {
    name: string
    fullPath: string
    isDirectory: boolean
    isSymlink: boolean
    mode: number
    size: number
    modified: Date
}

export class SFTPSession {
    get closed$ (): Observable<void> { return this.closed }
    private closed = new Subject<void>()
    private logger: Logger
    private platform: PlatformService
    private activeTransfers = new Map<string, FileTransfer>()

    constructor (private sftpSessionId: string, injector: Injector) {
        this.logger = injector.get(LogService).create('sftp')
        this.platform = injector.get(PlatformService)

        window['require']('electron').ipcRenderer.on('sftp:progress', this.onProgress)
    }

    async readdir (p: string): Promise<SFTPFile[]> {
        this.logger.debug('readdir', p)
        const entries = await window['require']('electron').ipcRenderer.invoke('sftp:list', { sessionId: this.sftpSessionId, path: p })
        return entries.map((entry: any) => ({
            name: entry.name,
            fullPath: posixPath.join(p, entry.name),
            isDirectory: entry.isDir,
            isSymlink: entry.isSymlink,
            mode: entry.mode || 0,
            size: entry.size || 0,
            modified: new Date(entry.modTime),
        }))
    }

    readlink (p: string): Promise<string> {
        this.logger.debug('readlink', p)
        return window['require']('electron').ipcRenderer.invoke('sftp:readlink', { sessionId: this.sftpSessionId, path: p })
    }

    async stat (p: string): Promise<SFTPFile> {
        this.logger.debug('stat', p)
        const stats = await window['require']('electron').ipcRenderer.invoke('sftp:stat', { sessionId: this.sftpSessionId, path: p })
        return {
            name: posixPath.basename(p),
            fullPath: p,
            isDirectory: stats.isDir,
            isSymlink: stats.isSymlink,
            mode: stats.mode || 0,
            size: stats.size || 0,
            modified: new Date(stats.modTime),
        }
    }

    async rmdir (p: string): Promise<void> {
        await window['require']('electron').ipcRenderer.invoke('sftp:delete', { sessionId: this.sftpSessionId, path: p })
    }

    async mkdir (p: string): Promise<void> {
        await window['require']('electron').ipcRenderer.invoke('sftp:mkdir', { sessionId: this.sftpSessionId, path: p })
    }

    async rename (oldPath: string, newPath: string): Promise<void> {
        this.logger.debug('rename', oldPath, newPath)
        await window['require']('electron').ipcRenderer.invoke('sftp:rename', { sessionId: this.sftpSessionId, oldPath, newPath })
    }

    async unlink (p: string): Promise<void> {
        await window['require']('electron').ipcRenderer.invoke('sftp:delete', { sessionId: this.sftpSessionId, path: p })
    }

    async chmod (p: string, mode: string|number): Promise<void> {
        this.logger.debug('chmod', p, mode)
        const numericMode = typeof mode === 'string' ? parseInt(mode, 8) : mode
        await window['require']('electron').ipcRenderer.invoke('sftp:chmod', { sessionId: this.sftpSessionId, path: p, mode: numericMode })
    }

    async upload (path: string, transfer: FileUpload): Promise<void> {
        this.logger.info('Uploading into', path)

        const transferId = Math.random().toString(36).substring(7)
        this.activeTransfers.set(transferId, transfer)
        ;(this.platform as any)._registerFileTransfer?.(transfer)

        try {
            const localPath = transfer.getFilePath()
            let base64Data: string|undefined

            if (!localPath) {
                // For now we delegate the full read logic to the main process
                // Note: in a true production app, we would chunk this stream.
                // For parity with Go backend proxying, we send it as a base64 payload.
                const chunks: Uint8Array[] = []
                while (true) {
                    const chunk = await transfer.read()
                    if (!chunk.length) { break }
                    chunks.push(chunk)
                }

                const totalLength = chunks.reduce((acc, val) => acc + val.length, 0)
                const result = new Uint8Array(totalLength)
                let offset = 0
                for (const chunk of chunks) {
                    result.set(chunk, offset)
                    offset += chunk.length
                }

                base64Data = Buffer.from(result).toString('base64')
            }

            await window['require']('electron').ipcRenderer.invoke('sftp:upload', {
                sessionId: this.sftpSessionId,
                remotePath: path,
                localPath,
                data: base64Data,
                transferId,
            })

            transfer.setCompleted(true)
            transfer.close()
        } catch (e) {
            transfer.cancel()
            throw e
        } finally {
            this.activeTransfers.delete(transferId)
        }
    }

    async download (path: string, transfer: FileDownload): Promise<void> {
        this.logger.info('Downloading', path)

        const transferId = Math.random().toString(36).substring(7)
        this.activeTransfers.set(transferId, transfer)
        ;(this.platform as any)._registerFileTransfer?.(transfer)

        try {
            const localPath = transfer.getFilePath()

            const res = await window['require']('electron').ipcRenderer.invoke('sftp:download', {
                sessionId: this.sftpSessionId,
                remotePath: path,
                localPath,
                transferId,
            })

            if (res.data) {
                const buffer = Buffer.from(res.data, 'base64')
                await transfer.write(buffer)
            }

            transfer.setCompleted(true)
            transfer.close()
        } catch (e) {
            transfer.cancel()
            throw e
        } finally {
            this.activeTransfers.delete(transferId)
        }
    }

    private onProgress = (_event: any, params: any) => {
        const transfer = this.activeTransfers.get(params.transferId)
        if (transfer) {
            transfer.updateProgress(params.bytesTransferred, params.totalBytes)
            if (params.complete) {
                transfer.setCompleted(true)
            }
            if (params.error) {
                this.logger.error(`Transfer ${params.transferId} failed: ${params.error}`)
                transfer.cancel()
            }
        }
    }
}
