import { Component } from '@angular/core';

@Component({
  selector: 'app-climber-loader',
  standalone: true,
  template: `
    <div class="loader" role="status" aria-label="Загрузка">
      <svg viewBox="0 0 160 100" aria-hidden="true">
        <path class="mountain" d="M8 92 L52 34 L68 50 L104 16 L152 92 Z" />
        <g class="climber">
          <circle cx="0" cy="-7" r="3.2" />
          <path d="M0 -4 L0 6 M0 0 L-6 4 M0 0 L5 -6 M0 6 L-5 14 M0 6 L5 13" />
        </g>
      </svg>
    </div>
  `,
  styles: [`
    .loader {
      position: fixed;
      inset: 0;
      z-index: 90;
      display: flex;
      align-items: center;
      justify-content: center;
      background: var(--bg);
    }
    svg { width: min(240px, 70vw); height: auto; }
    .mountain {
      fill: none;
      stroke: var(--muted);
      stroke-width: 2;
      stroke-linejoin: round;
    }
    .climber {
      fill: var(--text);
      stroke: var(--text);
      stroke-width: 2;
      stroke-linecap: round;
      transform-box: fill-box;
      transform-origin: center;
      animation: climb 1.8s ease-in-out infinite;
    }
    @keyframes climb {
      0% { transform: translate(22px, 80px) rotate(-32deg); }
      40% { transform: translate(40px, 58px) rotate(-24deg); }
      70%, 100% { transform: translate(56px, 42px) rotate(-12deg); }
    }
  `]
})
export class ClimberLoaderComponent {}
