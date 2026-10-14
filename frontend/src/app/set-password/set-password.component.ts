import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { AuthGate } from '../auth-gate';
import { httpErrorMessage } from '../utils/http-error';

@Component({
  selector: 'app-set-password',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  template: `
    <div class="page">
      <h2>Задайте пароль</h2>
      <p *ngIf="checking">Проверка ссылки...</p>
      <p *ngIf="!checking && !valid" class="error">{{ error || 'Ссылка недействительна' }}</p>
      <form *ngIf="valid" (ngSubmit)="submit()">
        <p>Аккаунт: <strong>{{ name }}</strong></p>
        <p *ngIf="!needsLogin">Логин: <strong>{{ username }}</strong></p>
        <input *ngIf="needsLogin" [(ngModel)]="username" name="username" placeholder="Придумайте логин" required autocomplete="username" />
        <input type="password" [(ngModel)]="password" name="password" placeholder="Пароль (мин. 8)" required autocomplete="new-password" />
        <button type="submit" [disabled]="loading">{{ loading ? 'Сохранение...' : 'Сохранить и войти' }}</button>
        <p *ngIf="error" class="error">{{ error }}</p>
      </form>
    </div>
  `,
  styles: [`
    .page { max-width: 420px; margin: 48px auto; padding: 16px; color: var(--text); }
    form { display: flex; flex-direction: column; gap: 10px; }
    input, button { padding: 10px; font-size: 14px; }
    .error { color: var(--danger-text); }
  `]
})
export class SetPasswordComponent implements OnInit {
  token = '';
  checking = true;
  valid = false;
  loading = false;
  needsLogin = false;
  name = '';
  username = '';
  password = '';
  error: string | null = null;

  constructor(
    private route: ActivatedRoute,
    private router: Router,
    private api: CalendarService,
    gate: AuthGate
  ) {
    gate.end();
  }

  ngOnInit(): void {
    this.token = this.route.snapshot.queryParamMap.get('token') || '';
    if (!this.token) {
      this.checking = false;
      this.error = 'Нет ссылки';
      return;
    }
    this.api.validatePasswordLink(this.token).subscribe({
      next: (res) => {
        this.checking = false;
        this.valid = !!res?.ok;
        this.name = res?.name || '';
        this.username = res?.username || '';
        this.needsLogin = !!res?.needs_login;
        if (!this.valid) {
          this.error = res?.error || 'Ссылка недействительна';
        }
      },
      error: (err) => {
        this.checking = false;
        this.error = httpErrorMessage(err, 'Не удалось проверить ссылку');
      }
    });
  }

  submit(): void {
    this.loading = true;
    this.error = null;
    this.api.acceptPasswordLink({
      token: this.token,
      password: this.password,
      username: this.needsLogin ? this.username.trim() : undefined
    }).subscribe({
      next: () => this.router.navigateByUrl('/calendar'),
      error: (err) => {
        this.loading = false;
        this.error = httpErrorMessage(err, 'Не удалось сохранить пароль');
      }
    });
  }
}
