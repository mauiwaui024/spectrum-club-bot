# Spectrum Club Calendar - Angular Frontend

Angular приложение для календаря тренировок.

## Разработка

### Установка зависимостей

```bash
npm install
```

### Запуск dev сервера

```bash
npm start
```

Приложение будет доступно на `http://localhost:4200`

**Важно**: Go API сервер должен быть запущен на другом порту (по умолчанию 8080). Angular dev server будет проксировать запросы к API.

### Сборка для production

```bash
npm run build
```

Собранные файлы будут в `dist/spectrum-club-calendar/browser/`. Go сервер раздает эти файлы в production.

## Структура проекта

```
src/
├── app/
│   ├── calendar/          # Компонент календаря
│   │   ├── calendar.component.ts
│   │   ├── calendar.component.html
│   │   └── calendar.component.css
│   ├── services/          # Сервисы для работы с API
│   │   └── calendar.service.ts
│   ├── models/            # TypeScript модели
│   │   └── training.model.ts
│   ├── app.component.ts    # Корневой компонент
│   └── app.routes.ts       # Роутинг
├── index.html
├── main.ts
└── styles.css
```

## API Endpoints

Приложение использует следующие API endpoints:

- `GET /api/calendar` - получение данных календаря
- `GET /api/training/{id}` - детали тренировки
- `POST /api/register` - запись на тренировку
- `POST /api/cancel` - отмена записи
- `GET /api/check-registration` - проверка статуса записи
