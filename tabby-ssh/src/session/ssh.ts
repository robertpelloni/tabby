import colors from 'ansi-colors'
import stripAnsi from 'strip-ansi'
import { Injector } from '@angular/core'
import { NgbModal } from '@ng-bootstrap/ng-bootstrap'
import { PromptModalComponent, LogService, Logger, HostAppService, Platform } from 'tabby-core'
import { Subject, Observable } from 'rxjs'
import { HostKeyPromptModalComponent } from '../components/hostKeyPromptModal.component'
import { PasswordStorageService } from '../services/passwordStorage.service'
import { SFTPSession } from './sftp'
import { SSHProfile } from '../api'
import { ForwardedPort } from './forwards'

const WINDOWS_OPENSSH_AGENT_PIPE = '\\.\pipe\openssh-ssh-agent'

export interface Prompt {
    prompt: string
    echo?: boolean
}

export class KeyboardInteractivePrompt {
    name: string
    instruction: string
    prompts: Prompt[]

    constructor (name: string, instruction: string, prompts: Prompt[]) {
        this.name = name
        this.instruction = instruction
        this.prompts = prompts
    }
}


export class MockChannel {
    id = 0
    inner: any = null
    extendedData$ = new Subject<Buffer>()
    closed$ = new Subject<void>()
    data$ = new Subject<Buffer>()
    eof$ = new Subject<void>()

    constructor (private sessionId: string, private connectionId: string) {}

    write (data: Uint8Array): void {
        window['require']('electron').ipcRenderer.send('ssh:write', this.connectionId, this.sessionId, Buffer.from(data).toString('base64'))
    }

    resizePTY (opts: { columns: number, rows: number, pixWidth: number, pixHeight: number }): void {
        window['require']('electron').ipcRenderer.invoke('ssh:resize', {
            connectionId: this.connectionId,
            sessionId: this.sessionId,
            columns: opts.columns,
            rows: opts.rows,
        }).catch(() => {})
    }

    close (): void {
        window['require']('electron').ipcRenderer.invoke('ssh:close', { connectionId: this.connectionId, sessionId: this.sessionId }).catch(() => {})
    }

    eof (): void {
        this.close()
    }

    take (n: number) { return this as any }
    requestShell () { return Promise.resolve() }
    requestExec (cmd: string) { return Promise.resolve() }
    requestAgentForwarding () { return Promise.resolve() }
    requestPTY (terminal: string, opts: any): Promise<void> { return Promise.resolve() }
    requestWindowChange (opts: any): Promise<void> { return Promise.resolve() }
    requestX11Forwarding (opts: any) { return Promise.resolve() }
    send (data: any): Promise<void> { return Promise.resolve() }
    sendEof (): Promise<void> { return Promise.resolve() }
    destructed = false
    destruct () {}
    assertNotDestructed () {}
    setEnv (key: string, value: string) { return Promise.resolve() }
    signal (signal: string) { return Promise.resolve() }
    requestSubsystem (subsystem: string) { return Promise.resolve() }
}

export class SSHSession {
    profile: SSHProfile

    private passwordStorage: PasswordStorageService
    private ngbModal: NgbModal
    private hostApp: HostAppService
    protected logger: Logger

    connectionId: string | null = null

    private serviceMessage = new Subject<string>()
    keyboardInteractivePrompt$ = new Subject<KeyboardInteractivePrompt>()
    willDestroy$ = new Subject<void>()
    output$ = new Subject<Buffer>()
    destroyed$ = new Subject<void>()
    authUsername = ''

    forwardedPorts: ForwardedPort[] = []

    // Mock properties to satisfy sshTab.component.ts and shell.ts
    ssh: any = null
    jumpChannel: any = null
    activePrivateKey: any = null

    get serviceMessage$ (): Observable<string> { return this.serviceMessage }

    // Provide these getters for sshTab compatibility
    get open (): boolean { return this.connectionId !== null }
    get fullyAuthenticated (): boolean { return this.ssh !== null && this.ssh._authenticated }

    constructor (
        private injector: Injector,
        profile: SSHProfile,
    ) {
        this.logger = injector.get(LogService).create(`ssh-${profile.options.host}-${profile.options.port}`)
        this.profile = profile
        this.passwordStorage = injector.get(PasswordStorageService)
        this.ngbModal = injector.get(NgbModal)
        this.hostApp = injector.get(HostAppService)

        this.willDestroy$.subscribe(() => {
            for (const port of this.forwardedPorts) {
                port.stopLocalListener()
            }
        })

        // Register global IPC handlers
        window['require']('electron').ipcRenderer.on('ssh:hostKeyPrompt', this.onHostKeyPrompt)
        window['require']('electron').ipcRenderer.on('ssh:keyboardInteractive', this.onKeyboardInteractive)
        window['require']('electron').ipcRenderer.on('ssh:banner', this.onBanner)
        window['require']('electron').ipcRenderer.on('ssh:serviceMessage', this.onServiceMessage)
        window['require']('electron').ipcRenderer.on('ssh:portForwardEvent', this.onPortForwardEvent)
    }

