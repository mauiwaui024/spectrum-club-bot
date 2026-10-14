export const THEME_KEY = 'spectrum_theme';
export type AppTheme = 'light' | 'dark';

export function readTheme(): AppTheme {
  try {
    return localStorage.getItem(THEME_KEY) === 'dark' ? 'dark' : 'light';
  } catch {
    return 'light';
  }
}

export function applyTheme(theme: AppTheme): void {
  document.documentElement.dataset['theme'] = theme;
  try {
    localStorage.setItem(THEME_KEY, theme);
  } catch {
    /* ignore private mode */
  }
}
