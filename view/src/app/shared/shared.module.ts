import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';

import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';

import { MatIconModule } from '@angular/material/icon';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatButtonModule } from '@angular/material/button';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDialogModule } from '@angular/material/dialog';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatSlideToggleModule } from '@angular/material/slide-toggle';
import { MatInputModule } from '@angular/material/input';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';

import { NavigationBarComponent } from './navigation-bar/navigation-bar.component';
import { LoginComponent } from './login/login.component';
import { ConfirmComponent } from './dialog/confirm/confirm.component';
import { PasswordComponent } from './password/password.component';



@NgModule({
  declarations: [NavigationBarComponent, LoginComponent, ConfirmComponent, PasswordComponent],
  imports: [
    CommonModule, RouterModule, FormsModule,
    MatIconModule, MatToolbarModule, MatButtonModule,
    MatTooltipModule, MatProgressSpinnerModule, MatDialogModule,
    MatProgressBarModule, MatFormFieldModule, MatSlideToggleModule,
    MatInputModule, MatMenuModule, MatDividerModule,
  ],
  exports: [
    NavigationBarComponent,
    ConfirmComponent,
  ],
  entryComponents: [
    LoginComponent, ConfirmComponent, PasswordComponent,
  ],
})
export class SharedModule { }
