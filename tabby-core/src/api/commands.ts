<<<<<<< HEAD
import slugify from 'slugify'
=======
>>>>>>> upstream/master
import { BaseTabComponent } from '../components/baseTab.component'
import { MenuItemOptions } from './menu'
import { ToolbarButton } from './toolbarButtonProvider'

export enum CommandLocation {
    LeftToolbar = 'left-toolbar',
    RightToolbar = 'right-toolbar',
    StartPage = 'start-page',
<<<<<<< HEAD
    TabHeaderMenu = 'tab-header-menu',
    TabBodyMenu = 'tab-body-menu',
}

export class Command {
    id: string
    label: string
    fullLabel?: string
    locations: CommandLocation[]
    run?: () => Promise<any>
=======
}

export class Command {
    id?: string
    label: string
    sublabel?: string
    locations?: CommandLocation[]
    run: () => Promise<void>
>>>>>>> upstream/master

    /**
     * Raw SVG icon code
     */
    icon?: string

<<<<<<< HEAD
    weight?: number

    parent?: string

    group?: string

    checked?: boolean

    static fromToolbarButton (button: ToolbarButton): Command {
        const command = new Command()
        command.id = `legacy:${slugify(button.title)}`
=======
    /**
     * Optional Touch Bar icon ID
     */
    touchBarNSImage?: string

    /**
     * Optional Touch Bar button label
     */
    touchBarTitle?: string

    weight?: number

    static fromToolbarButton (button: ToolbarButton): Command {
        const command = new Command()
>>>>>>> upstream/master
        command.label = button.title
        command.run = async () => button.click?.()
        command.icon = button.icon
        command.locations = [CommandLocation.StartPage]
        if ((button.weight ?? 0) <= 0) {
            command.locations.push(CommandLocation.LeftToolbar)
        }
        if ((button.weight ?? 0) > 0) {
            command.locations.push(CommandLocation.RightToolbar)
        }
<<<<<<< HEAD
=======
        command.touchBarNSImage = button.touchBarNSImage
        command.touchBarTitle = button.touchBarTitle
>>>>>>> upstream/master
        command.weight = button.weight
        return command
    }

<<<<<<< HEAD
    static fromMenuItem (item: MenuItemOptions): Command[] {
        if (item.type === 'separator') {
            return []
        }
        const commands: Command[] = [{
            id: `legacy:${slugify(item.commandLabel ?? item.label).toLowerCase()}`,
            label: item.commandLabel ?? item.label,
            run: async () => item.click?.(),
            locations: [CommandLocation.TabBodyMenu, CommandLocation.TabHeaderMenu],
            checked: item.checked,
        }]
        for (const submenu of item.submenu ?? []) {
            commands.push(...Command.fromMenuItem(submenu).map(x => ({
                ...x,
                id: `${commands[0].id}:${slugify(x.label).toLowerCase()}`,
                parent: commands[0].id,
            })))
        }
        return commands
=======
    static fromMenuItem (item: MenuItemOptions): Command {
        const command = new Command()
        command.label = item.commandLabel ?? item.label ?? ''
        command.sublabel = item.sublabel
        command.run = async () => item.click?.()
        return command
>>>>>>> upstream/master
    }
}

export interface CommandContext {
    tab?: BaseTabComponent,
}

/**
 * Extend to add commands
 */
export abstract class CommandProvider {
    abstract provide (context: CommandContext): Promise<Command[]>
}
