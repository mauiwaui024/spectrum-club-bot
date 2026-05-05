import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { AllStudentsSubscriptionsResponse, StudentWithSubscriptions, Subscription } from '../models/student.model';

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
  saving: boolean = false;

  constructor(private calendarService: CalendarService) {}

  ngOnInit() {
    this.loadData();
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

  getMaxAddableLessons(sub: Subscription): number {
    // Определяем максимальный лимит для типа абонемента
    // Определяем тип по точным значениям или ближайшему нижнему
    let maxTypeLimit: number;
    if (sub.total_lessons === 1) {
      maxTypeLimit = 1;
    } else if (sub.total_lessons === 12) {
      maxTypeLimit = 12;
    } else if (sub.total_lessons === 16) {
      maxTypeLimit = 16;
    } else if (sub.total_lessons < 12) {
      // Между 1 и 12, предполагаем тип 12
      maxTypeLimit = 12;
    } else {
      // >= 13 или > 16, предполагаем тип 16
      maxTypeLimit = 16;
    }

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
    if (sub.total_lessons === 1) {
      return '1 занятие';
    } else if (sub.total_lessons === 12) {
      return '12 занятий - несгораемый';
    } else if (sub.total_lessons === 16) {
      return '16 занятий - 30 дней';
    } else if (sub.total_lessons < 12) {
      return '12 занятий - несгораемый';
    } else {
      return '16 занятий - 30 дней';
    }
  }

  startEdit(subscription: Subscription) {
    this.editingSubscriptionId = subscription.id;
    this.addLessonsCount = 1;
    this.removeLessonsCount = 1;
    this.editError = null;
  }

  cancelEdit() {
    this.editingSubscriptionId = null;
    this.addLessonsCount = 1;
    this.removeLessonsCount = 1;
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
      // Определяем лимит типа для сообщения об ошибке
      let maxTypeLimit: number;
      if (subscription.total_lessons <= 1) {
        maxTypeLimit = 1;
      } else if (subscription.total_lessons <= 12) {
        maxTypeLimit = 12;
      } else {
        maxTypeLimit = 16;
      }
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
      this.editError = `Недостаточно занятий. Доступно: ${maxRemovable}`;
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
}
