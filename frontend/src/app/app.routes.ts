import { Routes } from '@angular/router';
import { CalendarComponent } from './calendar/calendar.component';
import { LoginComponent } from './login/login.component';
import { AppShellComponent } from './layout/app-shell.component';
import { shellGuard } from './auth.guard';

export const routes: Routes = [
  {
    path: 'login',
    component: LoginComponent
  },
  {
    path: 'register',
    loadComponent: () => import('./register/register.component').then(m => m.RegisterComponent)
  },
      {
        path: 'privacy',
        loadComponent: () => import('./privacy/privacy.component').then(m => m.PrivacyComponent)
      },
      {
        path: 'set-password',
        loadComponent: () => import('./set-password/set-password.component').then(m => m.SetPasswordComponent)
      },
  {
    path: '',
    component: AppShellComponent,
    canActivate: [shellGuard],
    children: [
      {
        path: '',
        pathMatch: 'full',
        redirectTo: 'calendar'
      },
      {
        path: 'calendar',
        component: CalendarComponent
      },
      {
        path: 'my-registrations',
        loadComponent: () => import('./my-registrations/my-registrations.component').then(m => m.MyRegistrationsComponent)
      },
      {
        path: 'my-subscription',
        loadComponent: () => import('./my-subscription/my-subscription.component').then(m => m.MySubscriptionComponent)
      },
      {
        path: 'profile',
        loadComponent: () => import('./profile/profile.component').then(m => m.ProfileComponent)
      },
      {
        path: 'students-subscriptions',
        loadComponent: () => import('./students-subscriptions/students-subscriptions.component').then(m => m.StudentsSubscriptionsComponent)
      },
      {
        path: 'schedule-manage',
        loadComponent: () => import('./schedule-manage/schedule-manage.component').then(m => m.ScheduleManageComponent)
      },
      {
        path: 'schedule-templates',
        loadComponent: () => import('./schedule-templates/schedule-templates.component').then(m => m.ScheduleTemplatesComponent)
      },
      {
        path: 'groups',
        loadComponent: () => import('./groups/groups.component').then(m => m.GroupsComponent)
      },
      {
        path: 'training-plans',
        loadComponent: () => import('./training-plans/training-plans.component').then(m => m.TrainingPlansComponent)
      },
      {
        path: 'exercises',
        loadComponent: () => import('./exercises/exercises.component').then(m => m.ExercisesComponent)
      },
      {
        path: 'climbers',
        loadComponent: () => import('./climbers/climbers.component').then(m => m.ClimbersComponent)
      },
      {
        path: 'invite',
        loadComponent: () => import('./invite/invite.component').then(m => m.InviteComponent)
      },
      {
        path: 'students',
        loadComponent: () => import('./students/students.component').then(m => m.StudentsComponent)
      },
      {
        path: 'analytics',
        loadComponent: () => import('./analytics/analytics.component').then(m => m.AnalyticsComponent)
      }
    ]
  }
];
