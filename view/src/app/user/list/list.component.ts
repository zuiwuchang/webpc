import { Component, OnInit, OnDestroy } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialog } from '@angular/material/dialog';
import { SessionService } from 'src/app/core/session/session.service';
import { User } from '../user';
import { ServerAPI } from 'src/app/core/core/api';
import { ConfirmComponent } from 'src/app/shared/dialog/confirm/confirm.component';
import { AddComponent } from '../add/add.component';
import { PasswordComponent } from '../password/password.component';
import { ChangeComponent } from '../change/change.component';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss']
})
export class ListComponent implements OnInit, OnDestroy {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialog: MatDialog,
    private sessionService: SessionService,
  ) { }
  private _ready = false
  get ready(): boolean {
    return this._ready
  }
  err: any
  private _closed = false
  private _disabled = false
  get disabled(): boolean {
    return this._disabled
  }
  private _source = new Array<User>()
  get source(): Array<User> {
    return this._source
  }
  ngOnInit(): void {
    this.sessionService.ready.then(() => {
      if (this._closed) {
        return
      }
      this.load()
    })
  }
  ngOnDestroy(): void {
    this._closed = true
  }
  load() {
    this.err = null
    this._ready = false
    ServerAPI.v1.users.get<Array<User>>(this.httpClient).then((data) => {
      if (this._closed) {
        return
      }
      if (data && data.length > 0) {
        this._source.push(...data)
        this._source.sort(User.compare)
      }
    }, (e) => {
      if (this._closed) {
        return
      }
      this.err = e
    }).finally(() => {
      this._ready = true
    })
  }
  onClickEdit(node: User) {
    this.matDialog.open(ChangeComponent, {
      data: node,
      disableClose: true,
    })
  }
  onClickPassword(node: User) {
    this.matDialog.open(PasswordComponent, {
      data: node.name,
      disableClose: true,
    })
  }

  onClickDelete(node: User) {
    this.matDialog.open(ConfirmComponent, {
      data: {
        title: this.i18nService.get("delete user"),
        content: `${this.i18nService.get("delete user")} : ${node.name}`,
      },
    }).afterClosed().toPromise().then((data) => {
      if (this._closed || !data) {
        return
      }
      this._delete(node)
    })
  }
  private _delete(node: User) {
    this._disabled = true
    ServerAPI.v1.users.deleteOne(this.httpClient, node.name)
      .then(() => {
        const index = this._source.indexOf(node)
        this._source.splice(index, 1)
        this.toasterService.pop('success',
          this.i18nService.get('success'),
          this.i18nService.get('user has been deleted'),
        )
      }, (e) => {
        console.warn(e)
        this.toasterService.pop('error',
          this.i18nService.get('error'),
          e,
        )
      }).finally(() => {
        this._disabled = false
      })
  }
  onClickAdd() {
    this.matDialog.open(AddComponent, {
      disableClose: true,
    }).afterClosed().toPromise().then((data) => {
      if (this._closed || !data) {
        return
      }
      this._source.push(data)
      this._source.sort(User.compare)
    })
  }
}
