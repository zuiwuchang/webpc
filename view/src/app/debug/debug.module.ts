import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatListModule } from '@angular/material/list';

import { DebugRoutingModule } from './debug-routing.module';
import { ListComponent } from './list/list.component';



@NgModule({
  declarations: [ListComponent],
  imports: [
    CommonModule,

    MatButtonModule, MatIconModule, MatCardModule,
    MatProgressSpinnerModule, MatListModule,

    DebugRoutingModule
  ]
})
export class DebugModule { }
