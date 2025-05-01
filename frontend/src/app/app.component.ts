import { Component } from '@angular/core';
import { MatIcon } from '@angular/material/icon';
import { MatIconAnchor } from '@angular/material/button';
import { RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { MAT_FORM_FIELD_DEFAULT_OPTIONS, MatFormFieldDefaultOptions } from '@angular/material/form-field';
import { MatNavList } from '@angular/material/list';
import { MatTooltip } from '@angular/material/tooltip';

@Component({
  selector: 'app-root',
  imports: [
    MatIcon,
    RouterLink,
    RouterOutlet,
    MatNavList,
    RouterLinkActive,
    MatIconAnchor,
    MatTooltip,
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
}
