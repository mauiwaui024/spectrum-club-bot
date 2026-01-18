import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';
import { FormsModule } from '@angular/forms';
import { CalendarService } from '../services/calendar.service';
import { MyProfileResponse } from '../models/student.model';

@Component({
  selector: 'app-profile',
  standalone: true,
  imports: [CommonModule, RouterModule, FormsModule],
  templateUrl: './profile.component.html',
  styleUrls: ['./profile.component.css']
})
export class ProfileComponent implements OnInit {
  profile: MyProfileResponse | null = null;
  loading: boolean = false;
  error: string | null = null;
  isEditing: boolean = false;
  saving: boolean = false;

  // Form fields for editing
  editFirstName: string = '';
  editLastName: string = '';
  editAthleticTitle: string = '';
  editSpecialty: string = '';
  editExperience: string = '';
  editDescription: string = '';

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

  startEdit() {
    if (!this.profile) return;
    
    this.isEditing = true;
    this.editFirstName = this.profile.user.first_name || '';
    this.editLastName = this.profile.user.last_name || '';
    
    if (this.isStudent() && this.profile.student) {
      this.editAthleticTitle = this.profile.student.athletic_title || '';
    }
    
    if (this.isCoach() && this.profile.coach) {
      this.editSpecialty = this.profile.coach.specialty || '';
      this.editExperience = this.profile.coach.experience || '';
      this.editDescription = this.profile.coach.description || '';
    }
  }

  cancelEdit() {
    this.isEditing = false;
    this.error = null;
    // Reset form fields
    this.editFirstName = '';
    this.editLastName = '';
    this.editAthleticTitle = '';
    this.editSpecialty = '';
    this.editExperience = '';
    this.editDescription = '';
  }

  saveProfile() {
    if (!this.profile) return;

    this.saving = true;
    this.error = null;

    const updateData: any = {
      first_name: this.editFirstName.trim() || undefined,
      last_name: this.editLastName.trim() || undefined,
    };

    if (this.isStudent()) {
      updateData.athletic_title = this.editAthleticTitle.trim() || undefined;
    }

    if (this.isCoach()) {
      updateData.specialty = this.editSpecialty.trim() || undefined;
      updateData.experience = this.editExperience.trim() || undefined;
      updateData.description = this.editDescription.trim() || undefined;
    }

    // Remove undefined fields
    Object.keys(updateData).forEach(key => {
      if (updateData[key] === undefined) {
        delete updateData[key];
      }
    });

    this.calendarService.updateProfile(updateData).subscribe({
      next: (data) => {
        this.profile = data;
        this.isEditing = false;
        this.saving = false;
      },
      error: (err) => {
        this.error = 'Не удалось сохранить изменения. Попробуйте позже.';
        this.saving = false;
        console.error('Error updating profile:', err);
      }
    });
  }
}
