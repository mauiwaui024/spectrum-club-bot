package schedule_template

import (
	"database/sql"
	"errors"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type weekScheduleRepository struct {
	db *sqlx.DB
}

func NewWeekScheduleRepository(db *sqlx.DB) repository.WeekScheduleRepository {
	return &weekScheduleRepository{db: db}
}

func (r *weekScheduleRepository) GetAllActive() ([]models.WeekScheduleTemplate, error) {
	query := `
        SELECT id, group_id, day_of_week, start_time, end_time, 
               description, is_active, created_at, updated_at
        FROM spectrum.week_schedule_templates
        WHERE is_active = TRUE
        ORDER BY day_of_week, start_time
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.WeekScheduleTemplate
	for rows.Next() {
		var t models.WeekScheduleTemplate
		err := rows.Scan(
			&t.ID, &t.GroupID, &t.DayOfWeek, &t.StartTime, &t.EndTime,
			&t.Description, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	return templates, nil
}

func (r *weekScheduleRepository) GetByGroupID(groupID int) ([]models.WeekScheduleTemplate, error) {
	query := `
        SELECT id, group_id, day_of_week, start_time, end_time, 
               description, is_active, created_at, updated_at
        FROM spectrum.week_schedule_templates
        WHERE group_id = $1 AND is_active = TRUE
        ORDER BY day_of_week, start_time
    `

	rows, err := r.db.Query(query, groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var templates []models.WeekScheduleTemplate
	for rows.Next() {
		var t models.WeekScheduleTemplate
		err := rows.Scan(
			&t.ID, &t.GroupID, &t.DayOfWeek, &t.StartTime, &t.EndTime,
			&t.Description, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		templates = append(templates, t)
	}

	return templates, nil
}

func (r *weekScheduleRepository) GetByID(id int) (*models.WeekScheduleTemplate, error) {
	query := `
        SELECT id, group_id, day_of_week, start_time, end_time, 
               description, is_active, created_at, updated_at
        FROM spectrum.week_schedule_templates
        WHERE id = $1
    `

	var t models.WeekScheduleTemplate
	err := r.db.QueryRow(query, id).Scan(
		&t.ID, &t.GroupID, &t.DayOfWeek, &t.StartTime, &t.EndTime,
		&t.Description, &t.IsActive, &t.CreatedAt, &t.UpdatedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, errors.New("шаблон не найден")
		}
		return nil, err
	}

	return &t, nil
}

func (r *weekScheduleRepository) Create(template *models.WeekScheduleTemplate) error {
	query := `
        INSERT INTO spectrum.week_schedule_templates 
        (group_id, day_of_week, start_time, end_time, description, is_active)
        VALUES ($1, $2, $3, $4, $5, $6)
        RETURNING id, created_at, updated_at
    `

	return r.db.QueryRow(
		query,
		template.GroupID,
		template.DayOfWeek,
		template.StartTime,
		template.EndTime,
		template.Description,
		template.IsActive,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
}

func (r *weekScheduleRepository) UpdatePartial(id int, updates map[string]interface{}) error {
	// Базовый запрос
	query := `
        UPDATE spectrum.week_schedule_templates 
        SET updated_at = CURRENT_TIMESTAMP
    `

	// Динамически добавляем SET части
	args := []interface{}{}
	argIndex := 1

	if groupID, ok := updates["group_id"]; ok {
		query += fmt.Sprintf(", group_id = $%d", argIndex)
		args = append(args, groupID)
		argIndex++
	}

	if dayOfWeek, ok := updates["day_of_week"]; ok {
		query += fmt.Sprintf(", day_of_week = $%d", argIndex)
		args = append(args, dayOfWeek)
		argIndex++
	}

	if startTime, ok := updates["start_time"]; ok {
		query += fmt.Sprintf(", start_time = $%d", argIndex)
		args = append(args, startTime)
		argIndex++
	}

	if endTime, ok := updates["end_time"]; ok {
		query += fmt.Sprintf(", end_time = $%d", argIndex)
		args = append(args, endTime)
		argIndex++
	}

	if description, ok := updates["description"]; ok {
		query += fmt.Sprintf(", description = COALESCE(NULLIF($%d, ''), description)", argIndex)
		args = append(args, description)
		argIndex++
	}

	if isActive, ok := updates["is_active"]; ok {
		query += fmt.Sprintf(", is_active = $%d", argIndex)
		args = append(args, isActive)
		argIndex++
	}

	// Добавляем WHERE условие
	query += fmt.Sprintf(" WHERE id = $%d", argIndex)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}

func (r *weekScheduleRepository) Delete(id int) error {
	query := `DELETE FROM spectrum.week_schedule_templates WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *weekScheduleRepository) Activate(id int) error {
	query := `UPDATE spectrum.week_schedule_templates SET is_active = TRUE WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *weekScheduleRepository) Deactivate(id int) error {
	query := `UPDATE spectrum.week_schedule_templates SET is_active = FALSE WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}
