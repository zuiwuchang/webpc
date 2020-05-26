import { Component, OnInit, OnDestroy } from '@angular/core';
import { Router, ActivatedRoute } from '@angular/router';
import { SessionService } from 'src/app/core/session/session.service';
import { ServerAPI } from 'src/app/core/core/api';
import { isString } from 'util';
import { ToasterService } from 'angular2-toaster';
import { I18nService } from 'src/app/core/i18n/i18n.service';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { resolveError } from 'src/app/core/core/restful';

@Component({
  selector: 'app-text',
  templateUrl: './text.component.html',
  styleUrls: ['./text.component.scss']
})
export class TextComponent implements OnInit, OnDestroy {
  private _closed = false
  ready: boolean
  root: string
  filepath: string
  name: string
  dir: string
  loading: boolean
  saving: boolean
  val: string
  private _val: string
  constructor(private router: Router,
    private route: ActivatedRoute,
    private sessionService: SessionService,
    private toasterService: ToasterService,
    private i18nService: I18nService,
    private httpClient: HttpClient,
  ) { }

  ngOnInit(): void {
    this.sessionService.ready.then(() => {
      if (this._closed) {
        return
      }
      const param = this.route.snapshot.queryParamMap
      const root = param.get(`root`)
      const path = param.get(`path`)
      this.root = root
      this.filepath = path
      this.name = path
      const index = path.lastIndexOf('/')
      if (index != -1) {
        this.dir = path.substring(0, index)
        this.name = path.substring(index + 1)
      }
      this.ready = true
      this.onClickLoad()
    })
  }
  ngOnDestroy() {
    this._closed = true
  }
  onPathChange(path: string) {
    if (!isString(path)) {
      path = '/'
    }
    if (!path.startsWith('/')) {
      path = '/' + path
    }

    this.router.navigate(['fs', 'list'], {
      queryParams: {
        root: this.root,
        path: path,
      }
    })
  }
  canDeactivate(): boolean {
    if (this.saving) {
      this.toasterService.pop('warning',
        undefined,
        this.i18nService.get('Wait for data to be saved'),
      )
      return false
    }
    return true
  }
  onClickLoad() {
    this.loading = true
    ServerAPI.v1.fs.getOneText(this.httpClient,
      [this.root, this.filepath],
      {
        responseType: 'text',
      },
    ).then((data) => {
      this.val = data
      this._val = data
    }, (e) => {
      this.toasterService.pop('error', undefined, e)
    }).finally(() => {
      this.loading = false
    })
  }
  get isNotChanged(): boolean {
    return this.val == this._val
  }
  onCLickSave() {
    this.saving = true
    ServerAPI.v1.fs.putOne(this.httpClient,
      [
        this.root,
        this.filepath
      ],
      {
        val: this.val,
      },
    ).then(() => {
      console.log('ok')
    }, (e) => {
      this.toasterService.pop('error', undefined, e)
    }).finally(() => {
      this.saving = false
    })
  }
}
