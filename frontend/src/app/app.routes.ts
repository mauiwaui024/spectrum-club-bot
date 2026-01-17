import { Routes } from '@angular/router';
import { CalendarComponent } from './calendar/calendar.component';

console.log('app.routes.ts: CalendarComponent imported:', CalendarComponent);

export const routes: Routes = [
  {
    path: '',
    component: CalendarComponent
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
    path: 'debug/initdata',
    loadComponent: () => import('./debug-initdata.component').then(m => m.DebugInitDataComponent)
  }
];

console.log('app.routes.ts: Routes configured:', routes);
