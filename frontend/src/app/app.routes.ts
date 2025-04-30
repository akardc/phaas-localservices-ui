import { Routes } from '@angular/router';
import { SettingsComponent } from './settings/settings.component';
import { RepoListComponent } from './repo-list/repo-list.component';

export const routes: Routes = [
  {
    path: '',
    pathMatch: 'full',
    redirectTo: 'repos',
  },
  {
    path: 'settings',
    component: SettingsComponent,
  },
  {
    path: 'repos',
    component: RepoListComponent,
  }
];
