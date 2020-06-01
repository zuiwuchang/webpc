import { Completer, Completers } from 'src/app/core/core/completer';
import { sizeString, MB } from 'src/app/core/core/utils';
import { HttpClient } from '@angular/common/http';
import { isNumber, isObject, isArray } from 'util';
import { NetCommand } from '../command';
import { buf } from 'crc-32';
import { ServerAPI } from 'src/app/core/core/api';
import { MatDialog } from '@angular/material/dialog';
import { ExistChoiceComponent } from '../exist-choice/exist-choice.component';
import * as md5 from "js-md5";

const ChunkSize = 5 * MB
const ChunkCount = 10 * 2
export interface Data {
    root: string
    dir: string
}

export enum Status {
    Nil,
    Working,
    Ok,
    Error,
    Skip,
}

export class UploadFile {
    constructor(public file: File) {
    }
    status = Status.Nil
    progress: number = 0
    error: string
    get sizeString(): string {
        return sizeString(this?.file?.size)
    }
    get key(): string {
        const file = this.file
        return `${file.size}${file.lastModified}${file.name}`
    }
    isWorking(): boolean {
        return this.status == Status.Working
    }
    isOk(): boolean {
        return this.status == Status.Ok
    }
    isError(): boolean {
        return this.status == Status.Error
    }
    hash: string
    chunks: Array<Chunk>
}
export class Workers {
    private _workers = new Array<Worker>()
    private _idle = new Array<boolean>()
    constructor(count: number) {
        const workers = new Array<Worker>()

        let worker = new Worker('./hash.worker', { type: 'module' })
        workers.push(worker)
        this._idle.push(true)

        // worker = new Worker('./hash.worker', { type: 'module' })
        // workers.push(worker)
        // this._idle.push(true)

        this._workers = workers
    }
    done(chunks: Array<Chunk>): Promise<undefined> | null {
        const count = this._workers.length
        for (let i = 0; i < count; i++) {
            if (this._idle[i]) {
                return this._done(chunks, i)
            }
        }
        this._wait = new Completer<undefined>()
        return null
    }
    private _wait: Completer<undefined>
    wait() {
        const wait = this._wait
        if (wait) {
            return wait.promise
        }
        throw `wait nil`
    }
    private _done(chunks: Array<Chunk>, i: number): Promise<undefined> {
        this._idle[i] = false
        const worker = this._workers[i]
        const completer = new Completer<undefined>()
        this._calculate(chunks, worker).then((ok) => {
            this._idle[i] = true
            const wait = this._wait
            if (wait) {
                wait.resolve()
                this._wait = null
            }
            completer.resolve()
        }, (e) => {
            this._idle[i] = true
            const wait = this._wait
            if (wait) {
                wait.resolve()
                this._wait = null
            }
            completer.reject(e)
        })
        return completer.promise
    }
    private _calculate(chunks: Array<Chunk>, worker: Worker): Promise<undefined> {
        return new Promise((resolve, reject) => {
            try {
                worker.postMessage(
                    chunks.map((chunk) => {
                        return {
                            file: chunk.file,
                            start: chunk.start,
                            end: chunk.end,
                        }
                    })
                )
            } catch (e) {
                reject(e)
                return
            }
            worker.onmessage = ({ data }) => {
                if (data) {
                    if (data.error) {
                        reject(data.error)
                    } else if (isArray(data.val)) {
                        for (let i = 0; i < data.val.length; i++) {
                            chunks[i].hash = data.val[i]
                        }
                        resolve()
                    } else {
                        console.warn('unknow worker result', data)
                        reject(`unknow worker result`)
                    }
                } else {
                    console.warn('unknow worker result', data)
                    reject(`unknow worker result`)
                }
            }
        })
    }
}

