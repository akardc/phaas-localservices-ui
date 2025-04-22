import { Component } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { ListRepos, StartRepo, StopRepo } from '../../../wailsjs/go/repobrowser/RepoBrowser';
import { fromPromise } from 'rxjs/internal/observable/innerFrom';
import { DatePipe } from '@angular/common';
import { Observable } from 'rxjs';
import { repobrowser } from '../../../wailsjs/go/models';
import { MatButton } from '@angular/material/button';

@Component({
  selector: 'app-repo-list',
  imports: [
    MatTableModule,
    DatePipe,
    MatButton,
  ],
  templateUrl: './repo-list.component.html',
  styleUrl: './repo-list.component.scss'
})
export class RepoListComponent {

  displayedColumns = ['name', 'lastModified', 'branch', 'start', 'stop'];

  repos: Observable<repobrowser.RepoInfo[]>
  // statuses = new Map<string, repobrowser.RepoStatus>();

  constructor() {
    this.repos = fromPromise(ListRepos(new repobrowser.ListReposOptions({ nameRegex: 'phaas-.*-((ui)|(api))' }))).pipe(
      // tap((repos) => this.getRepoStatuses(repos)),
    );
  }

  startRepo(repo: repobrowser.RepoInfo) {
    StartRepo(repo.name);
  }

  stopRepo(repo: repobrowser.RepoInfo) {
    StopRepo(repo.name);
  }

  // private getRepoStatuses(repos: RepoInfo[]) {
  //   repos.forEach((r) => {
  //     RepoStatus(r.path).then((s) => this.statuses.set(r.name, s))
  //   });
  // }
}
