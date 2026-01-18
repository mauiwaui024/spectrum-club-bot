import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CalendarAPIResponse, TrainingDetails } from '../models/training.model';
import { MyRegistrationsResponse, MySubscriptionResponse, MyProfileResponse, AllStudentsSubscriptionsResponse } from '../models/student.model';

@Injectable({
  providedIn: 'root'
})
export class CalendarService {
  private apiUrl = '/api';
  private initData: string | null = null;

  constructor(private http: HttpClient) {
    // Получаем initData из Telegram WebApp при инициализации
    this.updateInitData();
  }

  // Обновляем initData (можно вызывать после инициализации Telegram WebApp)
  updateInitData(): void {
    try {
      const tg = (window as any).Telegram?.WebApp;
      if (tg) {
        // Пробуем получить initData (может быть пустым, если не в Telegram)
        // Важно: initData может появиться не сразу, поэтому проверяем несколько раз
        this.initData = tg.initData || null;
        
        // Если initData пустой, но WebApp готов, пробуем подождать и проверить еще раз
        if (!this.initData && tg.ready) {
          // Даем Telegram WebApp время на инициализацию
          setTimeout(() => {
            this.initData = tg.initData || null;
            if (this.initData) {
              console.log('✅ initData получен после задержки:', this.initData.substring(0, 50) + '...');
            }
          }, 500);
        }
        
        // Логирование для отладки
        console.log('🔍 Telegram WebApp Debug:', {
          available: true,
          initData: this.initData ? `${this.initData.substring(0, 50)}...` : 'null',
          initDataLength: this.initData ? this.initData.length : 0,
          initDataUnsafe: tg.initDataUnsafe || null,
          ready: tg.ready || false,
          version: tg.version || 'unknown',
          platform: tg.platform || 'unknown'
        });
        
        // Если initData пустой, но WebApp доступен, возможно нужно подождать инициализации
        if (!this.initData && tg.ready) {
          // WebApp готов, но initData пустой - это нормально для тестирования вне Telegram
          console.warn('⚠️ Telegram WebApp доступен, но initData пустой. Проверьте, что страница открыта через Telegram WebApp (не через обычный браузер)');
        }
      } else {
        console.warn('⚠️ Telegram WebApp не доступен (открыто не через Telegram)');
      }
    } catch (e) {
      // Telegram WebApp не доступен
      console.error('❌ Telegram WebApp ошибка:', e);
    }
  }

  private getHeaders(): HttpHeaders {
    // Обновляем initData перед каждым запросом (на случай, если он появился позже)
    this.updateInitData();
    
    // Дополнительная проверка: получаем initData напрямую из Telegram WebApp
    let currentInitData = this.initData;
    try {
      const tg = (window as any).Telegram?.WebApp;
      if (tg && tg.initData) {
        currentInitData = tg.initData;
        // Обновляем кэш
        this.initData = currentInitData;
      }
    } catch (e) {
      // ignore
    }
    
    const headers = new HttpHeaders();
    if (currentInitData && currentInitData.length > 0) {
      console.log('📤 Отправка запроса с initData (длина:', currentInitData.length + ')');
      console.log('📤 initData (первые 100 символов):', currentInitData.substring(0, 100));
      return headers.set('X-Telegram-Init-Data', currentInitData);
    } else {
      console.warn('⚠️ Отправка запроса БЕЗ initData!');
      console.warn('⚠️ Проверка Telegram WebApp:', {
        available: !!(window as any).Telegram?.WebApp,
        initData: (window as any).Telegram?.WebApp?.initData || 'null',
        initDataLength: (window as any).Telegram?.WebApp?.initData?.length || 0,
        ready: (window as any).Telegram?.WebApp?.ready || false
      });
    }
    return headers;
  }

  getCalendar(
    userId: string | null,
    view: string = 'month',
    date: string | null = null
  ): Observable<CalendarAPIResponse> {
    let params = new HttpParams();
    // Если initData нет, используем fallback на user_id в query параметре (для обратной совместимости)
    if (!this.initData && userId) {
      params = params.set('user_id', userId);
    }
    params = params.set('view', view);
    if (date) {
      params = params.set('date', date);
    }

    return this.http.get<CalendarAPIResponse>(`${this.apiUrl}/calendar`, { 
      params,
      headers: this.getHeaders()
    });
  }

  getTrainingDetails(trainingId: number, userId: string | null): Observable<TrainingDetails> {
    // user_id больше не передаем в URL для безопасности
    // Используется initData из заголовков
    return this.http.get<TrainingDetails>(`${this.apiUrl}/training/${trainingId}`, { 
      headers: this.getHeaders()
    });
  }

  registerForTraining(trainingId: number, userId: string): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());
    // Если initData нет, используем fallback на user_id в форме (для обратной совместимости)
    if (!this.initData && userId) {
      formData.append('user_id', userId);
    }

    return this.http.post(`${this.apiUrl}/register`, formData, { 
      responseType: 'text',
      headers: this.getHeaders()
    });
  }

  cancelRegistration(trainingId: number, userId: string): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());
    // Если initData нет, используем fallback на user_id в форме (для обратной совместимости)
    if (!this.initData && userId) {
      formData.append('user_id', userId);
    }

    return this.http.post(`${this.apiUrl}/cancel`, formData, { 
      responseType: 'text',
      headers: this.getHeaders()
    });
  }

  checkRegistration(trainingId: number, userId: string): Observable<any> {
    let params = new HttpParams();
    params = params.set('training_id', trainingId.toString());
    // user_id больше не передаем в URL для безопасности
    // Используется initData из заголовков

    return this.http.get(`${this.apiUrl}/check-registration`, { 
      params,
      headers: this.getHeaders()
    });
  }

  // Получить userID через API используя initData
  getUserId(): Observable<{user_id: number}> {
    return this.http.post<{user_id: number}>(`${this.apiUrl}/auth`, null, {
      headers: this.getHeaders()
    });
  }

  // Подтвердить посещаемость тренировки (для тренеров)
  markAttendance(trainingId: number, studentIds: number[]): Observable<any> {
    const body = {
      training_id: trainingId,
      student_ids: studentIds
    };
    return this.http.post(`${this.apiUrl}/mark-attendance`, body, {
      headers: this.getHeaders()
    });
  }

  // Получить список записей ученика на тренировки
  getMyRegistrations(): Observable<MyRegistrationsResponse> {
    return this.http.get<MyRegistrationsResponse>(`${this.apiUrl}/my-registrations`, {
      headers: this.getHeaders()
    });
  }

  // Получить информацию об абонементе ученика
  getMySubscription(): Observable<MySubscriptionResponse> {
    return this.http.get<MySubscriptionResponse>(`${this.apiUrl}/my-subscription`, {
      headers: this.getHeaders()
    });
  }

  // Получить профиль ученика
  getMyProfile(): Observable<MyProfileResponse> {
    return this.http.get<MyProfileResponse>(`${this.apiUrl}/my-profile`, {
      headers: this.getHeaders()
    });
  }

  // Получить все абонементы всех учеников (для тренеров)
  getAllStudentsSubscriptions(): Observable<AllStudentsSubscriptionsResponse> {
    return this.http.get<AllStudentsSubscriptionsResponse>(`${this.apiUrl}/students-subscriptions`, {
      headers: this.getHeaders()
    });
  }
}
