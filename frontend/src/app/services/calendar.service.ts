import { Injectable } from '@angular/core';
import { HttpClient, HttpHeaders, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CalendarAPIResponse, TrainingDetails } from '../models/training.model';

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
        this.initData = tg.initData || null;
        
        // Если initData пустой, но WebApp доступен, возможно нужно подождать инициализации
        if (!this.initData && tg.ready) {
          // WebApp готов, но initData пустой - это нормально для тестирования вне Telegram
          console.log('Telegram WebApp доступен, но initData пустой');
        }
      }
    } catch (e) {
      // Telegram WebApp не доступен
      console.log('Telegram WebApp не доступен:', e);
    }
  }

  private getHeaders(): HttpHeaders {
    const headers = new HttpHeaders();
    if (this.initData) {
      return headers.set('X-Telegram-Init-Data', this.initData);
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
}
