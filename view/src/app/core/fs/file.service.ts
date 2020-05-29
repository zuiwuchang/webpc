import { Injectable } from '@angular/core';
import { isString, isArray, isObject } from 'util';
const Key = `copy-cut-files`
export interface Files {
  // copy or cut
  copy: boolean
  // 掛接點
  root: string
  // 所在路徑
  dir: string
  // 檔案名稱列表
  names: Array<string>
}

@Injectable({
  providedIn: 'root'
})
export class FileService {
  constructor() { }
  private _files: Files

  get files(): Files {
    try {
      if (localStorage) {
        const str = localStorage.getItem(Key)
        if (isString(str)) {
          return JSON.parse(str)
        }
      } else {
        return this._files
      }
    } catch (e) {
      console.warn(e)
    }
    return undefined
  }
  set files(info: Files) {
    if (!isObject(info)) {
      if (localStorage) {
        localStorage.removeItem(Key)
      } else {
        this._files = undefined
      }
      return
    } else {
      if (localStorage) {
        let names: Array<string>
        if (isArray(info.names) && info.names.length > 0) {
          for (let i = 0; i < info.names.length; i++) {
            const element = info.names[i]
            if (isString(element)) {
              if (!names) {
                names = new Array<string>()
              }
              names.push(element)
            }
          }
        }
        localStorage.setItem(Key, JSON.stringify({
          copy: info.copy,
          root: info.root,
          dir: info.dir,
          names: names,
        }))
      } else {
        this._files = this.files
      }
    }
  }
  clear(files: Files) {
    if (localStorage) {
      try {
        const str = localStorage.getItem(Key)
        if (isString(str)) {
          const old = JSON.parse(str)
          if (this._isEqual(files, old)) {
            localStorage.removeItem(Key)
            return
          }
        }
      } catch (e) {
        console.warn(e)
      }
      localStorage.removeItem(Key)
    } else {
      if (this._isEqual(this._files, files)) {
        this._files = undefined
      }
    }
  }
  private _isEqual(l: Files, r: Files): boolean {
    if (!isObject(l) || !isObject(r)) {
      return false
    }
    if (l.copy != r.copy || l.root != r.root && l.dir != r.dir) {
      return false
    }
    let lc = 0
    if (isArray(l.names)) {
      lc = l.names.length
    }
    let rc = 0
    if (isArray(r.names)) {
      rc = r.names.length
    }
    if (lc != rc) {
      return false
    }
    for (let i = 0; i < lc; i++) {
      if (l.names[i] != r.names[i]) {
        return false
      }
    }
    return true
  }
}
