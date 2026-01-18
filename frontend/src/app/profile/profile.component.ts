import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { CalendarService } from '../services/calendar.service';
import { MyProfileResponse } from '../models/student.model';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, RouterModule],
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileComponent implements OnInit {
  profile: MyProfileResponse | null = null;
  loading: boolean = false;
  error: string | null = null;

  constructor(private calendarService: CalendarService) {}

  ngOnInit() {
    this.loadProfile();
  }

  loadProfile() {
    this.loading = true;
    this.error = null;
    
    this.calendarService.getMyProfile().subscribe({
      next: (data) => {
        this.profile = data;
        this.loading = false;
      },
      error: (err) => {
        this.error = 'Не удалось загрузить профиль. Попробуйте позже.';
        this.loading = false;
        console.error('Error loading profile:', err);
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

  getFullName(): string {
    if (!this.profile) return '';
    const { first_name, last_name } = this.profile.user;
    return `${first_name} ${last_name}`.trim() || 'Не указано';
  }

  isCoach(): boolean {
    return this.profile?.user.role === 'coach';
  }

  isStudent(): boolean {
    return this.profile?.user.role === 'student';
  }
}
