import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { CalendarAPIResponse, CalendarDay, WeekDayData, CalendarEvent, ScheduleDay, TrainingDetails } from '../models/training.model';

@Component({
  selector: 'app-calendar',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './calendar.component.html',
  styleUrls: ['./calendar.component.css']
})
export class CalendarComponent implements OnInit {
  currentView: string = 'month'; // Только месяц
  currentDate: Date = new Date();
  calendarData: CalendarAPIResponse | null = null;
  loading: boolean = false;
  error: string | null = null;
  
  // Modal state
  showModal: boolean = false;
  selectedTraining: TrainingDetails | null = null;
  loadingTraining: boolean = false;

  constructor(
    private calendarService: CalendarService,
    public route: ActivatedRoute,
    private router: Router
  ) {}

  ngOnInit() {
    this.route.queryParams.subscribe(params => {
      const userId = params['user_id'] || null;
      const viewParam = params['view'] || 'month';
      // Разрешаем только месяц и расписание
      const view = (viewParam === 'month' || viewParam === 'schedule') ? viewParam : 'month';
      const date = params['date'] || null;

      this.currentView = view;
      if (date) {
        this.currentDate = new Date(date);
      }

      this.loadCalendar(userId, view, date);
    });
  }

  loadCalendar(userId: string | null, view: string, date: string | null) {
    this.loading = true;
    this.error = null;

    this.calendarService.getCalendar(userId, view, date).subscribe({
      next: (data) => {
        this.calendarData = data;
        this.currentDate = new Date(data.current_date);
        this.currentView = data.view;
        this.loading = false;
        
      },
      error: (err) => {
        this.error = 'Ошибка загрузки календаря: ' + err.message;
        this.loading = false;
      }
    });
  }

  changeView(view: string) {
    // Разрешаем только месяц и расписание
    if (view !== 'month' && view !== 'schedule') {
      view = 'month';
    }
    
    const userId = this.route.snapshot.queryParams['user_id'] || null;
    const dateStr = this.currentDate.toISOString().split('T')[0];
    
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: {
        user_id: userId,
        view: view,
        date: dateStr
      },
      queryParamsHandling: 'merge'
    });
  }

  navigateDate(direction: 'prev' | 'next') {
    if (!this.calendarData) return;

    const dateStr = direction === 'prev' 
      ? this.calendarData.prev_date 
      : this.calendarData.next_date;

    const userId = this.route.snapshot.queryParams['user_id'] || null;
    
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: {
        user_id: userId,
        view: this.currentView,
        date: dateStr
      },
      queryParamsHandling: 'merge'
    });
  }

  showEventDetails(trainingId: number) {
    this.loadingTraining = true;
    this.showModal = true;
    const userId = this.route.snapshot.queryParams['user_id'] || null;

    this.calendarService.getTrainingDetails(trainingId, userId).subscribe({
      next: (data) => {
        this.selectedTraining = data;
        this.loadingTraining = false;
      },
      error: (err) => {
        this.error = 'Ошибка загрузки деталей тренировки: ' + err.message;
        this.loadingTraining = false;
      }
    });
  }

  closeModal() {
    this.showModal = false;
    this.selectedTraining = null;
  }

  registerForTraining() {
    if (!this.selectedTraining) return;

    const userId = this.route.snapshot.queryParams['user_id'];
    if (!userId) {
      alert('Необходимо войти в систему');
      return;
    }

    if (!confirm('Записаться на эту тренировку?')) return;

    this.calendarService.registerForTraining(this.selectedTraining.training.id, userId).subscribe({
      next: () => {
        alert('✅ Вы успешно записались на тренировку!');
        this.closeModal();
        this.reloadCalendar();
      },
      error: (err) => {
        alert('❌ Ошибка: ' + (err.error || err.message || 'Не удалось записаться'));
      }
    });
  }

  cancelRegistration() {
    if (!this.selectedTraining) return;

    const userId = this.route.snapshot.queryParams['user_id'];
    if (!userId) {
      alert('Необходимо войти в систему');
      return;
    }

    if (!confirm('Вы уверены, что хотите отменить запись?')) return;

    this.calendarService.cancelRegistration(this.selectedTraining.training.id, userId).subscribe({
      next: () => {
        alert('✅ Запись отменена');
        this.closeModal();
        this.reloadCalendar();
      },
      error: (err) => {
        alert('❌ Ошибка: ' + (err.error || err.message || 'Не удалось отменить запись'));
      }
    });
  }

  reloadCalendar() {
    const userId = this.route.snapshot.queryParams['user_id'] || null;
    const dateStr = this.currentDate.toISOString().split('T')[0];
    this.loadCalendar(userId, this.currentView, dateStr);
  }

  getEventColorClass(colorIndex: number): string {
    return `event-color-${colorIndex}`;
  }

  formatDate(dateStr: string): string {
    const date = new Date(dateStr);
    const options: Intl.DateTimeFormatOptions = { 
      weekday: 'long', 
      year: 'numeric', 
      month: 'long', 
      day: 'numeric' 
    };
    return date.toLocaleDateString('ru-RU', options);
  }

  formatDateTime(dateTimeStr: string): string {
    if (!dateTimeStr) return '';
    const date = new Date(dateTimeStr);
    return date.toLocaleDateString('ru-RU') + ' ' + 
           date.toLocaleTimeString('ru-RU', { hour: '2-digit', minute: '2-digit' });
  }

  getMonthYear(): string {
    if (!this.calendarData) return '';
    const date = new Date(this.calendarData.current_date);
    const options: Intl.DateTimeFormatOptions = { month: 'long', year: 'numeric' };
    return date.toLocaleDateString('ru-RU', options);
  }

  getDateRange(): string {
    if (!this.calendarData) return '';
    const start = new Date(this.calendarData.start_date);
    const end = new Date(this.calendarData.end_date);
    const startStr = start.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
    const endStr = end.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
    return `${startStr} – ${endStr}`;
  }

  getDayFromDate(dateStr: string): number {
    return new Date(dateStr).getDate();
  }

  trackByEventId(index: number, event: any): number {
    return event.id;
  }

  // Вычисляем размер события для месяца (динамический)
  getEventHeightForMonth(dayEvents: CalendarEvent[]): string {
    if (!dayEvents || dayEvents.length === 0) {
      return 'auto';
    }
    // Равномерно распределяем доступное пространство
    const eventCount = dayEvents.length;
    const baseHeight = 20; // Базовая высота
    const maxHeight = 40; // Максимальная высота
    const height = Math.min(maxHeight, Math.max(baseHeight, 100 / eventCount));
    return `${height}%`;
  }
}