export class Uploader {
    constructor(
        private root: Data,
        private file: UploadFile,
        private httpClient: HttpClient,
        private matDialog: MatDialog,
        public style: number
    ) {
    }
    close() {
        if (this._completer) {
            const completer = this._completer
            this._completer = null
            completer.resolve()
        }
    }
    private _completer = new Completer<undefined>()
    get isClosed(): boolean {
        return this._completer ? false : true
    }
    done(): Promise<undefined> {
        const promise = this._completer.promise
        this._run().then(() => {
            if (this.isClosed) {
                return
            }
            if (this.file.status == Status.Working) {
                this.file.status = Status.Ok
            }
            this.close()
        }, (e) => {
            if (this.isClosed) {
                return
            }
            this.file.status = Status.Error
            this.file.error = e
            this.close()
        })
        return promise
    }
    async _run() {
        this.file.status = Status.Working
        this.file.error = null
        const completers = new Completers(
            ServerAPI.v1.fs.getOne(this.httpClient, [this.root.root, this.root.dir + `/${this.file.file.name}`, `whash`]),
            this._hash(),
        )
        const results = await completers.done()
        if (this.isClosed) {
            return
        }
        if (results[0] == results[1]) {
            return
        } else if (!results[0]) {
            // 服務器 不存在 直接上傳
            this._upload()
            return
        }
        if (this.style == NetCommand.YesAll) {
            this._upload()
            return
        }
        if (this.style == NetCommand.SkipAll) {
            this.file.status = Status.Skip
            return
        }
        const style = await this.matDialog.open(ExistChoiceComponent, {
            data: this.file.file.name,
            disableClose: true,
        }).afterClosed().toPromise()
        if (isNumber(style)) {
            switch (style) {
                case NetCommand.Yes:
                    await this._upload()
                    return
                case NetCommand.YesAll:
                    this.style = style
                    await this._upload()
                    return
                case NetCommand.SkipAll:
                    this.style = style
                    return
            }
        }
        this.file.status = Status.Skip
    }
    private async _hash(): Promise<string> {
        if (this.file.hash) {
            return this.file.hash
        }
        const file = this.file.file
        const size = file.size
        const chunks = new Array<Chunk>()

        let start = 0
        let index = -1
        while (start != size) {
            let end = start + ChunkSize
            if (end > size) {
                end = size
            }
            ++index
            const chunk = new Chunk(file, index, start, end)
            chunks.push(chunk)
            start = end
        }
        const last = new Date()
        if (typeof Worker !== 'undefined') {
            await this._webWorkers(chunks)
        } else {
            for (let i = 0; i < chunks.length; i++) {
                await chunks[i].calculate()
            }
        }
        this.file.chunks = chunks
        const hash = md5(chunks.map((chunk => chunk.hash)).join(","))
        this.file.hash = hash
        console.log(`calculate hash`, hash, (new Date().getTime() - last.getTime()) / 1000)
        return hash
    }
    private _getTasks(chunks: Array<Chunk>, index: number, count: number): Array<Chunk> {
        if (index >= chunks.length) {
            return null
        }
        const results = new Array<Chunk>()
        for (let i = index; i < chunks.length; i++) {
            results.push(chunks[i])
        }
        return results
    }
    async _webWorkers(chunks: Array<Chunk>): Promise<undefined> {
        const workers = this._getWorkers()
        let arrs: Array<Promise<undefined>>
        const count = ChunkCount
        let index = 0
        while (true) {
            const tasks = this._getTasks(chunks, index, count)
            if (!tasks) {
                // 沒有任務
                break
            }
            index += tasks.length

            const promise = workers.done(chunks)
            if (promise) {
                // 添加到 arrs
                if (!arrs) {
                    arrs = new Array<Promise<undefined>>()
                }
                arrs.push(promise)
                continue
            }

            // 執行 任務
            if (arrs.length == 1) {
                await arrs[0]
            } else {
                const completers = new Completers(...arrs)
                await completers.done()
            }

            // 投遞 緩存任務
            while (true) {
                await workers.wait()
                const promise = workers.done(tasks)
                if (promise) {
                    arrs = new Array<Promise<undefined>>()
                    arrs.push(promise)
                    break
                }
            }
        }
        if (arrs) {
            if (arrs.length == 1) {
                await arrs[0]
            } else {
                const completers = new Completers(...arrs)
                await completers.done()
            }
        }
        return
    }
    workers: Workers
    private _getWorkers(): Workers {
        if (this.workers) {
            return this.workers
        }
        let count = 1
        if (isObject(navigator) && isNumber(navigator.hardwareConcurrency) && navigator.hardwareConcurrency > 1) {
            //count = navigator.hardwareConcurrency
        }
        this.workers = new Workers(count)
        return this.workers
    }
    private async _upload() {
        const chunks = this.file.chunks
        console.log(chunks)
    }
}
export class Chunk {
    hash: number
    constructor(public file: File, public index: number, public start: number, public end: number) {
    }
    async calculate() {
        const data = await this.file.slice(this.start, this.end).arrayBuffer()
        this.hash = buf(new Uint8Array(data), 0)
    }
}