package group

import (
	"database/sql"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type trainingGroupRepository struct {
	db *sqlx.DB
}

func NewTrainingGroupRepository(db *sqlx.DB) repository.TrainingGroupRepository {
	return &trainingGroupRepository{db: db}
}

func (r *trainingGroupRepository) GetAllGroups() ([]models.TrainingGroup, error) {
	query := `
        SELECT id, name, code, age_min, age_max, description, created_at
        FROM spectrum.training_groups
        ORDER BY age_min, name
    `

	rows, err := r.db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var groups []models.TrainingGroup
	for rows.Next() {
		var group models.TrainingGroup
		var ageMax sql.NullInt64 // для обработки NULL значений

		err := rows.Scan(
			&group.ID,
			&group.Name,
			&group.Code,
			&group.AgeMin,
			&ageMax,
			&group.Description,
			&group.CreatedAt,
		)
		if err != nil {
			return nil, err
		}

		// Конвертируем sql.NullInt64 в *int
		if ageMax.Valid {
			ageMaxVal := int(ageMax.Int64)
			group.AgeMax = &ageMaxVal
		}

		groups = append(groups, group)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return groups, nil
}

func (r *trainingGroupRepository) GetGroupByID(id int) (*models.TrainingGroup, error) {
	query := `
        SELECT id, name, code, age_min, age_max, description, created_at
        FROM spectrum.training_groups
        WHERE id = $1
    `

	group := &models.TrainingGroup{}
	var ageMax sql.NullInt64

	err := r.db.QueryRow(query, id).Scan(
		&group.ID,
		&group.Name,
		&group.Code,
		&group.AgeMin,
		&ageMax,
		&group.Description,
		&group.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // группа не найдена
		}
		return nil, err
	}

	// Конвертируем sql.NullInt64 в *int
	if ageMax.Valid {
		ageMaxVal := int(ageMax.Int64)
		group.AgeMax = &ageMaxVal
	}

	return group, nil
}

func (r *trainingGroupRepository) GetGroupByCode(code string) (*models.TrainingGroup, error) {
	query := `
        SELECT id, name, code, age_min, age_max, description, created_at
        FROM spectrum.training_groups
        WHERE code = $1
    `

	group := &models.TrainingGroup{}
	var ageMax sql.NullInt64

	err := r.db.QueryRow(query, code).Scan(
		&group.ID,
		&group.Name,
		&group.Code,
		&group.AgeMin,
		&ageMax,
		&group.Description,
		&group.CreatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // группа не найдена
		}
		return nil, err
	}

	// Конвертируем sql.NullInt64 в *int
	if ageMax.Valid {
		ageMaxVal := int(ageMax.Int64)
		group.AgeMax = &ageMaxVal
	}

	return group, nil
}

func (r *trainingGroupRepository) Create(group *models.TrainingGroup) error {
	query := `
		INSERT INTO spectrum.training_groups (name, code, age_min, age_max, description)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, created_at
	`
	return r.db.QueryRow(query, group.Name, group.Code, group.AgeMin, group.AgeMax, group.Description).
		Scan(&group.ID, &group.CreatedAt)
}

func (r *trainingGroupRepository) Update(id int, name, code, description string, ageMin int, ageMax *int) error {
	_, err := r.db.Exec(`
		UPDATE spectrum.training_groups
		SET name = $1, code = $2, age_min = $3, age_max = $4, description = $5
		WHERE id = $6
	`, name, code, ageMin, ageMax, description, id)
	return err
}

func (r *trainingGroupRepository) Delete(id int) error {
	res, err := r.db.Exec(`DELETE FROM spectrum.training_groups WHERE id = $1`, id)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return fmt.Errorf("group not found")
	}
	return nil
}

func (r *trainingGroupRepository) CountFutureTrainings(groupID int) (int, error) {
	var n int
	err := r.db.Get(&n, `
		SELECT COUNT(*) FROM spectrum.training_schedule
		WHERE group_id = $1 AND training_date >= CURRENT_DATE
	`, groupID)
	return n, err
}
