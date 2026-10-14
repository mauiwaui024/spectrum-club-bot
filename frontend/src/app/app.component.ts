import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterOutlet } from '@angular/router';
import { AuthGate } from './auth-gate';
import { ClimberLoaderComponent } from './climber-loader/climber-loader.component';

@Component({
  selector: 'app-root',
  standalone: true,
  imports: [CommonModule, RouterOutlet, ClimberLoaderComponent],
  template: `
    <app-climber-loader *ngIf="gate.pending()"></app-climber-loader>
    <router-outlet></router-outlet>
  `
})
export class AppComponent {
  title = 'Spectrum Club Calendar';

  constructor(readonly gate: AuthGate) {}
}
