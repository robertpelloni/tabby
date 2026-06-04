<<<<<<< HEAD
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
=======
import { Component, Input, Output, EventEmitter, ViewChild, ElementRef, ChangeDetectionStrategy, OnInit, ChangeDetectorRef } from '@angular/core'
import { KeyboardInteractivePrompt } from '../session/ssh'
import { SSHProfile } from '../api'
import { PasswordStorageService } from '../services/passwordStorage.service'

@Component({
    selector: 'keyboard-interactive-auth-panel',
    templateUrl: './keyboardInteractiveAuthPanel.component.pug',
    styleUrls: ['./keyboardInteractiveAuthPanel.component.scss'],
    changeDetection: ChangeDetectionStrategy.OnPush,
})
export class KeyboardInteractiveAuthComponent implements OnInit {
    @Input() profile: SSHProfile
    @Input() prompt: KeyboardInteractivePrompt
    @Input() step = 0
    @Output() done = new EventEmitter()
    @ViewChild('input') input: ElementRef
    remember = false

    constructor (
        private passwordStorage: PasswordStorageService,
        private cdr: ChangeDetectorRef,
    ) {}

    async ngOnInit (): Promise<void> {
        const savedPassword = await this.passwordStorage.loadPassword(this.profile)
        if (savedPassword) {
            for (let i = 0; i < this.prompt.prompts.length; i++) {
                if (this.prompt.isAPasswordPrompt(i) && !this.prompt.responses[i]) {
                    this.prompt.responses[i] = savedPassword
                }
            }
            this.cdr.markForCheck()
        }
    }

    isPassword (): boolean {
        return this.prompt.isAPasswordPrompt(this.step)
    }

    shouldEcho (): boolean {
        return this.prompt.prompts[this.step].echo ?? false
    }

    previous (): void {
        if (this.step > 0) {
            this.step--
        }
        this.input.nativeElement.focus()
    }

    next (): void {
        if (this.isPassword() && this.remember) {
            this.passwordStorage.savePassword(this.profile, this.prompt.responses[this.step])
        }

        if (this.step === this.prompt.prompts.length - 1) {
            this.prompt.respond()
            this.done.emit()
            return
        }
        this.step++
        this.input.nativeElement.focus()
>>>>>>> upstream/master
    }
}
