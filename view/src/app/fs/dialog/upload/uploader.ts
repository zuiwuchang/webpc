import { Completer } from 'src/app/core/core/completer';
import { sizeString, MB } from 'src/app/core/core/utils';
import { HttpClient } from '@angular/common/http';
import { isNumber, isObject } from 'util';
import { NetCommand } from '../command';
import { buf } from 'crc-32';
import { ServerAPI } from 'src/app/core/core/api';
import { MatDialog } from '@angular/material/dialog';
import { ExistChoiceComponent } from '../exist-choice/exist-choice.component';
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
interface Message {
    cmd: number
    error: string
    val: string
    progress: number
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
    private _crc32: number
    private _chunks: Array<Chunk>
    get chunks(): Array<Chunk> {
        return this._chunks
    }
    async _webWorkers(count: number): Promise<number> {
        const workers = new Array<Worker>()
        for (let i = 0; i < count; i++) {
            const worker = new Worker('./md5.worker', { type: 'module' })
            workers.push(worker)
            worker.postMessage(this.file)
        }

        return 0
    }
    async crc32(): Promise<number> {
        if (isNumber(this._crc32)) {
            return new Promise(function (resolve, reject) {
                resolve(this._crc32)
            })
        }
        const file = this.file
        const size = file.size
        const chunks = new Array<Chunk>()

        let start = 0
        let seed = 0
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

        let count = 4
        if (isObject(navigator) && isNumber(navigator.hardwareConcurrency) && navigator.hardwareConcurrency > 1) {
            count = navigator.hardwareConcurrency
        }
        if (typeof Worker !== 'undefined') {

            //            return this._webWorkers(count)
        } else {
            for (let i = 0; i < chunks.length; i++) {
                chunks[i].run()
            }
        }
        this._chunks = chunks
        this._crc32 = seed
        return seed
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
            completer.resolve(true)
        }
    }
    private _completer = new Completer<boolean>()
    get isClosed(): boolean {
        return this._completer ? false : true
    }
    done(): Promise<boolean> {
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
        const results = await Promise.all([
            ServerAPI.v1.fs.getOne(this.httpClient, [this.root.root, this.root.dir + `/${this.file.file.name}`, `wcrc32`]),
            this.file.crc32(),
        ])
        if (this.isClosed) {
            return
        }
        if (results[0] == results[1]) {
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
    private async _upload() {
        const file = this.file.file
        const size = file.size
        // 計算分塊
        const chunks = new Array<Chunk>()
        let start = 0
        let end: number
        let index = -1
        while (start != size) {
            end = start + ChunkSize
            if (end > size) {
                end = size
            }

        }
        console.log(chunks)
    }
}
export class Chunk {
    constructor(public file: File, public index: number, public start: number, public end: number) {
    }
    run() {

    }
}