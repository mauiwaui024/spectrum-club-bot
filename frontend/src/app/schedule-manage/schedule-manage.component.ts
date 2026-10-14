import { Component, HostListener, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface PlanExercise {
  text: string;
  sets?: number | null;
  reps?: number | null;
  rest_seconds?: number | null;
  exercise_id?: number | null;
  legacy?: boolean;
}

@Component({
  selector: 'app-schedule-manage',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterModule],
  templateUrl: './schedule-manage.component.html',
  styleUrls: ['./schedule-manage.component.css']
})
export class ScheduleManageComponent implements OnInit {
  groups: any[] = [];
  trainings: any[] = [];
  loading = false;
  error: string | null = null;
  message: string | null = null;

  /** Monday 00:00 local of the visible week */
  weekStart: Date = this.mondayOf(new Date());

  groupId: number | null = null;
  date = '';
  startTime = '17:00';
  endTime = '19:00';
  description = '';
  weeksCount = 4;

  editingId: number | null = null;
  editDate = '';
  editStartTime = '';
  editEndTime = '';
  editDescription = '';

  menuOpenId: number | null = null;

  planTemplates: { id: number; title: string; description: string; items: PlanExercise[] }[] = [];
  planTrainingId: number | null = null;
  planTemplateId: number | null = null;
  planDescription = '';
  planItems: PlanExercise[] = [{ text: '', exercise_id: null }];
  planSaving = false;
  exercises: { id: number; name: string }[] = [];

  constructor(private calendarService: CalendarService) {}

  ngOnInit(): void {
    this.loadGroups();
    this.load();
    this.calendarService.getExercises().subscribe({
      next: (res) => { this.exercises = res.exercises || []; }
    });
  }

  get weekLabel(): string {
    const end = this.addDays(this.weekStart, 6);
    return `${this.formatDayMonth(this.weekStart)} — ${this.formatDayMonthYear(end)}`;
  }

  @HostListener('document:click')
  onDocumentClick(): void {
    this.menuOpenId = null;
  }

  prevWeek(): void {
    this.weekStart = this.addDays(this.weekStart, -7);
    this.menuOpenId = null;
    this.editingId = null;
    this.load();
  }

  nextWeek(): void {
    this.weekStart = this.addDays(this.weekStart, 7);
    this.menuOpenId = null;
    this.editingId = null;
    this.load();
  }

  goThisWeek(): void {
    this.weekStart = this.mondayOf(new Date());
    this.menuOpenId = null;
    this.editingId = null;
    this.load();
  }

