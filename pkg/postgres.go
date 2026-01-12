package database

import (
	"fmt"
	"log"
	"spectrum-club-bot/internal/models/config"

	"github.com/jmoiron/sqlx"
)

// POSTGRES_DB: postgres
// POSTGRES_USER: spectrum-dev
// POSTGRES_PASSWORD: 112233aa
// ports:
// - "5420:5432"

// func NewPostgres() (*sqlx.DB, error) {
// 	connStr := fmt.Sprintf(
// 		"host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
// 		"127.0.0.1",
// 		"5420",
// 		"spectrum-dev",
// 		"112233aa",
// 		"spectrum-db-m",
// 	)

// 	db, err := sqlx.Connect("postgres", connStr)
// 	if err != nil {
// 		return nil, err
// 	}

// 	if err = db.Ping(); err != nil {
// 		return nil, err
// 	}

// 	log.Println("✅ Connected to PostgreSQL")
// 	return db, nil
// }

func NewPostgres() (*sqlx.DB, error) {
	cfg := config.AppConfig.Database

	connStr := fmt.Sprintf(
		"host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
		cfg.Host,
		cfg.Port,
		cfg.Username,
		cfg.Password,
		cfg.Name,
		"disable",
	)

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("database connection failed: %w", err)
	}

	if err = db.Ping(); err != nil {
		return nil, fmt.Errorf("database ping failed: %w", err)
	}

	log.Printf("🗄️  Подключено к PostgreSQL: %s:%d/%s", cfg.Host, cfg.Port, cfg.Name)
	return db, nil
}
