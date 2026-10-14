package config

// AppConfig глобальная конфигурация приложения
var AppConfig *Config

// Config основной конфиг
type Config struct {
	Environment string
	BaseURL     string
	Database    DatabaseConfig
	HTTPPort    string `mapstructure:"HTTP_PORT" default:"8080"`
}
