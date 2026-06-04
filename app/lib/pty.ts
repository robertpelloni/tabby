import * as nodePTY from 'node-pty'
import { v4 as uuidv4 } from 'uuid'
import { ipcMain } from 'electron'
import { Application } from './app'
<<<<<<< HEAD
import { goBackend } from './goBackend'
=======
>>>>>>> upstream/master
import { UTF8Splitter } from './utfSplitter'
import { Subject, debounceTime } from 'rxjs'

class PTYDataQueue {
    private buffers: Buffer[] = []
    private delta = 0
    private maxChunk = 1024 * 100
    private maxDelta = this.maxChunk * 5
    private flowPaused = false
    private decoder = new UTF8Splitter()
    private output$ = new Subject<Buffer>()

    constructor (private pty: nodePTY.IPty, private onData: (data: Buffer) => void) {
        this.output$.pipe(debounceTime(500)).subscribe(() => {
            const remainder = this.decoder.flush()
            if (remainder.length) {
                this.onData(remainder)
            }
        })
    }

    push (data: Buffer) {
        this.buffers.push(data)
        this.maybeEmit()
    }

    ack (length: number) {
        this.delta -= length
        this.maybeEmit()
    }

    private maybeEmit () {
        if (this.delta <= this.maxDelta && this.flowPaused) {
            this.resume()
            return
        }
        if (this.buffers.length > 0) {
            if (this.delta > this.maxDelta && !this.flowPaused) {
                this.pause()
                return
            }

            const buffersToSend = []
            let totalLength = 0
            while (totalLength < this.maxChunk && this.buffers.length) {
                totalLength += this.buffers[0].length
                buffersToSend.push(this.buffers.shift())
            }

            if (buffersToSend.length === 0) {
                return
            }

            let toSend = Buffer.concat(buffersToSend)
            if (toSend.length > this.maxChunk) {
                this.buffers.unshift(toSend.slice(this.maxChunk))
                toSend = toSend.slice(0, this.maxChunk)
            }
            this.emitData(toSend)
            this.delta += toSend.length

            if (this.buffers.length) {
                setImmediate(() => this.maybeEmit())
            }
        }
    }

    private emitData (data: Buffer) {
        const validChunk = this.decoder.write(data)
        this.onData(validChunk)
        this.output$.next(validChunk)
    }

    private pause () {
        this.pty.pause()
        this.flowPaused = true
    }

    private resume () {
        this.pty.resume()
        this.flowPaused = false
        this.maybeEmit()
    }
}

<<<<<<< HEAD



export class PTY {
    private outputQueue: PTYDataQueue;
    exited = false;
    private _pid: number = 0;
    private goBackendId: string | null = null;

    constructor (private id: string, private app: Application, ...args: any[]) {
        // Create a mock nodePTY.IPty to satisfy PTYDataQueue's pause/resume
        const mockPty: any = {
            pause: () => { goBackend.request('pty.pause', { id: this.id }).catch(() => {}) },
            resume: () => { goBackend.request('pty.resume', { id: this.id }).catch(() => {}) }
        };

        this.outputQueue = new PTYDataQueue(mockPty, data => {
            setImmediate(() => this.emit('data', data));
        });

        let command = args[0];
        let commandArgs = args[1] || [];
        const opt = args[2] || {};
        let env = opt.env || {};
        let cwd = opt.cwd || '';
        let columns = opt.cols || 80;
        let rows = opt.rows || 24;

        const options = {
            command,
            args: commandArgs,
            env,
            cwd,
            columns,
            rows
        };

        goBackend.request('pty.spawn', options).then((result: any) => {
            this._pid = result.pid;
            this.goBackendId = result.id;
            PTYManager.registerGoBackendId(this.id, this.goBackendId!);
        }).catch((err: any) => {
            console.error('Failed to spawn PTY via Go Backend', err);
            this.emit('exit', -1);
            this.exited = true;
        });
    }

    getPID (): number {
        return this._pid;
    }

