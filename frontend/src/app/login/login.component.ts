import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { applyTheme, AppTheme, readTheme } from '../theme';
import { CookieNoticeComponent } from '../cookie-notice/cookie-notice.component';
import { AuthGate } from '../auth-gate';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule, CookieNoticeComponent],
  template: `
    <div class="login-page">
      <div class="login-top">
        <h2>Вход в Spectrum Club</h2>
        <button type="button" class="btn-ghost theme-icon" (click)="toggleTheme()" [attr.aria-label]="theme === 'dark' ? 'Светлая тема' : 'Тёмная тема'">{{ theme === 'dark' ? '☀️' : '🌙' }}</button>
      </div>
      <p>Войдите логином и паролем. Регистрация только по приглашению тренера.</p>
      <form (ngSubmit)="login()">
        <input type="text" [(ngModel)]="username" name="username" placeholder="Логин" required autocomplete="username" />
        <input type="password" [(ngModel)]="password" name="password" placeholder="Пароль" required autocomplete="current-password" />
        <button type="submit" class="btn-primary" [disabled]="loading">{{ loading ? 'Вход...' : 'Войти' }}</button>
      </form>
      <p *ngIf="error" class="error">{{ error }}</p>
    </div>
    <app-cookie-notice></app-cookie-notice>
  `,
  styles: [`
    .login-page { max-width: 420px; margin: 48px auto; padding: 16px; color: var(--text); }
    .login-top { display: flex; justify-content: space-between; align-items: center; gap: 12px; margin-bottom: 8px; }
    .login-top h2 { margin: 0; }
    form { display: flex; flex-direction: column; gap: 10px; }
    input, button { padding: 10px; font-size: 14px; }
    .error { color: var(--danger-text); margin-top: 10px; }
  `]
})
export class LoginComponent implements OnInit {
  username = '';
  password = '';
  loading = false;
  error: string | null = null;
  theme: AppTheme = readTheme();

  constructor(
    private calendarService: CalendarService,
    private router: Router,
    private route: ActivatedRoute,
    private gate: AuthGate
  ) {}

  ngOnInit(): void {
    this.gate.begin();
    this.calendarService.getAuthStatus().subscribe({
      next: (status) => {
        if (status?.authenticated) {
          this.router.navigateByUrl('/calendar');
          return;
        }
        this.gate.end();
      },
      error: () => this.gate.end()
    });
  }

  toggleTheme(): void {
    this.theme = this.theme === 'dark' ? 'light' : 'dark';
    applyTheme(this.theme);
  }

  login(): void {
    this.loading = true;
    this.error = null;
    this.calendarService.loginWithBrowserCredentials(this.username, this.password).subscribe({
      next: () => {
        const redirect = this.route.snapshot.queryParamMap.get('redirect') || '/calendar';
        this.router.navigateByUrl(redirect);
      },
      error: () => {
        this.loading = false;
        this.error = 'Неверный логин или пароль';
      }
    });
  }
}
