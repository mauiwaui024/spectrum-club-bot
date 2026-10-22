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

  constructor(private http: HttpClient) {}

  private getRequestOptions(params?: HttpParams) {
    return {
      params,
      headers: new HttpHeaders(),
      withCredentials: true
    };
  }

  getCalendar(view: string = 'month', date: string | null = null): Observable<CalendarAPIResponse> {
    let params = new HttpParams().set('view', view);
    if (date) {
      params = params.set('date', date);
    }
    return this.http.get<CalendarAPIResponse>(`${this.apiUrl}/calendar`, this.getRequestOptions(params));
  }

  getTrainingDetails(trainingId: number): Observable<TrainingDetails> {
    return this.http.get<TrainingDetails>(`${this.apiUrl}/training/${trainingId}`, this.getRequestOptions());
  }

  registerStudentByCoach(trainingId: number, studentId: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/trainings/register-student`, {
      training_id: trainingId,
      student_id: studentId
    }, this.getRequestOptions());
  }

  registerForTraining(trainingId: number): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());
    return this.http.post(`${this.apiUrl}/register`, formData, {
      responseType: 'text',
      withCredentials: true
    });
  }

  cancelRegistration(trainingId: number): Observable<any> {
    const formData = new FormData();
    formData.append('training_id', trainingId.toString());
    return this.http.post(`${this.apiUrl}/cancel`, formData, {
      responseType: 'text',
      withCredentials: true
    });
  }

  getAuthStatus(): Observable<any> {
    return this.http.get(`${this.apiUrl}/auth/status`, this.getRequestOptions());
  }

  loginWithBrowserCredentials(username: string, password: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/browser-auth/login`, { username, password }, {
      withCredentials: true
    });
  }

  logoutBrowserSession(): Observable<any> {
    return this.http.post(`${this.apiUrl}/browser-auth/logout`, {}, {
      withCredentials: true
    });
  }

  setBrowserPassword(password: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/browser-auth/set-password`, { password }, {
      withCredentials: true
    });
  }

  markAttendance(trainingId: number, studentIds: number[]): Observable<any> {
    return this.http.post(`${this.apiUrl}/mark-attendance`, {
      training_id: trainingId,
      student_ids: studentIds
    }, this.getRequestOptions());
  }

  getMyRegistrations(): Observable<MyRegistrationsResponse> {
    return this.http.get<MyRegistrationsResponse>(`${this.apiUrl}/my-registrations`, this.getRequestOptions());
  }

  getMySubscription(): Observable<MySubscriptionResponse> {
    return this.http.get<MySubscriptionResponse>(`${this.apiUrl}/my-subscription`, this.getRequestOptions());
  }

  getMyProfile(): Observable<MyProfileResponse> {
    return this.http.get<MyProfileResponse>(`${this.apiUrl}/my-profile`, this.getRequestOptions());
  }

  getAllStudentsSubscriptions(): Observable<AllStudentsSubscriptionsResponse> {
    return this.http.get<AllStudentsSubscriptionsResponse>(`${this.apiUrl}/students-subscriptions`, this.getRequestOptions());
  }

  updateProfile(data: UpdateProfileRequest): Observable<MyProfileResponse> {
    return this.http.put<MyProfileResponse>(`${this.apiUrl}/update-profile`, data, this.getRequestOptions());
  }

  addLessons(subscriptionId: number, count: number): Observable<Subscription> {
    return this.http.post<Subscription>(`${this.apiUrl}/subscription/add-lessons`, {
      subscription_id: subscriptionId,
      count
    }, this.getRequestOptions());
  }

  removeLessons(subscriptionId: number, count: number): Observable<Subscription> {
    return this.http.post<Subscription>(`${this.apiUrl}/subscription/remove-lessons`, {
      subscription_id: subscriptionId,
      count
    }, this.getRequestOptions());
  }

  setSubscriptionEndDate(subscriptionId: number, endDate: string): Observable<Subscription> {
    return this.http.post<Subscription>(`${this.apiUrl}/subscription/end-date`, {
      subscription_id: subscriptionId,
      end_date: endDate
    }, this.getRequestOptions());
  }

  getGroups(): Observable<{ groups: any[] }> {
    return this.http.get<{ groups: any[] }>(`${this.apiUrl}/groups`, this.getRequestOptions());
  }

  getCoachSchedule(start?: string, end?: string): Observable<{ trainings: any[] }> {
    let params = new HttpParams();
    if (start) params = params.set('start', start);
    if (end) params = params.set('end', end);
    return this.http.get<{ trainings: any[] }>(`${this.apiUrl}/coach/schedule`, this.getRequestOptions(params));
  }

  createTraining(body: {
    group_id: number;
    date: string;
    start_time: string;
    end_time: string;
    description?: string;
  }): Observable<any> {
    return this.http.post(`${this.apiUrl}/trainings`, body, this.getRequestOptions());
  }

  updateTraining(id: number, body: any): Observable<any> {
    return this.http.patch(`${this.apiUrl}/trainings/${id}`, body, this.getRequestOptions());
  }

  deleteTraining(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/trainings/${id}`, this.getRequestOptions());
  }

  createFromTemplates(weeksCount: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/trainings/from-templates`, { weeks_count: weeksCount }, this.getRequestOptions());
  }

  getStudents(filter = 'all', days = 14, q = ''): Observable<{ students: any[] }> {
    let params = new HttpParams().set('filter', filter).set('days', String(days));
    if (q) params = params.set('q', q);
    return this.http.get<{ students: any[] }>(`${this.apiUrl}/students`, this.getRequestOptions(params));
  }

  createGroup(body: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/groups`, body, this.getRequestOptions());
  }

  updateGroup(id: number, body: any): Observable<any> {
    return this.http.patch(`${this.apiUrl}/groups/${id}`, body, this.getRequestOptions());
  }

  deleteGroup(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/groups/${id}`, this.getRequestOptions());
  }

  getTemplates(): Observable<{ templates: any[] }> {
    return this.http.get<{ templates: any[] }>(`${this.apiUrl}/templates`, this.getRequestOptions());
  }

  createTemplate(body: any): Observable<any> {
    return this.http.post(`${this.apiUrl}/templates`, body, this.getRequestOptions());
  }

  updateTemplate(id: number, body: any): Observable<any> {
    return this.http.patch(`${this.apiUrl}/templates/${id}`, body, this.getRequestOptions());
  }

  deleteTemplate(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/templates/${id}`, this.getRequestOptions());
  }

  activateTemplate(id: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/templates/${id}/activate`, {}, this.getRequestOptions());
  }

  deactivateTemplate(id: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/templates/${id}/deactivate`, {}, this.getRequestOptions());
  }

  patchUser(id: number, body: { role?: string; is_active?: boolean; first_name?: string; last_name?: string }): Observable<any> {
    return this.http.patch(`${this.apiUrl}/users/${id}`, body, this.getRequestOptions());
  }

  deleteUser(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/users/${id}`, this.getRequestOptions());
  }

  unmarkAttendance(trainingId: number, studentId: number): Observable<any> {
    return this.http.post(`${this.apiUrl}/unmark-attendance`, {
      training_id: trainingId,
      student_id: studentId
    }, this.getRequestOptions());
  }

  getAnalyticsSummary(from?: string, to?: string): Observable<any> {
    let params = new HttpParams();
    if (from) params = params.set('from', from);
    if (to) params = params.set('to', to);
    return this.http.get(`${this.apiUrl}/analytics/summary`, this.getRequestOptions(params));
  }

  getInactiveStudents(weeks = 4): Observable<any> {
    const params = new HttpParams().set('weeks', String(weeks));
    return this.http.get(`${this.apiUrl}/analytics/inactive-students`, this.getRequestOptions(params));
  }

  attendanceCsvUrl(year: number, month: number): string {
    return `${this.apiUrl}/analytics/attendance.csv?year=${year}&month=${month}`;
  }

  createSubscription(studentUserId: number, lessons: number, endDate: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/subscriptions`, {
      student_user_id: studentUserId,
      lessons,
      end_date: endDate
    }, this.getRequestOptions());
  }

  deleteSubscription(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/subscriptions/${id}`, this.getRequestOptions());
  }

  createInvite(role: 'student' | 'coach', ttlHours = 72): Observable<any> {
    return this.http.post(`${this.apiUrl}/invites`, { role, ttl_hours: ttlHours }, this.getRequestOptions());
  }

  getUsersWithoutPassword(): Observable<{ users: { id: number; name: string; username: string; role: string }[] }> {
    return this.http.get<{ users: { id: number; name: string; username: string; role: string }[] }>(
      `${this.apiUrl}/users/without-password`,
      this.getRequestOptions()
    );
  }

  createPasswordLink(userId: number, ttlHours = 72): Observable<{ url: string; expires_at: string; name: string }> {
    return this.http.post<{ url: string; expires_at: string; name: string }>(
      `${this.apiUrl}/password-links`,
      { user_id: userId, ttl_hours: ttlHours },
      this.getRequestOptions()
    );
  }

  validatePasswordLink(token: string): Observable<any> {
    const params = new HttpParams().set('token', token);
    return this.http.get(`${this.apiUrl}/password-links/validate`, this.getRequestOptions(params));
  }

  acceptPasswordLink(body: { token: string; password: string; username?: string }): Observable<any> {
    return this.http.post(`${this.apiUrl}/password-links/accept`, body, { withCredentials: true });
  }

  validateInvite(token: string): Observable<any> {
    const params = new HttpParams().set('token', token);
    return this.http.get(`${this.apiUrl}/invites/validate`, this.getRequestOptions(params));
  }

  registerWithInvite(body: {
    token: string;
    username: string;
    password: string;
    first_name: string;
    last_name: string;
    pd_consent: boolean;
    policy_version: string;
  }): Observable<any> {
    return this.http.post(`${this.apiUrl}/register-invite`, body, { withCredentials: true });
  }

  acceptPDConsent(policyVersion: string): Observable<any> {
    return this.http.post(`${this.apiUrl}/pd-consent`, { policy_version: policyVersion }, this.getRequestOptions());
  }

  getPlanTemplates(): Observable<{ templates: any[] }> {
    return this.http.get<{ templates: any[] }>(`${this.apiUrl}/plan-templates`, this.getRequestOptions());
  }

  createPlanTemplate(body: { title: string; description: string; items: { text: string; sets: number | null; reps: number | null; rest_seconds: number | null; exercise_id: number | null }[] }): Observable<any> {
    return this.http.post(`${this.apiUrl}/plan-templates`, body, this.getRequestOptions());
  }

  updatePlanTemplate(id: number, body: { title: string; description: string; items: { text: string; sets: number | null; reps: number | null; rest_seconds: number | null; exercise_id: number | null }[] }): Observable<any> {
    return this.http.patch(`${this.apiUrl}/plan-templates/${id}`, body, this.getRequestOptions());
  }

  deletePlanTemplate(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/plan-templates/${id}`, this.getRequestOptions());
  }

  getExercises(): Observable<{ exercises: { id: number; name: string; description: string }[] }> {
    return this.http.get<{ exercises: { id: number; name: string; description: string }[] }>(`${this.apiUrl}/exercises`, this.getRequestOptions());
  }

  createExercise(body: { name: string; description: string }): Observable<any> {
    return this.http.post(`${this.apiUrl}/exercises`, body, this.getRequestOptions());
  }

  updateExercise(id: number, body: { name: string; description: string }): Observable<any> {
    return this.http.patch(`${this.apiUrl}/exercises/${id}`, body, this.getRequestOptions());
  }

  deleteExercise(id: number): Observable<any> {
    return this.http.delete(`${this.apiUrl}/exercises/${id}`, this.getRequestOptions());
  }

  getTrainingPlan(trainingId: number): Observable<{ plan: any }> {
    return this.http.get<{ plan: any }>(`${this.apiUrl}/trainings/${trainingId}/plan`, this.getRequestOptions());
  }

  saveTrainingPlan(trainingId: number, body: {
    source_template_id: number | null;
    description: string;
    items: { text: string; sets: number | null; reps: number | null; rest_seconds: number | null; exercise_id: number | null }[];
  }): Observable<any> {
    return this.http.put(`${this.apiUrl}/trainings/${trainingId}/plan`, body, this.getRequestOptions());
  }

  getClimbers(): Observable<{ climbers: any[]; is_coach: boolean }> {
    return this.http.get<{ climbers: any[]; is_coach: boolean }>(`${this.apiUrl}/climbers`, this.getRequestOptions());
  }

  updateClimber(studentId: number, body: {
    group_id: number | null;
    discipline: string;
    last_grade: string;
  }): Observable<any> {
    return this.http.patch(`${this.apiUrl}/climbers/${studentId}`, body, this.getRequestOptions());
  }
}
