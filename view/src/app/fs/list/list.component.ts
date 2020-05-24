import { Component, OnInit } from '@angular/core';
import { ActivatedRoute } from '@angular/router';
import { ServerAPI } from 'src/app/core/core/api';

@Component({
  selector: 'app-list',
  templateUrl: './list.component.html',
  styleUrls: ['./list.component.scss']
})
export class ListComponent implements OnInit {
  private _id = 0
  constructor(
    private route: ActivatedRoute,
  ) { }

  ngOnInit(): void {
    this.route.queryParamMap.subscribe((param) => {

      const root = param.get(`root`)
      const path = param.get(`path`)


      console.log(`root = ${root}`)
      console.log(`path = ${path}`)
    })
  }
}
