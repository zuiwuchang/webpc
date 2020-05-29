import { Component, OnInit, Inject, OnDestroy } from '@angular/core';
import { ServerAPI } from 'src/app/core/core/api';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialogRef, MAT_DIALOG_DATA, MatDialog } from '@angular/material/dialog';
import { FileInfo, Dir } from '../../fs';
import { isString } from 'util';
import { interval, Subscription } from 'rxjs';
import { ExistComponent } from '../exist/exist.component';

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

  }
}
