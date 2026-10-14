import { Component } from '@angular/core';
import { CommonModule } from '@angular/common';
import { RouterModule } from '@angular/router';

const NOTICE_KEY = 'spectrum_cookie_notice';

@Component({
  selector: 'app-cookie-notice',
  standalone: true,
  imports: [CommonModule, RouterModule],
  template: `
    <div class="cookie-notice" *ngIf="visible">
      <p>
        Сайт использует cookie, необходимые для входа и работы приложения.
        Рекламные и аналитические cookie не применяются.
        <a routerLink="/privacy">Политика</a>
      </p>
      <button type="button" (click)="dismiss()">Понятно</button>
    </div>
  `,
  styles: [`
    .cookie-notice {
      position: fixed;
      left: 12px;
      right: 12px;
      bottom: 12px;
      z-index: 70;
      display: flex;
      gap: 12px;
      align-items: center;
      justify-content: space-between;
      max-width: 720px;
      margin: 0 auto;
      padding: 12px 14px;
      background: var(--card);
      color: var(--text);
      border: 1px solid var(--border);
      border-radius: 10px;
      box-shadow: 0 4px 16px var(--shadow);
    }
    p { margin: 0; font-size: 14px; line-height: 1.4; }
    a { color: var(--primary); }
    button { flex-shrink: 0; padding: 8px 12px; cursor: pointer; }
    @media (max-width: 560px) {
      .cookie-notice { flex-direction: column; align-items: stretch; }
    }
  `]
})
export class CookieNoticeComponent {
  visible = localStorage.getItem(NOTICE_KEY) !== '1';

  dismiss(): void {
    localStorage.setItem(NOTICE_KEY, '1');
    this.visible = false;
  }
}