  loadGroups(): void {
    this.calendarService.getGroups().subscribe({
      next: (g) => {
        this.groups = g.groups || [];
        if (!this.groupId && this.groups.length) {
          this.groupId = this.groups[0].id;
        }
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось загрузить группы'))
    });
  }

  load(): void {
    this.loading = true;
    this.error = null;
    const start = this.localDateStr(this.weekStart);
    const end = this.localDateStr(this.addDays(this.weekStart, 6));

    this.calendarService.getCoachSchedule(start, end).subscribe({
      next: (s) => {
        this.trainings = (s.trainings || []).slice().sort((a, b) => {
          const ka = `${this.datePart(a)}T${this.timePart(a.start_time)}`;
          const kb = `${this.datePart(b)}T${this.timePart(b.start_time)}`;
          return ka.localeCompare(kb);
        });
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось загрузить расписание');
        this.loading = false;
      }
    });
  }

  createTraining(): void {
    if (!this.groupId || !this.date) {
      this.error = 'Укажите группу и дату';
      return;
    }
    this.message = null;
    this.calendarService.createTraining({
      group_id: this.groupId,
      date: this.date,
      start_time: this.startTime,
      end_time: this.endTime,
      description: this.description
    }).subscribe({
      next: () => {
        this.message = 'Тренировка создана';
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка создания'))
    });
  }

  createFromTemplates(): void {
    this.message = null;
    this.calendarService.createFromTemplates(this.weeksCount).subscribe({
      next: (res) => {
        this.message = `Создано тренировок: ${res.created_count} (${res.start} — ${res.end})`;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка создания из шаблонов'))
    });
  }

  toggleMenu(id: number, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = this.menuOpenId === id ? null : id;
  }

  startEdit(t: any, event?: Event): void {
    event?.stopPropagation();
    this.menuOpenId = null;
    this.planTrainingId = null;
    this.editingId = t.id;
    this.editDate = this.datePart(t);
    this.editStartTime = this.timePart(t.start_time) || '17:00';
    this.editEndTime = this.timePart(t.end_time) || '19:00';
    this.editDescription = this.displayDescription(t);
    this.error = null;
    this.message = null;
  }

  cancelEdit(): void {
    this.editingId = null;
  }

  saveEdit(): void {
    if (!this.editingId) return;
    this.calendarService.updateTraining(this.editingId, {
      date: this.editDate,
      start_time: this.editStartTime,
      end_time: this.editEndTime,
      description: this.editDescription
    }).subscribe({
      next: () => {
        this.message = 'Сохранено';
        this.editingId = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка сохранения'))
    });
  }

  deleteTraining(id: number, event?: Event): void {
    event?.stopPropagation();
    this.menuOpenId = null;
    if (!confirm('Удалить тренировку?')) return;
    this.calendarService.deleteTraining(id).subscribe({
      next: () => {
        this.message = 'Удалено';
        if (this.editingId === id) this.editingId = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка удаления'))
    });
  }

  formatSlot(t: any): string {
    const when = this.formatDate(t);
    const group = t.group_name || `группа #${t.group_id}`;
    const place = this.displayDescription(t);
    return place ? `${when} — ${group} · ${place}` : `${when} — ${group}`;
  }

  /** Strip legacy "place - GroupName" suffix if it duplicates group_name. */
  displayDescription(t: any): string {
    let desc = (t.description || '').trim();
    const group = (t.group_name || '').trim();
    if (group && desc.endsWith(` - ${group}`)) {
      desc = desc.slice(0, desc.length - ` - ${group}`.length).trim();
    } else if (group && desc.endsWith(` — ${group}`)) {
      desc = desc.slice(0, desc.length - ` — ${group}`.length).trim();
    }
    return desc;
  }

  formatDate(t: any): string {
    const dateStr = this.datePart(t);
    const start = this.timePart(t.start_time);
    const end = this.timePart(t.end_time);
    const timeStr = start && end ? `${start}–${end}` : start;
    const [y, m, d] = dateStr.split('-').map(Number);
    if (!y || !m || !d) {
      return timeStr || '—';
    }
    const dd = String(d).padStart(2, '0');
    const mm = String(m).padStart(2, '0');
    return timeStr ? `${dd}.${mm}, ${timeStr}` : `${dd}.${mm}`;
  }

  private datePart(t: any): string {
    const raw = t?.training_date || t?.date || '';
    if (typeof raw === 'string' && /^\d{4}-\d{2}-\d{2}/.test(raw)) {
      return raw.slice(0, 10);
    }
    if (raw) {
      const d = new Date(raw);
      if (!Number.isNaN(d.getTime())) {
        return this.localDateStr(d);
      }
    }
    return '';
  }

  private timePart(raw: unknown): string {
    if (raw == null || raw === '') return '';
    const s = String(raw);
    const iso = s.match(/T(\d{2}:\d{2})/);
    if (iso) return iso[1];
    const hhmm = s.match(/^(\d{1,2}:\d{2})/);
    if (hhmm) {
      const [h, m] = hhmm[1].split(':');
      return `${h.padStart(2, '0')}:${m}`;
    }
    return '';
  }

  private mondayOf(d: Date): Date {
    const x = new Date(d.getFullYear(), d.getMonth(), d.getDate());
    const day = x.getDay(); // 0=Sun … 6=Sat
    const offset = day === 0 ? -6 : 1 - day;
    x.setDate(x.getDate() + offset);
    return x;
  }

  private addDays(d: Date, n: number): Date {
    const x = new Date(d.getFullYear(), d.getMonth(), d.getDate());
    x.setDate(x.getDate() + n);
    return x;
  }

  private formatDayMonth(d: Date): string {
    const dd = String(d.getDate()).padStart(2, '0');
    const mm = String(d.getMonth() + 1).padStart(2, '0');
    return `${dd}.${mm}`;
  }

  private formatDayMonthYear(d: Date): string {
    return `${this.formatDayMonth(d)}.${d.getFullYear()}`;
  }

  private localDateStr(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  openPlan(t: { id: number }, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    this.editingId = null;
    this.planTrainingId = t.id;
    this.planTemplateId = null;
    this.planDescription = '';
    this.planItems = [{ text: '', exercise_id: null }];
    this.error = null;
    this.calendarService.getPlanTemplates().subscribe({
      next: (res) => { this.planTemplates = res.templates || []; }
    });
    this.calendarService.getTrainingPlan(t.id).subscribe({
      next: (res) => {
        const plan = res.plan;
        if (!plan) return;
        this.planTemplateId = plan.source_template_id ?? null;
        this.planDescription = plan.description || '';
        this.planItems = (plan.items || []).map((item: PlanExercise) => this.copyPlanItem(item));
        if (!this.planItems.length) this.planItems = [{ text: '', exercise_id: null }];
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось открыть план');
      }
    });
  }

  onPlanTemplatePick(id: number | null): void {
    this.planTemplateId = id;
    this.applyPlanTemplate();
  }

  applyPlanTemplate(): void {
    const tpl = this.planTemplates.find((item) => item.id === this.planTemplateId);
    if (!tpl) {
      this.planDescription = '';
      this.planItems = [{ text: '', exercise_id: null }];
      return;
    }
    this.planDescription = tpl.description || '';
    this.planItems = (tpl.items || []).map((item) => this.copyPlanItem(item));
    if (!this.planItems.length) this.planItems = [{ text: '', exercise_id: null }];
  }

  private copyPlanItem(item: PlanExercise): PlanExercise {
    return {
      text: item.text,
      sets: item.sets,
      reps: item.reps,
      rest_seconds: item.rest_seconds,
      exercise_id: item.exercise_id ?? null,
      legacy: item.exercise_id == null && !!item.text
    };
  }

  addPlanItem(): void {
    this.planItems.push({ text: '', exercise_id: null });
  }

  removePlanItem(index: number): void {
    this.planItems.splice(index, 1);
    if (!this.planItems.length) this.planItems = [{ text: '', exercise_id: null }];
  }

  cancelPlan(): void {
    this.planTrainingId = null;
  }

  savePlan(): void {
    if (this.planTrainingId == null) return;
    this.planSaving = true;
    this.calendarService.saveTrainingPlan(this.planTrainingId, {
      source_template_id: this.planTemplateId,
      description: this.planDescription.trim(),
      items: this.planItems
        .map((item) => ({
          text: item.exercise_id ? '' : item.text.trim(),
          sets: item.sets && item.sets > 0 ? item.sets : null,
          reps: item.reps && item.reps > 0 ? item.reps : null,
          rest_seconds: item.rest_seconds && item.rest_seconds > 0 ? item.rest_seconds : null,
          exercise_id: item.exercise_id ?? null
        }))
        .filter((item) => item.exercise_id != null || item.text.length > 0)
    }).subscribe({
      next: () => {
        this.planSaving = false;
        this.planTrainingId = null;
        this.message = 'План тренировки сохранён';
      },
      error: (err) => {
        this.planSaving = false;
        this.error = httpErrorMessage(err, 'Не удалось сохранить план');
      }
    });
  }
}
