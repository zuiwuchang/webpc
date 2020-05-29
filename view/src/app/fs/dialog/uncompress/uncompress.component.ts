import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialog } from '@angular/material/dialog';
import { FileInfo, Dir } from '../../fs';
import { isString, isNumber } from 'util';
import { interval, Subscription } from 'rxjs';
import { ExistComponent } from '../exist/exist.component';
import { ExistChoiceComponent } from '../exist-choice/exist-choice.component';

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
// CmdError 錯誤
const CmdError = 1
// CmdHeart websocket 心跳防止瀏覽器 關閉不獲取 websocket
const CmdHeart = 2
// CmdProgress 更新進度
const CmdProgress = 3
// CmdDone 操作完成
const CmdDone = 4
// CmdInit 初始化
const CmdInit = 5
// CmdYes 確認操作
const CmdYes = 6
// CmdNo 取消操作
const CmdNo = 7
// CmdExist 檔案已經存在
const CmdExist = 8
// CmdYesAll 覆蓋全部 重複檔案
const CmdYesAll = 9
// CmdSkip 跳過 重複檔案
const CmdSkip = 10
// CmdSkipAll 跳過全部 重複檔案
const CmdSkipAll = 11

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
        this._websocket.send(JSON.stringify({
          cmd: CmdHeart,
        }))
      }
    })

    this._init()
  }
  private _websocket: WebSocket
  private _closed: boolean
  progress: string
  ngOnDestroy() {
    this._closed = true
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
        if (isString(evt.data)) {
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
        'cmd': CmdInit,
        'name': this.target.source.name,
      }))
    }
  }
  _onMessage(websocket: WebSocket, msg: Message) {
    switch (msg.cmd) {
      case CmdError:
        this.toasterService.pop('error', undefined, msg.error)
        websocket.close()
        this._websocket = null
        this.matDialogRef.close()
        break;
      case CmdProgress:
        this.progress = msg.val
        break;
      case CmdDone:
        this._websocket.close()
        this._websocket = null
        this.toasterService.pop('success', undefined, this.i18nService.get('Uncompress done'))
        this.matDialogRef.close(true)
        break
      case CmdExist:
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
      if (!websocket || websocket != this._websocket || !isNumber(number)) {
        websocket.close()
        return
      }
      if (number == CmdYes) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else if (number == CmdYesAll) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else if (number == CmdSkip) {
        websocket.send(JSON.stringify({
          cmd: number,
        }))
      } else if (number == CmdSkipAll) {
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
          cmd: CmdNo,
        }))
        websocket.close()
        this.matDialogRef.close()
      }
    })
  }
}
