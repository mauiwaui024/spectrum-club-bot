package models

import (
	"time"
)

// Group - модель группы тренировок
type TrainingGroup struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Code        string    `json:"code"`
	AgeMin      int       `json:"age_min"`
	AgeMax      *int      `json:"age_max"`
	Description string    `json:"description"`
	CreatedAt   time.Time `json:"created_at"`
}

type TrainingSchedule struct {
	ID              int       `json:"id"`
	GroupID         int       `json:"group_id"`
	CoachID         *int64    `json:"coach_id"`
	TrainingDate    time.Time `json:"training_date"`
	StartTime       time.Time `json:"start_time"`
	EndTime         time.Time `json:"end_time"`
	Description     string    `json:"description"`
	MaxParticipants *int      `json:"max_participants"`
	CreatedBy       *int64    `json:"created_by"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`

	// Joined fields
	GroupName string `json:"group_name,omitempty"`
	CoachName string `json:"coach_name,omitempty"`
}

type Attendance struct {
	ID         int       `json:"id"`
	TrainingID int       `json:"training_id"`
	StudentID  int       `json:"student_id"`
	Attended   bool      `json:"attended"`
	Notes      string    `json:"notes"`
	RecordedBy *int      `json:"recorded_by"`
	RecordedAt time.Time `json:"recorded_at"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
	Status     string    `json:"status"` // registered, cancelled, attended

	// Joined fields
	StudentName string `json:"student_name,omitempty"`
}

type AttendanceWithStudent struct {
	Attendance
	Student struct {
		ID            int64     `json:"id"`
		UserID        int64     `json:"user_id"`
		AthleticTitle string    `json:"athletic_title"`
		CreatedAt     time.Time `json:"created_at"`
		UpdatedAt     time.Time `json:"updated_at"`
		StudentName   string    `json:"student_name"`
		User          User      `json:"user"`
	} `json:"student"`
}

type AttendanceWithTraining struct {
	Attendance
	Training TrainingSchedule `json:"training"`
}
