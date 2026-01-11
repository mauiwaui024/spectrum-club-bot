package models

import "time"

type WeekScheduleTemplate struct {
	ID          int       `db:"id"`
	GroupID     int       `db:"group_id"`
	DayOfWeek   int       `db:"day_of_week"` // 1=понедельник, 7=воскресенье
	StartTime   string    `db:"start_time"`  // "15:30:00"
	EndTime     string    `db:"end_time"`    // "17:00:00"
	Description string    `db:"description"` // "Тенгус (блдр)"
	IsActive    bool      `db:"is_active"`
	CreatedAt   time.Time `db:"created_at"`
	UpdatedAt   time.Time `db:"updated_at"`
}
