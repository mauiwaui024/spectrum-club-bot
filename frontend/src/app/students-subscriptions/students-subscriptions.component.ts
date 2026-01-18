import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { AllStudentsSubscriptionsResponse, StudentWithSubscriptions, Subscription } from '../models/student.model';

@Component({
  selector: 'app-students-subscriptions',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './students-subscriptions.component.html',
  styleUrls: ['./students-subscriptions.component.css']
})
export class StudentsSubscriptionsComponent implements OnInit {
  data: AllStudentsSubscriptionsResponse | null = null;
  loading: boolean = false;
  error: string | null = null;
  expandedStudents: Set<number> = new Set();

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
}
