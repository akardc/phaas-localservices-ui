import { Component } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { GetStatus, List, StartRepo, StopRepo } from '../../../wailsjs/go/repobrowser/RepoBrowser';
import { fromPromise } from 'rxjs/internal/observable/innerFrom';
import { AsyncPipe, DatePipe } from '@angular/common';
import { asyncScheduler, map, Observable, ReplaySubject } from 'rxjs';
import { MatButton } from '@angular/material/button';
import { repo } from '../../../wailsjs/go/models';
import { TypeSafeMatCellDef } from '../../lib/type-safe-mat-cell-def.directive';

export class Repo {
  basicDetails: repo.BasicDetails;

  private statusSub = new ReplaySubject<repo.Status>(1);

  constructor(basicDetails: repo.BasicDetails) {
    this.basicDetails = basicDetails;

    this.refreshStatus();
  }

  get status$(): Observable<repo.Status> {
    return this.statusSub.asObservable();
  }

  start() {
    StartRepo(this.basicDetails.name).then(
      () => console.log('Started repo', this.basicDetails.name),
      (err) => console.log('Failed to start repo', err),
    );
  }

  stop() {
    StopRepo(this.basicDetails.name).then(
      () => console.log('Stopped repo', this.basicDetails.name),
      (err) => console.log('Failed to stop repo', err),
    );
  }

  private refreshStatus() {
    GetStatus(this.basicDetails.name).then(
      (status) => {
        if (this.basicDetails.name === 'phaas-virtualevent-api') {
          console.log(this.basicDetails, status);
        }
        this.statusSub.next(status);
      }, (err) => console.log(`[repo:${this.basicDetails.name}:refreshStatus] failed to get status`, err)
    );
  }
}

@Component({
  selector: 'app-repo-list',
  imports: [
    MatTableModule,
    AsyncPipe,
    DatePipe,
    MatButton,
    TypeSafeMatCellDef,
  ],
  templateUrl: './repo-list.component.html',
  styleUrl: './repo-list.component.scss'
})
export class RepoListComponent {

  displayedColumns = ['running', 'name', 'lastModified', 'branch', 'button'];

  repos: Observable<Repo[]>

  constructor() {
    this.repos = fromPromise(List()).pipe(
      map((list) => list.map((repoDetails) => new Repo(repoDetails))),
    );
  }
}
