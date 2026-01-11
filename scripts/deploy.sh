#!/bin/bash
set -e  # Выход при ошибке

echo "🚀 Starting deployment process..."

# Переходим в директорию со скриптом
cd "$(dirname "$0")"

# Загружаем переменные окружения
if [ -f .env ]; then
    echo "📦 Loading environment variables..."
    export $(cat .env | grep -v '^#' | xargs)
else
    echo "❌ .env file not found!"
    exit 1
fi

# Проверяем обязательные переменные
required_vars=("BOT_TOKEN" "DB_PASSWORD" "DB_USER" "DB_NAME")
for var in "${required_vars[@]}"; do
    if [ -z "${!var}" ]; then
        echo "❌ Missing required variable: $var"
        exit 1
    fi
done

# Останавливаем и удаляем старый контейнер бота
echo "🛑 Stopping old bot container..."
docker stop spectrum-bot-prod 2>/dev/null || true
docker rm spectrum-bot-prod 2>/dev/null || true

# Pull образ бота (если используется registry)
# docker pull spectrum-club-bot:${IMAGE_TAG:-latest}

# Запускаем сервисы
echo "🚀 Starting services with Docker Compose..."
docker compose -f ../docker-compose.yml up -d --remove-orphans

# Проверяем здоровье контейнеров
echo "🏥 Checking container health..."
sleep 10  # Даем время на запуск

if ! docker ps --filter "name=spectrum-bot-prod" --filter "health=healthy" | grep -q "spectrum-bot-prod"; then
    echo "⚠️ Bot container is not healthy. Checking logs..."
    docker logs spectrum-bot-prod --tail 50
    echo "❌ Deployment failed!"
    exit 1
fi

echo "✅ Deployment completed successfully!"

# Очистка старых образов (опционально)
echo "🧹 Cleaning up old Docker images..."
docker image prune -f --filter "until=24h" 2>/dev/null || true

echo "📊 Current containers status:"
docker ps --filter "name=spectrum" --format "table {{.Names}}\t{{.Status}}\t{{.Ports}}"