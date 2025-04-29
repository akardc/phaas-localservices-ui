import { Injectable, signal } from '@angular/core';
import { ListRepos } from '../../../wailsjs/go/repobrowser/RepoBrowser';
import { RepoController } from './repo-controller';
import { from, retry } from 'rxjs';

@Injectable({
  providedIn: 'root'
})
export class ControllersService {

  controllers = signal<RepoController[]>([]);

  private allRepos: RepoController[] = [];

  constructor() {
    from(ListRepos()).pipe(
      retry(2),
    ).subscribe({
      next: (list) => {
        list.forEach((repoDetails) => this.allRepos.push(new RepoController(repoDetails)));
        this.sortByName();
      },
      error: (err) => {
        console.log('Failed to list repos', err);
      }
    });
  }

  sortByName() {
    this.allRepos.sort((a, b) => a.name.localeCompare(b.name));
    this.controllers.set(this.allRepos);
  }
}
