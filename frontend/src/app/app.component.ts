import { Component } from '@angular/core';
import { RepoListComponent } from './repo-list/repo-list.component';

@Component({
  selector: 'app-root',
  imports: [
    RepoListComponent
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {
}
