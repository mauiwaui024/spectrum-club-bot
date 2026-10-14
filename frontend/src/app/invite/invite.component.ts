import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

@Component({
  selector: 'app-invite',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="page">
      <div class="header">
        <h1>Пригласить</h1>
      </div>
      <p>Закрытая регистрация: сгенерируйте ссылку и отправьте человеку вне системы.</p>
      <label>
        Роль
        <select [(ngModel)]="role">
          <option value="student">Ученик</option>
          <option value="coach">Тренер</option>
        </select>
      </label>
      <label>
        Срок (часы)
        <input type="number" [(ngModel)]="ttlHours" min="1" max="336" />
      </label>
      <button (click)="create()" [disabled]="loading">{{ loading ? '...' : 'Создать ссылку' }}</button>
      <p *ngIf="error" class="error">{{ error }}</p>
      <div *ngIf="url" class="result">
        <p>Ссылка (до {{ expiresAt }}):</p>
        <code>{{ url }}</code>
        <button (click)="copy()">Копировать</button>
        <p *ngIf="copied" class="ok">Скопировано</p>
      </div>

      <h2>Уже в клубе, без пароля</h2>
      <p>Этим людям можно отправить ссылку, чтобы они задали пароль к своему аккаунту.</p>
      <p *ngIf="listError" class="error">{{ listError }}</p>
      <p *ngIf="!withoutPassword.length" class="muted">Таких людей нет</p>
      <ul class="people">
        <li *ngFor="let person of withoutPassword">
          <div class="person">
            <strong>{{ person.name }}</strong>
            <span class="muted" *ngIf="person.username">{{ person.username }}</span>
            <span class="muted" *ngIf="!person.username">логина ещё нет</span>
          </div>
          <button type="button" (click)="createPasswordLink(person)" [disabled]="linkLoadingId === person.id">
            Ссылка, чтобы задать пароль
          </button>
          <div *ngIf="passwordLinks[person.id]" class="result">
            <p>Ссылка (до {{ passwordLinks[person.id].expiresAt }}):</p>
            <code>{{ passwordLinks[person.id].url }}</code>
            <button type="button" (click)="copyLink(person.id)">Копировать</button>
            <p *ngIf="copiedId === person.id" class="ok">Скопировано</p>
          </div>
        </li>
      </ul>
    </div>
  `,
  styles: [`
    .page { max-width: 640px; margin: 0 auto; padding: 16px; }
    .header { display: flex; justify-content: space-between; align-items: center; }
    label { display: flex; flex-direction: column; gap: 4px; margin: 12px 0; }
    input, select, button { padding: 8px 10px; }
    code { display: block; word-break: break-all; background: var(--card); color: var(--text); border: 1px solid var(--border); padding: 10px; margin: 8px 0; }
    .error { color: var(--danger-text); }
    .ok { color: var(--ok); }
    .muted { color: var(--muted); }
    h2 { margin: 28px 0 8px; font-size: 18px; }
    .people { list-style: none; padding: 0; margin: 0; }
    .people li { border-top: 1px solid var(--border); padding: 12px 0; }
    .person { display: flex; flex-direction: column; gap: 2px; margin-bottom: 8px; }
  `]
})
export class InviteComponent implements OnInit {
  role: 'student' | 'coach' = 'student';
  ttlHours = 72;
  loading = false;
  error: string | null = null;
  url = '';
  expiresAt = '';
  copied = false;
  withoutPassword: { id: number; name: string; username: string }[] = [];
  listError: string | null = null;
  linkLoadingId: number | null = null;
  passwordLinks: Record<number, { url: string; expiresAt: string }> = {};
  copiedId: number | null = null;

  constructor(private calendarService: CalendarService) {}

  ngOnInit(): void {
    this.calendarService.getUsersWithoutPassword().subscribe({
      next: (res) => { this.withoutPassword = res.users || []; },
      error: (err) => { this.listError = httpErrorMessage(err, 'Не удалось загрузить список'); }
    });
  }

  create(): void {
    this.loading = true;
    this.error = null;
    this.copied = false;
    this.calendarService.createInvite(this.role, this.ttlHours).subscribe({
      next: (res) => {
        this.loading = false;
        this.url = res.url;
        this.expiresAt = res.expires_at;
      },
      error: (err) => {
        this.loading = false;
        this.error = err?.error || 'Не удалось создать приглашение';
      }
    });
  }

  copy(): void {
    navigator.clipboard.writeText(this.url).then(() => {
      this.copied = true;
    });
  }

  createPasswordLink(person: { id: number }): void {
    this.linkLoadingId = person.id;
    this.listError = null;
    this.copiedId = null;
    this.calendarService.createPasswordLink(person.id, this.ttlHours).subscribe({
      next: (res) => {
        this.linkLoadingId = null;
        this.passwordLinks[person.id] = { url: res.url, expiresAt: res.expires_at };
      },
      error: (err) => {
        this.linkLoadingId = null;
        this.listError = httpErrorMessage(err, 'Не удалось создать ссылку');
      }
    });
  }

  copyLink(id: number): void {
    const link = this.passwordLinks[id];
    if (!link) return;
    navigator.clipboard.writeText(link.url).then(() => {
      this.copiedId = id;
    });
  }
}
