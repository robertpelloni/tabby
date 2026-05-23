import { Injector } from '@angular/core'
import { Frontend, SearchOptions, SearchState } from './frontend'
import { BaseTerminalProfile } from '../api/interfaces'
import { marked } from 'marked'
import DOMPurify from 'dompurify'

const AnsiToHtml = require('ansi-to-html')

export class BlockFrontend extends Frontend {
    private container: HTMLElement
    private currentBlock: HTMLElement
    private ansiConverter: any

    constructor (injector: Injector) {
        super(injector)
        this.ansiConverter = new AnsiToHtml({ fg: '#cccccc', bg: '#1e1e1e', escapeXML: true })
    }

    async attach (host: HTMLElement, profile: BaseTerminalProfile): Promise<void> {
        this.container = document.createElement('div')
        this.container.style.display = 'flex'
        this.container.style.flexDirection = 'column'
        this.container.style.height = '100%'
        this.container.style.width = '100%'
        this.container.style.overflowY = 'auto'
        this.container.style.backgroundColor = '#1e1e1e'
        this.container.style.color = '#cccccc'
        this.container.style.fontFamily = 'monospace'

        host.appendChild(this.container)
        this.createNewBlock()
        setTimeout(() => {
            this.ready.next()
            this.ready.complete()
        }, 100)
    }

    private createNewBlock() {
        this.currentBlock = document.createElement('div')
        this.currentBlock.className = 'terminal-block'
        this.currentBlock.style.padding = '8px'
        this.currentBlock.style.borderBottom = '1px solid #333'
        this.currentBlock.style.position = 'relative'

        const outputContainer = document.createElement('div')
        outputContainer.className = 'block-output'

        const actionsContainer = document.createElement('div')
        actionsContainer.className = 'block-actions'
        actionsContainer.style.position = 'absolute'
        actionsContainer.style.top = '5px'
        actionsContainer.style.right = '5px'
        actionsContainer.style.display = 'none'

        const copyCmdBtn = document.createElement('button')
        copyCmdBtn.innerText = '📋 Copy Command'
        copyCmdBtn.style.background = '#44475a'
        copyCmdBtn.style.color = '#f8f8f2'
        copyCmdBtn.style.border = 'none'
        copyCmdBtn.style.borderRadius = '4px'
        copyCmdBtn.style.padding = '4px 8px'
        copyCmdBtn.style.cursor = 'pointer'
        copyCmdBtn.style.fontSize = '12px'
        copyCmdBtn.style.marginRight = '5px'

        copyCmdBtn.onclick = () => {
            // First line is generally the prompt + command
            const firstLine = outputContainer.innerText.split('\n')[0].replace(/^\$ /, '').trim()
            navigator.clipboard.writeText(firstLine);
            const originalText = copyCmdBtn.innerText;
            copyCmdBtn.innerText = '✅ Copied!';
            setTimeout(() => { copyCmdBtn.innerText = originalText }, 2000);
        }

        const copyBtn = document.createElement('button')
        copyBtn.innerText = '📋 Copy Output'
        copyBtn.style.background = '#44475a'
        copyBtn.style.color = '#f8f8f2'
        copyBtn.style.border = 'none'
        copyBtn.style.borderRadius = '4px'
        copyBtn.style.padding = '4px 8px'
        copyBtn.style.cursor = 'pointer'
        copyBtn.style.fontSize = '12px'
        copyBtn.style.marginRight = '5px'

        copyBtn.onclick = () => {
            // Output is everything after first line
            const outputLines = outputContainer.innerText.split('\n').slice(1).join('\n').trim()
            navigator.clipboard.writeText(outputLines);
            const originalText = copyBtn.innerText;
            copyBtn.innerText = '✅ Copied!';
            setTimeout(() => { copyBtn.innerText = originalText }, 2000);
        }

        const explainBtn = document.createElement('button')
        explainBtn.innerText = '✨ Explain Error'
        explainBtn.style.background = '#44475a'
        explainBtn.style.color = '#f8f8f2'
        explainBtn.style.border = 'none'
        explainBtn.style.borderRadius = '4px'
        explainBtn.style.padding = '4px 8px'
        explainBtn.style.cursor = 'pointer'
        explainBtn.style.fontSize = '12px'

        explainBtn.onclick = async () => {
            explainBtn.innerText = 'Analyzing...'
            try {
                const firstLine = outputContainer.innerText.split('\n')[0].replace(/^\$ /, '').trim()
                const outputLines = outputContainer.innerText.split('\n').slice(1).join('\n').trim()
                const response = await window['require']('electron').ipcRenderer.invoke('ai:explainError', {
                    command: firstLine || 'unknown',
                    errorOutput: outputLines || 'unknown error'
                });

                const explanationDiv = document.createElement('div')
                explanationDiv.style.background = '#282a36'
                explanationDiv.style.borderLeft = '3px solid #ffb86c'
                explanationDiv.style.padding = '8px'
                explanationDiv.style.marginTop = '8px'
                explanationDiv.innerHTML = response.explanation.replace(/\n/g, '<br/>')
                this.currentBlock.appendChild(explanationDiv)
            } catch (e) {
                console.error(e)
            } finally {
                explainBtn.innerText = '✨ Explain Error'
            }
        }

        actionsContainer.appendChild(copyCmdBtn)
        actionsContainer.appendChild(copyBtn)
        actionsContainer.appendChild(explainBtn)

        this.currentBlock.appendChild(outputContainer)
        this.currentBlock.appendChild(actionsContainer)

        this.currentBlock.onmouseenter = () => { actionsContainer.style.display = 'block' }
        this.currentBlock.onmouseleave = () => { actionsContainer.style.display = 'none' }

        this.container.appendChild(this.currentBlock)
    }

