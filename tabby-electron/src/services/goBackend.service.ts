/**
 * GoBackendService — Angular service for Tabby's Go backend
 * Communicates via JSON-RPC 2.0 over stdin/stdout with the Go backend process.
 *
 * This service mirrors the full Go backend API:
 * - SSH: connect, startShell, resize, write, close, listConnections
 * - SSH Port Forwarding: addForward, removeForward, listForwards
 * - SSH Auth: verifyHostKey, keyboardInteractiveResp
 * - SFTP: open, list, download, upload, delete, rename, mkdir, stat, chmod, readlink, symlink, close
 * - PTY: spawn, resize, write, kill
 * - Serial: open, write, close, listPorts
 */
import { Injectable, OnDestroy } from '@angular/core'
import { Subject, Observable } from 'rxjs'
import { spawn, ChildProcess } from 'child_process'
import * as path from 'path'

// ---- SSH Types ----
export interface GoSSHConnectParams {
    host: string
    port?: number
    user: string
    auth: {
        type: 'password' | 'publicKey' | 'agent' | 'keyboardInteractive' | 'none'
        password?: string
        privateKey?: string
        privateKeyPaths?: string[]
        agentSocketPath?: string
        agentType?: string
    }
    keepaliveInterval?: number
    keepaliveCountMax?: number
    readyTimeout?: number
    agentForward?: boolean
    x11?: boolean
    x11Display?: string
    jumpHost?: GoSSHConnectParams
    algorithms?: {
        hmac?: string[]
        kex?: string[]
        cipher?: string[]
        serverHostKey?: string[]
        compression?: string[]
    }
    proxyCommand?: string
    socksProxyHost?: string
    socksProxyPort?: number
    httpProxyHost?: string
    httpProxyPort?: number
    environment?: Record<string, string>
    verifyHostKey?: boolean
    knownHostsPath?: string
    skipBanner?: boolean
    password?: string
}

export interface GoSSHSessionParams {
    connectionId: string
    columns: number
    rows: number
    terminal?: string
    command?: string
    agentForward?: boolean
    x11?: boolean
}

export interface GoSSHConnectionResult {
    connectionId: string
    serverVersion: string
    remoteAddress: string
    banner?: string
    authMethods: string[]
}

export interface GoSSHSessionResult {
    sessionId: string
}

// ---- Port Forwarding Types ----
export type PortForwardType = 'local' | 'remote' | 'dynamic'

export interface GoPortForwardParams {
    connectionId: string
    type: PortForwardType
    host: string
    port: number
    targetAddress?: string
    targetPort?: number
}

export interface GoPortForwardResult {
    forwardId: string
}

export interface GoPortForwardInfo {
    id: string
    type: PortForwardType
    host: string
    port: number
    targetAddress?: string
    targetPort?: number
    active: boolean
}

// ---- SFTP Types ----
export interface GoSFTPFile {
    name: string
    fullPath: string
    size: number
    mode: number
    modTime: string
    isDir: boolean
    isSymlink: boolean
}

export interface GoSFTPTransferResult {
    bytesTransferred: number
}

// ---- Notification Types ----
export interface GoDataNotification {
    connectionId?: string
    sessionId?: string
    ptyId?: string
    serialId?: string
    data: string // Base64-encoded
}

export interface GoExitNotification {
    connectionId?: string
    sessionId?: string
    ptyId?: string
    serialId?: string
    exitCode?: number
    signal?: string
}

export interface GoHostKeyPromptNotification {
    connectionId: string
    host: string
    port: number
    keyType: string
    fingerprint: string
    keyBytes?: string
}

export interface GoKeyboardInteractiveNotification {
    connectionId: string
    name: string
    instruction: string
    prompts: { prompt: string; echo: boolean }[]
}

export interface GoBannerNotification {
    connectionId: string
    message: string
}

export interface GoServiceMessageNotification {
    connectionId: string
    message: string
}

