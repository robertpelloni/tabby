import { ipcRenderer } from 'electron'
import { LogService, NotificationsService } from 'tabby-core'
import { Subject, Observable } from 'rxjs'
import { Injector } from '@angular/core'
import { BaseSession, ConnectableTerminalProfile, InputProcessingOptions, InputProcessor, LoginScriptsOptions, SessionMiddleware, StreamProcessingOptions, TerminalStreamProcessor, UTF8SplitterMiddleware } from 'tabby-terminal'
import { SerialService } from './services/serial.service'

export interface SerialProfile extends ConnectableTerminalProfile {
    options: SerialProfileOptions
}

export interface SerialProfileOptions extends StreamProcessingOptions, LoginScriptsOptions {
    port: string
    baudrate: number | null
    databits: 5 | 6 | 7 | 8
    stopbits: 1 | 1.5 | 2
    parity: string
    rtscts: boolean
    xon: boolean
    xoff: boolean
    xany: boolean
    slowSend: boolean
    input: InputProcessingOptions,
}

export const BAUD_RATES = [
    110, 150, 300, 1200, 2400, 4800, 9600, 19200, 38400, 57600, 115200, 230400, 460800, 921600, 1500000,
]

export interface SerialPortInfo {
    name: string
    description?: string
}

class SlowFeedMiddleware extends SessionMiddleware {
    feedFromTerminal (data: Buffer): void {
        for (const byte of data) {
            this.outputToSession.next(Buffer.from([byte]))
        }
    }
}

export class SerialSession extends BaseSession {
    serialId: string | null = null

    get serviceMessage$ (): Observable<string> { return this.serviceMessage }
    private serviceMessage = new Subject<string>()
    private streamProcessor: TerminalStreamProcessor

    private notifications: NotificationsService
    private serialService: SerialService

    constructor (injector: Injector, public profile: SerialProfile) {
        super(injector.get(LogService).create(`serial-${profile.options.port}`))
        this.serialService = injector.get(SerialService)


        this.notifications = injector.get(NotificationsService)

        this.streamProcessor = new TerminalStreamProcessor(profile.options)
        this.middleware.push(this.streamProcessor)

        if (this.profile.options.slowSend) {
            this.middleware.unshift(new SlowFeedMiddleware())
        }

        this.middleware.push(new UTF8SplitterMiddleware())
        this.middleware.push(new InputProcessor(profile.options.input))

        this.setLoginScriptsOptions(profile.options)
    }

    async start (): Promise<void> {
        if (!this.profile.options.port) {
            const ports = await this.serialService.listPorts()
            if (ports.length > 0) {
                this.profile.options.port = ports[0].name
            }
        }

        const params = {
            port: this.profile.options.port,
            baudRate: parseInt(this.profile.options.baudrate as any) || 9600,
            dataBits: this.profile.options.databits || 8,
            stopBits: this.profile.options.stopbits || 1,
            parity: this.profile.options.parity || 'none',
            flowControl: this.profile.options.rtscts ? 'hardware' : 'none'
        }

        try {
            const result = await ipcRenderer.invoke('serial:open', params)
            this.serialId = result.id
            this.open = true
            setTimeout(() => this.streamProcessor.start())

            ipcRenderer.on('serial:data', (event, id, base64Data) => {
                if (id === this.serialId) {
                    this.emitOutput(Buffer.from(base64Data, 'base64'))
                }
            })

            ipcRenderer.on('serial:exit', (event, id) => {
                if (id === this.serialId) {
                    this.serviceMessage.next('Port closed')
                    if (this.open) {
                        this.destroy()
                    }
                }
            })

        } catch (e: any) {
            this.notifications.error(e.message)
            throw e
        }
    }

    resize (_, __) {
        this.streamProcessor.resize()
    }

    write (data: Buffer): void {
        if (this.open && this.serialId) {
            ipcRenderer.send('serial:write', this.serialId, data.toString('base64'))
        }
    }

    kill (_?: string): void {
        if (this.serialId) {
            ipcRenderer.send('serial:close', this.serialId)
        }
    }

    async destroy (): Promise<void> {
        this.open = false
        if (this.serialId) {
            ipcRenderer.send('serial:close', this.serialId)
            this.serialId = null
        }
        await super.destroy()
    }

    async getChildProcesses (): Promise<any[]> {
        return []
    }

    async gracefullyKillProcess (): Promise<void> {
        this.kill()
    }

    supportsWorkingDirectory (): boolean {
        return false
    }

    async getWorkingDirectory (): Promise<string|null> {
        return null
    }
}
