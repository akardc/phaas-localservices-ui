import { Component, ViewChild } from '@angular/core';
import { MatToolbar } from '@angular/material/toolbar';
import { MatIcon } from '@angular/material/icon';
import { MatIconButton } from '@angular/material/button';
import { MatSidenav, MatSidenavContainer, MatSidenavContent } from '@angular/material/sidenav';
import { NavigationEnd, Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { MAT_FORM_FIELD_DEFAULT_OPTIONS, MatFormFieldDefaultOptions } from '@angular/material/form-field';
import { MatListItem, MatNavList } from '@angular/material/list';
import { filter } from 'rxjs';
import { takeUntilDestroyed } from '@angular/core/rxjs-interop';

@Component({
  selector: 'app-root',
  imports: [
    MatToolbar,
    MatIcon,
    MatIconButton,
    MatSidenavContent,
    MatSidenavContainer,
    MatSidenav,
    RouterLink,
    RouterOutlet,
    MatNavList,
    MatListItem,
    RouterLinkActive
  ],
  providers: [
    {
      provide: MAT_FORM_FIELD_DEFAULT_OPTIONS,
      useValue: { appearance: 'outline' } satisfies MatFormFieldDefaultOptions
    }
  ],
  templateUrl: './app.component.html',
  styleUrl: './app.component.scss'
})
export class AppComponent {

  @ViewChild('sidenav') menuSidenav?: MatSidenav;

  constructor(
    private router: Router,
  ) {
    this.router.events.pipe(
      filter((ev): ev is NavigationEnd => ev instanceof NavigationEnd),
      takeUntilDestroyed(),
    ).subscribe(() => this.menuSidenav?.close());
  }
}