export interface GoPortForwardEventNotification {
    connectionId: string
    forwardId: string
    eventType: string
    message?: string
}

@Injectable({ providedIn: 'root' })
export class GoBackendService implements OnDestroy {
    private process: ChildProcess | null = null
    private requestId = 0
    private pendingRequests = new Map<number, {
        resolve: (value: any) => void
        reject: (reason: any) => void
    }>()
    private buffer = ''

    private dataSubject = new Subject<GoDataNotification>()
    readonly onData: Observable<GoDataNotification> = this.dataSubject.asObservable()

    private exitSubject = new Subject<GoExitNotification>()
    readonly onExit: Observable<GoExitNotification> = this.exitSubject.asObservable()

    private hostKeyPromptSubject = new Subject<GoHostKeyPromptNotification>()
    readonly onHostKeyPrompt: Observable<GoHostKeyPromptNotification> = this.hostKeyPromptSubject.asObservable()

    private keyboardInteractiveSubject = new Subject<GoKeyboardInteractiveNotification>()
    readonly onKeyboardInteractive: Observable<GoKeyboardInteractiveNotification> = this.keyboardInteractiveSubject.asObservable()

    private bannerSubject = new Subject<GoBannerNotification>()
    readonly onBanner: Observable<GoBannerNotification> = this.bannerSubject.asObservable()

    private serviceMessageSubject = new Subject<GoServiceMessageNotification>()
    readonly onServiceMessage: Observable<GoServiceMessageNotification> = this.serviceMessageSubject.asObservable()

    private portForwardEventSubject = new Subject<GoPortForwardEventNotification>()
    readonly onPortForwardEvent: Observable<GoPortForwardEventNotification> = this.portForwardEventSubject.asObservable()

    private _running = false
    get running(): boolean { return this._running }

    async start(binaryPath?: string): Promise<void> {
        if (this._running) return
        const exePath = binaryPath || this.getDefaultBinaryPath()

        return new Promise<void>((resolve, reject) => {
            try {
                this.process = spawn(exePath, [], {
                    stdio: ['pipe', 'pipe', 'pipe'],
                    env: { ...process.env },
                })

                this.process.on('error', (err) => {
                    console.error('[GoBackend] Failed to start:', err)
                    this._running = false
                    reject(err)
                })

                this.process.on('exit', (code) => {
                    console.log(`[GoBackend] Process exited with code ${code}`)
                    this._running = false
                    this.rejectAllPending(`Process exited with code ${code}`)
                })

                this.process.stdout!.on('data', (data: Buffer) => {
                    this.handleStdoutData(data.toString())
                })

                this.process.stderr!.on('data', (data: Buffer) => {
                    console.log(`[GoBackend] stderr: ${data.toString().trim()}`)
                })

                this._running = true
                console.log('[GoBackend] Started successfully')
                resolve()
            } catch (err) {
                reject(err)
            }
        })
    }

    stop(): void {
        if (this.process) {
            this.process.kill()
            this.process = null
        }
        this._running = false
        this.rejectAllPending('Backend stopped')
    }

    ngOnDestroy(): void { this.stop() }

    // ---- SSH Methods ----
    async sshConnect(params: GoSSHConnectParams): Promise<GoSSHConnectionResult> {
        return this.call('ssh.connect', params)
    }

    async sshStartShell(params: GoSSHSessionParams): Promise<GoSSHSessionResult> {
        return this.call('ssh.startShell', params)
    }

    async sshResize(connectionId: string, sessionId: string, columns: number, rows: number): Promise<void> {
        return this.call('ssh.resize', { connectionId, sessionId, columns, rows })
    }

    async sshWrite(connectionId: string, sessionId: string, data: Buffer): Promise<void> {
        return this.call('ssh.write', { connectionId, sessionId, data: data.toString('base64') })
    }

    async sshClose(connectionId: string, sessionId?: string): Promise<void> {
        return this.call('ssh.close', { connectionId, sessionId })
    }