    resize (columns: number, rows: number): void {
        if (this.goBackendId) {
            goBackend.request('pty.resize', { id: this.goBackendId, columns, rows }).catch(err => {
                console.error('Failed to resize PTY via Go Backend', err);
            });
=======
export class PTY {
    private pty: nodePTY.IPty
    private outputQueue: PTYDataQueue
    exited = false

    constructor (private id: string, private app: Application, ...args: any[]) {
        this.pty = (nodePTY as any).spawn(...args)
        for (const key of ['close', 'exit']) {
            (this.pty as any).on(key, (...eventArgs) => this.emit(key, ...eventArgs))
        }

        this.outputQueue = new PTYDataQueue(this.pty, data => {
            setImmediate(() => this.emit('data', data))
        })

        this.pty.onData(data => this.outputQueue.push(Buffer.from(data)))
        this.pty.onExit(() => {
            this.exited = true
        })
    }

    getPID (): number {
        return this.pty.pid
    }

    resize (columns: number, rows: number): void {
        if ((this.pty as any)._writable) {
            this.pty.resize(columns, rows)
>>>>>>> upstream/master
        }
    }

    write (buffer: Buffer): void {
<<<<<<< HEAD
        if (this.goBackendId) {
            goBackend.request('pty.write', { id: this.goBackendId, data: buffer.toString('base64') }).catch(err => {
                console.error('Failed to write PTY via Go Backend', err);
            });
=======
        if ((this.pty as any)._writable) {
            this.pty.write(buffer as any)
>>>>>>> upstream/master
        }
    }

    ackData (length: number): void {
<<<<<<< HEAD
        this.outputQueue.ack(length);
    }

    kill (signal?: string): void {
        if (this.goBackendId) {
            goBackend.request('pty.kill', { id: this.goBackendId, signal }).catch(err => {
                console.error('Failed to kill PTY via Go Backend', err);
            });
        }
    }

    private emit (event: string, ...args: any[]) {
        this.app.broadcast(`pty:${this.id}:${event}`, ...args);
    }

    handleData(data: string) {
        this.outputQueue.push(Buffer.from(data, 'base64'));
    }

    handleExit(code: number) {
        this.exited = true;
        this.emit('exit', code);
=======
        this.outputQueue.ack(length)
    }

    kill (signal?: string): void {
        this.pty.kill(signal)
    }

    private emit (event: string, ...args: any[]) {
        this.app.broadcast(`pty:${this.id}:${event}`, ...args)
>>>>>>> upstream/master
    }
}

export class PTYManager {
<<<<<<< HEAD
    private ptys: Record<string, PTY|undefined> = {};
    private static goIdToLocalId: Record<string, string> = {};

    static registerGoBackendId(localId: string, goId: string) {
        PTYManager.goIdToLocalId[goId] = localId;
    }

    init (app: Application): void {
        goBackend.on('pty.data', (params: any) => {
            const localId = PTYManager.goIdToLocalId[params.ptyId];
            if (localId && this.ptys[localId]) {
                this.ptys[localId]!.handleData(params.data);
            }
        });

        goBackend.on('pty.exit', (params: any) => {
            const localId = PTYManager.goIdToLocalId[params.ptyId];
            if (localId && this.ptys[localId]) {
                this.ptys[localId]!.handleExit(params.exitCode);
                delete PTYManager.goIdToLocalId[params.ptyId];
            }
        });

        ipcMain.on('pty:spawn', (event, ...options) => {
            const id = uuidv4().toString();
            event.returnValue = id;
            this.ptys[id] = new PTY(id, app, ...options);
        });

        ipcMain.on('pty:exists', (event, id) => {
            event.returnValue = this.ptys[id] && !this.ptys[id]!.exited;
        });

        ipcMain.on('pty:get-pid', (event, id) => {
            event.returnValue = this.ptys[id]?.getPID();
        });

        ipcMain.on('pty:resize', (_event, id, columns, rows) => {
            this.ptys[id]?.resize(columns, rows);
        });

        ipcMain.on('pty:write', (_event, id, data) => {
            this.ptys[id]?.write(Buffer.from(data));
        });

        ipcMain.on('pty:kill', (_event, id, signal) => {
            this.ptys[id]?.kill(signal);
        });

        ipcMain.on('pty:ack-data', (_event, id, length) => {
            this.ptys[id]?.ackData(length);
        });
=======
    private ptys: Record<string, PTY|undefined> = {}

    init (app: Application): void {
        ipcMain.on('pty:spawn', (event, ...options) => {
            const id = uuidv4().toString()
            event.returnValue = id
            this.ptys[id] = new PTY(id, app, ...options)
        })

        ipcMain.on('pty:exists', (event, id) => {
            event.returnValue = this.ptys[id] && !this.ptys[id].exited
        })

        ipcMain.on('pty:get-pid', (event, id) => {
            event.returnValue = this.ptys[id]?.getPID()
        })

        ipcMain.on('pty:resize', (_event, id, columns, rows) => {
            this.ptys[id]?.resize(columns, rows)
        })

        ipcMain.on('pty:write', (_event, id, data) => {
            this.ptys[id]?.write(Buffer.from(data))
        })

        ipcMain.on('pty:kill', (_event, id, signal) => {
            this.ptys[id]?.kill(signal)
        })

        ipcMain.on('pty:ack-data', (_event, id, length) => {
            this.ptys[id]?.ackData(length)
        })
>>>>>>> upstream/master
    }
}
