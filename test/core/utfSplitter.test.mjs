import test from 'node:test'
import assert from 'node:assert/strict'
import { UTF8Splitter } from '../../tabby-core/src/utfSplitter.js'

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
    const chunk1 = fullBytes.subarray(0, 7)
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
    const euro = Buffer.from('€', 'utf8')
    assert.equal(euro.length, 3)

    // Write first byte only
    const out1 = splitter.write(euro.subarray(0, 1))
    assert.equal(out1.length, 0)

    // Write remaining two bytes
    const out2 = splitter.write(euro.subarray(1))
    assert.deepEqual(out2, euro)

    // Flush should now be empty
    assert.equal(splitter.flush().length, 0)
})

test('UTF8Splitter - flush returns remaining buffered bytes', () => {
    const splitter = new UTF8Splitter()
    const euro = Buffer.from('€', 'utf8')

    splitter.write(euro.subarray(0, 2))
    const remaining = splitter.flush()
    assert.deepEqual(remaining, euro.subarray(0, 2))
    assert.equal(splitter.flush().length, 0)
})
