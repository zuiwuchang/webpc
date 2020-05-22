import { Component, OnInit, OnDestroy, AfterViewInit, ViewChild, ElementRef } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import { WebLinksAddon } from 'xterm-addon-web-links';
import { Subject, Subscription } from 'rxjs'
import { debounceTime } from 'rxjs/operators';
import { ServerAPI, getWebSocketAddr } from 'src/app/core/core/api';
interface Size {
  w: number
  h: number
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
  private _disabled = false
  private _subject = new Subject()
  private _subscription: Subscription
  get disabled(): boolean {
    return this._disabled
  }
  private _xterm: Terminal
  ngOnInit(): void {
  }
  ngOnDestroy() {
    this._closed = true
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

    xterm.write('wait connect server $ ')
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

    const id = this.route.snapshot.paramMap.get(`id`)
    this._connect(id)
  }
  private _connect(id: string) {

    const url = getWebSocketAddr(`/ws${ServerAPI.v1.shells.baseURL}/${id}/${this._xterm.cols}/${this._xterm.rows}`)
    const websocket = new WebSocket(url)
    console.log(websocket)
  }
  onResize() {
    this._subject.next(new Date())
  }
}
