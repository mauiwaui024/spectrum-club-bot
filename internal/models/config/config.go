package config

// AppConfig глобальная конфигурация приложения
var AppConfig *Config

// Config основной конфиг
type Config struct {
	Environment string
	Bot         BotConfig
	Database    DatabaseConfig
}

type BotConfig struct {
	Token    string
	Debug    bool
	AdminIDs []int64 // ID администраторов для уведомлений
}
