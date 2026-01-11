package models

import "time"

type User struct {
	ID           int64     `db:"id" json:"id"`
	TelegramID   int64     `db:"telegram_id" json:"telegram_id"`
	FirstName    string    `db:"first_name" json:"first_name"`
	LastName     string    `db:"last_name" json:"last_name"`
	Username     string    `db:"username" json:"username"`
	Role         string    `db:"role" json:"role"` // "student" или "coach"
	RegisteredAt time.Time `db:"registered_at" json:"registered_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
