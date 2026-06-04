/* eslint-disable @typescript-eslint/no-unused-vars */
import { Subject, Observable } from 'rxjs'
import { posix as posixPath } from 'path'
import { Injector } from '@angular/core'
import { FileDownload, FileUpload, Logger, LogService } from 'tabby-core'
<<<<<<< HEAD
=======
import * as russh from 'russh'
>>>>>>> upstream/master

export interface SFTPFile {
    name: string
    fullPath: string
    isDirectory: boolean
    isSymlink: boolean
    mode: number
    size: number
    modified: Date
}

<<<<<<< HEAD
=======
export class SFTPFileHandle {
    position = 0

    constructor (
        private inner: russh.SFTPFile|null,
    ) { }

    async read (): Promise<Uint8Array> {
        if (!this.inner) {
            return Promise.resolve(new Uint8Array(0))
        }
        return this.inner.read(256 * 1024)
    }

    async write (chunk: Uint8Array): Promise<void> {
        if (!this.inner) {
            throw new Error('File handle is closed')
        }
        await this.inner.writeAll(chunk)
    }

    async close (): Promise<void> {
        await this.inner?.shutdown()
        this.inner = null
    }
}

>>>>>>> upstream/master
export class SFTPSession {
    get closed$ (): Observable<void> { return this.closed }
    private closed = new Subject<void>()
    private logger: Logger

<<<<<<< HEAD
    constructor (private sftpSessionId: string, injector: Injector) {
        this.logger = injector.get(LogService).create('sftp')
=======
    constructor (private sftp: russh.SFTP, injector: Injector) {
        this.logger = injector.get(LogService).create('sftp')
        sftp.closed$.subscribe(() => {
            this.closed.next()
            this.closed.complete()
        })
>>>>>>> upstream/master
    }

    async readdir (p: string): Promise<SFTPFile[]> {
        this.logger.debug('readdir', p)
<<<<<<< HEAD
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
=======
        const entries = await this.sftp.readDirectory(p)
        return entries.map(entry => this._makeFile(
            posixPath.join(p, entry.name), entry,
        ))
>>>>>>> upstream/master
    }

    readlink (p: string): Promise<string> {
        this.logger.debug('readlink', p)
<<<<<<< HEAD
        return window['require']('electron').ipcRenderer.invoke('sftp:readlink', { sessionId: this.sftpSessionId, path: p })
=======
        return this.sftp.readlink(p)
>>>>>>> upstream/master
    }

    async stat (p: string): Promise<SFTPFile> {
        this.logger.debug('stat', p)
<<<<<<< HEAD
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
=======
        const stats = await this.sftp.stat(p)
        return {
            name: posixPath.basename(p),
            fullPath: p,
            isDirectory: stats.type === russh.SFTPFileType.Directory,
            isSymlink: stats.type === russh.SFTPFileType.Symlink,
            mode: stats.permissions ?? 0,
            size: stats.size,
            modified: new Date((stats.mtime ?? 0) * 1000),
        }
    }

    async open (p: string, mode: number): Promise<SFTPFileHandle> {
        this.logger.debug('open', p, mode)
        const handle = await this.sftp.open(p, mode)
        return new SFTPFileHandle(handle)
    }

    async rmdir (p: string): Promise<void> {
        await this.sftp.removeDirectory(p)
    }

    async mkdir (p: string): Promise<void> {
        await this.sftp.createDirectory(p)
>>>>>>> upstream/master
    }

    async rename (oldPath: string, newPath: string): Promise<void> {
        this.logger.debug('rename', oldPath, newPath)
<<<<<<< HEAD
        await window['require']('electron').ipcRenderer.invoke('sftp:rename', { sessionId: this.sftpSessionId, oldPath, newPath })
    }

    async unlink (p: string): Promise<void> {
        await window['require']('electron').ipcRenderer.invoke('sftp:delete', { sessionId: this.sftpSessionId, path: p })
=======
        await this.sftp.rename(oldPath, newPath)
    }

    async unlink (p: string): Promise<void> {
        await this.sftp.removeFile(p)
>>>>>>> upstream/master
    }

    async chmod (p: string, mode: string|number): Promise<void> {
        this.logger.debug('chmod', p, mode)
<<<<<<< HEAD
        const numericMode = typeof mode === 'string' ? parseInt(mode, 8) : mode
        await window['require']('electron').ipcRenderer.invoke('sftp:chmod', { sessionId: this.sftpSessionId, path: p, mode: numericMode })
=======
        await this.sftp.chmod(p, mode)
>>>>>>> upstream/master
    }

    async upload (path: string, transfer: FileUpload): Promise<void> {
        this.logger.info('Uploading into', path)
<<<<<<< HEAD

        try {
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

            const base64Data = Buffer.from(result).toString('base64')

            await window['require']('electron').ipcRenderer.invoke('sftp:upload', {
                sessionId: this.sftpSessionId,
                path: path,
                data: base64Data,
            })

            transfer.close()
        } catch (e) {
            transfer.cancel()
=======
        const tempPath = path + '.tabby-upload'
        try {
            const handle = await this.open(tempPath, russh.OPEN_WRITE | russh.OPEN_CREATE)
            while (true) {
                const chunk = await transfer.read()
                if (!chunk.length) {
                    break
                }
                await handle.write(chunk)
            }
            await handle.close()
            await this.unlink(path).catch(() => null)
            await this.rename(tempPath, path)
            transfer.close()
        } catch (e) {
            transfer.cancel()
            this.unlink(tempPath).catch(() => null)
>>>>>>> upstream/master
            throw e
        }
    }

    async download (path: string, transfer: FileDownload): Promise<void> {
        this.logger.info('Downloading', path)
        try {
<<<<<<< HEAD
            const base64Data = await window['require']('electron').ipcRenderer.invoke('sftp:download', {
                sessionId: this.sftpSessionId,
                path: path,
            })

            const buffer = Buffer.from(base64Data, 'base64')
            await transfer.write(buffer)
            transfer.close()
=======
            const handle = await this.open(path, russh.OPEN_READ)
            while (true) {
                const chunk = await handle.read()
                if (!chunk.length) {
                    break
                }
                await transfer.write(chunk)
            }
            transfer.close()
            handle.close()
>>>>>>> upstream/master
        } catch (e) {
            transfer.cancel()
            throw e
        }
    }
<<<<<<< HEAD
=======

    private _makeFile (p: string, entry: russh.SFTPDirectoryEntry): SFTPFile {
        return {
            fullPath: p,
            name: posixPath.basename(p),
            isDirectory: entry.metadata.type === russh.SFTPFileType.Directory,
            isSymlink: entry.metadata.type === russh.SFTPFileType.Symlink,
            mode: entry.metadata.permissions ?? 0,
            size: entry.metadata.size,
            modified: new Date((entry.metadata.mtime ?? 0) * 1000),
        }
    }
>>>>>>> upstream/master
}
