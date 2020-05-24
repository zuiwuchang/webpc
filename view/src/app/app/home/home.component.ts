import { Component, OnInit } from '@angular/core';
import { HttpClient } from '@angular/common/http';
import { SessionService } from 'src/app/core/session/session.service';
import { ServerAPI } from 'src/app/core/core/api';

@Component({
  selector: 'app-home',
  templateUrl: './home.component.html',
  styleUrls: ['./home.component.scss']
})
export class HomeComponent implements OnInit {

  constructor(private httpClient: HttpClient,
    private sessionService: SessionService,
  ) { }
  private _ready = false
  get ready(): boolean {
    return this._ready
  }
  err: any
  private _closed = false
  private _source = new Array<string>()
  get source(): Array<string> {
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
    this._source.splice(0, this._source.length)
    ServerAPI.v1.roots.get<Array<string>>(this.httpClient).then((data) => {
      if (this._closed) {
        return
      }
      if (data && data.length > 0) {
        this._source.push(...data)
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
}
