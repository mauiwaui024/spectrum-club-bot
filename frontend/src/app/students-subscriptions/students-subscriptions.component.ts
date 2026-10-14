import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { AllStudentsSubscriptionsResponse, StudentWithSubscriptions, Subscription } from '../models/student.model';
import { httpErrorMessage } from '../utils/http-error';

@Component({
  selector: 'app-students-subscriptions',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './students-subscriptions.component.html',
  styleUrls: ['./students-subscriptions.component.css']
})
export class StudentsSubscriptionsComponent implements OnInit {
  data: AllStudentsSubscriptionsResponse | null = null;
  loading: boolean = false;
  error: string | null = null; // Общая ошибка загрузки данных
  editError: string | null = null; // Ошибка редактирования (показывается только в форме)
  expandedStudents: Set<number> = new Set();
  editingSubscriptionId: number | null = null;
  addLessonsCount: number = 1;
  removeLessonsCount: number = 1;
  editEndDate = '';
  saving: boolean = false;
  allStudents: any[] = [];
  createStudentUserId: number | null = null;
  createCatalogKey = 'pack12';
  createEndDate = '';
  createMessage: string | null = null;

  readonly catalog: { key: string; label: string; lessons: number; defaultDays: number | 'years2' }[] = [
    { key: 'single1', label: 'Разовое — 1 посещение', lessons: 1, defaultDays: 30 },
    { key: 'single2', label: 'Разовое — 2 посещения', lessons: 2, defaultDays: 30 },
    { key: 'single3', label: 'Разовое — 3 посещения', lessons: 3, defaultDays: 30 },
    { key: 'single4', label: 'Разовое — 4 посещения', lessons: 4, defaultDays: 30 },
    { key: 'pack4', label: 'Абонемент на 4', lessons: 4, defaultDays: 30 },
    { key: 'pack8', label: 'Абонемент на 8', lessons: 8, defaultDays: 30 },
    { key: 'pack12', label: 'Абонемент на 12', lessons: 12, defaultDays: 'years2' },
    { key: 'pack16', label: 'Абонемент на 16', lessons: 16, defaultDays: 30 },
  ];

  constructor(private calendarService: CalendarService) {}

  ngOnInit() {
    this.applyDefaultEndDate();
    this.loadData();
    this.calendarService.getStudents().subscribe({
      next: (res) => (this.allStudents = res.students || []),
      error: () => {}
    });
  }

  onCatalogChange(): void {
    this.applyDefaultEndDate();
  }

  applyDefaultEndDate(): void {
    const item = this.catalog.find(c => c.key === this.createCatalogKey);
    const d = new Date();
    if (item?.defaultDays === 'years2') {
      d.setFullYear(d.getFullYear() + 2);
    } else {
      d.setDate(d.getDate() + (typeof item?.defaultDays === 'number' ? item.defaultDays : 30));
    }
    this.createEndDate = this.toDateInput(d);
  }

  private toDateInput(d: Date): string {
    const y = d.getFullYear();
    const m = String(d.getMonth() + 1).padStart(2, '0');
    const day = String(d.getDate()).padStart(2, '0');
    return `${y}-${m}-${day}`;
  }

  studentOptionLabel(s: any): string {
    return s.display_name || [s.last_name, s.first_name].filter(Boolean).join(' ') || s.username || ('#' + s.user_id);
  }

  createSubscription(): void {
    if (!this.createStudentUserId) {
      this.error = 'Выберите ученика';
      return;
    }
    if (!this.createEndDate) {
      this.error = 'Укажите срок действия';
      return;
    }
    const item = this.catalog.find(c => c.key === this.createCatalogKey);
    if (!item) {
      this.error = 'Выберите тип абонемента';
      return;
    }
    this.createMessage = null;
    this.error = null;
    this.calendarService.createSubscription(this.createStudentUserId, item.lessons, this.createEndDate).subscribe({
      next: () => {
        this.createMessage = 'Абонемент создан';
        this.loadData();
      },
      error: (err) => (this.error = err?.error || 'Не удалось создать абонемент')
    });
  }

  deleteSubscription(sub: Subscription): void {
    if (!confirm('Удалить абонемент #' + sub.id + '?')) return;
    this.calendarService.deleteSubscription(sub.id).subscribe({
      next: () => this.loadData(),
      error: (err) => (this.editError = err?.error || 'Не удалось удалить')
    });
  }

  loadData() {
    this.loading = true;
    this.error = null;
    
    this.calendarService.getAllStudentsSubscriptions().subscribe({
      next: (data) => {
        console.log('Students subscriptions data received:', data);
        this.data = data;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Не удалось загрузить информацию об абонементах учеников. Попробуйте позже.';
        this.loading = false;
        console.error('Error loading students subscriptions:', err);
        if (err.error) {
          console.error('Error details:', err.error);
        }
      }
    });
  }

