import { TerminalDecorator } from './decorator'
import { BaseTerminalTabComponent } from './baseTerminalTab.component'
import { Injectable } from '@angular/core'

/**
 * A wrapper to allow users to inject React/WebComponent plugins without Angular DI knowledge.
 * In a user's ~/.tabby-plugins/ dir, they can export simple functions returning DOM nodes or
 * React-rendered elements, and this decorator will dynamically mount them over the terminal.
 */
@Injectable()
export class ReactPluginDecorator extends TerminalDecorator {
    private plugins: Array<(tab: any) => HTMLElement | null> = []

    constructor() {
        super()
        this.loadUserPlugins()
    }

    private loadUserPlugins() {
        // Here we would dynamically require() from ~/.tabby-plugins/react-plugins
        // For demonstration of the Phase 5 API wrapper, we just expose the registry.
        window['tabbyReactPlugins'] = {
            register: (pluginFactory: (tab: any) => HTMLElement) => {
                this.plugins.push(pluginFactory)
            }
        }
    }

    attach(terminal: BaseTerminalTabComponent<any>): void {
        const container = document.createElement('div')
        container.className = 'react-plugins-container'
        container.style.position = 'absolute'
        container.style.top = '0'
        container.style.right = '0'
        container.style.pointerEvents = 'none' // Allow clicks to pass through to terminal
        container.style.zIndex = '100'

        this.plugins.forEach(pluginFactory => {
            try {
                const element = pluginFactory(terminal)
                if (element) {
                    element.style.pointerEvents = 'auto' // Re-enable clicks for the plugin itself
                    container.appendChild(element)
                }
            } catch (e) {
                console.error('Failed to load React plugin wrapper', e)
            }
        })

        if (container.childNodes.length > 0) {
            terminal.element.nativeElement.appendChild(container)
        }
    }
}
