import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild, ElementRef } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { Subject, Subscription } from 'rxjs'
import { debounceTime } from 'rxjs/operators';
import { ServerAPI, getWebSocketAddr } from 'src/app/core/core/api';

// DataTypeTTY tty 消息
const DataTypeTTY = 1
// DataTypeError 錯誤
const DataTypeError = 2
// DataTypeResize 更改大小
const DataTypeResize = 3

@Component({
  selector: 'app-view',
  templateUrl: './view.component.html',
  styleUrls: ['./view.component.scss']
})
export class ViewComponent implements OnInit, OnDestroy, AfterViewInit {
  constructor(private route: ActivatedRoute,
  ) { }
  private _closed = false
  private _disabled = false
  private _subject = new Subject()
  private _subscription: Subscription
  get disabled(): boolean {
    return this._disabled
  }
  private _xterm: Terminal
  private _websocket: WebSocket
  ngOnInit(): void {
  }
  ngOnDestroy() {
    this._closed = true
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

    // 加載插件
    const fitAddon = new FitAddon()
    xterm.loadAddon(fitAddon)
    xterm.loadAddon(new WebLinksAddon())

    xterm.open(this.xterm.nativeElement)
    fitAddon.fit()

    xterm.writeln('wait connect server')
    // 訂閱 窗口大小 改變
    this._subscription = this._subject.pipe(
      debounceTime(100)
    ).subscribe((_) => {
      if (this._closed) {
        return
      }
      fitAddon.fit()
    })
    xterm.onResize((evt) => {
      console.log(evt)
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
  private _connect(id: number) {
    const url = getWebSocketAddr(`/ws${ServerAPI.v1.shells.baseURL}/${id}/${this._xterm.cols}/${this._xterm.rows}`)
    console.log(url)
    const websocket = new WebSocket(url)
    this._websocket = websocket
    websocket.binaryType = "arraybuffer"
    websocket.onerror = (evt) => {
      this._xterm.writeln("websocket error")
      console.log(evt)
    }
    websocket.onopen = (evt) => {
      this._xterm.onData(function (data) {
        websocket.send(new TextEncoder().encode(data))
      })
      this._xterm.onResize(function (evt) {
        websocket.send(JSON.stringify({
          what: DataTypeResize,
          cols: evt.cols,
          rows: evt.rows,
        }))
      })
      let first = true
      websocket.onmessage = (evt) => {
        if (evt.data instanceof ArrayBuffer) {
          if (first) {
            first = false
            this._xterm.focus()
          }
          this._xterm.write(new Uint8Array(evt.data))
        } else {
          console.log(evt.data)
        }
      }
      websocket.onclose = (evt) => {
        this._xterm.writeln("\nSession terminated")
        this._xterm.setOption("cursorBlink", false)
      }
    }
  }
  onResize() {
    this._subject.next(new Date())
  }
}
