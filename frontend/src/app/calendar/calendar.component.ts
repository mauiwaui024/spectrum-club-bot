import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { ActivatedRoute, Router, RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { CalendarAPIResponse, CalendarDay, TrainingDetails } from '../models/training.model';
import { CalendarEvent, CalendarView, CalendarCommonModule, CalendarMonthModule, CalendarWeekModule, CalendarDayModule } from 'angular-calendar';
import { adapterFactory } from 'angular-calendar/date-adapters/date-fns';

@Component({
  selector: 'app-calendar',
  standalone: true,
  imports: [
    CommonModule, 
    RouterModule,
    CalendarCommonModule,
    CalendarMonthModule,
    CalendarWeekModule,
    CalendarDayModule
  ],
  templateUrl: './calendar.component.html',
  styleUrls: ['./calendar.component.css']
})
export class CalendarComponent implements OnInit {
  view: CalendarView = CalendarView.Month;
  viewDate: Date = new Date();
  events: CalendarEvent[] = [];
  
  calendarData: CalendarAPIResponse | null = null;
  loading: boolean = false;
  error: string | null = null;
  
  // User role
  isCoach: boolean = false;
  
  // Modal state
  showModal: boolean = false;
  selectedTraining: TrainingDetails | null = null;
  loadingTraining: boolean = false;
  
  // Attendance marking state (для тренеров)
  selectedStudents: Set<number> = new Set();
  isSelectingAttendance: boolean = false;

  readonly CalendarView = CalendarView;

  constructor(
    private calendarService: CalendarService,
    public route: ActivatedRoute,
    private router: Router
  ) {
  }

  // Получаем userID из разных источников (для обратной совместимости)
  // В идеале userID не должен использоваться на фронтенде, все запросы идут с initData в заголовках
  private getUserId(): string | null {
    // 1. Из query параметров (fallback для обратной совместимости)
    const queryUserId = this.route.snapshot.queryParams['user_id'];
    if (queryUserId) {
      return queryUserId;
    }

    // 2. Из localStorage (если был сохранен ранее через старый способ)
    const storedUserId = localStorage.getItem('user_id');
    if (storedUserId) {
      return storedUserId;
    }

    return null;
  }

  ngOnInit() {
    // Обновляем initData при инициализации компонента
    this.calendarService.updateInitData();
    
    // Проверяем наличие initData после небольшой задержки (для WebApp инициализации)
    setTimeout(() => {
      const tg = (window as any).Telegram?.WebApp;
      if (tg) {
        const hasInitData = tg.initData && tg.initData.length > 0;
        console.log('🔍 Проверка initData при загрузке календаря:', {
          hasInitData,
          initDataLength: tg.initData?.length || 0,
          platform: tg.platform || 'unknown',
          version: tg.version || 'unknown',
          ready: tg.ready || false
        });
        
        if (hasInitData) {
          console.log('✅ initData успешно получен через WebApp кнопку!');
        } else {
          console.warn('⚠️ initData пустой. Убедитесь, что страница открыта через WebApp кнопку в Telegram боте.');
        }
      }
    }, 1000);
    
    this.route.queryParams.subscribe(params => {
      const viewParam = params['view'] || 'month';
      const date = params['date'] || null;

      // Устанавливаем вид календаря
      if (viewParam === 'week') {
        this.view = CalendarView.Week;
      } else if (viewParam === 'day') {
        this.view = CalendarView.Day;
      } else {
        this.view = CalendarView.Month;
      }

      if (date) {
        this.viewDate = new Date(date);
      }

      const viewStr = this.view === CalendarView.Week ? 'week' : 
                     this.view === CalendarView.Day ? 'day' : 'month';
      // Пробуем получить user_id из query параметров для fallback (если initData нет)
      const userIdParam = params['user_id'] || null;
      this.loadCalendar(userIdParam, viewStr, date);
    });
  }

  loadCalendar(userId: string | null, view: string, date: string | null) {
    this.loading = true;
    this.error = null;

    this.calendarService.getCalendar(userId, view, date).subscribe({
      next: (data) => {
        this.calendarData = data;
        this.isCoach = data.is_coach || false;
        
        // Для недельного вида нужно установить viewDate на начало недели (понедельник)
        if (view === 'week' && data.start_date) {
          const startDateParts = data.start_date.split('-');
          const startYear = parseInt(startDateParts[0]);
          const startMonth = parseInt(startDateParts[1]) - 1;
          const startDay = parseInt(startDateParts[2]);
          this.viewDate = new Date(startYear, startMonth, startDay, 0, 0, 0, 0);
        } else if (view === 'day') {
          // Для дневного вида используем current_date
          const currentDateParts = data.current_date.split('-');
          const currentYear = parseInt(currentDateParts[0]);
          const currentMonth = parseInt(currentDateParts[1]) - 1;
          const currentDay = parseInt(currentDateParts[2]);
          this.viewDate = new Date(currentYear, currentMonth, currentDay, 0, 0, 0, 0);
        } else {
          // Для месячного вида используем current_date
          const currentDateParts = data.current_date.split('-');
          const currentYear = parseInt(currentDateParts[0]);
          const currentMonth = parseInt(currentDateParts[1]) - 1;
          const currentDay = parseInt(currentDateParts[2]);
          this.viewDate = new Date(currentYear, currentMonth, currentDay, 0, 0, 0, 0);
        }
        
        this.events = this.convertToCalendarEvents(data);
        this.loading = false;
      },
      error: (err) => {
        console.error('Calendar load error:', err);
        this.error = 'Ошибка загрузки календаря: ' + err.message;
        this.loading = false;
      }
    });
  }

  convertToCalendarEvents(data: CalendarAPIResponse): CalendarEvent[] {
    const events: CalendarEvent[] = [];
    
    // Для дневного вида используем events
    if (this.view === CalendarView.Day && data.events) {
      // Для дневного вида события уже содержат время в формате "HH:MM - HH:MM"
      // Используем current_date для определения даты
      const currentDateParts = data.current_date.split('-');
      const year = parseInt(currentDateParts[0]);
      const month = parseInt(currentDateParts[1]) - 1;
      const dayOfMonth = parseInt(currentDateParts[2]);
      
      data.events.forEach((event: any) => {
        // Парсим время из формата "HH:MM - HH:MM"
        const timeRangeMatch = event.time.match(/(\d{2}):(\d{2})\s*-\s*(\d{2}):(\d{2})/);
        
        if (timeRangeMatch) {
          const startHour = parseInt(timeRangeMatch[1]);
          const startMinute = parseInt(timeRangeMatch[2]);
          const endHour = parseInt(timeRangeMatch[3]);
          const endMinute = parseInt(timeRangeMatch[4]);
          
          // Создаем даты начала и окончания в локальном времени
          const startDate = new Date(year, month, dayOfMonth, startHour, startMinute, 0, 0);
          const endDate = new Date(year, month, dayOfMonth, endHour, endMinute, 0, 0);
          
          const calendarEvent: CalendarEvent = {
            start: startDate,
            end: endDate,
            title: event.title,
            color: this.getEventColor(event.color_index),
            meta: {
              id: event.id,
              color_index: event.color_index,
              coach: event.coach,
              user_id: event.user_id,
              time: event.time
            }
          };
          
          events.push(calendarEvent);
        } else {
          // Fallback для формата без диапазона (только начало)
          const timeMatch = event.time.match(/(\d{2}):(\d{2})/);
          if (timeMatch) {
            const startHour = parseInt(timeMatch[1]);
            const startMinute = parseInt(timeMatch[2]);
            const startDate = new Date(year, month, dayOfMonth, startHour, startMinute, 0, 0);
            
            const calendarEvent: CalendarEvent = {
              start: startDate,
              title: event.title,
              color: this.getEventColor(event.color_index),
              meta: {
                id: event.id,
                color_index: event.color_index,
                coach: event.coach,
                user_id: event.user_id,
                time: event.time
              }
            };
            
            events.push(calendarEvent);
          }
        }
      });
    }
    // Для недельного вида используем week_days_data
    else if (this.view === CalendarView.Week && data.week_days_data) {
      data.week_days_data.forEach((day: any) => {
        // Обрабатываем случай, когда events равен null - преобразуем в пустой массив
        let dayEvents = day.events;
        if (!dayEvents || dayEvents === null) {
          dayEvents = [];
        }
        
        // Проверяем, что dayEvents это массив
        if (!Array.isArray(dayEvents) || dayEvents.length === 0) {
          return;
        }
        
        // Обрабатываем события
        dayEvents.forEach((event: any) => {
          // Парсим дату в формате "YYYY-MM-DD" и создаем дату в локальном времени
          const dateParts = day.date.split('-');
          const year = parseInt(dateParts[0]);
          const month = parseInt(dateParts[1]) - 1; // месяцы в JS начинаются с 0
          const dayOfMonth = parseInt(dateParts[2]);
          
          // Парсим время из формата "HH:MM - HH:MM"
          const timeRangeMatch = event.time.match(/(\d{2}):(\d{2})\s*-\s*(\d{2}):(\d{2})/);
          
          if (timeRangeMatch) {
            const startHour = parseInt(timeRangeMatch[1]);
            const startMinute = parseInt(timeRangeMatch[2]);
            const endHour = parseInt(timeRangeMatch[3]);
            const endMinute = parseInt(timeRangeMatch[4]);
            
            // Создаем даты начала и окончания в локальном времени
            // Используем прямой конструктор Date для гарантии локального времени
            const startDate = new Date(year, month, dayOfMonth, startHour, startMinute, 0, 0);
            const endDate = new Date(year, month, dayOfMonth, endHour, endMinute, 0, 0);
            
            
            const calendarEvent: CalendarEvent = {
              start: startDate,
              end: endDate,
              title: event.title,
              color: this.getEventColor(event.color_index),
              meta: {
                id: event.id,
                color_index: event.color_index,
                coach: event.coach,
                user_id: event.user_id,
                time: event.time
              }
            };
            
            events.push(calendarEvent);
          } else {
            // Fallback для формата без диапазона (только начало)
            const timeMatch = event.time.match(/(\d{2}):(\d{2})/);
            if (timeMatch) {
              const startHour = parseInt(timeMatch[1]);
              const startMinute = parseInt(timeMatch[2]);
              const startDate = new Date(year, month, dayOfMonth, startHour, startMinute, 0, 0);
              
              const calendarEvent: CalendarEvent = {
                start: startDate,
                title: event.title,
                color: this.getEventColor(event.color_index),
                meta: {
                  id: event.id,
                  color_index: event.color_index,
                  coach: event.coach,
                  user_id: event.user_id,
                  time: event.time
                }
              };
              
              events.push(calendarEvent);
            }
          }
        });
      });
    } 
    // Для месячного вида используем calendar_days
    else if (this.view === CalendarView.Month && data.calendar_days) {
      data.calendar_days.forEach(day => {
        if (!day.events || day.events === null || day.events.length === 0) return;
        
        day.events.forEach(event => {
          const eventDate = new Date(day.date);
          // Парсим время из формата "HH:MM - HH:MM" или "HH:MM"
          const timeMatch = event.time.match(/(\d{2}):(\d{2})/);
          if (timeMatch) {
            eventDate.setHours(parseInt(timeMatch[1]), parseInt(timeMatch[2]), 0, 0);
          }
          
          const calendarEvent: CalendarEvent = {
            start: eventDate,
            title: event.title,
            color: this.getEventColor(event.color_index),
            meta: {
              id: event.id,
              color_index: event.color_index,
              coach: event.coach,
              user_id: event.user_id,
              time: event.time
            }
          };
          
          events.push(calendarEvent);
        });
      });
    }
    
    return events;
  }

  getEventColor(colorIndex: number): any {
    const colors = [
      { primary: '#6366f1', secondary: '#eef2ff' }, // Индиго - мягкий синий
      { primary: '#8b5cf6', secondary: '#f3e8ff' }, // Фиолетовый - мягкий
      { primary: '#ec4899', secondary: '#fce7f3' }, // Розовый - мягкий
      { primary: '#14b8a6', secondary: '#ccfbf1' }, // Бирюзовый - мягкий
      { primary: '#f59e0b', secondary: '#fef3c7' }, // Янтарный - мягкий
      { primary: '#06b6d4', secondary: '#cffafe' }, // Голубой - мягкий
      { primary: '#a855f7', secondary: '#f3e8ff' }, // Фиолетовый - светлый
      { primary: '#10b981', secondary: '#d1fae5' }  // Зеленый - мягкий
    ];
    return colors[colorIndex % colors.length] || colors[0];
  }

  onEventClicked(event: CalendarEvent) {
    const trainingId = (event.meta as any)?.id;
    if (trainingId) {
      this.showEventDetails(trainingId);
    }
  }

  onViewDateChange(date: Date) {
    this.viewDate = date;
    const dateStr = date.toISOString().split('T')[0];
    const viewStr = this.view === CalendarView.Week ? 'week' : 
                   this.view === CalendarView.Day ? 'day' : 'month';
    
    // user_id больше не передаем в URL для безопасности
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: {
        view: viewStr,
        date: dateStr
      },
      queryParamsHandling: 'merge'
    });
  }

  changeView(newView: CalendarView) {
    if (this.view === newView) return;
    
    this.view = newView;
    const dateStr = this.viewDate.toISOString().split('T')[0];
    const viewStr = newView === CalendarView.Week ? 'week' : 
                   newView === CalendarView.Day ? 'day' : 'month';
    
    // user_id больше не передаем в URL для безопасности
    this.router.navigate([], {
      relativeTo: this.route,
      queryParams: {
        view: viewStr,
        date: dateStr
      },
      queryParamsHandling: 'merge'
    });
  }

  showEventDetails(trainingId: number) {
    this.loadingTraining = true;
    this.showModal = true;
    // user_id больше не передаем, используется initData из заголовков

    this.calendarService.getTrainingDetails(trainingId, null).subscribe({
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
    this.isSelectingAttendance = false;
    this.selectedStudents.clear();
  }
  
  // Переключение режима выбора посещаемости
  toggleAttendanceSelection() {
    this.isSelectingAttendance = !this.isSelectingAttendance;
    if (this.isSelectingAttendance && this.selectedTraining) {
      // При включении режима выбираем только тех участников, у которых еще не отмечена посещаемость
      this.selectedStudents.clear();
      this.selectedTraining.participants.forEach(participant => {
        const isAlreadyAttended = participant.attended || participant.status === 'attended';
        if (!isAlreadyAttended) {
          this.selectedStudents.add(participant.student_id);
        }
      });
    } else {
      // При выключении очищаем выбор
      this.selectedStudents.clear();
    }
  }

  // Проверка, отмечена ли уже посещаемость для участника
  isAlreadyAttended(participant: any): boolean {
    return participant.attended === true || participant.status === 'attended';
  }
  
  // Переключение выбора конкретного ученика
  toggleStudentSelection(studentId: number) {
    if (this.selectedStudents.has(studentId)) {
      this.selectedStudents.delete(studentId);
    } else {
      this.selectedStudents.add(studentId);
    }
  }
  
  // Проверка, выбран ли ученик
  isStudentSelected(studentId: number): boolean {
    return this.selectedStudents.has(studentId);
  }
  
  // Подтверждение посещаемости
  confirmAttendance() {
    if (!this.selectedTraining) return;
    
    // Проверяем, что выбранные ученики еще не имеют отмеченной посещаемости
    const alreadyAttendedStudents: string[] = [];
    this.selectedStudents.forEach(studentId => {
      const participant = this.selectedTraining?.participants.find(p => p.student_id === studentId);
      if (participant && this.isAlreadyAttended(participant)) {
        alreadyAttendedStudents.push(participant.student_name || `Ученик ${studentId}`);
      }
    });

    if (alreadyAttendedStudents.length > 0) {
      alert('Невозможно отметить посещаемость для следующих учеников, так как она уже была отмечена:\n' + alreadyAttendedStudents.join('\n'));
      // Удаляем уже отмеченных из выбранных
      alreadyAttendedStudents.forEach(() => {
        // Находим и удаляем из selectedStudents
        this.selectedTraining?.participants.forEach(p => {
          if (this.isAlreadyAttended(p)) {
            this.selectedStudents.delete(p.student_id);
          }
        });
      });
      if (this.selectedStudents.size === 0) {
        return;
      }
    }
    
    if (this.selectedStudents.size === 0) {
      alert('Выберите хотя бы одного ученика, у которого еще не отмечена посещаемость');
      return;
    }
    
    if (!confirm(`Подтвердить посещаемость для ${this.selectedStudents.size} учеников?`)) {
      return;
    }
    
    const studentIds = Array.from(this.selectedStudents);
    this.calendarService.markAttendance(this.selectedTraining.training.id, studentIds).subscribe({
      next: (response: any) => {
        // Проверяем, есть ли ошибки в ответе
        if (response.success === false || (response.failed_count && response.failed_count > 0)) {
          // Частичный успех или полная ошибка
          let message = '';
          if (response.message) {
            message = response.message;
          } else if (response.failed_count > 0) {
            message = `Посещаемость подтверждена для ${response.marked_count || 0} из ${response.total_count || studentIds.length} учеников`;
          } else {
            message = 'Не удалось подтвердить посещаемость';
          }
          
          // Добавляем детали об ошибках
          if (response.errors && response.errors.length > 0) {
            const errorsList = response.errors.join('\n');
            message += '\n\nОшибки:\n' + errorsList;
          } else if (response.failed_students && response.failed_students.length > 0) {
            const failedNames = response.failed_students.map((s: any) => 
              `${s.student_name || 'Ученик ' + s.student_id}: ${s.error || 'Неизвестная ошибка'}`
            ).join('\n');
            message += '\n\nОшибки:\n' + failedNames;
          }
          
          alert('⚠️ ' + message);
          
          // Если хотя бы часть учеников отмечена, обновляем календарь
          if (response.marked_count > 0) {
            this.closeModal();
            this.reloadCalendar();
          }
        } else {
          // Полный успех
          alert(`✅ Посещаемость подтверждена для ${response.marked_count || studentIds.length} учеников`);
          this.closeModal();
          this.reloadCalendar();
        }
      },
      error: (err) => {
        let errorMessage = 'Не удалось подтвердить посещаемость';
        
        // Пытаемся извлечь детальную информацию об ошибке
        if (err.error) {
          if (typeof err.error === 'string') {
            errorMessage = err.error;
          } else if (err.error.message) {
            errorMessage = err.error.message;
          } else if (err.error.error) {
            errorMessage = err.error.error;
          }
          
          // Если есть информация о неудачных учениках
          if (err.error.failed_students && err.error.failed_students.length > 0) {
            const failedNames = err.error.failed_students.map((s: any) => 
              `${s.student_name || 'Ученик ' + s.student_id}: ${s.error || 'Неизвестная ошибка'}`
            ).join('\n');
            errorMessage += '\n\nДетали:\n' + failedNames;
          } else if (err.error.errors && err.error.errors.length > 0) {
            errorMessage += '\n\nДетали:\n' + err.error.errors.join('\n');
          }
        } else if (err.message) {
          errorMessage = err.message;
        }
        
        console.error('Ошибка при подтверждении посещаемости:', err);
        alert('❌ Ошибка: ' + errorMessage);
      }
    });
  }

  registerForTraining() {
    if (!this.selectedTraining) return;

    // Обновляем initData перед запросом (на случай, если он появился позже)
    this.calendarService.updateInitData();
    
    // Проверяем наличие initData
    const tg = (window as any).Telegram?.WebApp;
    const hasInitData = tg && tg.initData && tg.initData.length > 0;
    
    console.log('🔍 Проверка перед регистрацией:', {
      telegramWebAppAvailable: !!tg,
      initDataAvailable: hasInitData,
      initDataLength: tg?.initData?.length || 0,
      platform: tg?.platform || 'unknown'
    });
    
    if (!hasInitData) {
      console.error('❌ initData пустой! Невозможно записаться без авторизации.');
      console.error('💡 Убедитесь, что:');
      console.error('   1. Страница открыта через Telegram WebApp (кнопка в боте)');
      console.error('   2. URL использует HTTPS (не localhost)');
      console.error('   3. Telegram WebApp инициализирован');
      
      alert('Ошибка: Необходимо войти в систему.\n\nПожалуйста, откройте календарь через Telegram бота (кнопка "📅 Открыть календарь").\n\nЕсли проблема сохраняется, убедитесь, что вы открыли страницу через WebApp, а не через обычный браузер.');
      return;
    }

    if (!confirm('Записаться на эту тренировку?')) return;

    // Получаем user_id для fallback (если initData нет)
    const userId = this.getUserId();
    this.calendarService.registerForTraining(this.selectedTraining.training.id, userId || '').subscribe({
      next: () => {
        alert('✅ Вы успешно записались на тренировку!');
        this.closeModal();
        this.reloadCalendar();
      },
      error: (err) => {
        // Для текстовых ответов (responseType: 'text') ошибка может быть в err.error как строка
        let errorMessage = 'Не удалось записаться';
        if (err.error) {
          // Если err.error - строка (текстовый ответ)
          if (typeof err.error === 'string') {
            errorMessage = err.error;
          } else if (err.error.message) {
            // Если err.error - объект с message
            errorMessage = err.error.message;
          }
        } else if (err.message) {
          errorMessage = err.message;
        }
        alert('❌ Ошибка: ' + errorMessage);
      }
    });
  }

  cancelRegistration() {
    if (!this.selectedTraining) return;

    // Обновляем initData перед запросом (на случай, если он появился позже)
    this.calendarService.updateInitData();

    if (!confirm('Вы уверены, что хотите отменить запись?')) return;

    // Получаем user_id для fallback (если initData нет)
    const userId = this.getUserId();
    this.calendarService.cancelRegistration(this.selectedTraining.training.id, userId || '').subscribe({
      next: () => {
        alert('✅ Запись отменена');
        this.closeModal();
        this.reloadCalendar();
      },
      error: (err) => {
        // Для текстовых ответов (responseType: 'text') ошибка может быть в err.error как строка
        let errorMessage = 'Не удалось отменить запись';
        if (err.error) {
          // Если err.error - строка (текстовый ответ)
          if (typeof err.error === 'string') {
            errorMessage = err.error;
          } else if (err.error.message) {
            // Если err.error - объект с message
            errorMessage = err.error.message;
          }
        } else if (err.message) {
          errorMessage = err.message;
        }
        alert('❌ Ошибка: ' + errorMessage);
      }
    });
  }

  reloadCalendar() {
    // user_id больше не передаем, используется initData из заголовков
    const dateStr = this.viewDate.toISOString().split('T')[0];
    const viewStr = this.view === CalendarView.Week ? 'week' : 
                   this.view === CalendarView.Day ? 'day' : 'month';
    this.loadCalendar(null, viewStr, dateStr);
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
    const options: Intl.DateTimeFormatOptions = { month: 'long', year: 'numeric' };
    return this.viewDate.toLocaleDateString('ru-RU', options);
  }

  getDateRange(): string {
    if (!this.calendarData) return '';
    
    if (this.view === CalendarView.Day) {
      // Для дня показываем только одну дату
      const date = new Date(this.calendarData.current_date);
      return date.toLocaleDateString('ru-RU', { weekday: 'long', day: 'numeric', month: 'long', year: 'numeric' });
    } else if (this.view === CalendarView.Week) {
      // Для недели показываем даты начала и конца недели
      const start = new Date(this.calendarData.start_date);
      const end = new Date(this.calendarData.end_date);
      const startStr = start.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
      const endStr = end.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
      return `${startStr} – ${endStr}`;
    } else {
      // Для месяца показываем диапазон месяца
      const start = new Date(this.calendarData.start_date);
      const end = new Date(this.calendarData.end_date);
      const startStr = start.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long' });
      const endStr = end.toLocaleDateString('ru-RU', { day: 'numeric', month: 'long', year: 'numeric' });
      return `${startStr} – ${endStr}`;
    }
  }
}
