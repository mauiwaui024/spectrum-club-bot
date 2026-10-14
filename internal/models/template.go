package models

import "time"

type WeekScheduleTemplate struct {
	ID          int       `db:"id" json:"id"`
	GroupID     int       `db:"group_id" json:"group_id"`
	DayOfWeek   int       `db:"day_of_week" json:"day_of_week"` // 1=понедельник, 7=воскресенье
	StartTime   string    `db:"start_time" json:"start_time"`    // "15:30:00"
	EndTime     string    `db:"end_time" json:"end_time"`        // "17:00:00"
	Description string    `db:"description" json:"description"`  // "Тенгус (блдр)"
	IsActive    bool      `db:"is_active" json:"is_active"`
	CreatedAt   time.Time `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time `db:"updated_at" json:"updated_at"`
	// Enriched for SPA (not a DB column)
	GroupName string `db:"-" json:"group_name,omitempty"`
	GroupCode string `db:"-" json:"group_code,omitempty"`
}
