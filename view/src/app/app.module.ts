import { BrowserModule } from '@angular/platform-browser';
import { NgModule } from '@angular/core';
import { BrowserAnimationsModule } from '@angular/platform-browser/animations';
import { HttpClientModule } from '@angular/common/http';
import { ToasterModule, ToasterService } from 'angular2-toaster';

import { SharedModule } from './shared/shared.module';

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
    BrowserAnimationsModule, HttpClientModule,
    SharedModule,
    AppRoutingModule,
    ToasterModule.forRoot(),
  ],
  providers: [ToasterService],
  bootstrap: [AppComponent]
})
export class AppModule { }
