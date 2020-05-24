import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ServerAPI } from 'src/app/core/core/api';
import { HttpClient } from '@angular/common/http';
import { ToasterService } from 'angular2-toaster';
import { SessionService } from 'src/app/core/session/session.service';
import { Subscription } from 'rxjs';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { isArray, isString } from 'util';

interface Dir {
  root: string
  read?: boolean
  write?: boolean
  shared?: boolean
  dir?: string
}
class FileInfo {
  name: string
  mode: number
  size: number
  isDir: boolean

  filename: string
  root: string
  constructor(root: string, dir: string, other: FileInfo) {
    this.name = other.name
    this.mode = other.mode
    this.size = other.size
    this.isDir = other.isDir
    if (dir.endsWith('/')) {
      this.filename = dir + other.name
    } else {
      this.filename = dir + '/' + other.name
    }
    this.root = root
  }
}
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
