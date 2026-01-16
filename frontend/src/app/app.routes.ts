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
  }
];

console.log('app.routes.ts: Routes configured:', routes);
