import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild, ElementRef } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { Subject, Subscription } from 'rxjs'
import { debounceTime } from 'rxjs/operators';
import { ServerAPI, getWebSocketAddr } from 'src/app/core/core/api';
import { isString, isNumber } from 'util';
import { interval } from 'rxjs';

// CmdError 錯誤
const CmdError = 1
// CmdResize 更改窗口大小
const CmdResize = 2
// CmdInfo 返回終端信息
const CmdInfo = 3
// CmdHeart websocket 心跳防止瀏覽器 關閉不獲取 websocket
const CmdHeart = 4

interface Info {
  cmd: number
  id: number
  name: string
  started: number
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
  ) { }
  private _closed = false
  private _subject = new Subject()
  private _subscription: Subscription
  private _xterm: Terminal
  private _fitAddon: FitAddon
  private _websocket: WebSocket
  info: Info
  private _subscriptionInterval: Subscription
  private _subscriptionPing: Subscription
  duration: string = ''
  fontSize = 15
  get ok(): boolean {
    if (this._websocket) {
      return true
    }
    return false
  }
  ngOnInit(): void {
    this._subscriptionInterval = interval(1000).subscribe(() => {
      if (!this._websocket) {
        return
      }
      if (this.info && isNumber(this.info.started)) {
        const val = new Date().getTime() - this.info.started * 1000
        this.duration = durationToString(val / 1000)
      }
    })
    this._subscriptionPing = interval(1000 * 30).subscribe(() => {
      if (this._websocket) {
        this._websocket.send(JSON.stringify({
          cmd: CmdHeart,
        }))
      }
    })
  }
  ngOnDestroy() {
    this._closed = true
    this._subscriptionInterval.unsubscribe()
    this._subscriptionPing.unsubscribe()
    if (this._websocket) {
      this._websocket.close()
      this._websocket = null
    }
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
    if (this._xterm) {
      this._xterm.dispose()
    }
  }
  @ViewChild("xterm")
  xterm: ElementRef
  ngAfterViewInit() {
    // 創建 xterm
    const xterm = new Terminal({
      cursorBlink: true,
      screenReaderMode: true,
    })
    this._xterm = xterm
    this.fontSize = xterm.getOption("fontSize")

    // 加載插件
    const fitAddon = new FitAddon()
    this._fitAddon = fitAddon
    xterm.loadAddon(fitAddon)
    xterm.loadAddon(new WebLinksAddon())

    xterm.open(this.xterm.nativeElement)
    fitAddon.fit()

    // 訂閱 窗口大小 改變
    this._subscription = this._subject.pipe(
      debounceTime(100)
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
    const url = getWebSocketAddr(`/ws${ServerAPI.v1.shells.baseURL}/${id}/${this._xterm.cols}/${this._xterm.rows}`)
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

      this._xterm.onData((data) => {
        if (this._websocket != websocket) {
          return
        }
        websocket.send(new TextEncoder().encode(data))
      })
      this._xterm.onResize((evt) => {
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
            this.onMessage(JSON.parse(evt.data))
          } catch (e) {
            console.warn(e)
          }
        } else {
          console.warn(`unknow type`, evt.data)
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
  private onMessage(obj: any) {
    switch (obj.cmd) {
      case CmdInfo:
        this.info = obj
        break
      case CmdError:
        this._xterm.writeln("\n" + obj.error)
        break
      default:
        console.warn(`unknow command : `, obj)
    }
  }
  onClickFontSize() {
    if (!this._xterm || this.fontSize < 5 || !isNumber(this.fontSize) || !this._fitAddon) {
      return
    }
    if (this.fontSize == this._xterm.getOption("fontSize")) {
      return
    }
    this._xterm.setOption("fontSize", this.fontSize)
    this._fitAddon.fit()
  }
}
