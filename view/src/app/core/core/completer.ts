import { Exception } from './exception';
import { isUndefined } from 'util';
export class Completer<T>{
    private _promise: Promise<T>
    private _resolve: any
    private _reject: any
    constructor() {
        this._promise = new Promise<T>((resolve, reject) => {
            this._resolve = resolve
            this._reject = reject
        });
    }
    get promise(): Promise<T> {
        return this._promise
    }
    resolve(value?: T | PromiseLike<T>) {
        this._resolve(value)
    }
    reject(reason?: any) {
        this._reject(reason);
    }
}
export class Completers {
    private _results: Array<any>
    private _errors: Array<any>
    private _promises: Array<Promise<any>>
    private _wait: number
    private _error
    private _index
    constructor(...promises: Array<Promise<any>>) {
        const count = promises.length
        this._wait = count
        this._results = new Array<any>(count)
        this._promises = promises
        this._errors = new Array<any>(count)
        this._index = -1
    }
    done(): Promise<Array<any>> {
        const completer = new Completer<Array<any>>()
        const promises = this._promises
        const count = promises.length
        for (let i = 0; i < count; i++) {
            this._done(completer, i)
        }
        return completer.promise
    }
    private _done(completer: Completer<Array<any>>, i: number) {
        const promise = this._promises[i]
        promise.then((data) => {
            this._results[i] = data
            this._errors[i] = null
        }, (e) => {
            this._results[i] = null
            this._errors[i] = e
            if (this._index != -1) {
                this._error = e
                this._index = i
            }
        }).finally(() => {
            this._wait--
            if (this._wait) {
                return
            }
            if (this._index == -1) {
                completer.resolve(this._results)
            } else {
                completer.reject(this._error)
            }
        })
    }
    get results(): Array<any> {
        return this._results
    }
    get errors(): Array<any> {
        return this._errors
    }
}

export class Mutex {
    private _completer: Completer<void>
    async lock(): Promise<void> {
        while (true) {
            if (this._completer == null) {
                this._completer = new Completer<void>()
                break
            }
            await this._completer.promise
        }
    }
    tryLock(): boolean {
        if (this._completer == null) {
            this._completer = new Completer<void>()
            return true
        }
        return false
    }
    unlock() {
        if (this._completer == null) {
            throw new Exception('not locked')
        }

        const completer = this._completer
        this._completer = null
        completer.resolve()
    }
    get isLocked(): boolean {
        if (this._completer) {
            return true
        }
        return false
    }
    get isNotLocked(): boolean {
        if (this._completer) {
            return false
        }
        return true
    }
}

export class BlobReader {
    private _seek: number = 0
    private _mutex = new Mutex()
    constructor(private blob: Blob) {
    }
    async read(size: number): Promise<ArrayBuffer> {
        size = Math.floor(size)
        if (size < 1) {
            throw new Exception(`size not support ${size}`)
        }

        await this._mutex.lock()
        let result: ArrayBuffer
        try {
            result = await this._read(size)
        } finally {
            this._mutex.unlock()
        }
        return result
    }
    private async _read(size: number): Promise<ArrayBuffer> {
        const start = this._seek
        if (start >= this.blob.size) {
            return null
        }
        let end = start + size
        if (end > this.blob.size) {
            end = this.blob.size
        }
        let blob: Blob
        if (start == 0 && end == this.blob.size) {
            blob = this.blob
        } else {
            blob = this.blob.slice(start, end)
        }
        const result = await blob.arrayBuffer()
        if (result) {
            this._seek += result.byteLength
        }
        return result
    }
}