    emitServiceMessage (msg: string) {
        this.serviceMessage.next(msg)
        this.logger.info(stripAnsi(msg))
    }

    emitKeyboardInteractivePrompt (prompt: KeyboardInteractivePrompt): void {
        this.logger.info('Keyboard-interactive auth:', prompt.name, prompt.instruction)
        this.emitServiceMessage(colors.bgBlackBright(' ') + ` Keyboard-interactive auth requested: ${prompt.name}`)
        if (prompt.instruction) {
            for (const line of prompt.instruction.split('\n')) {
                this.emitServiceMessage(line)
            }
        }
        this.keyboardInteractivePrompt$.next(prompt)
    }

    async connect (): Promise<void> {
        this.emitServiceMessage(`Connecting to ${this.profile.options.host}`)

        let authType: 'password' | 'publicKey' | 'agent' | 'keyboardInteractive' | 'none' = 'none'
        let password = ''
        const privateKey = ''
        let privateKeyPaths: string[] = []
        let agentSocketPath = ''
        let agentType = ''

        if (this.profile.options.auth === 'password' || this.profile.options.auth === 'keyboardInteractive') {
            const savedPassword = await this.passwordStorage.loadPassword(this.profile)
            if (savedPassword) {
                authType = this.profile.options.auth
                password = savedPassword
            } else {
                authType = 'keyboardInteractive' // Go backend will emit interactive prompt if needed
            }
        } else if (this.profile.options.auth === 'publicKey') {
            authType = 'publicKey'
            if (this.profile.options.privateKeys && this.profile.options.privateKeys.length > 0) {
                privateKeyPaths = this.profile.options.privateKeys
            }
        } else if (this.profile.options.auth === 'agent') {
            authType = 'agent'
            if (this.hostApp.platform === Platform.Windows) {
                agentSocketPath = WINDOWS_OPENSSH_AGENT_PIPE
                agentType = 'pageant'
            } else {
                agentSocketPath = process.env.SSH_AUTH_SOCK || ''
            }
        }

        const connectParams = {
            host: this.profile.options.host,
            port: this.profile.options.port || 22,
            user: this.profile.options.user,
            auth: {
                type: authType,
                password,
                privateKey,
                privateKeyPaths,
                agentSocketPath,
                agentType,
            },
            keepaliveInterval: this.profile.options.keepaliveInterval,
            keepaliveCountMax: this.profile.options.keepaliveCountMax,
            readyTimeout: this.profile.options.readyTimeout,
            agentForward: (this.profile.options as any).forwardAgent,
            x11: this.profile.options.x11,
            proxyJump: (this.profile.options as any).proxyJump || '',
        }

        try {
            const res = await window['require']('electron').ipcRenderer.invoke('ssh:connect', connectParams)
            this.connectionId = res.connectionId

            if (res.serverVersion) {
                this.emitServiceMessage(`Multiplexed into active connection to ${this.profile.options.host}`)
            } else {
                this.emitServiceMessage(`Connected to ${this.profile.options.host}`)
            }

            // Fake the authenticated client to satisfy sshTab.component.ts
            this.ssh = { _authenticated: true }
        } catch (e: any) {
            this.emitServiceMessage(colors.bgRed.black(' X ') + ` Could not connect: ${e.message}`)
            throw e
        }
    }

    async handleAuth (): Promise<any> {
        // Go backend handles auth internally during connect(). If interactive prompts are needed,
        // it suspends and emits IPC events which we answer.
        // Returning true mock satisfies calling components.
        return this.ssh
    }

    async openShellChannel (options: { x11: boolean }): Promise<MockChannel> {
        if (!this.connectionId) { throw new Error('Not connected') }
        const shellRes = await window['require']('electron').ipcRenderer.invoke('ssh:startShell', {
            connectionId: this.connectionId,
            x11: options.x11,
        })
        const channel = new MockChannel(shellRes.sessionId, this.connectionId)

        // Listen to this specific session's data
        window['require']('electron').ipcRenderer.on('ssh:data', (_event: any, connId: string, sessId: string, data: string) => {
            if (connId === this.connectionId && sessId === shellRes.sessionId) {
                channel.data$.next(Buffer.from(data, 'base64'))
            }
        })
        window['require']('electron').ipcRenderer.on('ssh:exit', (_event: any, connId: string, sessId: string) => {
            if (connId === this.connectionId && sessId === shellRes.sessionId) {
                channel.eof$.next()
            }
        })

        return channel
    }

