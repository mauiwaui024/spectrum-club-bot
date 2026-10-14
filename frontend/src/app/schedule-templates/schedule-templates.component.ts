import { Component, HostListener, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface TemplateSlot {
  id: number;
  group_id: number;
  day_of_week: number;
  start_time: string;
  end_time: string;
  description: string;
  is_active: boolean;
  group_name?: string;
  group_code?: string;
}

@Component({
  selector: 'app-schedule-templates',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './schedule-templates.component.html',
  styleUrl: './schedule-templates.component.css'
})
export class ScheduleTemplatesComponent implements OnInit {
  readonly dayNames = ['', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота', 'Воскресенье'];
  readonly days = [1, 2, 3, 4, 5, 6, 7];

  groups: any[] = [];
  templates: TemplateSlot[] = [];
  loading = false;
  error: string | null = null;
  message: string | null = null;

  editing: TemplateSlot | null = null;
  editGroupId: number | null = null;
  editStart = '';
  editEnd = '';
  editDesc = '';
  editActive = true;

  addingDay: number | null = null;
  addGroupId: number | null = null;
  addStart = '17:00';
  addEnd = '19:00';
  addDesc = '';

  menuOpenId: number | null = null;

  constructor(private api: CalendarService) {}

  ngOnInit(): void {
    this.load();
  }

  @HostListener('document:click')
  closeMenus(): void {
    this.menuOpenId = null;
  }

  toggleMenu(id: number, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = this.menuOpenId === id ? null : id;
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getGroups().subscribe({
      next: (g) => {
        this.groups = g.groups || [];
        if (!this.addGroupId && this.groups.length) {
          this.addGroupId = this.groups[0].id;
        }
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось загрузить группы'))
    });
    this.api.getTemplates().subscribe({
      next: (t) => {
        this.templates = t.templates || [];
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось загрузить шаблоны');
        this.loading = false;
      }
    });
  }

  slotsForDay(day: number): TemplateSlot[] {
    return this.templates
      .filter((t) => t.day_of_week === day)
      .sort((a, b) => String(a.start_time).localeCompare(String(b.start_time)));
  }

  formatTime(t: string): string {
    if (!t) return '';
    // Driver may return "0000-01-01T15:30:00Z" for Postgres TIME
    const iso = t.match(/T(\d{2}:\d{2})/);
    if (iso) return iso[1];
    const hhmm = t.match(/^(\d{1,2}:\d{2})/);
    if (hhmm) return hhmm[1].padStart(5, '0');
    return t.length >= 5 ? t.slice(0, 5) : t;
  }

  startEdit(t: TemplateSlot, event?: Event): void {
    event?.stopPropagation();
    this.addingDay = null;
    this.menuOpenId = null;
    this.editing = { ...t };
    this.editGroupId = t.group_id;
    this.editStart = this.formatTime(t.start_time);
    this.editEnd = this.formatTime(t.end_time);
    this.editDesc = t.description || '';
    this.editActive = !!t.is_active;
    this.message = null;
    this.error = null;
  }

  cancelEdit(): void {
    this.editing = null;
  }

  saveEdit(): void {
    if (!this.editing || !this.editGroupId) return;
    this.api.updateTemplate(this.editing.id, {
      group_id: this.editGroupId,
      day_of_week: this.editing.day_of_week,
      start_time: this.editStart,
      end_time: this.editEnd,
      description: this.editDesc,
      is_active: this.editActive
    }).subscribe({
      next: () => {
        this.message = 'Шаблон сохранён';
        this.editing = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка сохранения'))
    });
  }

  toggleActive(t: TemplateSlot, event?: Event): void {
    event?.stopPropagation();
    this.menuOpenId = null;
    const req = t.is_active ? this.api.deactivateTemplate(t.id) : this.api.activateTemplate(t.id);
    req.subscribe({
      next: () => this.load(),
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось изменить статус'))
    });
  }

  deleteTemplate(t: TemplateSlot, event?: Event): void {
    event?.stopPropagation();
    this.menuOpenId = null;
    if (!confirm(`Удалить слот ${this.formatTime(t.start_time)}–${this.formatTime(t.end_time)}?`)) return;
    this.api.deleteTemplate(t.id).subscribe({
      next: () => {
        if (this.editing?.id === t.id) this.editing = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка удаления'))
    });
  }

  startAdd(day: number): void {
    this.editing = null;
    this.addingDay = day;
    this.addStart = '17:00';
    this.addEnd = '19:00';
    this.addDesc = '';
    if (!this.addGroupId && this.groups.length) {
      this.addGroupId = this.groups[0].id;
    }
    this.message = null;
    this.error = null;
  }

  cancelAdd(): void {
    this.addingDay = null;
  }

  saveAdd(): void {
    if (this.addingDay == null || !this.addGroupId) return;
    this.api.createTemplate({
      group_id: this.addGroupId,
      day_of_week: this.addingDay,
      start_time: this.addStart,
      end_time: this.addEnd,
      description: this.addDesc
    }).subscribe({
      next: () => {
        this.message = `Слот добавлен в ${this.dayNames[this.addingDay!]}`;
        this.addingDay = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка создания'))
    });
  }
}
