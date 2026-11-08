import test from 'node:test'
import assert from 'node:assert/strict'
import { ResettableTimeout } from '../../tabby-core/src/resettableTimeout.mjs'

test('ResettableTimeout - triggers callback if not cleared', async () => {
    let called = false
    const timeout = new ResettableTimeout(() => {
        called = true
    }, 20)

    timeout.set()
    assert.equal(called, false)
    await new Promise((resolve) => setTimeout(resolve, 35))
    assert.equal(called, true)
})

test('ResettableTimeout - does not trigger callback if cleared', async () => {
    let called = false
    const timeout = new ResettableTimeout(() => {
        called = true
    }, 25)

    timeout.set()
    timeout.clear()
    await new Promise((resolve) => setTimeout(resolve, 35))
    assert.equal(called, false)
})

test('ResettableTimeout - reschedules existing timeout when set() is called again', async () => {
    let counter = 0
    const timeout = new ResettableTimeout(() => {
        counter++
    }, 30)

    timeout.set()
    await new Promise((resolve) => setTimeout(resolve, 15))
    timeout.set() // reset timer before 30ms expires
    await new Promise((resolve) => setTimeout(resolve, 20))
    // 35ms passed since first set, but only 20ms since reset: shouldn't have fired yet
    assert.equal(counter, 0)
    await new Promise((resolve) => setTimeout(resolve, 20))
    // Now 40ms passed since reset: should have fired exactly once
    assert.equal(counter, 1)
})
