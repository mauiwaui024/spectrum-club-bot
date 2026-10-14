import { Component } from '@angular/core';
import { Location } from '@angular/common';
import { Router } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { AuthGate } from '../auth-gate';

@Component({
  selector: 'app-privacy',
  standalone: true,
  imports: [],
  templateUrl: './privacy.component.html',
  styleUrl: './privacy.component.css'
})
export class PrivacyComponent {
  constructor(
    private location: Location,
    private router: Router,
    private api: CalendarService,
    gate: AuthGate
  ) {
    gate.end();
  }

  close(): void {
    const navId = history.state?.navigationId;
    if (typeof navId === 'number' && navId > 1) {
      this.location.back();
      return;
    }
    this.api.getAuthStatus().subscribe({
      next: (status) => {
        this.router.navigateByUrl(status?.authenticated ? '/calendar' : '/login');
      },
      error: () => this.router.navigateByUrl('/login')
    });
  }
}
