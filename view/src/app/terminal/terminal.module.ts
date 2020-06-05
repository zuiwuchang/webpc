import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { TerminalRoutingModule } from './terminal-routing.module';

import { MatListModule } from '@angular/material/list';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatDialogModule } from '@angular/material/dialog';
import { MatProgressBarModule } from '@angular/material/progress-bar';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatMenuModule } from '@angular/material/menu';
import { MatAutocompleteModule } from '@angular/material/autocomplete';

import { ListComponent } from './list/list.component';
import { ViewComponent } from './view/view.component';
import { EditComponent } from './edit/edit.component';
import { SettingsComponent } from './dialog/settings/settings.component';


@NgModule({
  declarations: [ListComponent, ViewComponent, EditComponent, SettingsComponent],
  imports: [
    CommonModule, FormsModule,

    MatListModule, MatButtonModule, MatIconModule,
    MatTooltipModule, MatFormFieldModule, MatInputModule,
    MatCardModule, MatProgressSpinnerModule, MatDialogModule,
    MatProgressBarModule, MatCheckboxModule, MatToolbarModule,
    MatMenuModule, MatAutocompleteModule,

    TerminalRoutingModule,
  ],
  entryComponents: [EditComponent, SettingsComponent],
})
export class TerminalModule { }
