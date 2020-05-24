import { Component, OnInit, VERSION } from '@angular/core';
import { ServerAPI } from 'src/app/core/core/api';
import { HttpClient } from '@angular/common/http';
interface Version {
  tag: string
  commit: string
  date: string
}
@Component({
  selector: 'app-about',
  templateUrl: './about.component.html',
  styleUrls: ['./about.component.scss']
})
export class AboutComponent implements OnInit {
  VERSION = VERSION
  version: Version
  constructor(private httpClient: HttpClient,
  ) { }

  ngOnInit(): void {
    ServerAPI.v1.version.get<Version>(this.httpClient).then((data) => {
      this.version = data
    })
  }

}