    async getSftp (): Promise<SFTPSession> {
        if (!this.connectionId) { throw new Error('Not connected') }
        const sftpSessionId = await window['require']('electron').ipcRenderer.invoke('sftp:open', this.connectionId)
        return new SFTPSession(sftpSessionId, this.injector)
    }

    async getSftpPath (): Promise<string> {
        return '.'
    }

    async destroy (): Promise<void> {
        this.willDestroy$.next()
        this.willDestroy$.complete()
        this.destroyed$.next()
        this.destroyed$.complete()
        this.output$.complete()
        this.serviceMessage.complete()

        if (this.connectionId) {
            window['require']('electron').ipcRenderer.invoke('ssh:close', { connectionId: this.connectionId }).catch(() => {})
        }

        window['require']('electron').ipcRenderer.off('ssh:hostKeyPrompt', this.onHostKeyPrompt)
        window['require']('electron').ipcRenderer.off('ssh:keyboardInteractive', this.onKeyboardInteractive)
        window['require']('electron').ipcRenderer.off('ssh:banner', this.onBanner)
        window['require']('electron').ipcRenderer.off('ssh:serviceMessage', this.onServiceMessage)
        window['require']('electron').ipcRenderer.off('ssh:portForwardEvent', this.onPortForwardEvent)
    }

    async addPortForward (fw: ForwardedPort): Promise<void> {
        if (!this.connectionId) { throw new Error('Not connected') }
        try {
            await window['require']('electron').ipcRenderer.invoke('ssh:addForward', {
                connectionId: this.connectionId,
                type: fw.type,
                host: fw.host,
                port: fw.port,
                targetAddress: fw.targetAddress,
                targetPort: fw.targetPort,
            })
            this.forwardedPorts.push(fw)
            this.emitServiceMessage(colors.bgGreen.black(' <-> ') + ` Forwarded ${fw}`)
        } catch (e) {
            this.emitServiceMessage(colors.bgRed.black(' X ') + ` Failed to forward port ${fw}: ${e}`)
            throw e
        }
    }

    async removePortForward (fw: ForwardedPort): Promise<void> {
        if (!this.connectionId) { return }
        await window['require']('electron').ipcRenderer.invoke('ssh:removeForward', { connectionId: this.connectionId, forwardId: fw.port.toString() })
        this.forwardedPorts = this.forwardedPorts.filter(x => x !== fw)
        this.emitServiceMessage(`Stopped forwarding ${fw}`)
    }

    // --- IPC Listeners ---

    private onHostKeyPrompt = async (_event: any, params: any) => {
        if (params.connectionId === this.connectionId) {
            const modal = this.ngbModal.open(HostKeyPromptModalComponent, { backdrop: 'static' })
            modal.componentInstance.host = params.host
            modal.componentInstance.type = params.keyType
            modal.componentInstance.digest = params.fingerprint

            try {
                await modal.result
                await window['require']('electron').ipcRenderer.invoke('ssh:verifyHostKey', this.connectionId, true)
                // this.knownHosts.add(params.host, params.keyType, params.keyBytes)
            } catch {
                await window['require']('electron').ipcRenderer.invoke('ssh:verifyHostKey', this.connectionId, false)
            }
        }
    }

    private onKeyboardInteractive = async (_event: any, params: any) => {
        if (params.connectionId === this.connectionId) {
            const prompt = new KeyboardInteractivePrompt(params.name, params.instruction, params.prompts)
            this.emitKeyboardInteractivePrompt(prompt)

            const responses: string[] = []
            for (const p of params.prompts) {
                const modal = this.ngbModal.open(PromptModalComponent)
                modal.componentInstance.prompt = p.prompt
                modal.componentInstance.password = !p.echo
                try {
                    const response = await modal.result
                    responses.push(response ? response.value : '')
                } catch {
                    responses.push('')
                }
            }
            await window['require']('electron').ipcRenderer.invoke('ssh:keyboardInteractiveResp', this.connectionId, responses)
        }
    }

    private onBanner = (_event: any, connectionId: string, message: string) => {
        if (connectionId === this.connectionId) {
            this.emitServiceMessage(message)
        }
    }

    private onServiceMessage = (_event: any, connectionId: string, message: string) => {
        if (connectionId === this.connectionId) {
            this.emitServiceMessage(message)
        }
    }

    private onPortForwardEvent = (_event: any, params: any) => {
        if (params.connectionId === this.connectionId) {
            this.emitServiceMessage(`[${params.eventType}] Port Forward ${params.forwardId}: ${params.message || ''}`)
        }
    }

    private refCount = 0
    ref (): void { this.refCount++ }
    unref (): void {
        this.refCount--
        if (this.refCount === 0) {
            this.destroy()
        }
    }
}
