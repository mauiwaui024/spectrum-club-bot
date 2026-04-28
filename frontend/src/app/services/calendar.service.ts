import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CalendarAPIResponse, TrainingDetails } from '../models/training.model';
import { MyRegistrationsResponse, MySubscriptionResponse, MyProfileResponse, AllStudentsSubscriptionsResponse, UpdateProfileRequest, Subscription } from '../models/student.model';

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

  private getRequestOptions(params?: HttpParams) {
    return {
      params,
      headers: this.getHeaders(),
      withCredentials: true
    };
  }

  getCalendar(
    view: string = 'month',
    date: string | null = null
  ): Observable<CalendarAPIResponse> {
    let params = new HttpParams();
    params = params.set('view', view);
    if (date) {
      params = params.set('date', date);
    }

    return this.http.get<CalendarAPIResponse>(`${this.apiUrl}/calendar`, this.getRequestOptions(params));
  }

  getTrainingDetails(trainingId: number): Observable<TrainingDetails> {
    return this.http.get<TrainingDetails>(`${this.apiUrl}/training/${trainingId}`, this.getRequestOptions());
  }

  registerForTraining(trainingId: number): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());

    return this.http.post(`${this.apiUrl}/register`, formData, { 
      responseType: 'text',
      headers: this.getHeaders(),
      withCredentials: true
    });
  }

  cancelRegistration(trainingId: number): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());

    return this.http.post(`${this.apiUrl}/cancel`, formData, { 
      responseType: 'text',
      headers: this.getHeaders(),
      withCredentials: true
    });
  }

  checkRegistration(trainingId: number, userId: string): Observable<any> {
    let params = new HttpParams();
    params = params.set('training_id', trainingId.toString());
    // user_id больше не передаем в URL для безопасности
    // Используется initData из заголовков

    return this.http.get(`${this.apiUrl}/check-registration`, this.getRequestOptions(params));
  }

  // Получить userID через API используя initData
  getUserId(): Observable<{user_id: number}> {
    return this.http.post<{user_id: number}>(`${this.apiUrl}/auth`, null, {
      headers: this.getHeaders(),
      withCredentials: true
    });
  }

  getAuthStatus(): Observable<any> {
    return this.http.get(`${this.apiUrl}/auth/status`, this.getRequestOptions());
  }

  loginWithBrowserCredentials(username: string, password: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/browser-auth/login`, { username, password }, {
      headers: this.getHeaders(),
      withCredentials: true
    });
  }

  logoutBrowserSession(): Observable<any> {
    return this.http.post(`${this.apiUrl}/browser-auth/logout`, {}, {
      headers: this.getHeaders(),
      withCredentials: true
    });
  }

  setBrowserPassword(password: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/browser-auth/set-password`, { password }, {
      headers: this.getHeaders(),
      withCredentials: true
    });
  }

  // Подтвердить посещаемость тренировки (для тренеров)
  markAttendance(trainingId: number, studentIds: number[]): Observable<any> {
    const body = {
      training_id: trainingId,
      student_ids: studentIds
    };
    return this.http.post(`${this.apiUrl}/mark-attendance`, body, this.getRequestOptions());
  }

  // Получить список записей ученика на тренировки
  getMyRegistrations(): Observable<MyRegistrationsResponse> {
    return this.http.get<MyRegistrationsResponse>(`${this.apiUrl}/my-registrations`, this.getRequestOptions());
  }

  // Получить информацию об абонементе ученика
  getMySubscription(): Observable<MySubscriptionResponse> {
    return this.http.get<MySubscriptionResponse>(`${this.apiUrl}/my-subscription`, this.getRequestOptions());
  }

  // Получить профиль ученика
  getMyProfile(): Observable<MyProfileResponse> {
    return this.http.get<MyProfileResponse>(`${this.apiUrl}/my-profile`, this.getRequestOptions());
  }

  // Получить все абонементы всех учеников (для тренеров)
  getAllStudentsSubscriptions(): Observable<AllStudentsSubscriptionsResponse> {
    return this.http.get<AllStudentsSubscriptionsResponse>(`${this.apiUrl}/students-subscriptions`, this.getRequestOptions());
  }

  // Обновить профиль пользователя
  updateProfile(data: UpdateProfileRequest): Observable<MyProfileResponse> {
    return this.http.put<MyProfileResponse>(`${this.apiUrl}/update-profile`, data, this.getRequestOptions());
  }

  // Добавить занятия к абонементу (для тренеров)
  addLessons(subscriptionId: number, count: number): Observable<Subscription> {
    return this.http.post<Subscription>(`${this.apiUrl}/subscription/add-lessons`, {
      subscription_id: subscriptionId,
      count: count
    }, this.getRequestOptions());
  }

  // Снять занятия с абонемента (для тренеров)
  removeLessons(subscriptionId: number, count: number): Observable<Subscription> {
    return this.http.post<Subscription>(`${this.apiUrl}/subscription/remove-lessons`, {
      subscription_id: subscriptionId,
      count: count
    }, this.getRequestOptions());
  }
}
