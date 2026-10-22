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
      <p class="hint">Нажмите на ученика, чтобы открыть карточку под ним.</p>
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

      <ul class="list" *ngIf="!loading">
        <li *ngFor="let s of students" class="person" [class.inactive]="!s.is_active" [class.open]="isOpen(s)">
          <button type="button" class="row" (click)="toggleStudent(s)">
            <span>
              <strong>{{ s.display_name }}</strong>
              <span class="meta">{{ s.username || '—' }}</span>
            </span>
            <span class="chevron">{{ isOpen(s) ? '▼' : '▶' }}</span>
          </button>

          <div class="detail" *ngIf="isOpen(s)">
            <dl class="info">
              <div><dt>Логин</dt><dd>{{ s.username || '—' }}</dd></div>
              <div><dt>ID ученика</dt><dd>{{ s.student_id || '—' }}</dd></div>
              <div><dt>Роль</dt><dd>{{ roleLabel(s.role) }}</dd></div>
              <div><dt>Статус</dt><dd>{{ s.is_active ? 'активен' : 'деактивирован' }}</dd></div>
            </dl>

            <div class="name-edit">
              <label>
                Имя
                <input [(ngModel)]="editFirst" name="firstName" />
              </label>
              <label>
                Фамилия
                <input [(ngModel)]="editLast" name="lastName" />
              </label>
              <button type="button" (click)="saveName(s)">Сохранить имя</button>
            </div>

            <div class="safe-actions">
              <button type="button" (click)="toggleActive(s)">
                {{ s.is_active ? 'Деактивировать вход' : 'Активировать вход' }}
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
                <button type="button" class="btn-danger" (click)="promoteToCoach(s)">
                  Сделать тренером
                </button>
                <button type="button" class="btn-muted" (click)="demoteToStudent(s)" *ngIf="s.role === 'coach'">
                  Вернуть в ученики
                </button>
                <button type="button" class="btn-ghost" (click)="roleMenuOpen = false">Скрыть</button>
              </div>
            </details>

            <div class="purge">
              <p class="warn">
                Удаление стирает человека, абонементы, записи и явку.
                В статистике его больше не будет. Вернуть нельзя.
              </p>
              <button type="button" class="btn-danger" (click)="deleteStudent(s)">Удалить ученика</button>
            </div>
          </div>
        </li>
        <li *ngIf="!students.length" class="empty">Никого не найдено</li>
      </ul>
    </div>
  `,
  styles: [`
    .page { max-width: 720px; margin: 0 auto; padding: 16px; color: var(--text); }
    .hint, .muted { color: var(--muted); font-size: 14px; margin: 0 0 8px; }
    .filters { display: flex; flex-wrap: wrap; gap: 8px; margin: 12px 0; }
    .filters input, .filters select, button { padding: 8px; font-size: 14px; }
    .list { list-style: none; padding: 0; margin: 0; border: 1px solid var(--border); border-radius: 8px; overflow: hidden; background: var(--card); }
    .person { border-bottom: 1px solid var(--border); }
    .person:last-child { border-bottom: none; }
    .person.inactive .row { opacity: 0.55; }
    .person.open { background: var(--hover); }
    .row {
      width: 100%;
      display: flex;
      justify-content: space-between;
      align-items: center;
      gap: 12px;
      padding: 12px 14px;
      background: transparent;
      border: none;
      color: inherit;
      text-align: left;
      cursor: pointer;
    }
    .row:hover { background: var(--hover); }
    .chevron { color: var(--muted); }
    .meta { display: block; color: var(--muted); font-size: 13px; margin-top: 2px; }
    .empty { padding: 16px; color: var(--muted); }
    .detail { padding: 4px 14px 16px; }
    .info { margin: 0 0 12px; }
    .info > div { display: grid; grid-template-columns: 120px 1fr; gap: 8px; padding: 6px 0; border-bottom: 1px solid var(--border); font-size: 14px; }
    .info dt { color: var(--muted); margin: 0; }
    .info dd { margin: 0; }
    .name-edit, .safe-actions, .nested, .purge { display: flex; flex-direction: column; gap: 8px; margin-top: 12px; }
    .name-edit label { display: flex; flex-direction: column; gap: 4px; font-size: 13px; color: var(--muted); }
    .name-edit input { padding: 8px; font-size: 14px; color: var(--text); background: var(--input-bg); border: 1px solid var(--border); }
    .danger-zone, .purge { border: 1px solid var(--danger-border); border-radius: 8px; padding: 10px 12px; background: var(--danger-bg); }
    .danger-zone summary { cursor: pointer; font-weight: 600; color: var(--danger-text); }
    .warn { color: var(--danger-text); font-size: 13px; line-height: 1.4; margin: 10px 0; }
    .purge .warn { margin-top: 0; }
    .warn.tight { margin-top: 0; }
    .btn-ghost { background: transparent; border: 1px solid var(--border); }
    .btn-muted { background: var(--hover); border: 1px solid var(--border); }
    .btn-danger { background: var(--danger-bg); border: 1px solid var(--danger-border); color: var(--danger-text); }
    .error { color: var(--danger-text); } .ok { color: var(--ok); }
  `]
})
export class StudentsComponent implements OnInit {
  students: StudentRow[] = [];
  selected: StudentRow | null = null;
  editFirst = '';
  editLast = '';
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

  isOpen(s: StudentRow): boolean {
    return this.selected?.user_id === s.user_id;
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
          if (still) {
            this.editFirst = still.first_name || '';
            this.editLast = still.last_name || '';
          } else {
            this.roleMenuOpen = false;
          }
        }
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Ошибка загрузки');
        this.loading = false;
      }
    });
  }

  toggleStudent(s: StudentRow): void {
    if (this.isOpen(s)) {
      this.selected = null;
      this.roleMenuOpen = false;
      return;
    }
    this.selected = s;
    this.editFirst = s.first_name || '';
    this.editLast = s.last_name || '';
    this.roleMenuOpen = false;
    this.message = null;
    this.error = null;
  }

  saveName(s: StudentRow): void {
    const first = this.editFirst.trim();
    const last = this.editLast.trim();
    if (!first || !last) {
      this.error = 'Укажите имя и фамилию';
      return;
    }
    this.message = null;
    this.error = null;
    this.api.patchUser(s.user_id, { first_name: first, last_name: last }).subscribe({
      next: () => {
        this.message = 'Имя сохранено';
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось сохранить имя'))
    });
  }

  deleteStudent(s: StudentRow): void {
    const ok = confirm(
      `Удалить «${s.display_name}»?\n\n` +
        'Пропадут абонементы, записи и явка. В статистике его больше не будет. Вернуть нельзя.'
    );
    if (!ok) return;
    this.message = null;
    this.error = null;
    this.api.deleteUser(s.user_id).subscribe({
      next: () => {
        this.selected = null;
        this.roleMenuOpen = false;
        this.message = 'Ученик удалён';
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось удалить ученика'))
    });
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
