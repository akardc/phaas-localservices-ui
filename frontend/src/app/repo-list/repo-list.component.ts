import { Component, computed, signal, Signal } from '@angular/core';
import { MatTableModule } from '@angular/material/table';
import { MatButton, MatIconButton } from '@angular/material/button';
import { repo } from '../../../wailsjs/go/models';
import { TypeSafeMatCellDef } from '../../lib/type-safe-mat-cell-def.directive';
import { MatMenu, MatMenuTrigger } from '@angular/material/menu';
import { MatIcon } from '@angular/material/icon';
import { ControllersService } from '../repo-controller/controllers.service';
import { RepoController } from '../repo-controller/repo-controller';
import { FormControl, ReactiveFormsModule } from '@angular/forms';
import { MatFormField, MatInput, MatLabel, MatSuffix } from '@angular/material/input';
import { MAT_FORM_FIELD_DEFAULT_OPTIONS, MatFormFieldDefaultOptions } from '@angular/material/form-field';
import { debounceTime } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';
import State = repo.State;
import { NgClass } from '@angular/common';

@Component({
  selector: 'app-repo-list',
  imports: [
    MatTableModule,
    MatButton,
    TypeSafeMatCellDef,
    MatIconButton,
    MatIcon,
    MatMenu,
    MatMenuTrigger,
    MatFormField,
    MatInput,
    MatLabel,
    ReactiveFormsModule,
    MatSuffix,
    NgClass,
  ],
  templateUrl: './repo-list.component.html',
  styleUrl: './repo-list.component.scss'
})
export class RepoListComponent {

  readonly State = State;

  displayedColumns = ['running', 'name', 'lastModified', 'button', 'menu'];

  repos: Signal<RepoController[]>;

  nameFilter = new FormControl('');
  private filterSig = signal('');


  constructor(
    private controllers: ControllersService,
  ) {
    this.repos = computed(() => {
      const filter = this.filterSig();
      let repos = this.controllers.controllers();
      if (filter) {
        repos = repos.filter((r) => r.name.toLowerCase().includes(filter.toLowerCase()));
      }
      return repos;
    });
    this.nameFilter.valueChanges.pipe(
      takeUntilDestroyed(),
      debounceTime(200),
    ).subscribe((filter) => this.filterSig.set(filter || ''));
  }
}
