/* eslint-disable @typescript-eslint/explicit-module-boundary-types */
import { Injectable } from '@angular/core'

export abstract class Logger {
    constructor (protected name: string) { }
    debug (...args: any[]): void {
        this.doLog('debug', args)
    }
    info (...args: any[]): void {
        this.doLog('info', args)
    }
    warn (...args: any[]): void {
        this.doLog('warn', args)
    }
    error (...args: any[]): void {
        this.doLog('error', args)
    }
    log (...args: any[]): void {
        this.doLog('log', args)
    }

    protected abstract doLog (level: string, args: any[]): void
}

export class ConsoleLogger extends Logger {
    protected doLog (level: string, args: any[]): void {
        const consoleArgs = [`%c[${this.name}]`, 'color: #aaa']
        for (let i = 0; i < args.length; i++) {
            consoleArgs.push(args[i])
        }
        (console[level] as any).apply(console, consoleArgs)
    }
}

@Injectable({ providedIn: 'root' })
export class LogService {
    create (name: string): Logger {
        return new ConsoleLogger(name)
    }
}
