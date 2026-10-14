import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { CookieNoticeComponent } from '../cookie-notice/cookie-notice.component';
import { PD_POLICY_VERSION } from '../pd-policy';
import { AuthGate } from '../auth-gate';

@Component({
  selector: 'app-register',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule, CookieNoticeComponent],
  template: `
    <div class="page">
      <h2>Регистрация по приглашению</h2>
      <p *ngIf="checking">Проверка ссылки...</p>
      <p *ngIf="!checking && !valid" class="error">{{ error || 'Ссылка недействительна' }}</p>
      <div *ngIf="valid">
        <p>Роль: <strong>{{ roleLabel }}</strong></p>
        <form (ngSubmit)="submit()">
          <input [(ngModel)]="firstName" name="firstName" placeholder="Имя" required />
          <input [(ngModel)]="lastName" name="lastName" placeholder="Фамилия" />
          <input [(ngModel)]="username" name="username" placeholder="Логин" required autocomplete="username" />
          <input type="password" [(ngModel)]="password" name="password" placeholder="Пароль (мин. 8)" required autocomplete="new-password" />
          <label class="consent">
            <input type="checkbox" [(ngModel)]="pdConsent" name="pdConsent" />
            <span>Согласен на обработку персональных данных. <a routerLink="/privacy">Политика</a></span>
          </label>
          <button type="submit" [disabled]="loading || !pdConsent">{{ loading ? 'Создание...' : 'Зарегистрироваться' }}</button>
        </form>
        <p *ngIf="error" class="error">{{ error }}</p>
      </div>
      <a routerLink="/login">Уже есть аккаунт? Войти</a>
    </div>
    <app-cookie-notice></app-cookie-notice>
  `,
  styles: [`
    .page { max-width: 420px; margin: 48px auto; padding: 16px; color: var(--text); }
    form { display: flex; flex-direction: column; gap: 10px; margin: 16px 0; }
    input, button { padding: 10px; font-size: 14px; }
    .consent { display: flex; gap: 8px; align-items: flex-start; font-size: 14px; }
    .consent input { margin-top: 3px; }
    .error { color: #b91c1c; }
  `]
})
export class RegisterComponent implements OnInit {
  token = '';
  role = '';
  checking = true;
  valid = false;
  loading = false;
  error: string | null = null;
  username = '';
  password = '';
  firstName = '';
  lastName = '';
  pdConsent = false;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private calendarService: CalendarService,
    gate: AuthGate
  ) {
    gate.end();
  }

  get roleLabel(): string {
    return this.role === 'coach' ? 'Тренер' : 'Ученик';
  }

  ngOnInit(): void {
    this.token = this.route.snapshot.queryParamMap.get('token') || '';
    if (!this.token) {
      this.checking = false;
      this.valid = false;
      this.error = 'Нет токена приглашения';
      return;
    }
    this.calendarService.validateInvite(this.token).subscribe({
      next: (res) => {
        this.checking = false;
        this.valid = !!res?.ok;
        this.role = res?.role || '';
        if (!this.valid) {
          this.error = res?.error || 'Ссылка недействительна';
        }
      },
      error: () => {
        this.checking = false;
        this.valid = false;
        this.error = 'Не удалось проверить ссылку';
      }
    });
  }

  submit(): void {
    if (!this.pdConsent) {
      this.error = 'Нужно согласие на обработку персональных данных';
      return;
    }
    this.loading = true;
    this.error = null;
    this.calendarService.registerWithInvite({
      token: this.token,
      username: this.username,
      password: this.password,
      first_name: this.firstName,
      last_name: this.lastName,
      pd_consent: true,
      policy_version: PD_POLICY_VERSION
    }).subscribe({
      next: () => this.router.navigateByUrl('/calendar'),
      error: (err) => {
        this.loading = false;
        this.error = err?.error || 'Ошибка регистрации';
      }
    });
  }
}
