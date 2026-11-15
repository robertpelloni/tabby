export const metaKeyName = {
    darwin: '⌘',
    win32: 'Win',
    linux: 'Super',
}[process.platform] || 'Super'

export const altKeyName = {
    darwin: '⌥',
    win32: 'Alt',
    linux: 'Alt',
}[process.platform] || 'Alt'

const REGEX_LATIN_KEYNAME = /^[A-Za-z]$/

export function getKeyName (event) {
    if (event.eventName === 'mouseup' || event.eventName === 'auxclick') {
        if (event.button === 1) {
            return 'MiddleClick'
        }
        return 'Mouse'
    }

    if (event.eventName === 'wheel') {
        const deltaX = event.deltaX ?? 0
        const deltaY = event.deltaY ?? 0
        if (Math.abs(deltaY) >= Math.abs(deltaX) && deltaY !== 0) {
            return deltaY < 0 ? 'WheelUp' : 'WheelDown'
        }
        if (deltaX !== 0) {
            return deltaX < 0 ? 'WheelLeft' : 'WheelRight'
        }
        return 'Wheel'
    }

    let key
    if (event.key === 'Control') {
        key = 'Ctrl'
    } else if (event.key === 'Meta') {
        key = metaKeyName
    } else if (event.key === 'Alt') {
        key = altKeyName
    } else if (event.key === 'Shift') {
        key = 'Shift'
    } else if (event.key === '`') {
        key = '`'
    } else if (event.key === '~') {
        key = '~'
    } else {
        key = event.code ?? ''
        if (event.key && REGEX_LATIN_KEYNAME.test(event.key)) {
            key = event.key.toUpperCase()
        } else {
            key = key.replace('Key', '')
            key = key.replace('Arrow', '')
            key = key.replace('Digit', '')
            key = {
                Comma: ',',
                Period: '.',
                Slash: '/',
                Backslash: '\\',
                IntlBackslash: '`',
                Minus: '-',
                Equal: '=',
                Semicolon: ';',
                Quote: '\'',
                BracketLeft: '[',
                BracketRight: ']',
            }[key] ?? key
        }
    }
    return key
}

export function getKeystrokeName (keys) {
    const strictOrdering = ['Ctrl', metaKeyName, altKeyName, 'Shift']
    keys = [
        ...strictOrdering.map(x => keys.find(k => k === x)).filter(x => !!x),
        ...keys.filter(k => !strictOrdering.includes(k)),
    ]
    return keys.join('-')
}
