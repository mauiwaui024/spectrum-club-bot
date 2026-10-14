package models

import "time"

type PlanItem struct {
	ID                  int    `db:"id" json:"id"`
	SortOrder           int    `db:"sort_order" json:"sort_order"`
	Text                string `db:"text" json:"text"`
	Sets                *int   `db:"sets" json:"sets"`
	Reps                *int   `db:"reps" json:"reps"`
	RestSeconds         *int   `db:"rest_seconds" json:"rest_seconds"`
	ExerciseID          *int   `db:"exercise_id" json:"exercise_id"`
	ExerciseDescription string `db:"exercise_description" json:"exercise_description"`
}

type Exercise struct {
	ID          int    `db:"id" json:"id"`
	Name        string `db:"name" json:"name"`
	Description string `db:"description" json:"description"`
}

type PlanTemplate struct {
	ID          int        `db:"id" json:"id"`
	Title       string     `db:"title" json:"title"`
	Description string     `db:"description" json:"description"`
	CreatedAt   time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt   time.Time  `db:"updated_at" json:"updated_at"`
	Items       []PlanItem `db:"-" json:"items"`
}

type TrainingPlan struct {
	ID               int        `db:"id" json:"id"`
	TrainingID       int        `db:"training_id" json:"training_id"`
	SourceTemplateID *int       `db:"source_template_id" json:"source_template_id"`
	Description      string     `db:"description" json:"description"`
	CreatedAt        time.Time  `db:"created_at" json:"created_at"`
	UpdatedAt        time.Time  `db:"updated_at" json:"updated_at"`
	Items            []PlanItem `db:"-" json:"items"`
}
