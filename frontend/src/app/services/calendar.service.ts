import { Injectable } from '@angular/core';
import { HttpClient, HttpParams } from '@angular/common/http';
import { Observable } from 'rxjs';
import { CalendarAPIResponse, TrainingDetails } from '../models/training.model';

@Injectable({
  providedIn: 'root'
})
export class CalendarService {
  private apiUrl = '/api';

  constructor(private http: HttpClient) {}

  getCalendar(
    userId: string | null,
    view: string = 'month',
    date: string | null = null
  ): Observable<CalendarAPIResponse> {
    let params = new HttpParams();
    if (userId) {
      params = params.set('user_id', userId);
    }
    params = params.set('view', view);
    if (date) {
      params = params.set('date', date);
    }

    return this.http.get<CalendarAPIResponse>(`${this.apiUrl}/calendar`, { params });
  }

  getTrainingDetails(trainingId: number, userId: string | null): Observable<TrainingDetails> {
    let params = new HttpParams();
    if (userId) {
      params = params.set('user_id', userId);
    }

    return this.http.get<TrainingDetails>(`${this.apiUrl}/training/${trainingId}`, { params });
  }

  registerForTraining(trainingId: number, userId: string): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());
    formData.append('user_id', userId);

    return this.http.post(`${this.apiUrl}/register`, formData, { responseType: 'text' });
  }

  cancelRegistration(trainingId: number, userId: string): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());
    formData.append('user_id', userId);

    return this.http.post(`${this.apiUrl}/cancel`, formData, { responseType: 'text' });
  }

  checkRegistration(trainingId: number, userId: string): Observable<any> {
    let params = new HttpParams();
    params = params.set('training_id', trainingId.toString());
    params = params.set('user_id', userId);

    return this.http.get(`${this.apiUrl}/check-registration`, { params });
  }
}
