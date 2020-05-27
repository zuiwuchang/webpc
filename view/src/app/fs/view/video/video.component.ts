import { Component, OnInit, OnDestroy } from '@angular/core';
import { ActivatedRoute, Router } from '@angular/router';
import { SessionService } from 'src/app/core/session/session.service';
import { ServerAPI } from 'src/app/core/core/api';
import { isString } from 'util';

@Component({
  selector: 'app-video',
  templateUrl: './video.component.html',
  styleUrls: ['./video.component.scss']
})
export class VideoComponent implements OnInit, OnDestroy {
  constructor(private router: Router,
    private route: ActivatedRoute,
    private sessionService: SessionService,
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
      const index = path.lastIndexOf('/')
      if (index != -1) {
        this.dir = path.substring(0, index)
      }
      this.url = ServerAPI.v1.fs.oneURL([root, path])
      this.ready = true
    })
  }
  private _closed = false
  ready: boolean
  root: string
  filepath: string
  dir: string
  url: string
  stretch = true
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
}
