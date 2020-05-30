import { Component, OnInit, Inject, OnDestroy, AfterViewInit, ElementRef, ViewChild } from '@angular/core';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialog } from '@angular/material/dialog';
import { FileInfo, Dir } from '../../fs';
import { isString, isNumber, isArray } from 'util';
import { interval, Subscription, fromEvent, Subject } from 'rxjs';
import { ExistChoiceComponent } from '../exist-choice/exist-choice.component';
import { NetCommand } from '../command';
import { sizeString } from 'src/app/core/core/utils';
import { takeUntil } from 'rxjs/operators';
interface Data {
  root: string
  dir: string
}
class UploadFile {
  constructor(public file: File) {
  }
  progress: number
  get sizeString(): string {
    return sizeString(this?.file?.size)
  }
  get key(): string {
    const file = this.file
    return `${file.size}${file.lastModified}${file.name}`
  }
}
class Source {
  private _keys = new Set<string>()
  private _items = new Array<UploadFile>()

  push(uploadFile: UploadFile) {
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
    this._items.splice(0, this._items.length)
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
  ngOnDestroy() {
    this._clsoeSubject.next(true)
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
      const files = (evt as any).dataTransfer.files;
      console.log(files)
    })
  }
  private _drop(evt) {
    let dataTransfer = evt.dataTransfer
    if (!dataTransfer) {
      if (evt.originalEvent) {
        dataTransfer = evt.originalEvent.dataTransfer
      }
    }
    if (!dataTransfer) {
      console.log(`dataTransfer nil`)
      return
    }
    console.log(dataTransfer.files)
  }
  onClose() {
    this.matDialogRef.close(this._num)
  }
  onClickClear() {
    if (this._disabled) {
      return
    }
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
}
