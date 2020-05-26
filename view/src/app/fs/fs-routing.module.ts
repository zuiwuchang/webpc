import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { ListComponent } from './list/list.component';
import { ImageComponent } from './view/image/image.component';
import { VideoComponent } from './view/video/video.component';
import { AudioComponent } from './view/audio/audio.component';
import { TextComponent } from './view/text/text.component';

const routes: Routes = [
  {
    path: 'list',
    component: ListComponent,
  },
  {
    path: 'view/video',
    component: VideoComponent,
  },
  {
    path: 'view/audio',
    component: AudioComponent,
  },
  {
    path: 'view/image',
    component: ImageComponent,
  },
  {
    path: 'view/text',
    component: TextComponent,
  },
];

@NgModule({
  imports: [RouterModule.forChild(routes)],
  exports: [RouterModule]
})
export class FsRoutingModule { }
