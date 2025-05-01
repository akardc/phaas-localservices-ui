import { Injectable, signal } from '@angular/core';
import { InitRepos, ListRepos } from '../../../wailsjs/go/repobrowser/RepoBrowser';
import { RepoController } from './repo-controller';
import { from, retry, switchMap } from 'rxjs';
import { EventsOn } from '../../../wailsjs/runtime';

@Injectable({
  providedIn: 'root'
})
export class ControllersService {

  controllers = signal<RepoController[]>([]);

  private allRepos: RepoController[] = [];

  constructor() {
    this.rebuildList();
    EventsOn('repos-location-changed', () => this.rebuildList());
  }

  sortByName(direction?: 'asc' | 'desc' | '') {
    let asc = false;
    if (!direction || direction === 'asc') {
      asc = true;
    }
    this.allRepos.sort((a, b) => {
      if (asc) {
        return a.name.localeCompare(b.name);
      } else {
        return b.name.localeCompare(a.name);
      }
    });
    this.controllers.set(this.allRepos);
  }

  private rebuildList() {
    from(InitRepos()).pipe(
      switchMap(() => from(ListRepos()).pipe(
        retry(2),
      )),
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
}
