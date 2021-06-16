import { Completer, Completers, Channel, WriteChannel, ReadChannel } from 'src/app/core/core/completer';
import { sizeString, MB } from 'src/app/core/core/utils';
import { HttpClient } from '@angular/common/http';
import { NetCommand } from '../command';
import { ServerAPI } from 'src/app/core/core/api';
import { MatDialog } from '@angular/material/dialog';
import { ExistChoiceComponent } from '../exist-choice/exist-choice.component';
import * as SparkMD5 from 'spark-md5';

const ChunkSize = 5 * MB
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
    isSkip(): boolean {
        return this.status == Status.Skip
    }
    hash: string
    chunks: Array<Chunk>
}
export class Workers {
    private _worker: Worker
    constructor() {
        this._worker = new Worker(new URL('./hash.worker', import.meta.url), { type: 'module' })
    }
    done(chunks: Array<Chunk>): Promise<string> {
        const worker = this._worker
        return new Promise<string>((resolve, reject) => {
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
                    } else if (Array.isArray(data.vals) && typeof data.hash === "string") {
                        for (let i = 0; i < data.vals.length; i++) {
                            chunks[i].hash = data.vals[i]
                        }
                        resolve(data.hash)
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
        if (this.upload) {
            this.upload.close()
            this.upload = null
        }
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
                this.file.progress = 100
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
            ServerAPI.v1.fs.getOne(this.httpClient, [this.root.root, this.root.dir + `/${this.file.file.name}`, `whash`], {
                params: {
                    'chunk': ChunkSize.toString(),
                    'size': this.file.file.size.toString(),
                },
                responseType: 'json',
            }),
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
            return this._upload()
        }
        if (this.style == NetCommand.YesAll) {
            return this._upload()
        }
        if (this.style == NetCommand.SkipAll) {
            this.file.status = Status.Skip
            return
        }
        const style = await this.matDialog.open(ExistChoiceComponent, {
            data: this.file.file.name,
            disableClose: true,
        }).afterClosed().toPromise()
        if (this.isClosed) {
            return
        }
        if (typeof style === "number") {
            switch (style) {
                case NetCommand.Yes:
                    return this._upload()
                case NetCommand.YesAll:
                    this.style = style
                    return this._upload()
                case NetCommand.SkipAll:
                    this.style = style
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
        let hash: string
        if (typeof Worker !== 'undefined') {
            hash = await this._webWorkers(chunks)
        } else {
            const spark = new SparkMD5()
            for (let i = 0; i < chunks.length; i++) {
                await chunks[i].calculate()
                spark.append(chunks[i].hash)
            }
            hash = spark.end()
        }
        this.file.chunks = chunks
        this.file.hash = hash
        console.log(`calculate hash`, hash, (new Date().getTime() - last.getTime()) / 1000)
        return hash
    }
    async _webWorkers(chunks: Array<Chunk>): Promise<string> {
        const workers = this._getWorkers()
        return workers.done(chunks)
    }
    workers: Workers
    private _getWorkers(): Workers {
        if (this.workers) {
            return this.workers
        }
        this.workers = new Workers()
        return this.workers
    }
    private async _upload() {
        const upload = new Upload(this.root, this.file, this.httpClient)
        this.upload = upload
        return upload.done()
    }
    private upload: Upload
}
export class Chunk {
    hash: string
    constructor(public file: File, public index: number, public start: number, public end: number) {
    }
    async calculate() {
        const data = await this.file.slice(this.start, this.end).arrayBuffer()
        const spark = new SparkMD5.ArrayBuffer()
        spark.append(data)
        this.hash = spark.end()
    }
}
class Upload {
    private _chCheck = new Channel<ICheck>()
    private _closed: boolean
    private _num = 0
    constructor(private root: Data, private file: UploadFile, private httpClient: HttpClient) {
    }
    close(): boolean {
        if (this._closed) {
            return false
        }
        this._closed = true
        this._close()
        this._resolve()
        return true
    }
    get isClosed(): boolean {
        return this._closed
    }
    get isNotClosed(): boolean {
        return !this._closed
    }
    private _close() {
        if (this._chCheck) {
            this._chCheck.close()
            this._chCheck = null
        }
    }
    private _resolve() {
        if (this._completer) {
            const completer = this._completer
            this._completer = null
            completer.resolve()
        }
    }
    private _reject(e) {
        if (this._completer) {
            const completer = this._completer
            this._completer = null
            completer.reject(e)
        }
        this._close()
    }
    private _completer: Completer<undefined>
    done(): Promise<undefined> {
        const completer = new Completer<undefined>()

        const check = this._chCheck
        // 同時 3 個 併發 驗證 chunks
        for (let i = 0; i < 3; i++) {
            this._readCheck(check)
        }
        this._writeCheck(check)

        this._completer = completer
        return completer.promise
    }
    private async _readCheck(ch: ReadChannel<ICheck>) {
        try {
            while (this.isNotClosed) {
                const check = await ch.read()
                if (!check.ok || this.isClosed) {
                    break
                }
                await this._check(check.data)
            }
        } catch (e) {
            console.warn(e)
            this._reject(e)
        }
    }
    private async _check(ch: ICheck) {
        const results = await ServerAPI.v1.fs.getOne<Array<string>>(this.httpClient, [this.root.root, this.root.dir + `/${this.file.file.name}`, `wchunk`], {
            params: {
                start: ch.start.toString(),
                count: ch.count.toString(),
            },
        })
        if (this.isClosed) {
            return
        }
        if (!Array.isArray(results) || results.length != ch.count) {
            console.warn('check chunks unknow results', results)
            throw 'check chunks unknow results'
        }
        for (let i = 0; i < results.length; i++) {
            const hash = results[i]
            const index = ch.start + i
            const chunk = this.file.chunks[index]
            await this._put(chunk, hash)
        }
    }
    private async _writeCheck(ch: WriteChannel<ICheck>) {
        try {
            const chunks = this.file.chunks
            const length = chunks.length
            let start = 0
            const count = 100
            while (start != length) {
                let end = start + count
                if (end > length) {
                    end = length
                }
                await ch.write({
                    start: start,
                    count: end - start,
                })
                start = end
            }
        } catch (e) {
            this._reject(e)
        } finally {
            ch.close()
        }
    }
    private async _put(chunk: Chunk, hash: string) {
        if (chunk.hash == hash) {
            console.log(`${chunk.file.name} chunk match`, chunk.index)
            await this._update()
            return
        }
        console.log(`${chunk.file.name} chunk put`, chunk.index)
        const body = await chunk.file.slice(chunk.start, chunk.end).arrayBuffer()
        if (this.isClosed) {
            return
        }
        await ServerAPI.v1.fs.putOne(this.httpClient,
            [this.root.root, this.root.dir + `/${this.file.file.name}`, `wchunk`, chunk.index],
            body,
        )
        if (this.isClosed) {
            return
        }
        await this._update()
    }
    private async _update() {
        this._num++
        const file = this.file
        const chunks = file.chunks
        const val = Math.floor(this._num * 100 / chunks.length)
        if (file.progress != val) {
            file.progress = val
        }

        if (this._num == chunks.length) {
            await this._merge()
        }
    }
    private async _merge() {
        try {
            await ServerAPI.v1.fs.putOne(this.httpClient,
                [this.root.root, this.root.dir + `/${this.file.file.name}`, `merge`],
                {
                    hash: this.file.hash,
                    count: this.file.chunks.length,
                },
            )
            if (this.isClosed) {
                return
            }
            this.file.status = Status.Ok
            this._resolve()
        } catch (e) {
            this._reject(e)
        }
    }
}
interface ICheck {
    start: number
    count: number
}
