import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface ClimberCard {
  student_id: number;
  user_id: number;
  name: string;
  group_id: number | null;
  group_name: string;
  discipline: string;
  last_grade: string;
  is_self: boolean;
  remaining_lessons?: number | null;
}

interface GroupOption {
  id: number;
  name: string;
}

@Component({
  selector: 'app-climbers',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './climbers.component.html',
  styleUrl: './climbers.component.css'
})
export class ClimbersComponent implements OnInit {
  climbers: ClimberCard[] = [];
  groups: GroupOption[] = [];
  isCoach = false;
  loading = false;
  error: string | null = null;
  message: string | null = null;

  editingId: number | null = null;
  editGroupId: number | null = null;
  editDiscipline = '';
  editGrade = '';

  constructor(private api: CalendarService) {}

  ngOnInit(): void {
    this.load();
  }

  load(): void {
    this.loading = true;
    this.error = null;
    this.api.getClimbers().subscribe({
      next: (res) => {
        this.climbers = res.climbers || [];
        this.isCoach = !!res.is_coach;
        this.loading = false;
        if (this.isCoach && !this.groups.length) {
          this.api.getGroups().subscribe({
            next: (groups) => { this.groups = groups.groups || []; }
          });
        }
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось загрузить карточки');
        this.loading = false;
      }
    });
  }

  disciplineLabel(value: string): string {
    if (value === 'boulder') return 'боулдер';
    if (value === 'lead') return 'трудность';
    if (value === 'both') return 'боулдер и трудность';
    return '';
  }

  startEdit(card: ClimberCard): void {
    this.editingId = card.student_id;
    this.editGroupId = card.group_id || null;
    this.editDiscipline = card.discipline || '';
    this.editGrade = card.last_grade || '';
    this.message = null;
    this.error = null;
  }

  cancelEdit(): void {
    this.editingId = null;
  }

  saveEdit(card: ClimberCard): void {
    this.api.updateClimber(card.student_id, {
      group_id: this.editGroupId || 0,
      discipline: this.editDiscipline,
      last_grade: this.editGrade
    }).subscribe({
      next: () => {
        this.message = 'Карточка сохранена';
        this.editingId = null;
        this.load();
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось сохранить карточку');
      }
    });
  }
}
