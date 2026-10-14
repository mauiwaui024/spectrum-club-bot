# Spectrum Club — документация

Один процесс Go: HTTP API + раздача Angular SPA. Telegram-бот удалён; колонка `telegram_id` в БД сохранена для совместимости (синтетические отрицательные id у web-пользователей).

---

## 1. Архитектура

```
Browser (Angular SPA) ──cookie session──► Go HTTP (/api/*, /) ──► Postgres (schema spectrum)
```

| Слой | Путь |
|------|------|
| Entrypoint | `cmd/bot/main.go` |
| HTTP / auth / calendar | `internal/web/` |
| Business logic | `internal/service/` |
| SQL | `internal/repository/` |
| Models | `internal/models/` |
| SPA | `frontend/src/app/` |
| DB migrations | `../spectrum-club-database/migrations/` |

---

## 2. Аутентификация

### Cookie-сессия

1. `POST /api/browser-auth/login` `{username, password}`
2. Проверка `password_hash` (bcrypt; legacy sha256 ещё поддерживается) и `is_active`
3. Cookie `spectrum_web_session` (opaque, in-memory, 24ч)

### Invite-регистрация

1. Coach: `POST /api/invites` → URL `/register?token=...`
2. HMAC-SHA256 секретом `INVITE_SECRET` (dev fallback в `NewHandler`)
3. `POST /api/register-invite` создаёт user + student (+ coach row при роли coach)

Открытой регистрации нет.

### Роли

- `student` — календарь, запись/отмена (не позже чем за **2 часа** до старта), свои данные
- `coach` — расписание, шаблоны, группы, абонементы, ученики, аналитика, invite, посещаемость

---

## 3. HTTP API (кратко)

| Method | Path | Кто | Назначение |
|--------|------|-----|------------|
| GET | `/health` | — | liveness |
| POST | `/api/browser-auth/login` | public | cookie login |
| POST | `/api/browser-auth/logout` | session | logout |
| POST | `/api/browser-auth/set-password` | auth | сменить пароль |
| GET | `/api/auth/status` | soft | `{authenticated, role, is_coach, ...}` |
| GET | `/api/calendar` | soft | календарь |
| GET | `/api/training/{id}` | soft | детали |
| POST | `/api/register` | student | запись (до −2ч) |
| POST | `/api/cancel` | student | отмена (до −2ч) |
| POST | `/api/mark-attendance` | coach | отметка |
| POST | `/api/unmark-attendance` | coach | снять отметку + вернуть занятие |
| GET | `/api/groups` | coach | список |
| POST | `/api/groups` | coach | создать |
| PATCH/DELETE | `/api/groups/{id}` | coach | правка/удаление |
| GET/POST | `/api/templates` | coach | шаблоны |
| PATCH/DELETE | `/api/templates/{id}` | coach | правка/удаление |
| POST | `/api/templates/{id}/activate\|deactivate` | coach | вкл/выкл |
| POST | `/api/trainings` | coach | создать (+ `max_participants`) |
| PATCH/DELETE | `/api/trainings/{id}` | coach | правка/удаление |
| POST | `/api/trainings/from-templates` | coach | bulk |
| GET | `/api/students` | coach | `?filter=all\|no_subscription\|expiring&days=14&q=` |
| PATCH | `/api/users/{id}` | coach | `{role?, is_active?}` |
| GET | `/api/analytics/summary` | coach | дашборд |
| GET | `/api/analytics/attendance.csv` | coach | CSV за месяц |
| GET | `/api/analytics/inactive-students` | coach | `?weeks=4` |
| POST | `/api/invites` | coach | invite |
| POST | `/api/register-invite` | public+token | регистрация |

SPA routes (coach): `/students`, `/analytics`, `/schedule-manage`, …

---

## 4. Переменные окружения

| Var | Назначение |
|-----|------------|
| `ENVIRONMENT` | production / иначе |
| `HTTP_PORT` | default 8080 |
| `BASE_URL` | публичный origin для invite URL |
| `INVITE_SECRET` | HMAC invite |
| `DB_*` | Postgres |

---

## 5. Узкие места

1. In-memory sessions и used-invite jti
2. `username` не UNIQUE в Postgres
3. Нет колонки типа абонемента
4. Тесты без Postgres

---

## 6. Миграции

```bash
psql "$DSN" -f spectrum-club-database/migrations/001_init.sql
psql "$DSN" -f spectrum-club-database/migrations/002_users_is_active.sql
```

---

## 7. Тесты / сборка

```bash
go test ./...
cd frontend && npm ci && npm run build
```
