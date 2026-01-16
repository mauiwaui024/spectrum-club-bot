package schedule

import (
	"database/sql"
	"errors"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"time"

	"github.com/jmoiron/sqlx"
)

type trainingScheduleRepository struct {
	db *sqlx.DB
}

func NewTrainingScheduleRepository(db *sqlx.DB) repository.TrainingScheduleRepository {
	return &trainingScheduleRepository{db: db}
}

func (r *trainingScheduleRepository) CreateTraining(training *models.TrainingSchedule) error {
	query := `
		INSERT INTO spectrum.training_schedule 
		(group_id, coach_id, training_date, start_time, end_time, description, max_participants, created_by)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at
	`
	return r.db.QueryRow(
		query,
		training.GroupID,
		training.CoachID,
		training.TrainingDate,
		training.StartTime,
		training.EndTime,
		training.Description,
		training.MaxParticipants,
		training.CreatedBy,
	).Scan(&training.ID, &training.CreatedAt, &training.UpdatedAt)
}

func (r *trainingScheduleRepository) GetTrainingByID(id int) (*models.TrainingSchedule, error) {
	query := `
		SELECT 
			ts.id, ts.group_id, ts.coach_id, ts.training_date, ts.start_time, 
			ts.end_time, ts.description, ts.max_participants, ts.created_by,
			ts.created_at, ts.updated_at,
			tg.name as group_name,
			u.first_name || ' ' || u.last_name as coach_name
		FROM spectrum.training_schedule ts
		LEFT JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		LEFT JOIN spectrum.coaches c ON ts.coach_id = c.id
		LEFT JOIN spectrum.users u ON c.user_id = u.id
		WHERE ts.id = $1
	`

	training := &models.TrainingSchedule{}
	err := r.db.QueryRow(query, id).Scan(
		&training.ID, &training.GroupID, &training.CoachID, &training.TrainingDate,
		&training.StartTime, &training.EndTime, &training.Description, &training.MaxParticipants,
		&training.CreatedBy, &training.CreatedAt, &training.UpdatedAt,
		&training.GroupName, &training.CoachName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return training, nil
}

func (r *trainingScheduleRepository) GetTrainingsByDate(date time.Time) ([]models.TrainingSchedule, error) {
	query := `
		SELECT 
			ts.id, ts.group_id, ts.coach_id, ts.training_date, ts.start_time, 
			ts.end_time, ts.description, ts.max_participants, ts.created_by,
			ts.created_at, ts.updated_at,
			tg.name as group_name,
			u.first_name || ' ' || u.last_name as coach_name
		FROM spectrum.training_schedule ts
		LEFT JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		LEFT JOIN spectrum.coaches c ON ts.coach_id = c.id
		LEFT JOIN spectrum.users u ON c.user_id = u.id
		WHERE ts.training_date = $1
		ORDER BY ts.start_time ASC
	`

	rows, err := r.db.Query(query, date.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainings []models.TrainingSchedule
	for rows.Next() {
		var training models.TrainingSchedule
		err := rows.Scan(
			&training.ID, &training.GroupID, &training.CoachID, &training.TrainingDate,
			&training.StartTime, &training.EndTime, &training.Description, &training.MaxParticipants,
			&training.CreatedBy, &training.CreatedAt, &training.UpdatedAt,
			&training.GroupName, &training.CoachName,
		)
		if err != nil {
			return nil, err
		}
		trainings = append(trainings, training)
	}

	return trainings, nil
}

func (r *trainingScheduleRepository) GetTrainingsByDateRange(start, end time.Time) ([]models.TrainingSchedule, error) {
	query := `
		SELECT 
			ts.id, ts.group_id, ts.coach_id, ts.training_date, ts.start_time, 
			ts.end_time, ts.description, ts.max_participants, ts.created_by,
			ts.created_at, ts.updated_at,
			tg.name as group_name,
			u.first_name || ' ' || u.last_name as coach_name
		FROM spectrum.training_schedule ts
		LEFT JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		LEFT JOIN spectrum.coaches c ON ts.coach_id = c.id
		LEFT JOIN spectrum.users u ON c.user_id = u.id
		WHERE ts.training_date BETWEEN $1 AND $2
		ORDER BY ts.training_date ASC, ts.start_time ASC
	`

	rows, err := r.db.Query(query, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainings []models.TrainingSchedule
	for rows.Next() {
		var training models.TrainingSchedule
		err := rows.Scan(
			&training.ID, &training.GroupID, &training.CoachID, &training.TrainingDate,
			&training.StartTime, &training.EndTime, &training.Description, &training.MaxParticipants,
			&training.CreatedBy, &training.CreatedAt, &training.UpdatedAt,
			&training.GroupName, &training.CoachName,
		)
		if err != nil {
			return nil, err
		}
		trainings = append(trainings, training)
	}

	return trainings, nil
}

func (r *trainingScheduleRepository) GetTrainingsByGroup(groupID int, start, end time.Time) ([]models.TrainingSchedule, error) {
	query := `
		SELECT 
			ts.id, ts.group_id, ts.coach_id, ts.training_date, ts.start_time, 
			ts.end_time, ts.description, ts.max_participants, ts.created_by,
			ts.created_at, ts.updated_at,
			tg.name as group_name,
			u.first_name || ' ' || u.last_name as coach_name
		FROM spectrum.training_schedule ts
		LEFT JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		LEFT JOIN spectrum.coaches c ON ts.coach_id = c.id
		LEFT JOIN spectrum.users u ON c.user_id = u.id
		WHERE ts.group_id = $1 AND ts.training_date BETWEEN $2 AND $3
		ORDER BY ts.training_date ASC, ts.start_time ASC
	`

	rows, err := r.db.Query(query, groupID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainings []models.TrainingSchedule
	for rows.Next() {
		var training models.TrainingSchedule
		err := rows.Scan(
			&training.ID, &training.GroupID, &training.CoachID, &training.TrainingDate,
			&training.StartTime, &training.EndTime, &training.Description, &training.MaxParticipants,
			&training.CreatedBy, &training.CreatedAt, &training.UpdatedAt,
			&training.GroupName, &training.CoachName,
		)
		if err != nil {
			return nil, err
		}
		trainings = append(trainings, training)
	}

	return trainings, nil
}

func (r *trainingScheduleRepository) GetTrainingsByCoach(coachID int64, start, end time.Time) ([]models.TrainingSchedule, error) {
	query := `
		SELECT 
			ts.id, ts.group_id, ts.coach_id, ts.training_date, ts.start_time, 
			ts.end_time, ts.description, ts.max_participants, ts.created_by,
			ts.created_at, ts.updated_at,
			tg.name as group_name,
			u.first_name || ' ' || u.last_name as coach_name
		FROM spectrum.training_schedule ts
		LEFT JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		LEFT JOIN spectrum.coaches c ON ts.coach_id = c.id
		LEFT JOIN spectrum.users u ON c.user_id = u.id
		WHERE ts.coach_id = $1 AND ts.training_date BETWEEN $2 AND $3
		ORDER BY ts.training_date ASC, ts.start_time ASC
	`

	rows, err := r.db.Query(query, coachID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainings []models.TrainingSchedule
	for rows.Next() {
		var training models.TrainingSchedule
		err := rows.Scan(
			&training.ID, &training.GroupID, &training.CoachID, &training.TrainingDate,
			&training.StartTime, &training.EndTime, &training.Description, &training.MaxParticipants,
			&training.CreatedBy, &training.CreatedAt, &training.UpdatedAt,
			&training.GroupName, &training.CoachName,
		)
		if err != nil {
			return nil, err
		}
		trainings = append(trainings, training)
	}

	return trainings, nil
}

func (r *trainingScheduleRepository) GetAvailableTrainingsForStudent(studentID int, start, end time.Time) ([]models.TrainingSchedule, error) {
	query := `
		SELECT DISTINCT
			ts.id, ts.group_id, ts.coach_id, ts.training_date, ts.start_time, 
			ts.end_time, ts.description, ts.max_participants, ts.created_by,
			ts.created_at, ts.updated_at,
			tg.name as group_name,
			u.first_name || ' ' || u.last_name as coach_name
		FROM spectrum.training_schedule ts
		LEFT JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		LEFT JOIN spectrum.coaches c ON ts.coach_id = c.id
		LEFT JOIN spectrum.users u ON c.user_id = u.id
		WHERE ts.training_date BETWEEN $1 AND $2
		AND ts.training_date >= CURRENT_DATE
		AND (ts.max_participants IS NULL OR ts.max_participants > (
			SELECT COUNT(*) FROM spectrum.attendance a WHERE a.training_id = ts.id
		))
		AND NOT EXISTS (
			SELECT 1 FROM spectrum.attendance a 
			WHERE a.training_id = ts.id AND a.student_id = $3
		)
		ORDER BY ts.training_date ASC, ts.start_time ASC
	`

	rows, err := r.db.Query(query, start.Format("2006-01-02"), end.Format("2006-01-02"), studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var trainings []models.TrainingSchedule
	for rows.Next() {
		var training models.TrainingSchedule
		err := rows.Scan(
			&training.ID, &training.GroupID, &training.CoachID, &training.TrainingDate,
			&training.StartTime, &training.EndTime, &training.Description, &training.MaxParticipants,
			&training.CreatedBy, &training.CreatedAt, &training.UpdatedAt,
			&training.GroupName, &training.CoachName,
		)
		if err != nil {
			return nil, err
		}
		trainings = append(trainings, training)
	}

	return trainings, nil
}

func (r *trainingScheduleRepository) UpdateTraining(training *models.TrainingSchedule) error {
	query := `
		UPDATE spectrum.training_schedule 
		SET group_id = $1, coach_id = $2, training_date = $3, start_time = $4, 
		    end_time = $5, description = $6, max_participants = $7, updated_at = CURRENT_TIMESTAMP
		WHERE id = $8
		RETURNING updated_at
	`
	return r.db.QueryRow(
		query,
		training.GroupID,
		training.CoachID,
		training.TrainingDate,
		training.StartTime,
		training.EndTime,
		training.Description,
		training.MaxParticipants,
		training.ID,
	).Scan(&training.UpdatedAt)
}

func (r *trainingScheduleRepository) DeleteTraining(id int) error {
	query := `DELETE FROM spectrum.training_schedule WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *trainingScheduleRepository) IsCoachAvailable(coachID int64, date time.Time, startTime, endTime time.Time) (bool, error) {
	query := `
		SELECT NOT EXISTS (
			SELECT 1 FROM spectrum.training_schedule 
			WHERE coach_id = $1 
			AND training_date = $2
			AND (
				(start_time <= $3 AND end_time > $3) OR
				(start_time < $4 AND end_time >= $4) OR
				(start_time >= $3 AND end_time <= $4)
			)
		)
	`

	var available bool
	err := r.db.QueryRow(
		query,
		coachID,
		date.Format("2006-01-02"),
		startTime.Format("15:04:05"),
		endTime.Format("15:04:05"),
	).Scan(&available)

	return available, err
}

func (r *trainingScheduleRepository) GetTrainingParticipantsCount(trainingID int) (int, error) {
	query := `SELECT COUNT(*) FROM spectrum.attendance WHERE training_id = $1`
	var count int
	err := r.db.QueryRow(query, trainingID).Scan(&count)
	return count, err
}

func (r *trainingScheduleRepository) Exists(groupID int, startTime time.Time) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM spectrum.training_schedule 
            WHERE group_id = $1 
            AND training_date = $2 
            AND start_time = $3
        )
    `

	var exists bool
	err := r.db.QueryRow(
		query,
		groupID,
		startTime.Format("2006-01-02"),
		startTime.Format("15:04:05"),
	).Scan(&exists)

	return exists, err
}

func (r *trainingScheduleRepository) ExistsForCoach(groupID int, coachID int64, startTime time.Time) (bool, error) {
	query := `
        SELECT EXISTS(
            SELECT 1 FROM spectrum.training_schedule 
            WHERE group_id = $1 
            AND coach_id = $2
            AND training_date = $3 
            AND start_time = $4
        )
    `

	var exists bool
	err := r.db.QueryRow(
		query,
		groupID,
		coachID,
		startTime.Format("2006-01-02"),
		startTime.Format("15:04:05"),
	).Scan(&exists)

	return exists, err
}

func (r *trainingScheduleRepository) UpdateTrainingPartial(id int, updates map[string]interface{}) error {
	if len(updates) == 0 {
		return errors.New("нет полей для обновления")
	}

	// Белый список допустимых полей для обновления
	allowedFields := map[string]bool{
		"group_id":         true,
		"coach_id":         true,
		"training_date":    true,
		"start_time":       true,
		"end_time":         true,
		"description":      true,
		"max_participants": true,
	}

	// Начинаем построение запроса
	query := `UPDATE spectrum.training_schedule SET updated_at = CURRENT_TIMESTAMP`
	args := []interface{}{}
	argCounter := 1

	// Динамически добавляем SET для каждого поля
	for field, value := range updates {
		if !allowedFields[field] {
			continue // Пропускаем неразрешенные поля
		}

		query += fmt.Sprintf(", %s = $%d", field, argCounter)
		args = append(args, value)
		argCounter++
	}

	// Добавляем WHERE условие
	query += fmt.Sprintf(" WHERE id = $%d", argCounter)
	args = append(args, id)

	_, err := r.db.Exec(query, args...)
	return err
}
