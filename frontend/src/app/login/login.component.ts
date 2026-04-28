import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router } from '@angular/router';
import { CalendarService } from '../services/calendar.service';

@Component({
  selector: 'app-login',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="login-page">
      <h2>Вход в календарь</h2>
      <p>Если вы открыли страницу вне Telegram, войдите через логин и пароль.</p>
      <form (ngSubmit)="login()">
        <input type="text" [(ngModel)]="username" name="username" placeholder="Telegram username" required />
        <input type="password" [(ngModel)]="password" name="password" placeholder="Пароль" required />
        <button type="submit" [disabled]="loading">{{ loading ? 'Вход...' : 'Войти' }}</button>
      </form>
      <p *ngIf="error" class="error">{{ error }}</p>
    </div>
  `,
  styles: [`
    .login-page { max-width: 420px; margin: 48px auto; padding: 16px; }
    form { display: flex; flex-direction: column; gap: 10px; }
    input, button { padding: 10px; font-size: 14px; }
    .error { color: #b91c1c; margin-top: 10px; }
  `]
})
export class LoginComponent {
  username = '';
  password = '';
  loading = false;
  error: string | null = null;

  constructor(
    private calendarService: CalendarService,
    private router: Router,
    private route: ActivatedRoute
  ) {}

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
