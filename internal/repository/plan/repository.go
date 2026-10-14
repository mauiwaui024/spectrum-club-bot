package plan

import (
	"database/sql"
	"errors"
	"fmt"

	"spectrum-club-bot/internal/models"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) ListTemplates() ([]models.PlanTemplate, error) {
	var templates []models.PlanTemplate
	if err := r.db.Select(&templates, `
		SELECT id, title, description, created_at, updated_at
		FROM spectrum.training_plan_templates
		ORDER BY title, id
	`); err != nil {
		return nil, err
	}
	if templates == nil {
		templates = []models.PlanTemplate{}
	}
	items, err := r.allTemplateItems()
	if err != nil {
		return nil, err
	}
	for i := range templates {
		templates[i].Items = items[templates[i].ID]
		if templates[i].Items == nil {
			templates[i].Items = []models.PlanItem{}
		}
	}
	return templates, nil
}

func (r *Repository) GetTemplate(id int) (*models.PlanTemplate, error) {
	var t models.PlanTemplate
	err := r.db.Get(&t, `
		SELECT id, title, description, created_at, updated_at
		FROM spectrum.training_plan_templates WHERE id = $1
	`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	t.Items, err = r.templateItems(t.ID)
	if err != nil {
		return nil, err
	}
	return &t, nil
}

func (r *Repository) CreateTemplate(title, description string, items []models.PlanItem) (*models.PlanTemplate, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var id int
	if err := tx.QueryRow(`
		INSERT INTO spectrum.training_plan_templates (title, description)
		VALUES ($1, $2) RETURNING id
	`, title, description).Scan(&id); err != nil {
		return nil, err
	}
	if err := insertTemplateItems(tx, id, items); err != nil {
		return nil, err
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetTemplate(id)
}

func (r *Repository) UpdateTemplate(id int, title, description string, items []models.PlanItem) error {
	tx, err := r.db.Beginx()
	if err != nil {
		return err
	}
	defer tx.Rollback()

	res, err := tx.Exec(`
		UPDATE spectrum.training_plan_templates
		SET title = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, title, description, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	if _, err := tx.Exec(`DELETE FROM spectrum.training_plan_template_items WHERE template_id = $1`, id); err != nil {
		return err
	}
	if err := insertTemplateItems(tx, id, items); err != nil {
		return err
	}
	return tx.Commit()
}

func (r *Repository) DeleteTemplate(id int) error {
	res, err := r.db.Exec(`DELETE FROM spectrum.training_plan_templates WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) GetTrainingPlan(trainingID int) (*models.TrainingPlan, error) {
	var p models.TrainingPlan
	err := r.db.Get(&p, `
		SELECT id, training_id, source_template_id, description, created_at, updated_at
		FROM spectrum.training_plans WHERE training_id = $1
	`, trainingID)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	p.Items, err = r.planItems(p.ID)
	if err != nil {
		return nil, err
	}
	return &p, nil
}

func (r *Repository) SaveTrainingPlan(trainingID int, sourceTemplateID *int, description string, items []models.PlanItem) (*models.TrainingPlan, error) {
	tx, err := r.db.Beginx()
	if err != nil {
		return nil, err
	}
	defer tx.Rollback()

	var planID int
	if err := tx.QueryRow(`
		INSERT INTO spectrum.training_plans (training_id, source_template_id, description)
		VALUES ($1, $2, $3)
		ON CONFLICT (training_id) DO UPDATE
		SET source_template_id = EXCLUDED.source_template_id,
		    description = EXCLUDED.description,
		    updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`, trainingID, sourceTemplateID, description).Scan(&planID); err != nil {
		return nil, err
	}
	if _, err := tx.Exec(`DELETE FROM spectrum.training_plan_items WHERE plan_id = $1`, planID); err != nil {
		return nil, err
	}
	for i, item := range items {
		if _, err := tx.Exec(`
			INSERT INTO spectrum.training_plan_items (plan_id, sort_order, text, sets, reps, rest_seconds, exercise_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, planID, i, item.Text, item.Sets, item.Reps, item.RestSeconds, item.ExerciseID); err != nil {
			return nil, err
		}
	}
	if err := tx.Commit(); err != nil {
		return nil, err
	}
	return r.GetTrainingPlan(trainingID)
}

func (r *Repository) allTemplateItems() (map[int][]models.PlanItem, error) {
	type row struct {
		TemplateID int `db:"template_id"`
		models.PlanItem
	}
	var rows []row
	if err := r.db.Select(&rows, `
		SELECT i.id, i.template_id, i.sort_order, i.text, i.sets, i.reps, i.rest_seconds, i.exercise_id,
		       COALESCE(e.description, '') AS exercise_description
		FROM spectrum.training_plan_template_items i
		LEFT JOIN spectrum.exercises e ON e.id = i.exercise_id
		ORDER BY i.template_id, i.sort_order, i.id
	`); err != nil {
		return nil, err
	}
	out := map[int][]models.PlanItem{}
	for _, row := range rows {
		out[row.TemplateID] = append(out[row.TemplateID], row.PlanItem)
	}
	return out, nil
}

func (r *Repository) templateItems(templateID int) ([]models.PlanItem, error) {
	var items []models.PlanItem
	err := r.db.Select(&items, `
		SELECT i.id, i.sort_order, i.text, i.sets, i.reps, i.rest_seconds, i.exercise_id,
		       COALESCE(e.description, '') AS exercise_description
		FROM spectrum.training_plan_template_items i
		LEFT JOIN spectrum.exercises e ON e.id = i.exercise_id
		WHERE i.template_id = $1
		ORDER BY i.sort_order, i.id
	`, templateID)
	if items == nil {
		items = []models.PlanItem{}
	}
	return items, err
}

func (r *Repository) planItems(planID int) ([]models.PlanItem, error) {
	var items []models.PlanItem
	err := r.db.Select(&items, `
		SELECT i.id, i.sort_order, i.text, i.sets, i.reps, i.rest_seconds, i.exercise_id,
		       COALESCE(e.description, '') AS exercise_description
		FROM spectrum.training_plan_items i
		LEFT JOIN spectrum.exercises e ON e.id = i.exercise_id
		WHERE i.plan_id = $1
		ORDER BY i.sort_order, i.id
	`, planID)
	if items == nil {
		items = []models.PlanItem{}
	}
	return items, err
}

func insertTemplateItems(tx *sqlx.Tx, templateID int, items []models.PlanItem) error {
	for i, item := range items {
		if _, err := tx.Exec(`
			INSERT INTO spectrum.training_plan_template_items (template_id, sort_order, text, sets, reps, rest_seconds, exercise_id)
			VALUES ($1, $2, $3, $4, $5, $6, $7)
		`, templateID, i, item.Text, item.Sets, item.Reps, item.RestSeconds, item.ExerciseID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) ListExercises() ([]models.Exercise, error) {
	var list []models.Exercise
	err := r.db.Select(&list, `
		SELECT id, name, description
		FROM spectrum.exercises
		ORDER BY lower(name), id
	`)
	if list == nil {
		list = []models.Exercise{}
	}
	return list, err
}

func (r *Repository) GetExercise(id int) (*models.Exercise, error) {
	var ex models.Exercise
	err := r.db.Get(&ex, `SELECT id, name, description FROM spectrum.exercises WHERE id = $1`, id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &ex, nil
}

func (r *Repository) CreateExercise(name, description string) (*models.Exercise, error) {
	var id int
	err := r.db.QueryRow(`
		INSERT INTO spectrum.exercises (name, description)
		VALUES ($1, $2)
		RETURNING id
	`, name, description).Scan(&id)
	if err != nil {
		return nil, mapExerciseError(err)
	}
	return r.GetExercise(id)
}

func (r *Repository) UpdateExercise(id int, name, description string) error {
	res, err := r.db.Exec(`
		UPDATE spectrum.exercises
		SET name = $1, description = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, name, description, id)
	if err != nil {
		return mapExerciseError(err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *Repository) DeleteExercise(id int) error {
	var used int
	if err := r.db.Get(&used, `
		SELECT
			(SELECT COUNT(*) FROM spectrum.training_plan_template_items WHERE exercise_id = $1) +
			(SELECT COUNT(*) FROM spectrum.training_plan_items WHERE exercise_id = $1)
	`, id); err != nil {
		return err
	}
	if used > 0 {
		return fmt.Errorf("упражнение используется в плане")
	}
	res, err := r.db.Exec(`DELETE FROM spectrum.exercises WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func mapExerciseError(err error) error {
	var pqErr *pq.Error
	if errors.As(err, &pqErr) && pqErr.Code == "23505" {
		return fmt.Errorf("упражнение с таким названием уже есть")
	}
	return err
}
