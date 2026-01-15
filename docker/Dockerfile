# Multi-stage build для минимального образа

# Этап 1: Angular build
FROM node:20-alpine AS angular-builder

WORKDIR /app

# Копируем package файлы для кэширования зависимостей
COPY frontend/package*.json ./frontend/

# Устанавливаем зависимости Angular
RUN cd frontend && npm ci --only=production=false

# Копируем исходный код Angular
COPY frontend/ ./frontend/

# Собираем Angular приложение для production
RUN cd frontend && npm run build -- --configuration production

# Этап 2: Go build
FROM golang:1.24.3-alpine AS go-builder

WORKDIR /app

# Устанавливаем зависимости для сборки
RUN apk add --no-cache git

# Копируем и загружаем зависимости (кэшируем этот слой)
COPY go.mod go.sum ./
RUN go mod download

# Копируем исходный код
COPY . .

# Собираем статический бинарник
RUN CGO_ENABLED=0 GOOS=linux go build \
    -ldflags="-s -w -X main.buildVersion=$(git rev-parse --short HEAD)" \
    -o /app/bin/bot ./cmd/bot

# Этап 3: Финальный образ
FROM alpine:3.18

WORKDIR /app

# Устанавливаем только необходимые пакеты
RUN apk add --no-cache tzdata ca-certificates wget && \
    update-ca-certificates

# Копируем Go бинарник из go-builder
COPY --from=go-builder /app/bin/bot /app/bot

# Копируем собранные Angular файлы из angular-builder
COPY --from=angular-builder /app/frontend/dist/spectrum-club-calendar/browser /app/frontend/dist/spectrum-club-calendar/browser

# Создаем non-root пользователя для безопасности
RUN addgroup -g 1001 -S appuser && \
    adduser -u 1001 -S appuser -G appuser && \
    chown -R appuser:appuser /app

USER appuser

# HEALTHCHECK для мониторинга
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Запускаем приложение
ENTRYPOINT ["/app/bot"]