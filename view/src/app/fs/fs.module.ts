import { NgModule } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';

import { MatListModule } from '@angular/material/list';
import { MatCardModule } from '@angular/material/card';
import { MatProgressSpinnerModule } from '@angular/material/progress-spinner';
import { MatButtonModule } from '@angular/material/button';
import { MatIconModule } from '@angular/material/icon';
import { MatTooltipModule } from '@angular/material/tooltip';
import { MatFormFieldModule } from '@angular/material/form-field';
import { MatInputModule } from '@angular/material/input';
import { MatRippleModule } from '@angular/material/core';
import { MatToolbarModule } from '@angular/material/toolbar';
import { MatCheckboxModule } from '@angular/material/checkbox';
import { MatMenuModule } from '@angular/material/menu';
import { MatDividerModule } from '@angular/material/divider';
import { MatDialogModule } from '@angular/material/dialog';
import { MatRadioModule } from '@angular/material/radio';

import { FsRoutingModule } from './fs-routing.module';
import { ListComponent } from './list/list.component';
import { ManagerComponent } from './manager/manager.component';
import { PathComponent } from './path/path.component';
import { FileComponent } from './file/file.component';
import { ImageComponent } from './view/image/image.component';
import { TextComponent } from './view/text/text.component';
import { AudioComponent } from './view/audio/audio.component';
import { VideoComponent } from './view/video/video.component';
import { RenameComponent } from './dialog/rename/rename.component';
import { NewFileComponent } from './dialog/new-file/new-file.component';
import { NewFolderComponent } from './dialog/new-folder/new-folder.component';
import { PropertyComponent } from './dialog/property/property.component';
import { RemoveComponent } from './dialog/remove/remove.component';
import { CompressComponent } from './dialog/compress/compress.component';
import { ExistComponent } from './dialog/exist/exist.component';
import { UncompressComponent } from './dialog/uncompress/uncompress.component';
import { ExistChoiceComponent } from './dialog/exist-choice/exist-choice.component';
import { CutComponent } from './dialog/cut/cut.component';
import { CopyComponent } from './dialog/copy/copy.component';


@NgModule({
  declarations: [ListComponent, ManagerComponent, PathComponent,
    FileComponent, ImageComponent, TextComponent,
    AudioComponent, VideoComponent, RenameComponent, NewFileComponent, NewFolderComponent, PropertyComponent, RemoveComponent, CompressComponent, ExistComponent, UncompressComponent, ExistChoiceComponent, CutComponent, CopyComponent
  ],
  imports: [
    CommonModule, RouterModule, FormsModule,
    MatListModule, MatCardModule, MatProgressSpinnerModule,
    MatButtonModule, MatIconModule, MatTooltipModule,
    MatFormFieldModule, MatInputModule, MatRippleModule,
    MatToolbarModule, MatCheckboxModule, MatMenuModule,
    MatDividerModule, MatDialogModule, MatRadioModule,
    FsRoutingModule
  ],
  entryComponents: [
    RenameComponent, NewFileComponent, NewFolderComponent,
    PropertyComponent, RemoveComponent, CompressComponent,
    ExistComponent, UncompressComponent, ExistChoiceComponent,
    CutComponent, CopyComponent,
  ]
})
export class FsModule { }
