package config

import (
	"fmt"
	"strconv"
	"strings"
)

// DatabaseConfig конфигурация БД
type DatabaseConfig struct {
	Host     string
	Port     int
	Username string
	Password string
	Name     string
	SSLMode  string
}

// Load загружает конфигурацию
func Load() error {
	env := getEnv("APP_ENV", "development")

	AppConfig = &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		Environment: getEnv("ENVIROMENT", "development"),
		Bot: BotConfig{
			Token:    getEnv("BOT_TOKEN", ""),
			Debug:    getEnvAsBool("BOT_DEBUG", env != "production"),
			AdminIDs: parseAdminIDs(getEnv("ADMIN_IDS", "")),
			BaseURL:  getEnv("BASE_URL", ""),
		},
		Database: DatabaseConfig{
			Host:     getEnv("DB_HOST", "localhost"),
			Port:     getEnvAsInt("DB_PORT", 5432),
			Username: getEnv("DB_USER", ""),
			Password: getEnv("DB_PASSWORD", ""),
			Name:     getEnv("DB_NAME", "spectrum-db"),
			SSLMode:  getSSLMode(env),
		},
	}

	return validate()
}

// validate проверяет обязательные параметры
func validate() error {
	var errors []string

	if AppConfig.Bot.Token == "" {
		errors = append(errors, "BOT_TOKEN is required")
	}

	if AppConfig.Database.Username == "" {
		errors = append(errors, "DB_USER is required")
	}

	if AppConfig.Database.Password == "" && AppConfig.Environment == "production" {
		errors = append(errors, "DB_PASSWORD is required in production")
	}

	if len(errors) > 0 {
		return fmt.Errorf("config validation failed: %s", strings.Join(errors, ", "))
	}

	return nil
}

// getSSLMode возвращает режим SSL в зависимости от окружения
func getSSLMode(env string) string {
	if env == "production" {
		return "require" // В продакшене всегда SSL
	}
	return "disable" // В разработке можно отключить
}

// parseAdminIDs парсит список ID администраторов
func parseAdminIDs(ids string) []int64 {
	if ids == "" {
		return []int64{}
	}

	var result []int64
	for _, idStr := range strings.Split(ids, ",") {
		if id, err := strconv.ParseInt(strings.TrimSpace(idStr), 10, 64); err == nil {
			result = append(result, id)
		}
	}
	return result
}
