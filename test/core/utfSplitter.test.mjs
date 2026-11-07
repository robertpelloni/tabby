import test from 'node:test'
import assert from 'node:assert/strict'

// Direct reference implementation of UTF8Splitter logic from tabby-core/src/utfSplitter.ts
const partials = [
    [0b110, 5, 0],
    [0b1110, 4, 1],
    [0b11110, 3, 2],
]

export class UTF8Splitter {
    internal = Buffer.alloc(0)

    write (data) {
        this.internal = Buffer.concat([this.internal, data])

        let keep = 0
        for (const [pattern, shift, maxOffset] of partials) {
            for (let offset = 0; offset < maxOffset + 1; offset++) {
                if (this.internal[this.internal.length - offset - 1] >> shift === pattern) {
                    keep = Math.max(keep, offset + 1)
                }
            }
        }

        const result = this.internal.slice(0, this.internal.length - keep)
        this.internal = this.internal.slice(this.internal.length - keep)
        return result
    }

    flush () {
        const result = this.internal
        this.internal = Buffer.alloc(0)
        return result
    }
}

test('UTF8Splitter - complete ASCII bytes pass through directly', () => {
    const splitter = new UTF8Splitter()
    const input = Buffer.from('hello world\n', 'utf8')
    const out = splitter.write(input)
    assert.deepEqual(out, input)
    assert.equal(splitter.flush().length, 0)
})

test('UTF8Splitter - splits and reconstructs multi-byte unicode sequences across chunks', () => {
    const splitter = new UTF8Splitter()
    const fullText = 'Tabby 🐱 🚀 Terminal'
    const fullBytes = Buffer.from(fullText, 'utf8')

    // Split at various byte offsets to simulate fragmented stream chunks
    const chunk1 = fullBytes.subarray(0, 7) // ends inside or before rocket
    const chunk2 = fullBytes.subarray(7, 13)
    const chunk3 = fullBytes.subarray(13)

    const out1 = splitter.write(chunk1)
    const out2 = splitter.write(chunk2)
    const out3 = splitter.write(chunk3)
    const flushed = splitter.flush()

    const combined = Buffer.concat([out1, out2, out3, flushed])
    assert.equal(combined.toString('utf8'), fullText)
})

test('UTF8Splitter - buffers partial multi-byte UTF-8 character until next chunk', () => {
    const splitter = new UTF8Splitter()
    // '€' is 3 bytes in UTF-8: 0xE2 0x82 0xAC
    const euro = Buffer.from('€', 'utf8')
    assert.equal(euro.length, 3)

    // Write first byte only
    const out1 = splitter.write(euro.subarray(0, 1))
    assert.equal(out1.length, 0) // Should buffer incomplete leading byte

    // Write remaining two bytes
    const out2 = splitter.write(euro.subarray(1))
    assert.deepEqual(out2, euro)

    // Flush should now be empty
    assert.equal(splitter.flush().length, 0)
})

test('UTF8Splitter - flush returns remaining buffered bytes', () => {
    const splitter = new UTF8Splitter()
    const euro = Buffer.from('€', 'utf8')

    splitter.write(euro.subarray(0, 2)) // Partial 2 bytes
    const remaining = splitter.flush()
    assert.deepEqual(remaining, euro.subarray(0, 2))
    assert.equal(splitter.flush().length, 0)
})
