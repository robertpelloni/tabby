<<<<<<< HEAD
export type MenuItemOptions = {
=======
export interface MenuItemOptions {
    type?: 'normal' | 'separator' | 'submenu' | 'checkbox' | 'radio'
    label?: string
>>>>>>> upstream/master
    sublabel?: string
    enabled?: boolean
    checked?: boolean
    submenu?: MenuItemOptions[]
    click?: () => void

    /** @hidden */
    commandLabel?: string
<<<<<<< HEAD
} & ({
    type: 'separator',
    label?: string,
} | {
    type?: 'normal' | 'submenu' | 'checkbox' | 'radio',
    label: string,
})
=======
}
>>>>>>> upstream/master
