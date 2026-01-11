package models

import "time"

type Coach struct {
	ID          int64     `db:"id" json:"id"`
	UserID      int64     `db:"user_id" json:"user_id"`
	Specialty   string    `db:"specialty" json:"specialty"`
	Experience  string    `db:"experience" json:"experience"`
	Description string    `db:"description" json:"description"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
}
