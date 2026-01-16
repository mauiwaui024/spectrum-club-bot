import { Component, OnInit } from '@angular/core';
import { CommonModule } from '@angular/common';
import { HttpClient, HttpHeaders } from '@angular/common/http';
import { FormsModule } from '@angular/forms';

@Component({
  selector: 'app-debug-initdata',
  standalone: true,
  imports: [CommonModule, FormsModule],
  template: `
    <div style="padding: 20px; font-family: monospace;">
      <h2>🔍 Debug: Telegram WebApp initData</h2>
      
      <div style="margin: 20px 0;">
        <h3>1. Telegram WebApp Status</h3>
        <pre>{{ webAppStatus | json }}</pre>
      </div>
      
      <div style="margin: 20px 0;">
        <h3>2. initData (raw)</h3>
        <textarea readonly style="width: 100%; height: 100px; font-family: monospace; font-size: 12px;">{{ initDataRaw }}</textarea>
      </div>
      
      <div style="margin: 20px 0;">
        <h3>3. initDataUnsafe (parsed)</h3>
        <pre>{{ initDataUnsafe | json }}</pre>
      </div>
      
      <div style="margin: 20px 0; padding: 15px; background: #e7f3ff; border-radius: 5px; border: 1px solid #b3d9ff;">
        <h3>🔧 Вставить initData вручную (для тестирования)</h3>
        <p style="font-size: 12px; color: #666; margin-bottom: 10px;">
          Если вы открыли страницу в обычном браузере (не через Telegram), вставьте сюда initData для тестирования.
          Получить initData можно: открыв страницу через Telegram бота и скопировав <code>window.Telegram.WebApp.initData</code> из консоли.
        </p>
        <textarea 
          [(ngModel)]="manualInitData" 
          placeholder="Вставьте сюда initData (например: user=%7B%22id%3A123456%22%7D&hash=...)"
          style="width: 100%; height: 80px; font-family: monospace; font-size: 12px; padding: 8px; border: 1px solid #ccc; border-radius: 4px;"
        ></textarea>
        <div style="margin-top: 10px;">
          <button (click)="applyManualInitData()" style="padding: 8px 16px; background: #4CAF50; color: white; border: none; border-radius: 4px; cursor: pointer;">
            Применить initData
          </button>
          <button (click)="clearManualInitData()" style="padding: 8px 16px; margin-left: 10px; background: #f44336; color: white; border: none; border-radius: 4px; cursor: pointer;">
            Очистить
          </button>
        </div>
        <div *ngIf="manualInitDataApplied" style="margin-top: 10px; padding: 8px; background: #d4edda; border-radius: 4px; color: #155724;">
          ✅ initData применен! Теперь можно использовать кнопки "Проверить на сервере" и "Проверить авторизацию"
        </div>
      </div>
      
      <div style="margin: 20px 0;">
        <h3>4. Server Response (from /api/debug/initdata)</h3>
        <button (click)="checkServer()" style="padding: 10px; margin-bottom: 10px;">Проверить на сервере</button>
        <pre *ngIf="serverResponse">{{ serverResponse | json }}</pre>
        <div *ngIf="serverError" style="color: red;">{{ serverError }}</div>
      </div>
      
      <div style="margin: 20px 0;">
        <h3>5. Auth Test (from /api/auth)</h3>
        <button (click)="testAuth()" style="padding: 10px; margin-bottom: 10px;">Проверить авторизацию</button>
        <pre *ngIf="authResponse">{{ authResponse | json }}</pre>
        <div *ngIf="authError" style="color: red;">{{ authError }}</div>
      </div>
      
      <div style="margin: 20px 0; padding: 10px; background: #f0f0f0; border-radius: 5px;">
        <h3>📝 Инструкции:</h3>
        <ol>
          <li><strong>Важно:</strong> Откройте эту страницу через Telegram бота (кнопка "Открыть календарь")</li>
          <li>Проверьте, что Telegram WebApp доступен</li>
          <li>Проверьте, что initData не пустой</li>
          <li>Нажмите "Проверить на сервере" - должен вернуться initData</li>
          <li>Нажмите "Проверить авторизацию" - должен вернуться user_id</li>
        </ol>
        <div style="margin-top: 10px; padding: 10px; background: #fff3cd; border-radius: 5px;">
          <strong>⚠️ Внимание:</strong> Если вы видите эту страницу в обычном браузере (не через Telegram), 
          то <code>initData</code> будет пустым. Это нормально! 
          <code>initData</code> доступен только при открытии через Telegram WebApp.
        </div>
      </div>
      
      <div style="margin: 20px 0; padding: 10px; background: #e7f3ff; border-radius: 5px;">
        <h3>🔧 Для локального тестирования (без Telegram):</h3>
        <p>Если нужно протестировать локально без Telegram, можно использовать query параметр <code>user_id</code>:</p>
        <code>http://localhost:4200/calendar?user_id=YOUR_USER_ID</code>
        <p style="margin-top: 10px; font-size: 12px; color: #666;">
          В этом случае будет использован fallback механизм - user_id из URL вместо initData.
        </p>
      </div>
    </div>
  `
})
export class DebugInitDataComponent implements OnInit {
  webAppStatus: any = {};
  initDataRaw: string = '';
  initDataUnsafe: any = null;
  serverResponse: any = null;
  serverError: string = '';
  authResponse: any = null;
  authError: string = '';
  manualInitData: string = '';
  manualInitDataApplied: boolean = false;

