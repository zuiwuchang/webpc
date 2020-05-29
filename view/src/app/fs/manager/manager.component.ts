import { Component, OnInit, Input, Output, EventEmitter, OnDestroy, ViewChild, ElementRef } from '@angular/core';
import { Dir, FileInfo } from '../fs';
import { Router } from '@angular/router';
import { isString } from 'util';
import { fromEvent, Subscription } from 'rxjs';
import { takeUntil, first } from 'rxjs/operators';
import { CheckEvent, NativeEvent } from '../file/file.component';
import { MatMenuTrigger } from '@angular/material/menu';
import { MatDialog } from '@angular/material/dialog';
import { Box, Point } from './box';
import { SessionService, Session } from 'src/app/core/session/session.service';
import { RenameComponent } from '../dialog/rename/rename.component';
import { NewFileComponent } from '../dialog/new-file/new-file.component';
import { NewFolderComponent } from '../dialog/new-folder/new-folder.component';
import { PropertyComponent } from '../dialog/property/property.component';
import { RemoveComponent } from '../dialog/remove/remove.component';
import { CompressComponent } from '../dialog/compress/compress.component';
import { UncompressComponent } from '../dialog/uncompress/uncompress.component';

@Component({
  selector: 'fs-manager',
  templateUrl: './manager.component.html',
  styleUrls: ['./manager.component.scss']
})
export class ManagerComponent implements OnInit, OnDestroy {
  constructor(private router: Router,
    private matDialog: MatDialog,
    private sessionService: SessionService,
  ) { }
  private _subscription: Subscription
  private _session: Session
  private _sessionSubscription: Subscription
  @Input()
  folder: Dir
  private _closed: boolean
  private _source: Array<FileInfo>
  private _hide: Array<FileInfo>
  @Output()
  sourceChange = new EventEmitter<Array<FileInfo>>()
  @Input('source')
  set source(arrs: Array<FileInfo>) {
    this._source = arrs
    this._hide = null
    if (arrs && arrs.length > 0) {
      const items = new Array<FileInfo>()
      for (let i = 0; i < arrs.length; i++) {
        if (arrs[i].name.startsWith('.')) {
          continue
        }
        items.push(arrs[i])
      }
      this._hide = items
    }
  }
  get source(): Array<FileInfo> {
    return this.all ? this._source : this._hide
  }
  ngOnInit(): void {
    this._sessionSubscription = this.sessionService.observable.subscribe((session) => {
      this._session = session
    })
  }
  ngOnDestroy() {
    this._closed = true
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
    this._sessionSubscription.unsubscribe()
  }
  @ViewChild('fs')
  fs: ElementRef
  @ViewChild('box')
  box: ElementRef
  _trigger: MatMenuTrigger
  @ViewChild(MatMenuTrigger)
  set trigger(trigger: MatMenuTrigger) {
    if (this._trigger) {
      return
    }
    this._trigger = trigger
  }
  get trigger(): MatMenuTrigger {
    return this._trigger
  }

