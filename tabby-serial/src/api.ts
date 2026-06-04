import { HostAppService, LogService, NotificationsService, Platform } from 'tabby-core'
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
        private hostApp: HostAppService

    constructor (injector: Injector, public profile: SerialProfile) {
        super(injector.get(LogService).create(`serial-${profile.options.port}`))
        this.serialService = injector.get(SerialService)
                this.hostApp = injector.get(HostAppService)


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

    private idGen = () => Math.random().toString(36).substring(7);

    private handleSerialData = (_event: any, id: string, base64Data: string) => {
        if (id === this.serialId) {
            this.emitOutput(Buffer.from(base64Data, 'base64'));
        }
    };

    private handleSerialExit = (_event: any, id: string) => {
        if (id === this.serialId) {
            this.serviceMessage.next('Port closed');
            if (this.open) {
                this.destroy();
            }
        }
    };

    async start (): Promise<void> {
        if (!this.profile.options.port) {
            const ports = await this.serialService.listPorts();
            if (ports.length > 0) {
                this.profile.options.port = ports[0].name;
            }
        }

        if (this.hostApp.platform === Platform.Web) {
            // Restore WebSerial for Browser
            const serial = this.serialId = new (this.serialService.detectBinding() as any)({
                path: this.profile.options.port,
                autoOpen: false,
                baudRate: parseInt(this.profile.options.baudrate as any),
                dataBits: this.profile.options.databits,
                stopBits: this.profile.options.stopbits,
                parity: this.profile.options.parity,
                rtscts: this.profile.options.rtscts,
                xon: this.profile.options.xon,
                xoff: this.profile.options.xoff,
                xany: this.profile.options.xany,
            })
            let connected = false
            await new Promise(async (resolve, reject) => {
                serial.on('open', () => {
                    connected = true
                    resolve(null)
                })
                serial.on('error', error => {
                    if (connected) {
                        this.notifications.error(error.message)
                    } else {
                        reject(error)
                    }
                    this.destroy()
                })
                serial.on('close', () => {
                    this.serviceMessage.next('Port closed')
                    this.destroy()
                })

                try {
                    serial.open()
                } catch (e: any) {
                    this.notifications.error(e.message)
                    reject(e)
                }
            })

            this.open = true
            setTimeout(() => this.streamProcessor.start())

            serial.on('readable', () => {
                this.emitOutput(serial.read())
            })

            serial.on('end', () => {
                if (this.open) {
                    this.destroy()
                }
            })
            return
        }

        // Electron native proxy via Go backend
        this.serialId = this.idGen();
        const params = {
            id: this.serialId,
            port: this.profile.options.port,
            baudRate: parseInt(this.profile.options.baudrate as any) || 9600,
            dataBits: this.profile.options.databits || 8,
            stopBits: this.profile.options.stopbits || 1,
            parity: this.profile.options.parity || 'none',
            flowControl: this.profile.options.rtscts ? 'hardware' : 'none'
        };

        // Register handlers BEFORE invoking open to prevent dropped initial output
        window['require']('electron').ipcRenderer.on('serial:data', this.handleSerialData);
        window['require']('electron').ipcRenderer.on('serial:exit', this.handleSerialExit);

        try {
            await window['require']('electron').ipcRenderer.invoke('serial:open', params);
            this.open = true;
            setTimeout(() => this.streamProcessor.start());
        } catch (e: any) {
            this.notifications.error(e.message);
            // Cleanup on failure
            window['require']('electron').ipcRenderer.off('serial:data', this.handleSerialData);
            window['require']('electron').ipcRenderer.off('serial:exit', this.handleSerialExit);
            throw e;
        }
    }

    resize (_, __) {
        this.streamProcessor.resize()
    }

    write (data: Buffer): void {
        if (!this.open || !this.serialId) return;
        if (this.hostApp.platform === Platform.Web) {
            (this.serialId as any).write(data);
        } else {
            window['require']('electron').ipcRenderer.send('serial:write', this.serialId, data.toString('base64'));
        }
    }

    kill (_?: string): void {
        if (this.serialId) {
            window['require']('electron').ipcRenderer.send('serial:close', this.serialId)
        }
    }

    async destroy (): Promise<void> {
        this.open = false
        if (this.serialId) {
            if (this.hostApp.platform === Platform.Web) {
                (this.serialId as any).close();
            } else {
                window['require']('electron').ipcRenderer.send('serial:close', this.serialId);
                window['require']('electron').ipcRenderer.off('serial:data', this.handleSerialData);
                window['require']('electron').ipcRenderer.off('serial:exit', this.handleSerialExit);
            }
            this.serialId = null;
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
