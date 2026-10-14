import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface StudentRow {
  user_id: number;
  student_id: number;
  first_name: string;
  last_name: string;
  username: string;
  display_name: string;
  is_active: boolean;
  role: string;
}

@Component({
  selector: 'app-students',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div class="page">
      <h1>Ученики</h1>
      <p class="hint">Нажмите на ученика, чтобы открыть карточку.</p>
      <div class="filters">
        <select [(ngModel)]="filter" (change)="load()">
          <option value="all">Все</option>
          <option value="no_subscription">Без абонемента</option>
          <option value="expiring">Истекает</option>
        </select>
        <input type="number" [(ngModel)]="days" min="1" max="90" (change)="load()" title="Дней до истечения" />
        <input type="search" [(ngModel)]="q" (keyup.enter)="load()" placeholder="Поиск" />
        <button type="button" (click)="load()">Найти</button>
      </div>
      <p *ngIf="error" class="error">{{ error }}</p>
      <p *ngIf="message" class="ok">{{ message }}</p>
      <div *ngIf="loading">Загрузка...</div>

      <div class="layout" *ngIf="!loading">
        <ul class="list">
          <li
            *ngFor="let s of students"
            class="row"
            [class.selected]="selected?.user_id === s.user_id"
            [class.inactive]="!s.is_active"
            (click)="openStudent(s)">
            <strong>{{ s.display_name }}</strong>
            <span class="meta">{{ s.username || '—' }}</span>
          </li>
          <li *ngIf="!students.length" class="empty">Никого не найдено</li>
        </ul>

        <aside class="detail" *ngIf="selected">
          <header class="detail-header">
            <h2>{{ selected.display_name }}</h2>
            <button type="button" class="btn-ghost" (click)="closeStudent()">Закрыть</button>
          </header>

          <dl class="info">
            <div><dt>Логин</dt><dd>{{ selected.username || '—' }}</dd></div>
            <div><dt>Имя</dt><dd>{{ selected.first_name || '—' }}</dd></div>
            <div><dt>Фамилия</dt><dd>{{ selected.last_name || '—' }}</dd></div>
            <div><dt>ID ученика</dt><dd>{{ selected.student_id || '—' }}</dd></div>
            <div><dt>Роль</dt><dd>{{ roleLabel(selected.role) }}</dd></div>
            <div><dt>Статус</dt><dd>{{ selected.is_active ? 'активен' : 'деактивирован' }}</dd></div>
          </dl>

          <div class="safe-actions">
            <button type="button" (click)="toggleActive(selected)">
              {{ selected.is_active ? 'Деактивировать вход' : 'Активировать вход' }}
            </button>
          </div>

          <details class="danger-zone" (toggle)="onDangerToggle($event)">
            <summary>Опасная зона · доступ и роли</summary>
            <p class="warn">
              Смена роли меняет права во всём приложении. Ученик, ставший тренером,
              пропадает из этого списка. Не используйте без необходимости.
            </p>

            <div *ngIf="!roleMenuOpen" class="nested">
              <button type="button" class="btn-muted" (click)="unlockRoleMenu()">
                Показать смену роли…
              </button>
            </div>

            <div *ngIf="roleMenuOpen" class="nested role-menu">
              <p class="warn tight">Два подтверждения обязательны.</p>
              <button type="button" class="btn-danger" (click)="promoteToCoach(selected)">
                Сделать тренером
              </button>
              <button type="button" class="btn-muted" (click)="demoteToStudent(selected)" *ngIf="selected.role === 'coach'">
                Вернуть в ученики
              </button>
              <button type="button" class="btn-ghost" (click)="roleMenuOpen = false">Скрыть</button>
            </div>
          </details>
        </aside>

        <aside class="detail placeholder" *ngIf="!selected && students.length">
          <p class="muted">Выберите ученика слева</p>
        </aside>
      </div>
    </div>
  `,
  styles: [`
    .page { max-width: 960px; margin: 0 auto; padding: 16px; color: var(--text); }
    .hint, .muted { color: var(--muted); font-size: 14px; margin: 0 0 8px; }
    .filters { display: flex; flex-wrap: wrap; gap: 8px; margin: 12px 0; }
    .filters input, .filters select, button { padding: 8px; font-size: 14px; }
    .layout { display: grid; grid-template-columns: 1fr; gap: 16px; }
    @media (min-width: 720px) {
      .layout { grid-template-columns: 1fr 1fr; align-items: start; }
    }
    .list { list-style: none; padding: 0; margin: 0; border: 1px solid var(--border); border-radius: 8px; overflow: hidden; background: var(--card); }
    .row { padding: 12px 14px; border-bottom: 1px solid var(--border); cursor: pointer; }
    .row:last-child { border-bottom: none; }
    .row:hover { background: var(--hover); }
    .row.selected { background: var(--primary-soft); }
    .row.inactive { opacity: 0.55; }
    .meta { display: block; color: var(--muted); font-size: 13px; margin-top: 2px; }
    .empty { padding: 16px; color: var(--muted); }
    .detail { border: 1px solid var(--border); border-radius: 8px; padding: 16px; background: var(--card); }
    .detail.placeholder { display: flex; align-items: center; justify-content: center; min-height: 160px; }
    .detail-header { display: flex; justify-content: space-between; gap: 12px; align-items: flex-start; margin-bottom: 12px; }
    .detail-header h2 { margin: 0; font-size: 1.15rem; }
    .info { margin: 0 0 16px; }
    .info > div { display: grid; grid-template-columns: 120px 1fr; gap: 8px; padding: 6px 0; border-bottom: 1px solid var(--border); font-size: 14px; }
    .info dt { color: var(--muted); margin: 0; }
    .info dd { margin: 0; }
    .safe-actions { margin-bottom: 16px; }
    .danger-zone { border: 1px solid var(--danger-border); border-radius: 8px; padding: 10px 12px; background: var(--danger-bg); }
    .danger-zone summary { cursor: pointer; font-weight: 600; color: var(--danger-text); }
    .warn { color: var(--danger-text); font-size: 13px; line-height: 1.4; margin: 10px 0; }
    .warn.tight { margin-top: 0; }
    .nested { display: flex; flex-direction: column; gap: 8px; margin-top: 8px; }
    .btn-ghost { background: transparent; border: 1px solid var(--border); }
    .btn-muted { background: var(--hover); border: 1px solid var(--border); }
    .btn-danger { background: var(--danger-bg); border: 1px solid var(--danger-border); color: var(--danger-text); }
    .error { color: var(--danger-text); } .ok { color: var(--ok); }
  `]
})
export class StudentsComponent implements OnInit {
  students: StudentRow[] = [];
  selected: StudentRow | null = null;
  filter = 'all';
  days = 14;
  q = '';
  loading = false;
  error: string | null = null;
  message: string | null = null;
  roleMenuOpen = false;

  constructor(private api: CalendarService) {}

  ngOnInit(): void {
    this.load();
  }

  roleLabel(role: string): string {
    if (role === 'coach') return 'тренер';
    if (role === 'student') return 'ученик';
    return role || '—';
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getStudents(this.filter, this.days, this.q).subscribe({
      next: (res) => {
        this.students = res.students || [];
        if (this.selected) {
          const still = this.students.find((s) => s.user_id === this.selected!.user_id);
          this.selected = still || null;
          if (!still) this.roleMenuOpen = false;
        }
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Ошибка загрузки');
        this.loading = false;
      }
    });
  }

  openStudent(s: StudentRow): void {
    this.selected = s;
    this.roleMenuOpen = false;
    this.message = null;
    this.error = null;
  }

  closeStudent(): void {
    this.selected = null;
    this.roleMenuOpen = false;
  }

  onDangerToggle(event: Event): void {
    const el = event.target as HTMLDetailsElement;
    if (!el.open) {
      this.roleMenuOpen = false;
    }
  }

  unlockRoleMenu(): void {
    const ok = confirm(
      'Смена роли — опасное действие.\n\nОткрыть меню смены роли? (это ещё не меняет роль)'
    );
    if (ok) {
      this.roleMenuOpen = true;
    }
  }

  promoteToCoach(s: StudentRow): void {
    const first = confirm(
      `ВНИМАНИЕ\n\nСделать «${s.display_name}» тренером?\n\n` +
        'У него появятся права тренера (расписание, шаблоны, ученики).\n' +
        'Он пропадёт из списка учеников.\n\nПродолжить?'
    );
    if (!first) return;

    const second = confirm(
      `Последнее подтверждение\n\nТочно назначить тренером:\n${s.display_name} (@${s.username || '—'})\n\nОтменить будет сложнее.`
    );
    if (!second) return;

    this.patchRole(s, 'coach');
  }

  demoteToStudent(s: StudentRow): void {
    const first = confirm(`Вернуть «${s.display_name}» в роль ученика?`);
    if (!first) return;
    const second = confirm(`Подтвердите ещё раз: роль ученик для ${s.display_name}.`);
    if (!second) return;
    this.patchRole(s, 'student');
  }

  private patchRole(s: StudentRow, role: string): void {
    this.message = null;
    this.api.patchUser(s.user_id, { role }).subscribe({
      next: () => {
        this.message = role === 'coach'
          ? 'Роль изменена на тренера. Ученик больше не в этом списке.'
          : 'Роль обновлена';
        this.roleMenuOpen = false;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка смены роли'))
    });
  }

  toggleActive(s: StudentRow): void {
    const next = !s.is_active;
    const ok = confirm(
      next
        ? `Активировать вход для «${s.display_name}»?`
        : `Деактивировать вход для «${s.display_name}»? Логин будет запрещён.`
    );
    if (!ok) return;

    this.message = null;
    this.api.patchUser(s.user_id, { is_active: next }).subscribe({
      next: () => {
        this.message = 'Статус обновлён';
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка'))
    });
  }
}
