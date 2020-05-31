import { Component, OnInit, Inject, OnDestroy, AfterViewInit, ElementRef, ViewChild } from '@angular/core';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialog } from '@angular/material/dialog';
import { fromEvent, Subject } from 'rxjs';
import { takeUntil } from 'rxjs/operators';
import { HttpClient } from '@angular/common/http';
import { Status, UploadFile, Uploader, Data } from './uploader';


class Source {
  private _keys = new Set<string>()
  private _items = new Array<UploadFile>()

  push(uploadFile: UploadFile) {
    if (!uploadFile.file.size) {
      console.warn(`not support size 0`, uploadFile.file)
      return
    }
    const key = uploadFile.key
    if (this._keys.has(key)) {
      return
    }
    this._keys.add(key)
    this._items.push(uploadFile)
  }
  get source(): Array<UploadFile> {
    return this._items
  }
  clear() {
    const arrs = this._items
    for (let i = arrs.length - 1; i >= 0; i--) {
      const node = arrs[i]
      if (node.isWorking()) {
        continue
      }
      arrs.splice(i, 1)
      this._keys.delete(node.key)
    }
  }
  delete(uploadFile: UploadFile) {
    if (!uploadFile || uploadFile.isWorking()) {
      return
    }
    const index = this._items.indexOf(uploadFile)
    if (index == -1) {
      return
    }
    this._items.splice(index, 1)
    this._keys.delete(uploadFile.key)
  }
  get(): UploadFile {
    const arrs = this._items
    let find: UploadFile
    for (let i = 0; i < arrs.length; i++) {
      const element = arrs[i]
      if (element.status == Status.Nil) {
        find = element
        break
      }
    }
    return find
  }
  prepare() {
    for (let i = 0; i < this._items.length; i++) {
      const element = this._items[i]
      if (element.status == Status.Error) {
        element.status = Status.Nil
      }
    }
  }
}
@Component({
  selector: 'app-upload',
  templateUrl: './upload.component.html',
  styleUrls: ['./upload.component.scss']
})
export class UploadComponent implements OnInit, OnDestroy, AfterViewInit {
  constructor(private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialog: MatDialog,
    private matDialogRef: MatDialogRef<UploadComponent>,
    @Inject(MAT_DIALOG_DATA) public data: Data,
    private httpClient: HttpClient,
  ) { }
  private _num = 0
  private _source = new Source()
  get source(): Array<UploadFile> {
    return this._source.source
  }
  private _disabled: boolean
  get disabled(): boolean {
    return this._disabled
  }
  private _clsoeSubject = new Subject<boolean>()
  ngOnInit(): void {
    // 禁用 瀏覽器 檔案拖動
    fromEvent(document, 'drop').pipe(
      takeUntil(this._clsoeSubject),
    ).subscribe((evt) => {
      evt.preventDefault()
    })
    fromEvent(document, 'dragover').pipe(
      takeUntil(this._clsoeSubject),
    ).subscribe((evt) => {
      evt.preventDefault()
    })
  }
  dragover: boolean
  private _closed: boolean
  ngOnDestroy() {
    this._closed = true
    this._clsoeSubject.next(true)
    if (this._uploader) {
      this._uploader.close()
      this._uploader = null
    }
  }
  @ViewChild("drop")
  private drop: ElementRef
  ngAfterViewInit() {
    fromEvent(this.drop.nativeElement, 'dragover').pipe(
      takeUntil(this._clsoeSubject),
    ).subscribe((evt: Event) => {
      this.dragover = true
      evt.stopPropagation()
      evt.preventDefault()
    })
    fromEvent(this.drop.nativeElement, 'dragenter').pipe(
      takeUntil(this._clsoeSubject),
    ).subscribe((evt: Event) => {
      this.dragover = true
      evt.stopPropagation()
      evt.preventDefault()
    })
    fromEvent(this.drop.nativeElement, 'dragexit').pipe(
      takeUntil(this._clsoeSubject),
    ).subscribe((evt: Event) => {
      this.dragover = false
      evt.stopPropagation()
      evt.preventDefault()
    })
    fromEvent(this.drop.nativeElement, 'drop').pipe(
      takeUntil(this._clsoeSubject),
    ).subscribe((evt: Event) => {
      this.dragover = false
      evt.stopPropagation()
      evt.preventDefault()
      this._drop(evt)
    })
  }
  private _drop(evt) {
    let dataTransfer = evt.dataTransfer
    if (!dataTransfer) {
      if (evt.originalEvent) {
        dataTransfer = evt.originalEvent.dataTransfer
        console.warn(`use evt.originalEvent`)
      }
    }
    if (!dataTransfer) {
      console.warn(`dataTransfer nil`)
      return
    }
    if (!dataTransfer.files) {
      return
    }
    for (let i = 0; i < dataTransfer.files.length; i++) {
      this._source.push(new UploadFile(dataTransfer.files[i]))
    }
  }
  onClose() {
    this.matDialogRef.close(this._num)
  }
  onClickClear() {
    this._source.clear()
  }
  onAdd(evt) {
    if (evt.target.files) {
      for (let i = 0; i < evt.target.files.length; i++) {
        const element = evt.target.files[i]
        this._source.push(new UploadFile(element))
      }
    }
  }
  onClickDelete(uploadFile: UploadFile) {
    this._source.delete(uploadFile)
  }
  onClickStart() {
    if (this._disabled) {
      return
    }
    this._disabled = true
    this._run().finally(() => {
      this._disabled = false
    })
  }
  private async _run() {
    this._source.prepare()
    let style: number
    while (!this._closed) {
      const uploadFile = this._source.get()
      if (!uploadFile) {
        break
      }
      const uploader = new Uploader(
        this.data, uploadFile,
        this.httpClient,
        this.matDialog,
        style,
      )
      this._uploader = uploader
      await uploader.done()
      style = uploader.style
    }
  }
  private _uploader: Uploader
}
