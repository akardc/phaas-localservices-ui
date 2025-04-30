import { Component, OnInit } from '@angular/core';
import { GetSettings, SaveSettings } from '../../../wailsjs/go/app/Settings';
import { FormArray, FormControl, FormGroup, ReactiveFormsModule, Validators } from '@angular/forms';
import { MatCheckbox } from '@angular/material/checkbox';
import { MatFormField, MatHint, MatInput, MatLabel } from '@angular/material/input';
import { MatIcon } from '@angular/material/icon';
import { MatButton, MatIconButton } from '@angular/material/button';
import { app } from '../../../wailsjs/go/models';

@Component({
  selector: 'app-settings',
  imports: [
    ReactiveFormsModule,
    MatCheckbox,
    MatFormField,
    MatLabel,
    MatInput,
    MatFormField,
    MatIconButton,
    MatIcon,
    MatButton,
    MatHint,
  ],
  templateUrl: './settings.component.html',
  styleUrl: './settings.component.scss'
})
export class SettingsComponent implements OnInit {

  form = new FormGroup({
    dataDirPath: new FormControl('', [Validators.required]),
    reposDirPath: new FormControl('', [Validators.required]),
    envParams: new FormArray<FormGroup<{
      key: FormControl<string | null>,
      value: FormControl<string | null>,
      enabled: FormControl<boolean | null>,
    }>>([]),
  });

  ngOnInit(): void {
    GetSettings().then(
      (settings) => {
        console.log('[Settings] Loaded settings', settings)
        this.form.controls.dataDirPath.setValue(settings?.dataDirPath || '');
        this.form.controls.reposDirPath.setValue(settings?.reposDirPath || '');
        if (settings?.envParams) {
          settings.envParams.forEach((param) => {
            this.form.controls.envParams.push(new FormGroup({
              key: new FormControl(param.key),
              value: new FormControl(param.value),
              enabled: new FormControl(param.enabled),
            }));
          });
        }
      },
      (err) => console.log('[Settings] Failed to load settings', err),
    )
  }

  addParam() {
    this.form.controls.envParams.push(new FormGroup({
      key: new FormControl(''),
      value: new FormControl(''),
      enabled: new FormControl(true),
    }));
  }

  save() {
    SaveSettings(new app.Settings(this.form.value)).then(
      () => console.log(`[Settings] Saved settings`),
      (err) => console.log(`[Settings] Failed to save settings`, err),
    );
  }
}
