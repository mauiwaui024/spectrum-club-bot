import { Component, HostListener, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface PlanItem {
  id?: number;
  sort_order?: number;
  text: string;
  sets?: number | null;
  reps?: number | null;
  rest_seconds?: number | null;
  exercise_id?: number | null;
  legacy?: boolean;
}

interface PlanTemplate {
  id: number;
  title: string;
  description: string;
  items: PlanItem[];
}

@Component({
  selector: 'app-training-plans',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './training-plans.component.html',
  styleUrl: './training-plans.component.css'
})
export class TrainingPlansComponent implements OnInit {
  templates: PlanTemplate[] = [];
  loading = false;
  error: string | null = null;
  message: string | null = null;

  menuOpenId: number | null = null;
  editingId: number | null = null;
  editTitle = '';
  editDescription = '';
  editItems: PlanItem[] = [];

  newTitle = '';
  newDescription = '';
  newItems: PlanItem[] = [{ text: '', exercise_id: null }];

  exercises: { id: number; name: string }[] = [];

  constructor(private api: CalendarService) {}

  ngOnInit(): void {
    this.load();
    this.api.getExercises().subscribe({
      next: (res) => { this.exercises = res.exercises || []; },
      error: (err) => { this.error = httpErrorMessage(err, 'Не удалось загрузить упражнения'); }
    });
  }

  @HostListener('document:click')
  closeMenu(): void {
    this.menuOpenId = null;
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getPlanTemplates().subscribe({
      next: (res) => {
        this.templates = res.templates || [];
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось загрузить планы');
        this.loading = false;
      }
    });
  }

  toggleMenu(id: number, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = this.menuOpenId === id ? null : id;
  }

  startEdit(t: PlanTemplate, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    this.editingId = t.id;
    this.editTitle = t.title;
    this.editDescription = t.description || '';
    this.editItems = (t.items || []).map((item) => ({
      text: item.text,
      sets: item.sets,
      reps: item.reps,
      rest_seconds: item.rest_seconds,
      exercise_id: item.exercise_id ?? null,
      legacy: item.exercise_id == null && !!item.text
    }));
    if (!this.editItems.length) this.editItems = [{ text: '', exercise_id: null }];
    this.message = null;
    this.error = null;
  }

  cancelEdit(): void {
    this.editingId = null;
  }

  saveEdit(): void {
    if (this.editingId == null) return;
    this.api.updatePlanTemplate(this.editingId, {
      title: this.editTitle.trim(),
      description: this.editDescription.trim(),
      items: this.payload(this.editItems)
    }).subscribe({
      next: () => {
        this.message = 'Заготовка сохранена';
        this.editingId = null;
        this.load();
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось сохранить');
      }
    });
  }

  deleteTemplate(t: PlanTemplate, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    if (!confirm(`Удалить заготовку «${t.title}»? Уже назначенные тренировки не изменятся.`)) return;
    this.api.deletePlanTemplate(t.id).subscribe({
      next: () => {
        this.message = 'Заготовка удалена';
        if (this.editingId === t.id) this.editingId = null;
        this.load();
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось удалить');
      }
    });
  }

  create(): void {
    this.api.createPlanTemplate({
      title: this.newTitle.trim(),
      description: this.newDescription.trim(),
      items: this.payload(this.newItems)
    }).subscribe({
      next: () => {
        this.message = 'Заготовка создана';
        this.newTitle = '';
        this.newDescription = '';
        this.newItems = [{ text: '', exercise_id: null }];
        this.load();
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось создать');
      }
    });
  }

  addItem(list: PlanItem[]): void {
    list.push({ text: '', exercise_id: null });
  }

  removeItem(list: PlanItem[], index: number): void {
    list.splice(index, 1);
    if (!list.length) list.push({ text: '', exercise_id: null });
  }

  private payload(list: PlanItem[]): { text: string; sets: number | null; reps: number | null; rest_seconds: number | null; exercise_id: number | null }[] {
    return list
      .map((item) => ({
        text: item.exercise_id ? '' : item.text.trim(),
        sets: this.count(item.sets),
        reps: this.count(item.reps),
        rest_seconds: this.count(item.rest_seconds),
        exercise_id: item.exercise_id ?? null
      }))
      .filter((item) => item.exercise_id != null || item.text.length > 0);
  }

  private count(n: number | null | undefined): number | null {
    if (n == null || n <= 0) return null;
    return n;
  }
}