  ctrl: boolean
  shift: boolean
  all: boolean
  onPathChange(path: string) {
    const folder = this.folder
    if (!folder) {
      return
    }

    if (!isString(path)) {
      path = '/'
    }
    if (!path.startsWith('/')) {
      path = '/' + path
    }

    this.router.navigate(['fs', 'list'], {
      queryParams: {
        root: folder.root,
        path: path,
      }
    })
  }
  menuLeft = 0
  menuTop = 0
  onContextmenu(evt) {
    if (!this.ctrl && !this.shift && !evt.ctrlKey && !evt.shiftKey) {
      this._clearChecked()
    }
    if (this.trigger) {
      this._openMenu(this.trigger, evt.clientX, evt.clientY)
    }
    return false
  }
  onContextmenuNode(evt: CheckEvent) {
    if (!evt.target.checked) {
      if (!this.ctrl && !this.shift && !evt.event.ctrlKey && !evt.event.shiftKey) {
        this._clearChecked()
      }
      evt.target.checked = true
    }
    if (this.trigger) {
      this._openMenu(this.trigger, (evt.event as any).clientX, (evt.event as any).clientY)
    }
    return false
  }
  private _box: Box = new Box()
  onStart(evt) {
    if (evt.button == 2 || evt.ctrlKey || evt.shiftKey || this.ctrl || this.shift) {
      return
    }
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
    this._displayBox = false
    const doc = this.box.nativeElement
    let start = new Date()
    this._subscription = fromEvent(document, 'mousemove').pipe(
      takeUntil(fromEvent(document, 'mouseup').pipe(first()))
    ).subscribe({
      next: (evt: any) => {
        if (start) {
          const now = new Date()
          const diff = now.getTime() - start.getTime()
          if (diff < 100) {
            return
          }
          this._displayBox = true
          start = null
          this._box.setRange(doc)
          this._box.start = new Point(evt.clientX, evt.clientY)
          this._box.stop = this._box.start
          return;
        }
        this._box.setRange(doc)
        this._box.stop = new Point(evt.clientX, evt.clientY)
        this._box.calculate()
      },
      complete: () => {
        this._select()
      },
    })
  }
  private _displayBox = false
  onClick(evt: NativeEvent) {
    evt.stopPropagation()
    if (this._displayBox || evt.ctrlKey || evt.shiftKey || this.ctrl || this.shift) {
      return
    }
    // 清空選項
    this._clearChecked()
  }
  private _clearChecked() {
    const source = this._source
    if (!source) {
      return
    }
    for (let i = 0; i < source.length; i++) {
      if (source[i].checked) {
        source[i].checked = false
      }
    }
  }
  private _select() {
    const arrs = this._box.checked(this.fs.nativeElement)
    this._clearChecked()
    const source = this.source
    if (source) {
      for (let i = 0; i < arrs.length; i++) {
        const index = arrs[i]
        if (index < source.length) {
          source[index].checked = true
        }
      }
    }
    this._box.reset()
  }
  get x(): number {
    return this._box.x
  }
  get y(): number {
    return this._box.y
  }
  get w(): number {
    return this._box.w
  }
  get h(): number {
    return this._box.h
  }
  onCheckChange(evt: CheckEvent) {
    if (evt.event.ctrlKey || this.ctrl) {
      evt.target.checked = !evt.target.checked
      return
    }
    let start = -1
    let stop = -1
    let index = -1
    // 清空選項
    const source = this.source
    if (source) {
      for (let i = 0; i < source.length; i++) {
        if (source[i] == evt.target) {
          index = i
        }
        if (source[i].checked) {
          if (start == -1) {
            start = i
          }
          stop = i
        }
        if (source[i].checked) {
          source[i].checked = false
        }
      }
    }
    if (index == -1) {
      return
    }
    // 設置選項
    if ((evt.event.shiftKey || this.shift) && start != -1) {
      if (index <= start) {
        for (let i = index; i <= stop; i++) {
          source[i].checked = true
        }
      } else if (index >= stop) {
        for (let i = start; i <= index; i++) {
          source[i].checked = true
        }
      } else {
        for (let i = start; i <= stop; i++) {
          source[i].checked = true
        }
      }
      return
    }
    source[index].checked = true
  }
  toggleDisplay() {
    this.all = !this.all
    this._clearChecked()
  }
  // 爲 彈出菜單 緩存 選中目標
  target = new Array<FileInfo>()
  private _openMenu(trigger: MatMenuTrigger, x: number, y: number) {
    this.menuLeft = x
    this.menuTop = y
    trigger.openMenu()
    const target = new Array<FileInfo>()
    const source = this.source
    if (source) {
      for (let i = 0; i < this.source.length; i++) {
        if (this.source[i].checked) {
          target.push(this.source[i])
        }
      }
    }
    this.target = target
  }
  get isNotCanWrite(): boolean {
    if (this._session) {
      if (this._session.root) {
        return false
      }
      if (this._session.write && this.folder.write) {
        return false
      }
    }
    return true
  }
  onClickRename() {
    if (this.target && this.target.length == 1) {
      const node = this.target[0]
      const name = node.name
      this.matDialog.open(RenameComponent, {
        data: node,
        disableClose: true,
      }).afterClosed().toPromise().then(() => {
        const current = node.name;
        if (name == current) {
          return
        }
        if (name.startsWith(`.`)) {
          if (!current.startsWith(`.`)) {
            if (!this._hide) {
              this._hide = new Array<FileInfo>()
            }
            this._hide.push(node)
            this._hide.sort(FileInfo.compare)
          }
        } else {
          if (current.startsWith(`.`)) {
            if (this._hide) {
              const index = this._hide.indexOf(node)
              if (index != -1) {
                this._hide.splice(index, 1)
              }
            }
          }
        }
      })
    }
  }
  onClickNewFile() {
    if (!this.folder || this._closed) {
      return
    }
    this.matDialog.open(NewFileComponent, {
      data: this.folder,
      disableClose: true,
    }).afterClosed().toPromise().then((fileinfo: FileInfo) => {
      if (fileinfo && fileinfo instanceof FileInfo) {
        this._pushNode(fileinfo)
      }
    })
  }
  onClickNewFolder() {
    if (!this.folder || this._closed) {
      return
    }
    this.matDialog.open(NewFolderComponent, {
      data: this.folder,
      disableClose: true,
    }).afterClosed().toPromise().then((fileinfo: FileInfo) => {
      if (fileinfo && fileinfo instanceof FileInfo) {
        this._pushNode(fileinfo)
      }
    })
  }
  private _pushNode(fileinfo: FileInfo) {
    if (!this._source) {
      this._source = new Array<FileInfo>()
      this.sourceChange.emit(this._source)
    }
    this._source.push(fileinfo)
    this._source.sort(FileInfo.compare)
    if (isString(fileinfo.name) && fileinfo.name.startsWith('.')) {
      return
    }
    if (!this._hide) {
      this._hide = new Array<FileInfo>()
    }
    this._hide.push(fileinfo)
    this._hide.sort(FileInfo.compare)
  }
  onClickProperty() {
    if (!this.target || this.target.length == 0) {
      return
    }
    this.matDialog.open(PropertyComponent, {
      data: this.target,
    })
  }
  onClickRemove() {
    const target = this.target
    if (!target || target.length == 0) {
      return
    }
    const dir = this.folder
    this.matDialog.open(RemoveComponent, {
      data: {
        dir: dir,
        source: target,
      },
      disableClose: true,
    }).afterClosed().toPromise().then((ok) => {
      if (ok) {
        for (let i = 0; i < target.length; i++) {
          const element = target[i]
          if (this._source) {
            const index = this._source.indexOf(element)
            if (index != -1) {
              this._source.splice(index, 1)
            }
          }
          if (this._hide) {
            const index = this._hide.indexOf(element)
            if (index != -1) {
              this._hide.splice(index, 1)
            }
          }
        }
      }
    })
  }
  onClickCompress() {
    const target = this.target
    if (!target || target.length == 0) {
      return
    }
    const dir = this.folder
    this.matDialog.open(CompressComponent, {
      data: {
        dir: dir,
        source: target,
      },
      disableClose: true,
    }).afterClosed().toPromise().then((fileinfo: FileInfo) => {
      if (fileinfo instanceof FileInfo) {
        this._pushOrUpdate(fileinfo)
      }
    })
  }
  private _pushOrUpdate(fileinfo: FileInfo) {
    if (!this._source) {
      this._source = new Array<FileInfo>()
      this.sourceChange.emit(this._source)
    }
    if (this._source.length == 0) {
      this._source.push(fileinfo)
    } else {
      let ok = false
      for (let i = 0; i < this._source.length; i++) {
        if (this._source[i].name == fileinfo.name) {
          ok = true
          this._source[i] = fileinfo
          break
        }
      }
      if (!ok) {
        this._source.push(fileinfo)
        this._source.sort(FileInfo.compare)
      }
    }

    if (!this._hide) {
      this._hide = new Array<FileInfo>()
    }
    if (this._hide.length == 0) {
      this._hide.push(fileinfo)
    } else {
      let ok = false
      for (let i = 0; i < this._hide.length; i++) {
        if (this._hide[i].name == fileinfo.name) {
          ok = true
          this._hide[i] = fileinfo
          break
        }
      }
      if (!ok) {
        this._hide.push(fileinfo)
        this._hide.sort(FileInfo.compare)
      }
    }
  }
  onClickRefresh() {
    const folder = this.folder
    if (!folder) {
      return
    }
    this.router.navigate(['fs', 'list'], {
      queryParams: {
        root: folder.root,
        path: folder.dir,
        tick: new Date().getTime(),
      }
    })
  }
  onClickUncompress() {
    const target = this.target
    if (!target || target.length == 0) {
      return
    }
    const dir = this.folder
    this.matDialog.open(UncompressComponent, {
      data: {
        dir: dir,
        source: target[0],
      },
      disableClose: true,
    }).afterClosed().toPromise().then((ok) => {
      if (ok) {
        this.onClickRefresh()
      }
    })
  }
}
