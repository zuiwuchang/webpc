import { Component, OnInit, Input, Output, EventEmitter } from '@angular/core';
interface Dir {
  name: string
  path: string
}

@Component({
  selector: 'fs-path',
  templateUrl: './path.component.html',
  styleUrls: ['./path.component.scss']
})
export class PathComponent implements OnInit {
  constructor() { }
  @Output() pathChange = new EventEmitter<string>()
  dirs: Array<Dir>
  private _path: string = ''
  @Input()
  disabled: boolean
  @Input()
  set path(val: string) {
    if (typeof val !== "string") {
      val = ''
    }
    if (val == this._path) {
      return
    }
    this._path = val
    this.val = val
    const strs = this._path.split('/')
    const dirs = new Array<Dir>()
    let path = ''
    for (let i = 0; i < strs.length; i++) {
      const str = strs[i]
      if (str != "") {
        path += '/' + str
        dirs.push({
          name: str,
          path: path,
        })
      }
    }
    this.dirs = dirs
  }
  get path(): string {
    return this._path
  }
  edit = false
  val: string = ''
  ngOnInit(): void {
  }
  onClickDone() {
    this.pathChange.emit(this.val)
  }
  onClickDir(node: Dir) {
    this.pathChange.emit(node.path)
  }
}
