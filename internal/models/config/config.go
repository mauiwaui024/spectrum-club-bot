package config

// AppConfig глобальная конфигурация приложения
var AppConfig *Config

// Config основной конфиг
type Config struct {
	Environment string
	Bot         BotConfig
	Database    DatabaseConfig
	HTTPPort    string `mapstructure:"HTTP_PORT" default:"8080"`
}

type BotConfig struct {
	Token    string
	Debug    bool
	BaseURL  string
	AdminIDs []int64 // ID администраторов для уведомлений
}
