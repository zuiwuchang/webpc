import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialog } from '@angular/material/dialog';
import { FileInfo, Dir } from '../../fs';
import { interval, Subscription } from 'rxjs';
import { ExistChoiceComponent } from '../exist-choice/exist-choice.component';
import { NetCommand, NetHeartMessage } from '../command';
interface Target {
  dir: Dir
  source: FileInfo
}
interface Message {
  cmd: number
  error: string
  val: string
  fileInfo: FileInfo
}


@Component({
  selector: 'app-uncompress',
  templateUrl: './uncompress.component.html',
  styleUrls: ['./uncompress.component.scss']
})
export class UncompressComponent implements OnInit, OnDestroy {
  constructor(private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialog: MatDialog,
    private matDialogRef: MatDialogRef<UncompressComponent>,
    @Inject(MAT_DIALOG_DATA) public target: Target,
  ) { }
  private _subscriptionPing: Subscription
  ngOnInit(): void {
    if (!this.target || !this.target.source) {
      this.matDialogRef.close()
      return
    }
    this._subscriptionPing = interval(1000 * 30).subscribe(() => {
      if (this._websocket) {
        this._websocket.send(NetHeartMessage)
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
      this.target.dir.root, this.target.dir.dir,
      'uncompress',
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
        if (typeof evt.data === "string") {
          try {
            this._onMessage(websocket, JSON.parse(evt.data))
          } catch (e) {
            console.warn('ws-compress', e)
          }
        } else {
          console.warn(`ws-compress unknow type`, evt.data)
        }
      }
      // send names
      websocket.send(JSON.stringify({
        'cmd': NetCommand.Init,
        'name': this.target.source.name,
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
        this.toasterService.pop('success', undefined, this.i18nService.get('Uncompress done'))
        this.matDialogRef.close(true)
        break
      case NetCommand.Exist:
        this._exist(websocket, msg.val)
        break
      default:
        console.warn(`ws-compress unknow msg`, msg)
        break;
    }
  }
  private _exist(websocket: WebSocket, name: string) {
    this.matDialog.open(ExistChoiceComponent, {
      data: name,
      disableClose: true,
    }).afterClosed().toPromise().then((number) => {
      if (!websocket || websocket != this._websocket || typeof number !== "number") {
        websocket.close()
        return
      }
      if (number == NetCommand.Yes) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else if (number == NetCommand.YesAll) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else if (number == NetCommand.Skip) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else if (number == NetCommand.SkipAll) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else {
        if (websocket != this._websocket) {
          websocket.close()
          return
        }
        this._websocket = null
        websocket.send(JSON.stringify({
          cmd: NetCommand.No,
        }))
        websocket.close()
        this.matDialogRef.close()
      }
    })
  }
}
