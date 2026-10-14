import { Component, HostListener, OnDestroy, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { FormsModule } from '@angular/forms';
import { NavigationEnd, Router, RouterLink, RouterLinkActive, RouterOutlet } from '@angular/router';
import { Subscription } from 'rxjs';
import { filter } from 'rxjs/operators';
import { CalendarService } from '../services/calendar.service';
import { applyTheme, AppTheme, readTheme } from '../theme';
import { CookieNoticeComponent } from '../cookie-notice/cookie-notice.component';
import { PD_POLICY_VERSION } from '../pd-policy';
import { httpErrorMessage } from '../utils/http-error';

@Component({
  selector: 'app-shell',
  standalone: true,
  imports: [CommonModule, FormsModule, RouterOutlet, RouterLink, RouterLinkActive, CookieNoticeComponent],
  templateUrl: './app-shell.component.html',
  styleUrl: './app-shell.component.css'
})
export class AppShellComponent implements OnInit, OnDestroy {
  isCoach = false;
  menuOpen = false;
  theme: AppTheme = readTheme();
  needsConsent = false;
  pdAccepted = false;
  consentSaving = false;
  consentError: string | null = null;
  showFooter = true;
  private subs = new Subscription();

  constructor(
    private calendarService: CalendarService,
    private router: Router
  ) {}

  ngOnInit(): void {
    this.updateFooter(this.router.url);
    this.subs.add(
      this.calendarService.getAuthStatus().subscribe({
        next: (status) => {
          this.isCoach = !!status?.is_coach;
          this.needsConsent = !!status?.authenticated && !status?.pd_consent_ok;
        },
        error: () => {
          this.isCoach = false;
        }
      })
    );

    this.subs.add(
      this.router.events
        .pipe(filter((e): e is NavigationEnd => e instanceof NavigationEnd))
        .subscribe((event) => {
          this.closeMenu();
          this.updateFooter(event.urlAfterRedirects);
        })
    );
  }

  ngOnDestroy(): void {
    this.subs.unsubscribe();
    document.body.style.overflow = '';
  }

  get logoSrc(): string {
    return this.theme === 'dark' ? 'assets/spect_on_black.png' : 'assets/spectr_club_logo.png';
  }

  private updateFooter(url: string): void {
    const path = url.split('?')[0];
    this.showFooter = path !== '/calendar' && path !== '/login';
  }

  toggleTheme(): void {
    this.theme = this.theme === 'dark' ? 'light' : 'dark';
    applyTheme(this.theme);
  }

  acceptConsent(): void {
    if (!this.pdAccepted || this.consentSaving) return;
    this.consentSaving = true;
    this.consentError = null;
    this.calendarService.acceptPDConsent(PD_POLICY_VERSION).subscribe({
      next: () => {
        this.consentSaving = false;
        this.needsConsent = false;
      },
      error: (err) => {
        this.consentSaving = false;
        this.consentError = httpErrorMessage(err, 'Не удалось сохранить согласие');
      }
    });
  }

  toggleMenu(): void {
    this.menuOpen = !this.menuOpen;
    this.syncBodyScroll();
  }

  closeMenu(): void {
    if (!this.menuOpen) {
      return;
    }
    this.menuOpen = false;
    this.syncBodyScroll();
  }

  private syncBodyScroll(): void {
    document.body.style.overflow = this.menuOpen ? 'hidden' : '';
  }

  @HostListener('document:keydown.escape')
  onEscape(): void {
    if (this.menuOpen) {
      this.closeMenu();
    }
  }
}