  toggleStudent(studentId: number) {
    if (this.expandedStudents.has(studentId)) {
      this.expandedStudents.delete(studentId);
    } else {
      this.expandedStudents.add(studentId);
    }
  }

  isExpanded(studentId: number): boolean {
    return this.expandedStudents.has(studentId);
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const year = date.getFullYear();
    return `${day}.${month}.${year}`;
  }

  getStatusText(sub: Subscription): string {
    if (sub.status === 'expired') {
      return '⏰ Истек';
    }
    if (sub.status === 'used') {
      return '🔄 Использован';
    }
    return '✅ Активен';
  }

  getStatusClass(sub: Subscription): string {
    if (sub.status === 'expired' || sub.status === 'used') {
      return 'status-expired';
    }
    return 'status-active';
  }

  getProgressPercentage(sub: Subscription): number {
    if (sub.total_lessons === 0) return 0;
    return Math.round((sub.used_lessons / sub.total_lessons) * 100);
  }

  getActiveSubscriptions(student: StudentWithSubscriptions): Subscription[] {
    return student.subscriptions.filter(sub => sub.status === 'active');
  }

  getActiveSubscriptionsTypes(student: StudentWithSubscriptions): string {
    const activeSubs = this.getActiveSubscriptions(student);
    if (activeSubs.length === 0) {
      return '';
    }
    // Получаем уникальные типы абонементов
    const types = activeSubs.map(sub => this.getSubscriptionTypeName(sub));
    const uniqueTypes = [...new Set(types)];
    return uniqueTypes.join(', ');
  }

  getExpiredSubscriptions(student: StudentWithSubscriptions): Subscription[] {
    return student.subscriptions.filter(sub => sub.status !== 'active');
  }

  getTotalLessonsCount(student: StudentWithSubscriptions): number {
    if (student.total_lessons_count !== undefined) {
      return student.total_lessons_count;
    }
    // Fallback: подсчитываем из активных абонементов
    return this.getActiveSubscriptions(student).reduce((sum, sub) => sum + Math.max(sub.remaining_lessons, 0), 0);
  }

  getMaxRemovableLessons(sub: Subscription): number {
    // Разрешаем ручное снятие до порога -2 включительно.
    return Math.max(0, sub.remaining_lessons + 2);
  }

  private maxLessonsForType(totalLessons: number): number {
    const limits = [1, 2, 3, 4, 8, 12, 16];
    if (limits.includes(totalLessons)) {
      return totalLessons;
    }
    for (const limit of limits) {
      if (totalLessons <= limit) {
        return limit;
      }
    }
    return 16;
  }

  getMaxAddableLessons(sub: Subscription): number {
    const maxTypeLimit = this.maxLessonsForType(sub.total_lessons);

    // Максимум, который можно добавить до исходного количества (использованные занятия)
    const maxToOriginal = sub.total_lessons - sub.remaining_lessons;

    // Если total_lessons превышает лимит типа, можно добавить только до лимита типа
    if (sub.total_lessons > maxTypeLimit) {
      // Максимум до лимита типа по total_lessons
      const maxTotalAtLimit = maxTypeLimit;
      const currentTotal = sub.total_lessons;
      const possibleToAddToTotalLimit = maxTotalAtLimit - currentTotal;
      
      // Максимум до лимита типа по remaining_lessons
      const maxRemainingAtLimit = maxTypeLimit;
      const currentRemaining = sub.remaining_lessons;
      const possibleToAddToRemainingLimit = maxRemainingAtLimit - currentRemaining;
      
      // Берем минимум из всех ограничений
      let possibleToAdd = maxToOriginal;
      if (possibleToAddToTotalLimit < possibleToAdd) {
        possibleToAdd = possibleToAddToTotalLimit;
      }
      if (possibleToAddToRemainingLimit < possibleToAdd && possibleToAddToRemainingLimit >= 0) {
        possibleToAdd = possibleToAddToRemainingLimit;
      }
      return Math.max(0, possibleToAdd);
    } else if (sub.total_lessons === maxTypeLimit) {
      // Точно на лимите типа, можно только вернуть использованные занятия
      return maxToOriginal;
    } else {
      // Ниже лимита типа, проверяем оба ограничения
      const maxToTypeLimit = maxTypeLimit - sub.total_lessons;
      return Math.min(maxToTypeLimit, maxToOriginal);
    }
  }

  getSubscriptionTypeName(sub: Subscription): string {
    const n = sub.total_lessons;
    if (n === 1) return '1 занятие';
    if (n >= 2 && n <= 4) return `${n} занятия`;
    return `${n} занятий`;
  }

  startEdit(subscription: Subscription) {
    this.editingSubscriptionId = subscription.id;
    this.addLessonsCount = 1;
    this.removeLessonsCount = 1;
    this.editEndDate = (subscription.end_date || '').slice(0, 10);
    this.editError = null;
  }

