/** Extract a human-readable message from Angular HttpClient errors. */
export function httpErrorMessage(err: unknown, fallback = 'Ошибка запроса'): string {
  if (err == null) {
    return fallback;
  }
  const e = err as {
    error?: unknown;
    message?: string;
    status?: number;
    statusText?: string;
  };
  const body = e.error;
  if (typeof body === 'string' && body.trim()) {
    return body.trim();
  }
  if (body && typeof body === 'object') {
    const o = body as Record<string, unknown>;
    if (typeof o['message'] === 'string' && o['message'].trim()) {
      return o['message'].trim();
    }
    if (typeof o['error'] === 'string' && o['error'].trim()) {
      return o['error'].trim();
    }
  }
  if (typeof e.message === 'string' && e.message && !e.message.startsWith('Http failure')) {
    return e.message;
  }
  if (e.status) {
    return `${fallback} (${e.status}${e.statusText ? ' ' + e.statusText : ''})`;
  }
  return fallback;
}
