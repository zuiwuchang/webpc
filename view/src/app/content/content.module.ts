import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { MatCardModule } from '@angular/material/card';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';

import { ContentRoutingModule } from './content-routing.module';
import { LicenseComponent } from './license/license.component';
import { AboutComponent } from './about/about.component';


@NgModule({
  declarations: [LicenseComponent, AboutComponent],
  imports: [
    CommonModule,
    MatCardModule, MatListModule, MatIconModule,
    ContentRoutingModule
  ]
})
export class ContentModule { }
