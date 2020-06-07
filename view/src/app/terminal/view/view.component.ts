import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild, ElementRef } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { Subject, fromEvent } from 'rxjs'
import { debounceTime, takeUntil } from 'rxjs/operators';
import { ServerAPI } from 'src/app/core/core/api';
import { isString, isNumber } from 'util';
import { interval } from 'rxjs';
import { MatDialog } from '@angular/material/dialog';
import { SettingsComponent } from '../dialog/settings/settings.component';
import { FullscreenService } from 'src/app/core/fullscreen/fullscreen.service';

// CmdError 錯誤
const CmdError = 1
// CmdResize 更改窗口大小
const CmdResize = 2
// CmdInfo 返回終端信息
const CmdInfo = 3
// CmdHeart websocket 心跳防止瀏覽器 關閉不獲取 websocket
const CmdHeart = 4
// CmdFontSize 設置字體大小
const CmdFontSize = 5
// CmdFontFamily 設置字體
const CmdFontFamily = 6
const HeartMessage = JSON.stringify({
  'cmd': CmdHeart,
})

const DefaultFontFamily = "monospace"
interface Info {
  cmd: number
  id: number
  name: string
  started: number
  fontSize: number
  fontFamily: string
}
const Second = 1
const Minute = 60 * Second
const Hour = 60 * Minute
const Day = 60 * Hour
function pushStep(arrs: Array<string>, v: number, step: number, flag: string): number {
  if (v > step) {
    const tmp = Math.floor(v / step)
    arrs.push(`${tmp}${flag}`)
    v -= tmp * step
  }
  return v
}
function durationToString(v: number): string {
  const arrs = new Array<string>()
  v = pushStep(arrs, v, Day, "d")
  v = pushStep(arrs, v, Hour, "h")
  v = pushStep(arrs, v, Minute, "m")
  v = pushStep(arrs, v, Second, "s")
  return arrs.join(``)
}
@Component({
  selector: 'app-view',
  templateUrl: './view.component.html',
  styleUrls: ['./view.component.scss']
})
export class ViewComponent implements OnInit, OnDestroy, AfterViewInit {
  constructor(private route: ActivatedRoute,
    private matDialog: MatDialog,
    private fullscreenService: FullscreenService,
  ) {
    this.fullscreen = false
  }
  private _closed = false
  private _subject = new Subject()
  private _xterm: Terminal
  private _fitAddon: FitAddon
  private _websocket: WebSocket
  info: Info
  private _closeSubject = new Subject<boolean>()
  duration: string = ''
  fontSize = 15
  fontFamily = DefaultFontFamily
  ctrl: boolean
  shift: boolean
  alt: boolean
  private _fullscreen: boolean
  set fullscreen(val: boolean) {
    this._fullscreen = val
    this.fullscreenService.fullscreen = val
  }
  get fullscreen(): boolean {
    return this._fullscreen
  }
  onClickFullscreen(val: boolean) {
    this._fullscreen = val
    this.fullscreenService.fullscreen = val
    this.onResize()
  }
  get ok(): boolean {
    if (this._websocket) {
      return true
    }
    return false
  }
  ngOnInit(): void {
    interval(1000).pipe(
      takeUntil(this._closeSubject),
    ).subscribe(() => {
      if (!this._websocket) {
        return
      }
      if (this.info && isNumber(this.info.started)) {
        const val = new Date().getTime() - this.info.started * 1000
        this.duration = durationToString(val / 1000)
      }
    })
    interval(1000 * 30).pipe(
      takeUntil(this._closeSubject),
    ).subscribe(() => {
      if (this._websocket) {
        this._websocket.send(HeartMessage)
      }
    })
  }
  ngOnDestroy() {
    this._closed = true
    this._closeSubject.next(true)
    this._closeSubject.complete()
    if (this._websocket) {
      this._websocket.close()
      this._websocket = null
    }
    if (this._xterm) {
      this._xterm.dispose()
    }
    this.fullscreen = false
  }
  @ViewChild("xterm")
  xterm: ElementRef
  @ViewChild("view")
  view: ElementRef
  private _getFontFamily(name: string) {
    if (isString(name) && name != '') {
      return name
    }
    return DefaultFontFamily
  }
  ngAfterViewInit() {
    // 屏蔽瀏覽器快捷鍵
    fromEvent(this.view.nativeElement, 'keydown').pipe(
      takeUntil(this._closeSubject)
    ).subscribe((evt: KeyboardEvent) => {
      // console.log('document', evt.keyCode, '****', evt)
      if (evt.ctrlKey || evt.shiftKey || evt.altKey) {
        if (evt.keyCode != 45) {
          evt.returnValue = false
        }
      }
    })

    // 創建 xterm
    const xterm = new Terminal({
      cursorBlink: true,
      screenReaderMode: true,
      fontFamily: this.fontFamily,
      rendererType: 'canvas',
    })
    this._xterm = xterm
    this.fontSize = xterm.getOption("fontSize")

    // 加載插件
    const fitAddon = new FitAddon()
    this._fitAddon = fitAddon
    xterm.loadAddon(fitAddon)
    xterm.loadAddon(new WebLinksAddon())

    xterm.open(this.xterm.nativeElement)
    this._textarea = this.xterm.nativeElement.querySelector('textarea')
    fitAddon.fit()

    // 訂閱 窗口大小 改變
    this._subject.pipe(
      debounceTime(100),
      takeUntil(this._closeSubject),
    ).subscribe((_) => {
      if (this._closed) {
        return
      }
      fitAddon.fit()
    })

    let id = 0;
    try {
      id = parseInt(this.route.snapshot.paramMap.get(`id`))
      if (isNaN(id)) {
        id = 0
      }
    } catch (e) {
      console.warn(e)
    }
    this._connect(id)
  }
  connect = false
  private _connect(id: number) {
    setTimeout(() => {
      this._connectDelay(id)
    }, 0)
  }
  private _connectDelay(id: number) {
    if (this.connect) {
      return
    }
    this.connect = true
    const url = ServerAPI.v1.shells.websocketURL([
      id,
      this._xterm.cols,
      this._xterm.rows,
    ])
    console.log(url)
    const websocket = new WebSocket(url)
    this._websocket = websocket
    websocket.binaryType = "arraybuffer"
    websocket.onerror = (evt) => {
      console.log(evt)
      websocket.close()
      if (this._websocket != websocket) {
        return
      }
      this.connect = false
      this._xterm.writeln("websocket error")
      this._websocket = null
    }
    websocket.onopen = (evt) => {
      if (this._websocket != websocket) {
        websocket.close()
        return
      }
      this.connect = false
      this._xterm.attachCustomKeyEventHandler((evt: any) => {
        const opt: any = {}
        let ok = true
        if (!evt.ctrlKey && this.ctrl) {
          opt.ctrlKey = true
          ok = false
        }
        if (!evt.shiftKey && this.shift) {
          opt.shiftKey = true
          ok = false
        }
        if (!evt.altKey && this.alt) {
          opt.altKey = true
          ok = false
        }
        if (!ok) {
          opt.keyCode = evt.keyCode
          opt.key = evt.key
          opt.code = evt.code
          // this.alt = false
          // this.shift = false
          // this.ctrl = false
          const textarea = this._textarea
          textarea.dispatchEvent(new KeyboardEvent('keydown', opt))
        }
        return ok
      })
      this._xterm.onData((data) => {
        if (this._websocket != websocket) {
          return
        }
        websocket.send(new TextEncoder().encode(data))
      })
      this._xterm.onResize((evt) => {
        console.log(`onResize`, evt)
        if (this._websocket != websocket) {
          return
        }
        websocket.send(JSON.stringify({
          cmd: CmdResize,
          cols: evt.cols,
          rows: evt.rows,
        }))
      })
      let first = true
      websocket.onmessage = (evt) => {
        if (this._websocket != websocket) {
          websocket.close()
          return
        }
        this.connect = false

        if (evt.data instanceof ArrayBuffer) {
          if (first) {
            first = false
            this._xterm.focus()
            this._xterm.setOption("cursorBlink", true)
          }
          this._xterm.write(new Uint8Array(evt.data))
        } else if (isString(evt.data)) {
          try {
            this._onMessage(JSON.parse(evt.data))
          } catch (e) {
            console.warn('ws-shell', e)
          }
        } else {
          console.warn(`ws-shell unknow type`, evt.data)
        }
      }
      websocket.onclose = (evt) => {
        websocket.close()

        if (this._websocket != websocket) {
          return
        }
        this._xterm.writeln("\r\nSession terminated")
        this._xterm.setOption("cursorBlink", false)
        this._websocket = null
        this.connect = false
      }
    }
  }
  onClickConnect() {
    if (isNumber(this.info.id) && !this._websocket) {
      this._xterm.clear()
      this._connect(this.info.id)
    }
  }
  onResize() {
    this._subject.next(new Date())
  }
  private _onMessage(obj: any) {
    switch (obj.cmd) {
      case CmdInfo:
        this.info = obj
        this._onFontsize(this.info.fontSize)
        if (this.info.fontFamily && this.info.fontFamily != this.fontFamily) {
          this.fontFamily = this.info.fontFamily
          console.log(`set font`, this.fontFamily)
          this._xterm.setOption("fontFamily", this.fontFamily)
          this._xterm.resize(1, 1)
          this._fitAddon.fit()
        }
        break
      case CmdError:
        this._xterm.writeln("\n" + obj.error)
        break
      default:
        console.warn(`ws-shell unknow msg : `, obj)
    }
  }
  private _onFontsize(fontSize: number) {
    if (!isNumber(fontSize)) {
      return
    }
    fontSize = Math.floor(fontSize)
    if (fontSize < 5 || fontSize == this.fontSize) {
      return
    }
    this.fontSize = fontSize
    this.onClickFontSize()
  }
  onClickFontSize() {
    if (!this._xterm || this.fontSize < 5 || !isNumber(this.fontSize) || !this._fitAddon || !this._websocket) {
      return
    }
    if (this.fontSize == this._xterm.getOption("fontSize")) {
      return
    }
    this._xterm.setOption("fontSize", this.fontSize)
    this._fitAddon.fit()
    this._websocket.send(JSON.stringify({
      cmd: CmdFontSize,
      val: this.fontSize,
    }))
  }
  onClickFontFamily() {
    if (!this._xterm || !this._fitAddon || !this._websocket) {
      return
    }
    const l = this._getFontFamily(this.fontFamily)
    const r = this._getFontFamily(this._xterm.getOption("fontFamily"))
    if (l == r) {
      return
    }
    this._xterm.setOption("fontFamily", l)
    this._xterm.resize(1, 1)
    this._xterm.clear()
    this._fitAddon.fit()
    this._websocket.send(JSON.stringify({
      cmd: CmdFontFamily,
      str: this.fontFamily,
    }))
  }
  onClickTab(evt: MouseEvent) {
    this._keyboardKeyDown(9, 'Tab', evt)
  }
  onClickCDHome(evt: MouseEvent) {
    this._keyboardKeyDown(192, '~', evt)
  }
  onClickESC(evt: MouseEvent) {
    this._keyboardKeyDown(27, 'Escape', evt)
  }
  onClickArrowUp(evt: MouseEvent) {
    this._keyboardKeyDown(38, 'ArrowUp', evt)
  }
  onClickArrowDown(evt: MouseEvent) {
    this._keyboardKeyDown(40, 'ArrowDown', evt)
  }
  onClickArrowLeft(evt: MouseEvent) {
    this._keyboardKeyDown(37, 'ArrowLeft', evt)
  }
  onClickArrowRight(evt: MouseEvent) {
    this._keyboardKeyDown(39, 'ArrowRight', evt)
  }
  onClickF1(evt: MouseEvent) {
    this._keyboardKeyDown(112, 'F1', evt)
  }
  onClickF2(evt: MouseEvent) {
    this._keyboardKeyDown(113, 'F2', evt)
  }
  onClickF3(evt: MouseEvent) {
    this._keyboardKeyDown(114, 'F3', evt)
  }
  onClickF4(evt: MouseEvent) {
    this._keyboardKeyDown(115, 'F4', evt)
  }
  onClickF5(evt: MouseEvent) {
    this._keyboardKeyDown(116, 'F5', evt)
  }
  onClickF6(evt: MouseEvent) {
    this._keyboardKeyDown(117, 'F6', evt)
  }
  onClickF7(evt: MouseEvent) {
    this._keyboardKeyDown(118, 'F7', evt)
  }
  onClickF8(evt: MouseEvent) {
    this._keyboardKeyDown(119, 'F8', evt)
  }
  onClickF9(evt: MouseEvent) {
    this._keyboardKeyDown(120, 'F9', evt)
  }
  onClickF10(evt: MouseEvent) {
    this._keyboardKeyDown(121, 'F10', evt)
  }
  onClickF11(evt: MouseEvent) {
    this._keyboardKeyDown(122, 'F11', evt)
  }
  onClickF12(evt: MouseEvent) {
    this._keyboardKeyDown(123, 'F12', evt)
  }
  onClickInsert(evt: MouseEvent) {
    this._keyboardKeyDown(45, 'Insert', evt)
  }
  onClickPause(evt: MouseEvent) {
    this._keyboardKeyDown(19, 'Pause', evt)
  }
  onClickPageUp(evt: MouseEvent) {
    this._keyboardKeyDown(33, 'PageUp', evt)
  }
  onClickPageDown(evt: MouseEvent) {
    this._keyboardKeyDown(34, 'PageDown', evt)
  }
  private _textarea: Document
  private _keyboardKeyDown(keyCode: number, key: string, evt: any) {
    if (!this._textarea) {
      return
    }
    this._textarea.dispatchEvent(new KeyboardEvent('keydown', {
      keyCode: keyCode,
      key: key,
      code: key,
      altKey: evt.altKey || this.alt ? true : false,
      shiftKey: evt.shiftKey || this.shift ? true : false,
      ctrlKey: evt.ctrlKey || this.ctrl ? true : false,
    } as any))
    // this.alt = false
    // this.shift = false
    // this.ctrl = false
    setTimeout(() => {
      this._xterm.focus()
    }, 0)
  }
  toggleAlt() {
    this.alt = !this.alt
    setTimeout(() => {
      this._xterm.focus()
    }, 0)
  }
  toggleShift() {
    this.shift = !this.shift
    setTimeout(() => {
      this._xterm.focus()
    }, 0)
  }
  toggleCtrl() {
    this.ctrl = !this.ctrl
    setTimeout(() => {
      this._xterm.focus()
    }, 0)
  }
  onClickSettings() {
    this.matDialog.open(SettingsComponent, {
      data: {
        fontFamily: this.fontFamily,
        onFontFamily: (str: string) => {
          this.fontFamily = str
          this.onClickFontFamily()
        },
        fontSize: this.fontSize,
        onFontSize: (size: number) => {
          this.fontSize = size
          this.onClickFontSize()
        },
      },
    })
  }
}
