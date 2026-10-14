import { Injectable, signal } from '@angular/core';

@Injectable({ providedIn: 'root' })
export class AuthGate {
  readonly pending = signal(true);

  begin(): void {
    this.pending.set(true);
  }

  end(): void {
    this.pending.set(false);
  }
}
