package models

import "time"

type Subscription struct {
	ID               int64     `db:"id" json:"id"`
	StudentID        int64     `db:"student_id" json:"student_id"`
	StartDate        time.Time `db:"start_date" json:"start_date"`
	EndDate          time.Time `db:"end_date" json:"end_date,omitempty"`
	TotalLessons     int       `db:"total_lessons" json:"total_lessons"`
	RemainingLessons int       `db:"remaining_lessons" json:"remaining_lessons"`
	CreatedAt        time.Time `db:"created_at" json:"created_at"`
}

// CREATE TABLE spectrum.subscriptions (
//     id SERIAL PRIMARY KEY,
//     student_id BIGINT REFERENCES students(id) ON DELETE CASCADE,
//     start_date TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     end_date TIMESTAMP,
//     total_lessons INT NOT NULL,
//     remaining_lessons INT NOT NULL,
//     days_left INT,
//     created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// );
