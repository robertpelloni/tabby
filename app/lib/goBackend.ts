import { spawn, ChildProcess } from 'child_process'
import { EventEmitter } from 'events'
import * as path from 'path'
import split2 from 'split2'

let nextId = 1

interface RequestPending {
    resolve: (val: any) => void
    reject: (err: any) => void
}

export class GoBackendService extends EventEmitter {
    private proc: ChildProcess | null = null
    private pending = new Map<number, RequestPending>()

    start(appPath: string) {
        // Find binary
        const binName = process.platform === 'win32' ? 'tabby-backend.exe' : 'tabby-backend'

        // In dev, it's at <root>/build/tabby-backend
        // In prod, it should be in the resources folder
        const isDev = process.env.TABBY_DEV === '1' || process.env.TABBY_DEV === 'true'
        let binPath = path.join(appPath, '..', 'build', binName)
        if (!isDev && process.resourcesPath) {
            binPath = path.join(process.resourcesPath, binName)
        }

        console.log('[go-backend] Spawning', binPath)

        this.proc = spawn(binPath, [], {
            stdio: ['pipe', 'pipe', 'pipe']
        })

        this.proc.on('error', err => {
            console.error('[go-backend] Failed to start Go backend:', err)
            this.proc = null
        })

        this.proc.stdout.pipe(split2()).on('data', (line: string) => {
            try {
                const msg = JSON.parse(line)
                if (msg.id !== undefined && msg.id !== null) {
                    // Response
                    const p = this.pending.get(msg.id)
                    if (p) {
                        this.pending.delete(msg.id)
                        if (msg.error) {
                            p.reject(new Error(msg.error.message))
                        } else {
                            p.resolve(msg.result)
                        }
                    }
                } else if (msg.method) {
                    // Notification
                    this.emit(msg.method, msg.params)
                }
            } catch (err) {
                console.error('[go-backend] Parse error:', err, line)
            }
        })

        this.proc.stderr.pipe(split2()).on('data', (line: string) => {
            console.log('[go-backend] STDERR:', line)
        })

        this.proc.on('exit', (code) => {
            console.log('[go-backend] Exited with code', code)
            this.proc = null
        })
    }

    async request(method: string, params?: any): Promise<any> {
        if (!this.proc) {
            throw new Error('Go backend not running')
        }

        const id = nextId++
        const payload = JSON.stringify({
            jsonrpc: '2.0',
            id,
            method,
            params
        })

        return new Promise((resolve, reject) => {
            this.pending.set(id, { resolve, reject })
            this.proc!.stdin.write(payload + '\n')
        })
    }
}

export const goBackend = new GoBackendService()