    async sshListConnections(): Promise<string[]> {
        return this.call('ssh.listConnections', {})
    }

    // ---- SSH Port Forwarding ----
    async sshAddForward(params: GoPortForwardParams): Promise<GoPortForwardResult> {
        return this.call('ssh.addForward', params)
    }

    async sshRemoveForward(connectionId: string, forwardId: string): Promise<void> {
        return this.call('ssh.removeForward', { connectionId, forwardId })
    }

    async sshListForwards(connectionId: string): Promise<GoPortForwardInfo[]> {
        const result = await this.call<{ forwards: GoPortForwardInfo[] }>('ssh.listForwards', { connectionId })
        return result.forwards || []
    }

    // ---- SSH Auth Callbacks ----
    async sshVerifyHostKey(connectionId: string, accepted: boolean): Promise<void> {
        return this.call('ssh.verifyHostKey', { connectionId, accepted })
    }

    async sshKeyboardInteractiveResp(connectionId: string, responses: string[]): Promise<void> {
        return this.call('ssh.keyboardInteractiveResp', { connectionId, responses })
    }

    // ---- SFTP Methods ----
    async sftpOpen(connectionId: string): Promise<string> {
        const result = await this.call<{ sessionId: string }>('sftp.open', { connectionId })
        return result.sessionId
    }

    async sftpList(sessionId: string, remotePath: string): Promise<GoSFTPFile[]> {
        return this.call('sftp.list', { sessionId, path: remotePath })
    }

    async sftpReadDir(sessionId: string, remotePath: string): Promise<GoSFTPFile[]> {
        return this.call('sftp.readDir', { sessionId, path: remotePath })
    }

    async sftpDownload(sessionId: string, remotePath: string, localPath: string): Promise<GoSFTPTransferResult> {
        return this.call('sftp.download', { sessionId, remotePath, localPath })
    }

    async sftpUpload(sessionId: string, localPath: string, remotePath: string): Promise<GoSFTPTransferResult> {
        return this.call('sftp.upload', { sessionId, localPath, remotePath })
    }

    async sftpDelete(sessionId: string, remotePath: string): Promise<void> {
        return this.call('sftp.delete', { sessionId, remotePath })
    }

    async sftpRename(sessionId: string, oldPath: string, newPath: string): Promise<void> {
        return this.call('sftp.rename', { sessionId, oldPath, newPath })
    }

    async sftpMkdir(sessionId: string, dirPath: string): Promise<void> {
        return this.call('sftp.mkdir', { sessionId, path: dirPath })
    }

    async sftpMkdirAll(sessionId: string, dirPath: string): Promise<void> {
        return this.call('sftp.mkdirAll', { sessionId, path: dirPath })
    }

    async sftpStat(sessionId: string, filePath: string): Promise<GoSFTPFile> {
        return this.call('sftp.stat', { sessionId, path: filePath })
    }

    async sftpLstat(sessionId: string, filePath: string): Promise<GoSFTPFile> {
        return this.call('sftp.lstat', { sessionId, path: filePath })
    }

    async sftpChmod(sessionId: string, filePath: string, mode: number): Promise<void> {
        return this.call('sftp.chmod', { sessionId, path: filePath, mode })
    }

    async sftpReadlink(sessionId: string, linkPath: string): Promise<string> {
        const result = await this.call<{ target: string }>('sftp.readlink', { sessionId, path: linkPath })
        return result.target
    }

    async sftpSymlink(sessionId: string, oldPath: string, newPath: string): Promise<void> {
        return this.call('sftp.symlink', { sessionId, oldPath, newPath })
    }

    async sftpRmdir(sessionId: string, dirPath: string): Promise<void> {
        return this.call('sftp.rmdir', { sessionId, path: dirPath })
    }

    async sftpClose(sessionId: string): Promise<void> {
        return this.call('sftp.close', { sessionId })
    }

