import { NgModule } from '@angular/core';
import { Routes, RouterModule } from '@angular/router';
import { HomeComponent } from './app/home/home.component';

const routes: Routes = [
  {
    path: '',
    component: HomeComponent,
  },
  {
    path: 'content',
    loadChildren: () => import('./content/content.module').then(m => m.ContentModule),
  },
  {
    path: 'terminal',
    loadChildren: () => import('./terminal/terminal.module').then(m => m.TerminalModule),
  },
];

@NgModule({
  imports: [RouterModule.forRoot(routes)],
  exports: [RouterModule]
})
export class AppRoutingModule { }
