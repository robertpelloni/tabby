import { Component, Input } from '@angular/core'
import { SSHProfile } from '../api'
import { KeyboardInteractivePrompt } from '../session/ssh'
import { PasswordStorageService } from '../services/passwordStorage.service'

/** @hidden */
@Component({
    selector: 'keyboard-interactive-auth-panel',
    template: require('./keyboardInteractiveAuthPanel.component.pug'),
})
export class KeyboardInteractiveAuthPanelComponent {
    @Input() profile: SSHProfile
    @Input() prompt: KeyboardInteractivePrompt

    step = 0
    savePassword = true

    constructor (public passwordStorage: PasswordStorageService) {}

    get isPasswordPrompt (): boolean {
        return this.prompt.prompts[this.step] && this.prompt.prompts[this.step].prompt.toLowerCase().includes("password")
    }

    get canSavePassword (): boolean {
        return this.isPasswordPrompt
    }

    submit () {
        if (this.savePassword) {
            // this.passwordStorage.savePassword(this.profile, this.prompt.responses[this.step])
        }
        // this.prompt.respond()
    }

    private isPromptUrl (url: string): boolean {
        try {
            const parsedUrl = new URL(url)
            return parsedUrl.protocol === 'http:' || parsedUrl.protocol === 'https:'
        } catch {
            return false
        }
    }
}
