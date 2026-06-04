import { Component } from '@angular/core'
import { DomSanitizer } from '@angular/platform-browser'
<<<<<<< HEAD
import { firstBy } from 'thenby'
=======
>>>>>>> upstream/master
import { HomeBaseService } from '../services/homeBase.service'
import { CommandService } from '../services/commands.service'
import { Command, CommandLocation } from '../api/commands'

/** @hidden */
@Component({
    selector: 'start-page',
    templateUrl: './startPage.component.pug',
    styleUrls: ['./startPage.component.scss'],
})
export class StartPageComponent {
    version: string
    commands: Command[] = []

    constructor (
        private domSanitizer: DomSanitizer,
        public homeBase: HomeBaseService,
        commands: CommandService,
    ) {
        commands.getCommands({}).then(c => {
<<<<<<< HEAD
            this.commands = c.filter(x => x.locations.includes(CommandLocation.StartPage))
            this.commands.sort(firstBy(x => x.weight ?? 0))
=======
            this.commands = c.filter(x => x.locations?.includes(CommandLocation.StartPage))
>>>>>>> upstream/master
        })
    }

    sanitizeIcon (icon?: string): any {
        return this.domSanitizer.bypassSecurityTrustHtml(icon ?? '')
    }

    // eslint-disable-next-line @typescript-eslint/explicit-module-boundary-types
    buttonsTrackBy (_, btn: Command): any {
        return btn.label + btn.icon
    }
}
