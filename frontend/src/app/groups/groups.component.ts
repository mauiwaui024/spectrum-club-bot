import { Component, HostListener, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { httpErrorMessage } from '../utils/http-error';

interface GroupRow {
  id: number;
  name: string;
  code: string;
  age_min: number;
}

@Component({
  selector: 'app-groups',
  standalone: true,
  imports: [CommonModule, FormsModule],
  templateUrl: './groups.component.html',
  styleUrl: './groups.component.css'
})
export class GroupsComponent implements OnInit {
  groups: GroupRow[] = [];
  loading = false;
  error: string | null = null;
  message: string | null = null;

  menuOpenId: number | null = null;
  editingId: number | null = null;
  editName = '';
  editCode = '';
  editAgeMin = 8;

  newName = '';
  newCode = '';
  newAgeMin = 8;

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
    this.api.getGroups().subscribe({
      next: (res) => {
        this.groups = res.groups || [];
        this.loading = false;
      },
      error: (err) => {
        this.error = httpErrorMessage(err, 'Не удалось загрузить группы');
        this.loading = false;
      }
    });
  }

  toggleMenu(id: number, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = this.menuOpenId === id ? null : id;
  }

  startEdit(g: GroupRow, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    this.editingId = g.id;
    this.editName = g.name;
    this.editCode = g.code;
    this.editAgeMin = g.age_min ?? 8;
    this.message = null;
    this.error = null;
  }

  cancelEdit(): void {
    this.editingId = null;
  }

  saveEdit(): void {
    if (this.editingId == null) return;
    this.api.updateGroup(this.editingId, {
      name: this.editName.trim(),
      code: this.editCode.trim(),
      age_min: Number(this.editAgeMin)
    }).subscribe({
      next: () => {
        this.message = 'Группа сохранена';
        this.editingId = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка сохранения'))
    });
  }

  deleteGroup(g: GroupRow, event: Event): void {
    event.stopPropagation();
    this.menuOpenId = null;
    if (!confirm(`Удалить группу «${g.name}»?`)) return;
    this.api.deleteGroup(g.id).subscribe({
      next: () => {
        this.message = 'Группа удалена';
        if (this.editingId === g.id) this.editingId = null;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Не удалось удалить'))
    });
  }

  createGroup(): void {
    if (!this.newName.trim() || !this.newCode.trim()) {
      this.error = 'Укажите название и код';
      return;
    }
    this.error = null;
    this.api.createGroup({
      name: this.newName.trim(),
      code: this.newCode.trim(),
      age_min: Number(this.newAgeMin)
    }).subscribe({
      next: () => {
        this.message = 'Группа создана';
        this.newName = '';
        this.newCode = '';
        this.newAgeMin = 8;
        this.load();
      },
      error: (err) => (this.error = httpErrorMessage(err, 'Ошибка создания'))
    });
  }
}
