import test from 'node:test'
import assert from 'node:assert/strict'
import { getKeyName, getKeystrokeName, metaKeyName, altKeyName } from '../../tabby-core/src/services/hotkeys.util.mjs'

test('getKeyName - maps modifier keys correctly', () => {
    assert.equal(getKeyName({ eventName: 'keydown', key: 'Control' }), 'Ctrl')
    assert.equal(getKeyName({ eventName: 'keydown', key: 'Shift' }), 'Shift')
    assert.equal(getKeyName({ eventName: 'keydown', key: 'Alt' }), altKeyName)
    assert.equal(getKeyName({ eventName: 'keydown', key: 'Meta' }), metaKeyName)
})

test('getKeyName - handles mouse buttons and wheel events', () => {
    assert.equal(getKeyName({ eventName: 'mouseup', button: 1 }), 'MiddleClick')
    assert.equal(getKeyName({ eventName: 'mouseup', button: 0 }), 'Mouse')
    assert.equal(getKeyName({ eventName: 'wheel', deltaY: -50, deltaX: 0 }), 'WheelUp')
    assert.equal(getKeyName({ eventName: 'wheel', deltaY: 50, deltaX: 0 }), 'WheelDown')
})

test('getKeyName - parses latin letters and code replacements', () => {
    assert.equal(getKeyName({ eventName: 'keydown', key: 'p', code: 'KeyP' }), 'P')
    assert.equal(getKeyName({ eventName: 'keydown', key: '', code: 'ArrowUp' }), 'Up')
    assert.equal(getKeyName({ eventName: 'keydown', key: '', code: 'Digit1' }), '1')
    assert.equal(getKeyName({ eventName: 'keydown', key: '', code: 'Equal' }), '=')
    assert.equal(getKeyName({ eventName: 'keydown', key: '', code: 'Minus' }), '-')
})

test('getKeystrokeName - orders modifiers strictly before non-modifier keys', () => {
    assert.equal(getKeystrokeName(['P', 'Ctrl', 'Shift']), 'Ctrl-Shift-P')
    assert.equal(getKeystrokeName(['A', altKeyName, 'Ctrl']), `Ctrl-${altKeyName}-A`)
})
