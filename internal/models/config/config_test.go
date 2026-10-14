package config

import (
	"os"
	"testing"
)

func TestValidate_RequiresDBUser(t *testing.T) {
	AppConfig = &Config{
		Environment: "development",
		Database:    DatabaseConfig{Username: "", Password: "p"},
	}
	if err := validate(); err == nil {
		t.Fatal("expected error when DB_USER empty")
	}
}

func TestValidate_OKDevelopment(t *testing.T) {
	AppConfig = &Config{
		Environment: "development",
		Database:    DatabaseConfig{Username: "u", Password: ""},
	}
	if err := validate(); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
}

func TestValidate_ProductionRequiresDBPassword(t *testing.T) {
	_ = os.Unsetenv("DISABLE_TELEGRAM_BOT")
	AppConfig = &Config{
		Environment: "production",
		Database:    DatabaseConfig{Username: "u", Password: ""},
	}
	if err := validate(); err == nil {
		t.Fatal("expected error when DB_PASSWORD empty in production")
	}
}
