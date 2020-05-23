import { Component, OnInit, OnDestroy } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { MatDialog } from '@angular/material/dialog';
import { SessionService } from 'src/app/core/session/session.service';
import { ServerAPI } from 'src/app/core/core/api';

interface Source {
  platform: string
  maxprocs: number
  cgos: number
  cpus: number
  goroutines: number
}

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss']
})
export class ListComponent implements OnInit, OnDestroy {
  constructor(private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private sessionService: SessionService,
  ) { }
  err: any
  private _closed = false
  private _disabled = false
  get disabled(): boolean {
    return this._disabled
  }
  source: Source
  ngOnInit(): void {
    this._disabled = true
    this.sessionService.ready.then(() => {
      if (this._closed) {
        return
      }
      this.load()
    })
  }
  ngOnDestroy() {
    this._closed = true
  }
  load() {
    this._disabled = true
    ServerAPI.v1.debug.get<Source>(this.httpClient)
      .then((data) => {
        if (this._closed) {
          return
        }
        this.source = data
      }, (e) => {
        if (this._closed) {
          return
        }
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
