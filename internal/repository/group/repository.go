package group

import (
	"database/sql"
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
