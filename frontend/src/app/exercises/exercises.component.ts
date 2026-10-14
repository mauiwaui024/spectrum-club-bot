import { Component, HostListener, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface ExerciseRow {
  id: number;
  name: string;
  description: string;
}

@Component({
  selector: 'app-exercises',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './exercises.component.html',
  styleUrl: './exercises.component.css'
})
export class ExercisesComponent implements OnInit {
  exercises: ExerciseRow[] = [];
  loading = false;
  error: string | null = null;
  message: string | null = null;

  menuOpenId: number | null = null;
  editingId: number | null = null;
  editName = '';
  editDescription = '';

  newName = '';
  newDescription = '';

  constructor(private api: CalendarService) {}

  ngOnInit(): void {
    this.load();
  }

  @HostListener('document:click')
  closeMenu(): void {
    this.menuOpenId = null;
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getExercises().subscribe({
      next: (res) => {
        this.exercises = res.exercises || [];
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось загрузить упражнения');
        this.loading = false;
      }
    });
  }

  toggleMenu(id: number, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = this.menuOpenId === id ? null : id;
  }

  startEdit(ex: ExerciseRow, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    this.editingId = ex.id;
    this.editName = ex.name;
    this.editDescription = ex.description || '';
    this.message = null;
    this.error = null;
  }

  cancelEdit(): void {
    this.editingId = null;
  }

  saveEdit(): void {
    if (this.editingId == null) return;
    this.api.updateExercise(this.editingId, {
      name: this.editName.trim(),
      description: this.editDescription.trim()
    }).subscribe({
      next: () => {
        this.message = 'Упражнение сохранено';
        this.editingId = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось сохранить'))
    });
  }

  deleteExercise(ex: ExerciseRow, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    if (!confirm(`Удалить упражнение «${ex.name}»?`)) return;
    this.api.deleteExercise(ex.id).subscribe({
      next: () => {
        this.message = 'Упражнение удалено';
        if (this.editingId === ex.id) this.editingId = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось удалить'))
    });
  }

  create(): void {
    if (!this.newName.trim()) {
      this.error = 'Укажите название';
      return;
    }
    this.api.createExercise({
      name: this.newName.trim(),
      description: this.newDescription.trim()
    }).subscribe({
      next: () => {
        this.message = 'Упражнение создано';
        this.newName = '';
        this.newDescription = '';
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось создать'))
    });
  }
}
