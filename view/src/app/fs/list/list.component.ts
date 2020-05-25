import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ServerAPI } from 'src/app/core/core/api';
import { HttpClient } from '@angular/common/http';
import { ToasterService } from 'angular2-toaster';
import { SessionService } from 'src/app/core/session/session.service';
import { Subscription } from 'rxjs';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { isArray } from 'util';
import { FileInfo, Dir } from '../fs';


interface LSResponse {
  dir: Dir
  items: Array<FileInfo>
}

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss']
})
export class ListComponent implements OnInit, OnDestroy {
  constructor(
    private route: ActivatedRoute,
    private httpClient: HttpClient,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private sessionService: SessionService,
  ) { }
  private _id = 0
  private _closed = false
  private _subscription: Subscription
  query: boolean
  dir: Dir
  private _source: Array<FileInfo>
  get source(): Array<FileInfo> {
    return this._source
  }
  ngOnInit(): void {
    this.sessionService.ready.then(() => {
      if (this._closed) {
        return
      }
      this._subscription = this.route.queryParamMap.subscribe((param) => {
        const root = param.get(`root`)
        const path = param.get(`path`)
        if (!this.dir) {
          this.dir = {
            root: root,
            read: false,
            write: false,
            shared: false,
            dir: '',
          }
        }
        this._id++
        const id = this._id
        this.query = true
        ServerAPI.v1.fs.get<LSResponse>(this.httpClient, {
          params: {
            root: root,
            path: path || '/',
          },
        }).then((response) => {
          if (this._closed || this._id != id) {
            return
          }

          this.dir = response.dir
          if (isArray(response.items) && response.items.length > 0) {
            this._source = new Array<FileInfo>()
            for (let i = 0; i < response.items.length; i++) {
              this._source.push(new FileInfo(this.dir.root, this.dir.dir, response.items[i]))
            }
            this._source.sort(FileInfo.compare)
          } else {
            this._source = null
          }
        }, (e) => {
          if (this._closed || this._id != id) {
            return
          }
          this.toasterService.pop('error',
            this.i18nService.get('error'),
            e,
          )
        }).finally(() => {
          if (this._closed || this._id != id) {
            return
          }
          this.query = false
        })
      })
    })
  }
  get root(): string {
    if (this.dir) {
      return this.dir.root
    }
    return ''
  }
  ngOnDestroy() {
    this._closed = true
    if (this._subscription) {
      this._subscription.unsubscribe()
    }
  }
}
