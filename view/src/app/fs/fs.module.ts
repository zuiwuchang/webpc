import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

import { MatListModule } from '@angular/material/list';

import { FsRoutingModule } from './fs-routing.module';
import { ListComponent } from './list/list.component';


@NgModule({
  declarations: [ListComponent],
  imports: [
    CommonModule, RouterModule,
    MatListModule,
    FsRoutingModule
  ]
})
export class FsModule { }
