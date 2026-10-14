import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface Summary {
  trainings_count: number;
  bookings_count: number;
  attended_count: number;
  unmarked_count: number;
  active_subscriptions: number;
  remaining_lessons_sum: number;
  expiring_subscriptions_14d: number;
}

interface InactiveStudent {
  display_name: string;
  group_name: string;
}

@Component({
  selector: 'app-analytics',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="page">
      <h1>Аналитика</h1>
      <p *ngIf="error" class="error">{{ error }}</p>

      <section class="card">
        <div class="period">
          <label>
            Месяц
            <select [(ngModel)]="month" (ngModelChange)="loadSummary()">
              <option *ngFor="let name of monthNames; let i = index" [ngValue]="i + 1">{{ name }}</option>
            </select>
          </label>
          <label>
            Год
            <input type="number" [(ngModel)]="year" (change)="loadSummary()" />
          </label>
        </div>

        <h2>Занятия за {{ monthTitle }}</h2>
        <table *ngIf="summary">
          <tbody>
            <tr><th>Тренировок</th><td>{{ summary.trainings_count }}</td></tr>
            <tr><th>Записались</th><td>{{ summary.bookings_count }}</td></tr>
            <tr><th>Пришли</th><td>{{ summary.attended_count }}</td></tr>
            <tr><th>Не отмечены</th><td>{{ summary.unmarked_count }}</td></tr>
          </tbody>
        </table>
        <p class="hint">«Не отмечены» — прошедшие тренировки, где явку не поставили.</p>

        <h2>Абонементы на сегодня</h2>
        <table *ngIf="summary">
          <tbody>
            <tr><th>Действующих</th><td>{{ summary.active_subscriptions }}</td></tr>
            <tr><th>Занятий осталось, всего</th><td>{{ summary.remaining_lessons_sum }}</td></tr>
            <tr><th>Закончатся в ближайшие 14 дней</th><td>{{ summary.expiring_subscriptions_14d }}</td></tr>
          </tbody>
        </table>

        <a class="btn" [href]="csvHref()">Скачать посещаемость за {{ monthTitle }}</a>
      </section>

      <section class="card">
        <h2>Давно не были</h2>
        <label>
          Не были дольше
          <select [(ngModel)]="weeks" (ngModelChange)="loadInactive()">
            <option [ngValue]="2">2 недель</option>
            <option [ngValue]="4">4 недель</option>
            <option [ngValue]="8">8 недель</option>
          </select>
        </label>
        <table *ngIf="inactive.length">
          <thead>
            <tr><th>Ученик</th><th>Группа</th></tr>
          </thead>
          <tbody>
            <tr *ngFor="let s of inactive">
              <td>{{ s.display_name }}</td>
              <td>{{ s.group_name || '—' }}</td>
            </tr>
          </tbody>
        </table>
        <p *ngIf="!inactive.length" class="hint">Таких учеников нет</p>
      </section>
    </div>
  `,
  styles: [`
    .page { max-width: 720px; margin: 0 auto; padding: 16px; color: var(--text); }
    h1 { margin: 0 0 12px; font-size: 22px; }
    h2 { margin: 16px 0 8px; font-size: 16px; }
    .card { border: 1px solid var(--border); border-radius: 8px; padding: 16px; margin: 12px 0; background: var(--card); }
    .period { display: flex; gap: 12px; flex-wrap: wrap; }
    label { display: flex; flex-direction: column; gap: 4px; font-size: 14px; }
    input, select, .btn { padding: 8px 10px; font-size: 14px; }
    table { width: 100%; border-collapse: collapse; margin-bottom: 8px; }
    th, td { text-align: left; padding: 8px 10px; border-bottom: 1px solid var(--border); font-weight: 400; }
    th { color: var(--muted); width: 70%; }
    thead th { color: var(--text); font-weight: 600; }
    .hint { color: var(--muted); font-size: 13px; margin: 0 0 8px; }
    .btn { display: inline-block; margin-top: 12px; text-decoration: none; background: var(--card); color: var(--primary); border: 1px solid var(--border); border-radius: 8px; }
    .error { color: var(--danger-text); }
  `]
})
export class AnalyticsComponent implements OnInit {
  readonly monthNames = [
    'январь', 'февраль', 'март', 'апрель', 'май', 'июнь',
    'июль', 'август', 'сентябрь', 'октябрь', 'ноябрь', 'декабрь'
  ];

  summary: Summary | null = null;
  inactive: InactiveStudent[] = [];
  weeks = 4;
  year = new Date().getFullYear();
  month = new Date().getMonth() + 1;
  error: string | null = null;

  constructor(private api: CalendarService) {}

  ngOnInit(): void {
    this.loadSummary();
    this.loadInactive();
  }

  get monthTitle(): string {
    const name = this.monthNames[this.month - 1] || '';
    return `${name} ${this.year}`;
  }

  loadSummary(): void {
    const from = `${this.year}-${String(this.month).padStart(2, '0')}-01`;
    const last = new Date(this.year, this.month, 0).getDate();
    const to = `${this.year}-${String(this.month).padStart(2, '0')}-${String(last).padStart(2, '0')}`;
    this.api.getAnalyticsSummary(from, to).subscribe({
      next: (s) => {
        this.summary = s;
        this.error = null;
      },
      error: (e) => (this.error = httpErrorMessage(e, 'Не удалось загрузить сводку'))
    });
  }

  loadInactive(): void {
    this.api.getInactiveStudents(this.weeks).subscribe({
      next: (res) => {
        this.inactive = res.students || [];
      },
      error: (e) => (this.error = httpErrorMessage(e, 'Не удалось загрузить список'))
    });
  }

  csvHref(): string {
    return this.api.attendanceCsvUrl(this.year, this.month);
  }
}
