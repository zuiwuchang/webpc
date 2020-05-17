import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { ContentRoutingModule } from './content-routing.module';
import { LicenseComponent } from './license/license.component';
import { AboutComponent } from './about/about.component';


@NgModule({
  declarations: [LicenseComponent, AboutComponent],
  imports: [
    CommonModule,
    ContentRoutingModule
  ]
})
export class ContentModule { }
