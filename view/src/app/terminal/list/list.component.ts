import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialog } from '@angular/material/dialog';
import { SessionService } from 'src/app/core/session/session.service';
import { Shell } from '../shell';
import { ServerAPI } from 'src/app/core/core/api';

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
  err: any
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
        this.err = e
      }).finally(() => {
        this._disabled = false
      })
  }
  onClickAdd() {

  }
}
