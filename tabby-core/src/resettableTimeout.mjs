export class ResettableTimeout {
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
