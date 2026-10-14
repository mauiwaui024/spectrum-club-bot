package models

import "time"

type Student struct {
	ID           int64     `db:"id" json:"id"`
	UserID       int64     `db:"user_id" json:"user_id"`
	AtleticTitle string    `db:"athletic_title" json:"athletic_title"`
	GroupID      *int64    `db:"group_id" json:"group_id"`
	Discipline   *string   `db:"discipline" json:"discipline"`
	LastGrade    *string   `db:"last_grade" json:"last_grade"`
	CreatedAt    time.Time `db:"created_at" json:"created_at"`
	UpdatedAt    time.Time `db:"updated_at" json:"updated_at"`
}
