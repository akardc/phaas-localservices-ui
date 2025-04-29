import { Component, computed, OnInit, Signal } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { GetRepoStatus, StartRepo, StopRepo } from '../../../wailsjs/go/repobrowser/RepoBrowser';
import { AsyncPipe, DatePipe } from '@angular/common';
import { Observable, ReplaySubject } from 'rxjs';
import { MatButton, MatIconButton } from '@angular/material/button';
import { repo } from '../../../wailsjs/go/models';
import { TypeSafeMatCellDef } from '../../lib/type-safe-mat-cell-def.directive';
import { MatMenu, MatMenuItem, MatMenuTrigger } from '@angular/material/menu';
import { MatIcon } from '@angular/material/icon';
import { MatTooltip } from '@angular/material/tooltip';
import { ControllersService } from '../repo-controller/controllers.service';
import { RepoController } from '../repo-controller/repo-controller';
import State = repo.State;

export class Repo {
  basicDetails: repo.BasicDetails;
  status = {
    status: '',
    description: '',
  };


  private statusSub = new ReplaySubject<repo.Status>(1);

  constructor(basicDetails: repo.BasicDetails) {
    this.basicDetails = basicDetails;

    this.refreshBranchInfo();
  }

  get branchInfo$(): Observable<repo.Status> {
    return this.statusSub.asObservable();
  }

  start() {
    this.status = {
      status: 'Starting',
      description: '',
    };
    StartRepo(this.basicDetails.name).then(
      () => {
        console.log('Started repo', this.basicDetails.name);
        this.refreshBranchInfo();
        this.status = {
          status: 'Running',
          description: '',
        };
      },
      (err) => {
        console.log('Failed to start repo', err);
        this.status = {
          status: 'Failed',
          description: err,
        };
      },
    );
  }

  stop() {
    StopRepo(this.basicDetails.name).then(
      () => console.log('Stopped repo', this.basicDetails.name),
      (err) => console.log('Failed to stop repo', err),
    );
  }

  watch() {
    // RegisterRunningRepoStatusWatcher(this.basicDetails.name).then(
    //   (channel) => EventsOn(channel, (status: any) => this.status = {
    //     status: status?.status || '',
    //     description: status?.description || '',
    //   }),
    // );
  }

  private refreshBranchInfo() {
    GetRepoStatus(this.basicDetails.name).then(
      (status) => {
        console.log(this.basicDetails, status);
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
    MatIconButton,
    MatIcon,
    MatMenu,
    MatMenuItem,
    MatMenuTrigger,
    MatTooltip,
  ],
  templateUrl: './repo-list.component.html',
  styleUrl: './repo-list.component.scss'
})
export class RepoListComponent implements OnInit {

  readonly State = State;

  displayedColumns = ['running', 'name', 'lastModified', 'branch', 'runningStatus', 'button', 'menu'];

  repos: Signal<RepoController[]>;

  constructor(
    private controllers: ControllersService,
  ) {
    this.repos = computed(() => this.controllers.controllers());
  }

  ngOnInit() {
  }
}