  cancelEdit() {
    this.editingSubscriptionId = null;
    this.addLessonsCount = 1;
    this.removeLessonsCount = 1;
    this.editEndDate = '';
    this.editError = null;
  }

  isEditing(subscriptionId: number): boolean {
    return this.editingSubscriptionId === subscriptionId;
  }

  addLessons(subscription: Subscription) {
    if (this.addLessonsCount <= 0) {
      this.editError = 'Количество занятий должно быть больше 0';
      return;
    }

    // Валидация: проверяем максимальное количество, которое можно добавить
    const maxToAdd = this.getMaxAddableLessons(subscription);
    if (this.addLessonsCount > maxToAdd) {
      const maxTypeLimit = this.maxLessonsForType(subscription.total_lessons);
      this.editError = `Можно добавить максимум ${maxToAdd} занятий (до исходного лимита абонемента ${maxTypeLimit} занятий)`;
      return;
    }

    this.saving = true;
    this.editError = null;

    this.calendarService.addLessons(subscription.id, this.addLessonsCount).subscribe({
      next: (updatedSub) => {
        // Обновляем абонемент в данных
        if (this.data) {
          for (const student of this.data.students) {
            const subIndex = student.subscriptions.findIndex(s => s.id === subscription.id);
            if (subIndex !== -1) {
              student.subscriptions[subIndex] = updatedSub;
              // Пересчитываем total_lessons_count
              student.total_lessons_count = this.getTotalLessonsCount(student);
            }
          }
        }
        this.editingSubscriptionId = null;
        this.saving = false;
        this.addLessonsCount = 1;
        this.editError = null;
      },
      error: (err) => {
        let errorMessage = 'Не удалось добавить занятия. Попробуйте позже.';
        if (err.error && err.error.error) {
          errorMessage = err.error.error;
        } else if (err.error && typeof err.error === 'string') {
          errorMessage = err.error;
        } else if (err.message) {
          errorMessage = err.message;
        }
        this.editError = errorMessage;
        this.saving = false;
        console.error('Error adding lessons:', err);
      }
    });
  }

  removeLessons(subscription: Subscription) {
    if (this.removeLessonsCount <= 0) {
      this.editError = 'Количество занятий должно быть больше 0';
      return;
    }

    // Валидация: проверяем, что достаточно занятий для снятия
    const maxRemovable = this.getMaxRemovableLessons(subscription);
    if (this.removeLessonsCount > maxRemovable) {
      if (maxRemovable === 0) {
        this.editError = 'Снятие недоступно при текущем балансе (достигнут порог -2)';
      } else {
        this.editError = `Можно снять максимум: ${maxRemovable}`;
      }
      return;
    }

    this.saving = true;
    this.editError = null;

    this.calendarService.removeLessons(subscription.id, this.removeLessonsCount).subscribe({
      next: (updatedSub) => {
        // Обновляем абонемент в данных
        if (this.data) {
          for (const student of this.data.students) {
            const subIndex = student.subscriptions.findIndex(s => s.id === subscription.id);
            if (subIndex !== -1) {
              student.subscriptions[subIndex] = updatedSub;
              // Пересчитываем total_lessons_count
              student.total_lessons_count = this.getTotalLessonsCount(student);
            }
          }
        }
        this.editingSubscriptionId = null;
        this.saving = false;
        this.removeLessonsCount = 1;
        this.editError = null;
      },
      error: (err) => {
        let errorMessage = 'Не удалось снять занятия. Попробуйте позже.';
        if (err.error && err.error.error) {
          errorMessage = err.error.error;
        } else if (err.error && typeof err.error === 'string') {
          errorMessage = err.error;
        } else if (err.message) {
          errorMessage = err.message;
        }
        this.editError = errorMessage;
        this.saving = false;
        console.error('Error removing lessons:', err);
      }
    });
  }

  saveEndDate(subscription: Subscription) {
    if (!this.editEndDate) {
      this.editError = 'Укажите дату окончания';
      return;
    }
    this.saving = true;
    this.editError = null;
    this.calendarService.setSubscriptionEndDate(subscription.id, this.editEndDate).subscribe({
      next: (updatedSub) => {
        if (this.data) {
          for (const student of this.data.students) {
            const subIndex = student.subscriptions.findIndex(s => s.id === subscription.id);
            if (subIndex !== -1) {
              student.subscriptions[subIndex] = updatedSub;
              student.total_lessons_count = this.getTotalLessonsCount(student);
            }
          }
        }
        this.editingSubscriptionId = null;
        this.saving = false;
        this.editError = null;
      },
      error: (err) => {
        this.editError = httpErrorMessage(err, 'Не удалось сохранить срок');
        this.saving = false;
      }
    });
  }
}