  constructor(private http: HttpClient) {}

  ngOnInit() {
    this.checkWebApp();
  }

  checkWebApp() {
    try {
      const tg = (window as any).Telegram?.WebApp;
      if (tg) {
        this.webAppStatus = {
          available: true,
          version: tg.version || 'unknown',
          platform: tg.platform || 'unknown',
          ready: tg.ready || false,
          initData: tg.initData || null,
          initDataLength: tg.initData ? tg.initData.length : 0
        };
        this.initDataRaw = tg.initData || '';
        this.initDataUnsafe = tg.initDataUnsafe || null;
      } else {
        this.webAppStatus = {
          available: false,
          message: 'Telegram WebApp не доступен (открыто не через Telegram)'
        };
      }
    } catch (e: any) {
      this.webAppStatus = {
        available: false,
        error: e.message
      };
    }
  }

  checkServer() {
    this.serverError = '';
    this.serverResponse = null;
    
    // Получаем initData из Telegram WebApp или ручного ввода
    const headers: any = {};
    const initData = this.getInitData();
    if (initData) {
      headers['X-Telegram-Init-Data'] = initData;
    }
    
    this.http.get('/api/debug/initdata', { headers }).subscribe({
      next: (data) => {
        this.serverResponse = data;
      },
      error: (err) => {
        console.error('Server check error:', err);
        // Обработка разных типов ошибок
        if (err.status === 404) {
          this.serverError = 'Debug endpoint недоступен (работает только в development режиме). Проверьте, что ENVIRONMENT=development';
        } else if (err.status === 0) {
          this.serverError = 'Не удалось подключиться к серверу. Убедитесь, что бэкенд запущен на порту 8080';
        } else if (err.error) {
          if (typeof err.error === 'string') {
            this.serverError = err.error;
          } else if (err.error.message) {
            this.serverError = err.error.message;
          } else {
            this.serverError = `Ошибка ${err.status}: ${JSON.stringify(err.error)}`;
          }
        } else {
          this.serverError = err.message || `Ошибка ${err.status || 'unknown'}`;
        }
      }
    });
  }

  testAuth() {
    this.authError = '';
    this.authResponse = null;
    
    const headers: any = {};
    const initData = this.getInitData();
    if (!initData) {
      this.authError = 'initData не доступен. Откройте страницу через Telegram бота или вставьте initData вручную.';
      return;
    }
    headers['X-Telegram-Init-Data'] = initData;
    
    this.http.post('/api/auth', null, { headers }).subscribe({
      next: (data) => {
        this.authResponse = data;
      },
      error: (err) => {
        console.error('Auth error:', err);
        // Обработка разных типов ошибок
        if (err.error) {
          if (typeof err.error === 'string') {
            this.authError = err.error;
          } else if (err.error.message) {
            this.authError = err.error.message;
          } else {
            this.authError = JSON.stringify(err.error);
          }
        } else {
          this.authError = err.message || 'Ошибка авторизации';
        }
      }
    });
  }

  // Получить initData из Telegram WebApp или ручного ввода
  private getInitData(): string | null {
    // Сначала пробуем получить из ручного ввода
    if (this.manualInitDataApplied && this.manualInitData) {
      return this.manualInitData;
    }
    
    // Затем пробуем получить из Telegram WebApp
    try {
      const tg = (window as any).Telegram?.WebApp;
      if (tg && tg.initData) {
        return tg.initData;
      }
    } catch (e) {
      // ignore
    }
    
    return null;
  }

  applyManualInitData() {
    if (this.manualInitData && this.manualInitData.trim()) {
      this.manualInitDataApplied = true;
      // Обновляем отображение
      this.initDataRaw = this.manualInitData;
      console.log('✅ Ручной initData применен:', this.manualInitData.substring(0, 100) + '...');
    } else {
      alert('Введите initData перед применением');
    }
  }

  clearManualInitData() {
    this.manualInitData = '';
    this.manualInitDataApplied = false;
    this.initDataRaw = '';
    console.log('🗑️ Ручной initData очищен');
  }
}
