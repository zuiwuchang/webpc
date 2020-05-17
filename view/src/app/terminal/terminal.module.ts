import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { TerminalRoutingModule } from './terminal-routing.module';
import { ListComponent } from './list/list.component';


@NgModule({
  declarations: [ListComponent],
  imports: [
    CommonModule,
    TerminalRoutingModule
  ]
})
export class TerminalModule { }
