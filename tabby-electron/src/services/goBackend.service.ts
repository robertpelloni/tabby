/**
 * GoBackendService — Angular service that communicates with the Tabby Go backend
 * via JSON-RPC 2.0 over a child process's stdin/stdout.
 *
 * This service spawns the Go backend binary and provides a TypeScript API
 * that mirrors the Go backend's JSON-RPC methods. It can be used as an
 * alternative to the russh-based SSH implementation.
 *
 * Usage:
 *   constructor(private goBackend: GoBackendService) {}
 *
 *   // Connect via Go backend
 *   const conn = await this.goBackend.sshConnect({
 *     host: 'example.com',
 *     port: 22,
 *     user: 'testuser',
 *     auth: { type: 'password', password: 'secret' }
 *   })
 *
 * Architecture:
 *   Angular Service → JSON-RPC → Child Process (Go binary) → SSH/PTY/Serial
 */
import { Injectable, OnDestroy } from '@angular/core'
import { Subject, Observable } from 'rxjs'
import { spawn, ChildProcess } from 'child_process'
import * as path from 'path'

// ---- Type definitions matching Go backend API types ----

/** SSH connection parameters */
export interface GoSSHConnectParams {
    host: string
    port?: number
    user: string
    auth: {
        type: 'password' | 'publicKey' | 'agent' | 'keyboardInteractive'
        password?: string
        privateKey?: string
        privateKeyPaths?: string[]
        agentSocketPath?: string
    }
    keepaliveInterval?: number
    keepaliveCountMax?: number
    readyTimeout?: number
    agentForward?: boolean
    x11?: boolean
    jumpHost?: GoSSHConnectParams
    algorithms?: {
        hmac?: string[]
        kex?: string[]
        cipher?: string[]
        serverHostKey?: string[]
        compression?: string[]
    }
    proxyCommand?: string
    environment?: Record<string, string>
}

/** SSH session parameters */
export interface GoSSHSessionParams {
    connectionId: string
    columns: number
    rows: number
    terminal?: string
    command?: string
}

/** SSH connection result */
export interface GoSSHConnectionResult {
    connectionId: string
    serverVersion: string
    remoteAddress: string
    banner?: string
    authMethods: string[]
}

/** SSH session result */
export interface GoSSHSessionResult {
    sessionId: string
}

/** Data notification from the Go backend */
export interface GoDataNotification {
    connectionId?: string
    sessionId?: string
    ptyId?: string
    serialId?: string
    data: string // Base64-encoded
}

/** Exit notification from the Go backend */
export interface GoExitNotification {
    connectionId?: string
    sessionId?: string
    ptyId?: string
    serialId?: string
    exitCode?: number
    signal?: string
}

/** SFTP file listing entry */
export interface GoSFTPFile {
    name: string
    size: number
    mode: number
    modTime: string
    isDir: boolean
}

/** JSON-RPC request */
interface JSONRPCRequest {
    jsonrpc: '2.0'
    id: number
    method: string
    params?: any
}

/** JSON-RPC response */
interface JSONRPCResponse {
    jsonrpc: '2.0'
    id: number
    result?: any
    error?: {
        code: number
        message: string
        data?: any
    }
}

/** JSON-RPC notification */
interface JSONRPCNotification {
    jsonrpc: '2.0'
    method: string
    params?: any
}

/** Callback for data notifications */
export type DataCallback = (notification: GoDataNotification) => void

/** Callback for exit notifications */
export type ExitCallback = (notification: GoExitNotification) => void

@Injectable({ providedIn: 'root' })
export class GoBackendService implements OnDestroy {
    private process: ChildProcess | null = null
    private requestId = 0
    private pendingRequests = new Map<number, {
        resolve: (value: any) => void
        reject: (reason: any) => void
    }>()
    private buffer = ''

    /** Observable for SSH data notifications */
    private dataSubject = new Subject<GoDataNotification>()
    readonly onData: Observable<GoDataNotification> = this.dataSubject.asObservable()

    /** Observable for exit notifications */
    private exitSubject = new Subject<GoExitNotification>()
    readonly onExit: Observable<GoExitNotification> = this.exitSubject.asObservable()

    private _running = false

    /** Whether the Go backend process is running */
    get running(): boolean {
        return this._running
    }

