import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { MyRegistrationsResponse, TrainingRegistration } from '../models/student.model';

@Component({
  selector: 'app-my-registrations',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './my-registrations.component.html',
  styleUrls: ['./my-registrations.component.css']
})
export class MyRegistrationsComponent implements OnInit {
  registrations: MyRegistrationsResponse | null = null;
  loading: boolean = false;
  error: string | null = null;

  constructor(private calendarService: CalendarService) {}

  ngOnInit() {
    this.loadRegistrations();
  }

  loadRegistrations() {
    this.loading = true;
    this.error = null;
    
    this.calendarService.getMyRegistrations().subscribe({
      next: (data) => {
        this.registrations = data;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Не удалось загрузить записи. Попробуйте позже.';
        this.loading = false;
        console.error('Error loading registrations:', err);
      }
    });
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    const days = ['Воскресенье', 'Понедельник', 'Вторник', 'Среда', 'Четверг', 'Пятница', 'Суббота'];
    const dayName = days[date.getDay()];
    const day = date.getDate().toString().padStart(2, '0');
    const month = (date.getMonth() + 1).toString().padStart(2, '0');
    const year = date.getFullYear();
    return `${dayName}, ${day}.${month}.${year}`;
  }

  getStatusText(registration: TrainingRegistration): string {
    if (registration.attended) {
      return '✅ Посещена';
    }
    if (registration.status === 'registered') {
      return '✅ Записан';
    }
    if (registration.status === 'cancelled') {
      return '❌ Отменена';
    }
    return '❌ Пропущена';
  }

  getStatusClass(registration: TrainingRegistration): string {
    if (registration.attended) {
      return 'status-attended';
    }
    if (registration.status === 'registered') {
      return 'status-registered';
    }
    return 'status-missed';
  }
}
