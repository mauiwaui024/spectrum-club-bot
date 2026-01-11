# Multi-stage build для минимального образа
FROM golang:1.24.3-alpine AS builder

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

# Финальный образ
FROM alpine:3.18

WORKDIR /app

# Устанавливаем только необходимые пакеты
RUN apk add --no-cache tzdata ca-certificates && \
    update-ca-certificates

# Копируем бинарник из builder
COPY --from=builder /app/bin/bot /app/bot

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