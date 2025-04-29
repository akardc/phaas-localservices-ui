import { repo } from '../../../wailsjs/go/models';
import { EventsOn } from '../../../wailsjs/runtime';
import {
  GetRepoStatus,
  RegisterRepoStatusWatcher,
  StartRepo,
  StopRepo
} from '../../../wailsjs/go/repobrowser/RepoBrowser';
import { signal } from '@angular/core';

export class RepoController {

  status = signal(new repo.Status());

  constructor(private basicDetails: repo.BasicDetails) {
    this.listenForStatusChanges();
    this.refreshStatus();
  }

  get name(): string {
    return this.basicDetails.name;
  }

  start() {
    StartRepo(this.name).then(
      () => {
        console.log(`[${this.name}] Starting`);
        this.refreshStatus();
      },
      (err) => {
        console.log(`[${this.name}] Failed to start`, err);
      }
    );
  }

  stop() {
    StopRepo(this.name).then(
      () => console.log(`[${this.name}] Stopped`),
      (err) => {
        console.log(`[${this.name}] Failed to stop`, err);
      }
    )
  }

  private refreshStatus() {
    GetRepoStatus(this.name).then(
      (status) => this.status.set(status),
      (err) => console.log(`[${this.name}] Failed to get repo status`, err),
    );
  }

  private listenForStatusChanges() {
    EventsOn(this.basicDetails.statusNotificationChannel, (status: repo.Status) => {
      console.log(`[${this.name}] Status notification received`, status);
      this.status.set(status);
    });
    RegisterRepoStatusWatcher(this.name).then(
      () => console.log(`[${this.name}] Registered status watcher`),
      (err) => console.log(`[${this.name}] Failed to register status watcher`, err),
    );
  }
}
