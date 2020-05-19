import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { UserRoutingModule } from './user-routing.module';
import { ListComponent } from './list/list.component';
import { AddComponent } from './add/add.component';
import { PasswordComponent } from './password/password.component';
import { ChangeComponent } from './change/change.component';


@NgModule({
  declarations: [ListComponent, AddComponent, PasswordComponent, ChangeComponent],
  imports: [
    CommonModule,
    UserRoutingModule
  ]
})
export class UserModule { }
