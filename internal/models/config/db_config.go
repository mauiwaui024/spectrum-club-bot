package config

import (
	"fmt"
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
	env := getEnv("ENVIRONMENT", "development")

	AppConfig = &Config{
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		Environment: env,
		BaseURL:     getEnv("BASE_URL", ""),
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

func validate() error {
	var errors []string

	if AppConfig.Database.Username == "" {
		errors = append(errors, "DB_USER is required")
	}

	if AppConfig.Database.Password == "" && AppConfig.Environment == "production" {
		errors = append(errors, "DB_PASSWORD is required in production")
	}

	if len(errors) > 0 {
		return fmt.Errorf("config validation failed: %s", joinErrors(errors))
	}

	return nil
}

func joinErrors(errs []string) string {
	out := ""
	for i, e := range errs {
		if i > 0 {
			out += ", "
		}
		out += e
	}
	return out
}

func getSSLMode(env string) string {
	if env == "production" {
		return "require"
	}
	return "disable"
}
