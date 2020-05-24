import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule } from '@angular/common/http';
import { RouterModule } from '@angular/router';

import { ToasterModule, ToasterService } from 'angular2-toaster';

import { SharedModule } from './shared/shared.module';

import { MatCardModule } from '@angular/material/card';
import { MatListModule } from '@angular/material/list';
import { MatIconModule } from '@angular/material/icon';
import { MatButtonModule } from '@angular/material/button';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';

import { AppRoutingModule } from './app-routing.module';
import { AppComponent } from './app.component';
import { HomeComponent } from './app/home/home.component';

@NgModule({
  declarations: [
    AppComponent,
    HomeComponent
  ],
  imports: [
    BrowserModule,
    BrowserAnimationsModule, HttpClientModule, RouterModule,
    SharedModule,

    MatCardModule, MatProgressSpinnerModule, MatListModule,
    MatIconModule, MatButtonModule,

    AppRoutingModule,
    ToasterModule.forRoot(),
  ],
  providers: [ToasterService],
  bootstrap: [AppComponent]
})
export class AppModule { }