    // ---- PTY Methods ----
    async ptySpawn(params: any): Promise<any> { return this.call('pty.spawn', params) }
    async ptyResize(id: string, columns: number, rows: number): Promise<void> {
        return this.call('pty.resize', { id, columns, rows })
    }
    async ptyWrite(id: string, data: Buffer): Promise<void> {
        return this.call('pty.write', { id, data: data.toString('base64') })
    }
    async ptyKill(id: string, signal?: string): Promise<void> {
        return this.call('pty.kill', { id, signal })
    }

    // ---- Serial Methods ----
    async serialOpen(params: any): Promise<any> { return this.call('serial.open', params) }
    async serialWrite(id: string, data: Buffer): Promise<void> {
        return this.call('serial.write', { id, data: data.toString('base64') })
    }
    async serialClose(id: string): Promise<void> { return this.call('serial.close', { id }) }
    async serialListPorts(): Promise<any[]> {
        const result = await this.call<{ ports: any[] }>('serial.listPorts', {})
        return result.ports || []
    }

    // ---- Internal ----
    private call<T = any>(method: string, params: any): Promise<T> {
        if (!this._running || !this.process) {
            return Promise.reject(new Error('Go backend is not running'))
        }
        const id = ++this.requestId
        return new Promise<T>((resolve, reject) => {
            this.pendingRequests.set(id, { resolve, reject })
            const data = JSON.stringify({ jsonrpc: '2.0', id, method, params }) + '\n'
            this.process!.stdin!.write(data)
        })
    }

    private handleStdoutData(data: string): void {
        this.buffer += data
        let newlineIndex: number
        while ((newlineIndex = this.buffer.indexOf('\n')) !== -1) {
            const line = this.buffer.slice(0, newlineIndex)
            this.buffer = this.buffer.slice(newlineIndex + 1)
            if (line.trim()) this.handleMessage(line.trim())
        }
    }

    private handleMessage(line: string): void {
        let message: any
        try { message = JSON.parse(line) } catch { console.error('[GoBackend] Invalid JSON:', line); return }

        if (message.id !== undefined && message.id !== null) {
            this.handleResponse(message)
        } else if (message.method) {
            this.handleNotification(message)
        }
    }

    private handleResponse(response: any): void {
        const pending = this.pendingRequests.get(response.id)
        if (!pending) { console.warn(`[GoBackend] No pending request for id ${response.id}`); return }
        this.pendingRequests.delete(response.id)
        if (response.error) {
            pending.reject(new Error(`Go backend error [${response.error.code}]: ${response.error.message}`))
        } else {
            pending.resolve(response.result)
        }
    }

    private handleNotification(notification: any): void {
        switch (notification.method) {
            case 'ssh.data':
            case 'pty.data':
            case 'serial.data':
                this.dataSubject.next(notification.params)
                break
            case 'ssh.exit':
            case 'pty.exit':
            case 'serial.exit':
                this.exitSubject.next(notification.params)
                break
            case 'ssh.hostKeyPrompt':
                this.hostKeyPromptSubject.next(notification.params)
                break
            case 'ssh.keyboardInteractive':
                this.keyboardInteractiveSubject.next(notification.params)
                break
            case 'ssh.banner':
                this.bannerSubject.next(notification.params)
                break
            case 'ssh.serviceMessage':
                this.serviceMessageSubject.next(notification.params)
                break
            case 'ssh.portForwardEvent':
                this.portForwardEventSubject.next(notification.params)
                break
            default:
                console.log(`[GoBackend] Unknown notification: ${notification.method}`)
        }
    }

    private rejectAllPending(reason: string): void {
        for (const [, pending] of this.pendingRequests) {
            pending.reject(new Error(reason))
        }
        this.pendingRequests.clear()
    }

    private getDefaultBinaryPath(): string {
        const isDev = !!process.env.TABBY_DEV
        if (isDev) return path.join(__dirname, '..', '..', 'build', 'tabby-backend')
        return path.join((process as any).resourcesPath, 'tabby-backend')
    }
}
