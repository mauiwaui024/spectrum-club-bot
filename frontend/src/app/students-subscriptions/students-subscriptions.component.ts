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
  error: string | null = null;
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

  getExpiredSubscriptions(student: StudentWithSubscriptions): Subscription[] {
    return student.subscriptions.filter(sub => sub.status !== 'active');
  }

  getTotalLessonsCount(student: StudentWithSubscriptions): number {
    if (student.total_lessons_count !== undefined) {
      return student.total_lessons_count;
    }
    // Fallback: подсчитываем из активных абонементов
    return this.getActiveSubscriptions(student).reduce((sum, sub) => sum + sub.remaining_lessons, 0);
  }

  startEdit(subscription: Subscription) {
    this.editingSubscriptionId = subscription.id;
    this.addLessonsCount = 1;
    this.removeLessonsCount = 1;
    this.error = null;
  }

  cancelEdit() {
    this.editingSubscriptionId = null;
    this.addLessonsCount = 1;
    this.removeLessonsCount = 1;
    this.error = null;
  }

  isEditing(subscriptionId: number): boolean {
    return this.editingSubscriptionId === subscriptionId;
  }

  addLessons(subscription: Subscription) {
    if (this.addLessonsCount <= 0) {
      this.error = 'Количество занятий должно быть больше 0';
      return;
    }

    this.saving = true;
    this.error = null;

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
        this.error = errorMessage;
        this.saving = false;
        console.error('Error adding lessons:', err);
      }
    });
  }

  removeLessons(subscription: Subscription) {
    if (this.removeLessonsCount <= 0) {
      this.error = 'Количество занятий должно быть больше 0';
      return;
    }

    this.saving = true;
    this.error = null;

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
        this.error = errorMessage;
        this.saving = false;
        console.error('Error removing lessons:', err);
      }
    });
  }
}
