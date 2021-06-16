import { Exception } from './exception';
import { Subject, from } from 'rxjs';
import { takeUntil } from 'rxjs/operators';

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
            if (this._index == -1) {
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

export interface ChannelResult<T> {
    ok: boolean
    data: T
}
export interface WriteChannel<T> {
    close()
    write(data: T): Promise<undefined>
}
export interface ReadChannel<T> {
    read(): Promise<ChannelResult<T>>
}

export class Channel<T> implements WriteChannel<T>, ReadChannel<T>{
    private _signalWrite = new Subject<boolean>()
    private _signalRead = new Subject<boolean>()
    private _closed: boolean
    private _datas: Array<T>
    private _index: number
    private _length: number

    /**
     * 緩衝節點數量 最小爲1
     * @param size 
     */
    constructor(size?: number) {
        if (typeof size != "number" || isNaN(size) || size < 1) {
            size = 1
        } else {
            size = Math.floor(size)
        }
        this._datas = new Array<T>(size)
        this._index = 0
        this._length = 0
    }
    close(): boolean {
        if (this._closed) {
            return true
        }
        this._closed = true
        this._signalWrite.complete()
        this._signalRead.complete()
        return false
    }
    private _write(data: T): boolean {
        const datas = this._datas
        const length = this._length
        if (length == datas.length) {
            return false
        }
        let i = this._index + length
        if (i >= datas.length) {
            i -= datas.length
        }
        datas[i] = data
        this._length++
        return true
    }
    write(data: T): Promise<undefined> {
        return new Promise<undefined>((resolve, reject) => {
            if (this._closed) {
                reject(`channel closed`)
                return
            }
            if (this._write(data)) {
                this._signalWrite.next(true)
                resolve(undefined)
                return
            }
            const completer = new Completer<undefined>()
            this._signalRead.pipe(
                takeUntil(from(completer.promise))
            ).subscribe({
                next: () => {
                    if (this._write(data)) {
                        completer.resolve()
                        this._signalWrite.next(true)
                        resolve(undefined)
                        return
                    }
                },
                complete: () => {
                    completer.resolve()
                    reject(`send to closed channel`)
                },
            })

        })
    }
    private _read(): ChannelResult<T> {
        if (!this._length) {
            return {
                ok: false,
                data: null,
            }
        }
        const datas = this._datas
        const data = datas[this._index]
        const result = {
            ok: true,
            data: data,
        }
        this._length--
        this._index++
        if (this._index == datas.length) {
            this._index = 0
        }
        return result
    }
    read(): Promise<ChannelResult<T>> {
        return new Promise<ChannelResult<T>>((resolve, reject) => {
            const result = this._read()
            if (result.ok) {
                this._signalRead.next(true)
                resolve(result)
                return
            }
            if (this._closed) {
                this._signalRead.next(true)
                resolve(result)
                return
            }
            const completer = new Completer<boolean>()
            this._signalWrite.pipe(
                takeUntil(from(completer.promise))
            ).subscribe({
                next: () => {
                    const result = this._read()
                    if (result.ok) {
                        this._signalRead.next(true)
                        completer.resolve(true)
                        resolve(result)
                        return
                    }
                },
                complete: () => {
                    completer.resolve(true)
                    resolve(this._read())
                },
            })
        })
    }
}