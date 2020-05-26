import { Component, OnInit, Input, OnDestroy, ViewChild, ElementRef } from '@angular/core';
import { Dir, FileInfo } from '../fs';
import { Router } from '@angular/router';
import { isString } from 'util';
import { fromEvent, Subscription } from 'rxjs';
import { takeUntil, first } from 'rxjs/operators';
import { CheckEvent, NativeEvent } from '../file/file.component';
import { MatMenuTrigger } from '@angular/material/menu';

class Point {
  constructor(public x: number, public y: number) {
  }
  toView(): Point {
    if (document.compatMode == "BackCompat") {
      this.x -= document.body.scrollLeft
      this.y -= document.body.scrollTop
    } else {
      this.x -= document.documentElement.scrollLeft
      this.y -= document.documentElement.scrollTop
    }
    return this
  }
}
// 有效範圍
class Box {
  private _p0: Point
  private _p1: Point
  start: Point
  stop: Point
  setRange(element) {
    this._p0 = getViewPoint(element)
    this._p1 = new Point(this._p0.x + element.offsetWidth, this._p0.y + element.offsetHeight)
  }
  private _fixStart() {
    if (this.start.x < this._p0.x) {
      this.start.x = this._p0.x
    } else if (this.start.x > this._p1.x) {
      this.start.x = this._p1.x
    }

    if (this.start.y < this._p0.y) {
      this.start.y = this._p0.y
    } else if (this.start.y > this._p1.y) {
      this.start.y = this._p1.y
    }
  }

  private _fixStop() {
    if (this.stop.x < this._p0.x) {
      this.stop.x = this._p0.x
    } else if (this.stop.x > this._p1.x) {
      this.stop.x = this._p1.x
    }

    if (this.stop.y < this._p0.y) {
      this.stop.y = this._p0.y
    } else if (this.stop.y > this._p1.y) {
      this.stop.y = this._p1.y
    }
  }
  calculate() {
    if (!this.start || !this.stop) {
      return
    }
    if (this._p0 && this._p1) {
      this._fixStart()
      this._fixStop()
    }
    this.x = Math.min(this.start.x, this.stop.x)
    this.y = Math.min(this.start.y, this.stop.y)
    this.w = Math.abs(this.start.x - this.stop.x)
    this.h = Math.abs(this.start.y - this.stop.y)
  }
  x = 0
  y = 0
  w = 0
  h = 0
  reset() {
    this.x = 0
    this.y = 0
    this.w = 0
    this.h = 0
    this._p0 = null
    this._p1 = null
    this.start = null
    this.stop = null
  }

  checked(doc: Document): Array<number> {
    const result = new Array<number>()
    const nodes = doc.childNodes

    if (nodes && nodes.length > 0) {
      let parent: any
      for (let i = 0; i < nodes.length; i++) {
        let node = (nodes[i] as any)
        if (!node || !node.querySelector) {
          continue
        }
        node = node.querySelector('.wrapper')
        if (!node) {
          continue
        }
        const l = getViewPoint(node)
        const r = new Point(l.x + node.offsetWidth, l.y + node.offsetHeight)
        const ok = this.testView(l, r)
        if (ok) {
          result.push(i)
        }
      }
    }
    return result
  }
  testView(l: Point, r: Point): boolean {
    if (r.x < this.x || l.x > (this.x + this.w)) {
      return false
    }
    if (r.y < this.y || l.y > (this.y + this.h)) {
      return false
    }
    return true
  }
}
function getPagePoint(element): Point {
  let x = 0
  let y = 0
  while (element) {
    x += element.offsetLeft + element.clientLeft
    y += element.offsetTop + element.clientTop
    element = element.offsetParent
  }
  return new Point(x, y)
}

function getViewPoint(element): Point {
  return getPagePoint(element).toView()
}
@Component({
  selector: 'fs-manager',
  templateUrl: './manager.component.html',
  styleUrls: ['./manager.component.scss']
})
export class ManagerComponent implements OnInit, OnDestroy {

  constructor(private router: Router) { }
  private _subscription: Subscription
  @Input()
  folder: Dir

  private _source: Array<FileInfo>
  private _hide: Array<FileInfo>
  @Input('source')
  set source0(arrs: Array<FileInfo>) {
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
  }
  ngOnDestroy() {
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
  }
  @ViewChild('fs')
  fs: ElementRef
  @ViewChild('box')
  box: ElementRef
  @ViewChild(MatMenuTrigger)
  trigger: MatMenuTrigger

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
  menuLeft: 0
  menuTop: 0
  onClickMenu(evt) {
    if (!this.trigger) {
      return
    }
    this.menuLeft = evt.clientX
    this.menuTop = evt.clientY
    this.trigger.openMenu()
  }
  onContextmenu(evt) {
    this._clearChecked()
    if (this.trigger) {
      this.menuLeft = evt.clientX
      this.menuTop = evt.clientY
      this.trigger.openMenu()
    }
    return false
  }
  onContextmenuNode(evt: CheckEvent) {
    console.log(evt)
    if (!evt.target.checked) {
      this._clearChecked()
      evt.target.checked = true
    }
    if (this.trigger) {
      this.menuLeft = (evt.event as any).clientX
      this.menuTop = (evt.event as any).clientY
      this.trigger.openMenu()
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
    for (let i = 0; i < arrs.length; i++) {
      const index = arrs[i]
      if (index < source.length) {
        source[index].checked = true
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
      evt.target.checked = true
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
}
