import { bootstrapApplication } from '@angular/platform-browser';
import { AppComponent } from './app/app.component';
import { provideHttpClient } from '@angular/common/http';
import { provideRouter } from '@angular/router';
import { LOCALE_ID } from '@angular/core';
import { registerLocaleData, I18nPluralPipe } from '@angular/common';
import localeRu from '@angular/common/locales/ru';
import { routes } from './app/app.routes';
import { CalendarModule, DateAdapter } from 'angular-calendar';
import { adapterFactory } from 'angular-calendar/date-adapters/date-fns';

// Register locale data
registerLocaleData(localeRu);

console.log('main.ts: Starting bootstrap...');

const calendarProviders = CalendarModule.forRoot(
  { provide: DateAdapter, useFactory: adapterFactory }
).providers || [];

bootstrapApplication(AppComponent, {
  providers: [
    provideHttpClient(),
    provideRouter(routes),
    { provide: LOCALE_ID, useValue: 'ru' },
    I18nPluralPipe,
    ...calendarProviders
  ]
}).then(() => {
  console.log('main.ts: Bootstrap successful!');
}).catch(err => {
  console.error('main.ts: Bootstrap error:', err);
});
