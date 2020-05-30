import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA } from '@angular/material/dialog';
import { isString } from 'util';
import { interval, Subscription } from 'rxjs';
import { NetCommand } from '../command';
export interface Dir {
  root: string
  dir: string
}
export interface Data {
  names: Array<String>
  src: Dir
  dst: Dir
}
export interface Message {
  cmd: number
  error: string
  val: string
}
@Component({
  selector: 'app-cut',
  templateUrl: './cut.component.html',
  styleUrls: ['./cut.component.scss']
})
export class CutComponent implements OnInit, OnDestroy {
  constructor(private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialogRef: MatDialogRef<CutComponent>,
    @Inject(MAT_DIALOG_DATA) public data: Data, ) { }
  private _subscriptionPing: Subscription
  ngOnInit(): void {
    this._subscriptionPing = interval(1000 * 30).subscribe(() => {
      if (this._websocket) {
        this._websocket.send(JSON.stringify({
          cmd: NetCommand.Heart,
        }))
      }
    })

    this._init()
  }
  private _websocket: WebSocket
  progress: string
  ngOnDestroy() {
    if (this._websocket) {
      this._websocket.close()
      this._websocket = null
    }
    this._subscriptionPing.unsubscribe()
  }
  onClose() {
    this.matDialogRef.close()
  }
  _init() {
    const url = ServerAPI.v1.fs.websocketURL([
      this.data.dst.root, this.data.dst.dir,
      'cut',
      this.data.src.root, this.data.src.dir,
    ])
    const websocket = new WebSocket(url)
    this._websocket = websocket
    websocket.onerror = (evt) => {
      websocket.close()
      console.warn(evt)
      if (this._websocket != websocket) {
        return
      }
      this.toasterService.pop('error', undefined, 'connect websocket error')
      this._websocket = null
      this.matDialogRef.close()
    }

    websocket.onopen = (evt) => {
      if (this._websocket != websocket) {
        websocket.close()
        return
      }
      websocket.onclose = (evt) => {
        websocket.close()
        console.warn(evt, this._websocket != websocket)
        if (this._websocket != websocket) {
          return
        }
        this.toasterService.pop('error', undefined, 'websocket closed')
        this._websocket = null
        this.matDialogRef.close()
      }
      websocket.onmessage = (evt) => {
        if (this._websocket != websocket) {
          websocket.close()
          return
        }
        if (isString(evt.data)) {
          try {
            this._onMessage(websocket, JSON.parse(evt.data))
          } catch (e) {
            console.warn('ws-cut', e)
          }
        } else {
          console.warn(`ws-cut unknow type`, evt.data)
        }
      }
      // send names
      websocket.send(JSON.stringify({
        'cmd': NetCommand.Init,
        'names': this.data.names,
      }))
    }
  }
  _onMessage(websocket: WebSocket, msg: Message) {
    switch (msg.cmd) {
      case NetCommand.Error:
        this.toasterService.pop('error', undefined, msg.error)
        websocket.close()
        this._websocket = null
        this.matDialogRef.close()
        break;
      case NetCommand.Progress:
        this.progress = msg.val
        break;
      case NetCommand.Done:
        this._websocket.close()
        this._websocket = null
        this.toasterService.pop('success', undefined, this.i18nService.get('Move file done'))
        this.matDialogRef.close(true)
        break
      default:
        console.warn(`ws-cut unknow msg`, msg)
        break;
    }
  }
}
