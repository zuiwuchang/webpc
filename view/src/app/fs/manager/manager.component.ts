import { Component, OnInit, Input, OnDestroy, ViewChild, ElementRef } from '@angular/core';
import { Dir, FileInfo } from '../fs';
import { Router } from '@angular/router';
import { isString } from 'util';
import { fromEvent, Subscription } from 'rxjs';
import { takeUntil, first } from 'rxjs/operators';
class Point {
  constructor(public x: number, public y: number) {
  }
}
// 有效範圍
class Box {
  private _p0: Point
  private _p1: Point
  constructor() {

  }
  setRange(element) {
    this._p0 = getViewPoint(element)
    this._p1 = new Point(this._p0.x + element.offsetWidth, this._p0.y + element.offsetHeight)
    console.log(`p0`, this._p0)
    console.log(`p1`, this._p1)
    console.log(element)
  }
}
class Rect {
  x: number = 0
  y: number = 0
  x1: number = 0
  y1: number = 0

  reset(x: number, y: number) {
    this.x = x
    this.x1 = x
    this.y = y
    this.y1 = y
  }
  set(x: number, y: number) {
    this.x1 = x
    this.y1 = y
  }
  get l(): number {
    return Math.min(this.x, this.x1)
  }
  get t(): number {
    return Math.min(this.y, this.y1)
  }
  get w(): number {
    return Math.abs(this.x - this.x1)
  }
  get h(): number {
    return Math.abs(this.y - this.y1)
  }
}
function getPagePoint(element): Point {
  let x = element.offsetLeft
  let current = element.offsetParent
  while (current) {
    x += current.offsetLeft
    current = current.offsetParent
  }
  let y = element.offsetTop
  current = element.offsetParent
  while (current !== null) {
    y += (current.offsetTop + current.clientTop)
    current = current.offsetParent
  }
  return new Point(x, y)
}
function getViewPoint(element): Point {
  const point = getPagePoint(element)
  if (document.compatMode == "BackCompat") {
    point.x -= document.body.scrollLeft
    point.y -= document.body.scrollTop
  } else {
    point.x -= document.documentElement.scrollLeft
    point.y -= document.documentElement.scrollTop
  }
  return point
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

  @Input()
  source: Array<FileInfo>
  ngOnInit(): void {
  }
  ngOnDestroy() {
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
  }
  @ViewChild('box')
  box: ElementRef
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
  onContextmenu(evt) {
    console.log('onContextmenu', evt)
    return false
  }
  rect = new Rect()
  private _box: Box = new Box()
  onStart(evt) {
    if (evt.button == 2) {
      return
    }
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
    const doc = this.box.nativeElement
    doc.setCapture()
    this._box.setRange(doc)

    this.rect.reset(evt.clientX, evt.clientY)

    this._subscription = fromEvent(this.box.nativeElement, 'mousemove').pipe(
      takeUntil(fromEvent(this.box.nativeElement, 'mouseup').pipe(first()))
    ).subscribe({
      next: (evt: any) => {
        this._box.setRange(doc)
        this.rect.set(evt.clientX, evt.clientY)
      },
      complete: () => {
        doc.releaseCapture()
        this._select()
      },
    })
  }
  private _select() {
    //console.log(this.rect)
    this.rect.reset(0, 0)
  }
}