    /**
     * Start the Go backend process
     * @param binaryPath Optional path to the tabby-backend binary
     */
    async start(binaryPath?: string): Promise<void> {
        if (this._running) {
            return
        }

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

                // Read stdout line by line (JSON-RPC responses)
                this.process.stdout!.on('data', (data: Buffer) => {
                    this.handleStdoutData(data.toString())
                })

                // Log stderr for debugging
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

    /**
     * Stop the Go backend process
     */
    stop(): void {
        if (this.process) {
            this.process.kill()
            this.process = null
        }
        this._running = false
        this.rejectAllPending('Backend stopped')
    }

    ngOnDestroy(): void {
        this.stop()
    }

    // ---- SSH Methods ----

    /** Connect to an SSH server */
    async sshConnect(params: GoSSHConnectParams): Promise<GoSSHConnectionResult> {
        return this.call('ssh.connect', params)
    }

    /** Start a shell session on an SSH connection */
    async sshStartShell(params: GoSSHSessionParams): Promise<GoSSHSessionResult> {
        return this.call('ssh.startShell', params)
    }

    /** Resize the terminal for an SSH session */
    async sshResize(connectionId: string, sessionId: string, columns: number, rows: number): Promise<void> {
        return this.call('ssh.resize', { connectionId, sessionId, columns, rows })
    }

    /** Write data to an SSH session */
    async sshWrite(connectionId: string, sessionId: string, data: Buffer): Promise<void> {
        return this.call('ssh.write', {
            connectionId,
            sessionId,
            data: data.toString('base64'),
        })
    }

    /** Close an SSH session or connection */
    async sshClose(connectionId: string, sessionId?: string): Promise<void> {
        return this.call('ssh.close', { connectionId, sessionId })
    }

    /** List active SSH connections */
    async sshListConnections(): Promise<string[]> {
        return this.call('ssh.listConnections', {})
    }

    // ---- SFTP Methods ----

    /** Open an SFTP session over an SSH connection */
    async sftpOpen(connectionId: string): Promise<string> {
        const result = await this.call('sftp.open', { connectionId })
        return result.sessionId
    }

    /** List files in a directory */
    async sftpList(sessionId: string, remotePath: string): Promise<GoSFTPFile[]> {
        return this.call('sftp.list', { sessionId, path: remotePath })
    }

    /** Download a file */
    async sftpDownload(sessionId: string, remotePath: string, localPath: string): Promise<{ bytesTransferred: number }> {
        return this.call('sftp.download', { sessionId, remotePath, localPath })
    }

    /** Upload a file */
    async sftpUpload(sessionId: string, localPath: string, remotePath: string): Promise<{ bytesTransferred: number }> {
        return this.call('sftp.upload', { sessionId, localPath, remotePath })
    }

    /** Delete a file or directory */
    async sftpDelete(sessionId: string, remotePath: string): Promise<void> {
        return this.call('sftp.delete', { sessionId, remotePath })
    }

    /** Rename a file or directory */
    async sftpRename(sessionId: string, oldPath: string, newPath: string): Promise<void> {
        return this.call('sftp.rename', { sessionId, oldPath, newPath })
    }

    /** Create a directory */
    async sftpMkdir(sessionId: string, dirPath: string): Promise<void> {
        return this.call('sftp.mkdir', { sessionId, path: dirPath })
    }

    /** Get file information */
    async sftpStat(sessionId: string, filePath: string): Promise<GoSFTPFile> {
        return this.call('sftp.stat', { sessionId, path: filePath })
    }

    /** Close an SFTP session */
    async sftpClose(sessionId: string): Promise<void> {
        return this.call('sftp.close', { sessionId })
    }

    // ---- PTY Methods (stubs) ----

    /** Spawn a local PTY (not yet implemented in Go backend) */
    async ptySpawn(params: any): Promise<any> {
        return this.call('pty.spawn', params)
    }

    /** Resize a PTY (not yet implemented in Go backend) */
    async ptyResize(id: string, columns: number, rows: number): Promise<void> {
        return this.call('pty.resize', { id, columns, rows })
    }

    /** Write data to a PTY (not yet implemented in Go backend) */
    async ptyWrite(id: string, data: Buffer): Promise<void> {
        return this.call('pty.write', { id, data: data.toString('base64') })
    }

    /** Kill a PTY process (not yet implemented in Go backend) */
    async ptyKill(id: string, signal?: string): Promise<void> {
        return this.call('pty.kill', { id, signal })
    }

    // ---- Serial Methods (stubs) ----

    /** Open a serial port (not yet implemented in Go backend) */
    async serialOpen(params: any): Promise<any> {
        return this.call('serial.open', params)
    }

    /** Write data to a serial port (not yet implemented in Go backend) */
    async serialWrite(id: string, data: Buffer): Promise<void> {
        return this.call('serial.write', { id, data: data.toString('base64') })
    }

    /** Close a serial port (not yet implemented in Go backend) */
    async serialClose(id: string): Promise<void> {
        return this.call('serial.close', { id })
    }

    // ---- Internal Methods ----

    /**
     * Send a JSON-RPC call and wait for the response
     */
    private call<T = any>(method: string, params: any): Promise<T> {
        if (!this._running || !this.process) {
            return Promise.reject(new Error('Go backend is not running'))
        }

        const id = ++this.requestId
        const request: JSONRPCRequest = {
            jsonrpc: '2.0',
            id,
            method,
            params,
        }

        return new Promise<T>((resolve, reject) => {
            this.pendingRequests.set(id, { resolve, reject })
            const data = JSON.stringify(request) + '\n'
            this.process!.stdin!.write(data)
        })
    }

    /**
     * Handle data from the Go backend's stdout
     * Messages are newline-delimited JSON
     */
    private handleStdoutData(data: string): void {
        this.buffer += data

        // Process complete lines
        let newlineIndex: number
        while ((newlineIndex = this.buffer.indexOf('\n')) !== -1) {
            const line = this.buffer.slice(0, newlineIndex)
            this.buffer = this.buffer.slice(newlineIndex + 1)

            if (line.trim()) {
                this.handleMessage(line.trim())
            }
        }
    }

    /**
     * Handle a single JSON-RPC message
     */
    private handleMessage(line: string): void {
        let message: any
        try {
            message = JSON.parse(line)
        } catch {
            console.error('[GoBackend] Invalid JSON:', line)
            return
        }

        if (message.id !== undefined && message.id !== null) {
            // This is a response to a request
            this.handleResponse(message as JSONRPCResponse)
        } else if (message.method) {
            // This is a notification from the server
            this.handleNotification(message as JSONRPCNotification)
        }
    }

    /**
     * Handle a JSON-RPC response
     */
    private handleResponse(response: JSONRPCResponse): void {
        const pending = this.pendingRequests.get(response.id)
        if (!pending) {
            console.warn(`[GoBackend] No pending request for id ${response.id}`)
            return
        }

        this.pendingRequests.delete(response.id)

        if (response.error) {
            pending.reject(new Error(`Go backend error [${response.error.code}]: ${response.error.message}`))
        } else {
            pending.resolve(response.result)
        }
    }

    /**
     * Handle a JSON-RPC notification from the Go backend
     */
    private handleNotification(notification: JSONRPCNotification): void {
        switch (notification.method) {
            case 'ssh.data':
                this.dataSubject.next(notification.params as GoDataNotification)
                break
            case 'ssh.exit':
            case 'pty.exit':
            case 'serial.exit':
                this.exitSubject.next(notification.params as GoExitNotification)
                break
            default:
                console.log(`[GoBackend] Unknown notification: ${notification.method}`)
        }
    }

    /**
     * Reject all pending requests (used when the backend stops)
     */
    private rejectAllPending(reason: string): void {
        for (const [id, pending] of this.pendingRequests) {
            pending.reject(new Error(reason))
        }
        this.pendingRequests.clear()
    }

    /**
     * Get the default path to the Go backend binary
     */
    private getDefaultBinaryPath(): string {
        // In development: look in build/ directory
        // In production: look in resources directory
        const isDev = !!process.env.TABBY_DEV
        if (isDev) {
            return path.join(__dirname, '..', '..', 'build', 'tabby-backend')
        }
        // Production path (inside the app resources)
        return path.join((process as any).resourcesPath, 'tabby-backend')
    }
}
