import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { MySubscriptionResponse, Subscription } from '../models/student.model';

@Component({
  selector: 'app-my-subscription',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './my-subscription.component.html',
  styleUrls: ['./my-subscription.component.css']
})
export class MySubscriptionComponent implements OnInit {
  subscription: MySubscriptionResponse | null = null;
  loading: boolean = false;
  error: string | null = null;

  constructor(private calendarService: CalendarService) {}

  ngOnInit() {
    this.loadSubscription();
  }

  loadSubscription() {
    this.loading = true;
    this.error = null;
    
    this.calendarService.getMySubscription().subscribe({
      next: (data) => {
        this.subscription = data;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Не удалось загрузить информацию об абонементе. Попробуйте позже.';
        this.loading = false;
        console.error('Error loading subscription:', err);
      }
    });
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
    return '🔄 Использован';
  }

  getProgressPercentage(sub: Subscription): number {
    if (sub.total_lessons === 0) return 0;
    return Math.round((sub.used_lessons / sub.total_lessons) * 100);
  }
}
