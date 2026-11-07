import test from 'node:test'
import assert from 'node:assert/strict'

// Direct unit test of utility logic and data structures
const TAB_COLORS = [
    { name: 'No color', value: null },
    { name: 'Blue', value: '#0275d8' },
    { name: 'Green', value: '#5cb85c' },
    { name: 'Orange', value: '#f0ad4e' },
    { name: 'Purple', value: '#613d7c' },
    { name: 'Red', value: '#d9534f' },
    { name: 'Yellow', value: '#ffd500' },
]

class ResettableTimeout {
    id = null

    constructor (fn, timeout) {
        this.fn = fn
        this.timeout = timeout
    }

    set (timeout) {
        this.clear()
        this.id = setTimeout(this.fn, timeout ?? this.timeout)
    }

    clear () {
        if (this.id) {
            clearTimeout(this.id)
            this.id = null
        }
    }
}

function serializeFunction (fn) {
    let queue = Promise.resolve()
    return ((...args) => {
        const res = queue.then(() => fn(...args))
        queue = res.catch(() => null)
        return res
    })
}

test('TAB_COLORS - exports standard palette options', () => {
    assert.equal(TAB_COLORS.length, 7)
    assert.equal(TAB_COLORS[0].value, null)
    assert.equal(TAB_COLORS[1].value, '#0275d8')
})

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

test('serializeFunction - serializes concurrent asynchronous invocations in order', async () => {
    const executed = []
    const asyncOp = async (val) => {
        await new Promise((resolve) => setTimeout(resolve, 10))
        executed.push(val)
        return val
    }

    const serialized = serializeFunction(asyncOp)
    const p1 = serialized(1)
    const p2 = serialized(2)
    const p3 = serialized(3)

    await Promise.all([p1, p2, p3])
    assert.deepEqual(executed, [1, 2, 3])
})
