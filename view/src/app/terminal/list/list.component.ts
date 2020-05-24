import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialog } from '@angular/material/dialog';
import { SessionService } from 'src/app/core/session/session.service';
import { Shell } from '../shell';
import { ServerAPI } from 'src/app/core/core/api';
import { EditComponent } from '../edit/edit.component';
import { ConfirmComponent } from 'src/app/shared/dialog/confirm/confirm.component';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss']
})
export class ListComponent implements OnInit {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private matDialog: MatDialog,
    private sessionService: SessionService,
  ) { }
  private _closed = false
  private _disabled = false
  get disabled(): boolean {
    return this._disabled
  }
  private _source = new Array<Shell>()
  get source(): Array<Shell> {
    return this._source
  }
  ngOnInit(): void {
    this._disabled = true
    this.sessionService.ready.then(() => {
      if (this._closed) {
        return
      }
      this.load()
    })
  }
  load() {
    this._disabled = true
    this._source.splice(0, this._source.length)
    ServerAPI.v1.shells.get<Array<Shell>>(this.httpClient)
      .then((data) => {
        if (this._closed) {
          return
        }
        if (data && data.length > 0) {
          this._source.push(...data)
          this._source.sort(Shell.compare)
        }
      }, (e) => {
        if (this._closed) {
          return
        }
        this.toasterService.pop('error',
          this.i18nService.get('error'),
          e,
        )
      }).finally(() => {
        this._disabled = false
      })
  }
  onClickEdit(node: Shell) {
    this.matDialog.open(EditComponent, {
      data: node,
      disableClose: true,
    })
  }
  onClickDelete(node: Shell) {
    this.matDialog.open(ConfirmComponent, {
      data: {
        title: this.i18nService.get("delete terminal"),
        content: `${this.i18nService.get("delete terminal")} : ${node.name}`,
      },
    }).afterClosed().toPromise().then((data) => {
      if (this._closed || !data) {
        return
      }
      this._delete(node)
    })
  }
  private _delete(node: Shell) {
    this._disabled = true
    ServerAPI.v1.shells.deleteOne(this.httpClient, node.id)
      .then(() => {
        const index = this._source.indexOf(node)
        this._source.splice(index, 1)
        this.toasterService.pop('success',
          this.i18nService.get('success'),
          this.i18nService.get('terminal has been deleted'),
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
}