    detach (host: HTMLElement): void {
        if (this.container && this.container.parentElement) {
            this.container.parentElement.removeChild(this.container)
        }
    }

    getSelection (): string {
        return window.getSelection()?.toString() || ''
    }

    copySelection (): void {
        document.execCommand('copy')
    }

    selectAll (): void {}

    clearSelection (): void {
        window.getSelection()?.removeAllRanges()
    }

    focus (): void {
        // Focus logic will go to the IDE input box later
    }

    async write (data: string): Promise<void> {
        // Intercept WaveTerm-like OSC sequences
        if (data.includes('\x1b]1337;WaveTermWidget=')) {
            this.renderWidget(data)
        } else {
            // Detect common shell prompts and rotate blocks
            // Simple heuristic: if a line ends with $, # or > followed by a space
            const hasPrompt = /[\$#>]\s$/.test(data) || data.includes('\r\n$ ') || data.includes('\r\n# ') || data.includes('\r\n> ')

            if (hasPrompt && this.currentBlock && (this.currentBlock.querySelector('.block-output')?.innerHTML || '').length > 0) {
                 this.createNewBlock()
            }

            // Render basic ANSI control codes to HTML using ansi-to-html
            const span = document.createElement('span')
            span.innerHTML = this.ansiConverter.toHtml(data).replace(/\r\n/g, '<br/>').replace(/\n/g, '<br/>')
            (this.currentBlock.querySelector('.block-output') || this.currentBlock).appendChild(span)
        }
        this.container.scrollTop = this.container.scrollHeight
    }

    private renderWidget(data: string) {
        // Extract widget JSON
        const match = data.match(/\x1b\]1337;WaveTermWidget=([^\x07]+)\x07/)
        if (match && match[1]) {
            try {
                const widget = JSON.parse(match[1])
                const widgetContainer = document.createElement('div')
                widgetContainer.style.background = '#252526'
                widgetContainer.style.padding = '10px'
                widgetContainer.style.margin = '10px 0'
                widgetContainer.style.border = '1px solid #444'
                widgetContainer.style.borderRadius = '4px'

                if (widget.type === 'markdown') {
                    // Render markdown widget
                    const rawHtml = marked.parse(widget.content) as string;
                    const cleanHtml = DOMPurify.sanitize(rawHtml);
                    widgetContainer.innerHTML = `<div class="markdown-widget" style="font-family: sans-serif;">${cleanHtml}</div>`
                } else if (widget.type === 'image') {
                    // Render image widget
                    const cleanSrc = DOMPurify.sanitize(widget.content, { ALLOWED_TAGS: [] });
                    // Provide a safe default if sanitization stripped everything, else output image tag
                    if (cleanSrc) {
                         widgetContainer.innerHTML = `<div class="image-widget" style="text-align: center;"><img src="${cleanSrc}" style="max-width: 100%; border-radius: 4px;" /></div>`
                    } else {
                         widgetContainer.innerHTML = `<div style="font-family: sans-serif; color: #ff5555;">Invalid Image URL provided.</div>`
                    }
                } else if (widget.type === 'iframe') {
                    // Render iframe widget
                    const cleanSrc = DOMPurify.sanitize(widget.content, { ALLOWED_TAGS: [] });
                    if (cleanSrc) {
                         widgetContainer.innerHTML = `<div class="iframe-widget" style="width: 100%; height: 400px;"><iframe src="${cleanSrc}" style="width: 100%; height: 100%; border: none; border-radius: 4px;"></iframe></div>`
                    } else {
                         widgetContainer.innerHTML = `<div style="font-family: sans-serif; color: #ff5555;">Invalid Iframe URL provided.</div>`
                    }
                } else {
                    widgetContainer.innerHTML = `<div style="font-family: sans-serif;"><h3>Unknown Widget Type</h3><pre>${JSON.stringify(widget)}</pre></div>`
                }

                (this.currentBlock.querySelector('.block-output') || this.currentBlock).appendChild(widgetContainer)
                this.createNewBlock() // Prepare for next command
            } catch (e) {
                console.error("Failed to parse widget JSON", e)
            }
        }
    }

    clear (): void {
        this.container.innerHTML = ''
        this.createNewBlock()
    }

    visualBell (): void {
        this.bell.next()
    }

    scrollToTop (): void {
        this.container.scrollTop = 0
    }

    scrollLines (amount: number): void {
        this.container.scrollTop += amount * 20
    }

    scrollPages (pages: number): void {
        this.container.scrollTop += pages * this.container.clientHeight
    }

    scrollToBottom (): void {
        this.container.scrollTop = this.container.scrollHeight
    }

    configure (profile: BaseTerminalProfile): void {
        // Handle config updates
    }

    setZoom (zoom: number): void {
        this.container.style.setProperty("zoom", zoom.toString())
    }

    findNext (term: string, searchOptions?: SearchOptions): SearchState { return { resultCount: 0 } }
    findPrevious (term: string, searchOptions?: SearchOptions): SearchState { return { resultCount: 0 } }
    cancelSearch (): void {}

    saveState (): any { return null }
    restoreState (state: string): void {}

    supportsBracketedPaste (): boolean { return false }
    isAlternateScreenActive (): boolean { return false }

    destroy (): void {
        super.destroy()
        if (this.container) {
            this.container.innerHTML = ''
        }
    }
}
