import { inject } from '@angular/core';
import { CanActivateFn, Router } from '@angular/router';
import { catchError, map, of } from 'rxjs';
import { AuthGate } from './auth-gate';
import { CalendarService } from './services/calendar.service';

export const shellGuard: CanActivateFn = () => {
  const api = inject(CalendarService);
  const router = inject(Router);
  const gate = inject(AuthGate);
  gate.begin();
  return api.getAuthStatus().pipe(
    map((status) => {
      gate.end();
      if (status?.authenticated) {
        return true;
      }
      return router.createUrlTree(['/login']);
    }),
    catchError(() => {
      gate.end();
      return of(router.createUrlTree(['/login']));
    })
  );
};
