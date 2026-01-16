package attendance

import (
	"database/sql"
	"fmt"
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"time"

	"github.com/jmoiron/sqlx"
)

type attendanceRepository struct {
	db *sqlx.DB
}

func NewAttendanceRepository(db *sqlx.DB) repository.AttendanceRepository {
	return &attendanceRepository{db: db}
}

func (r *attendanceRepository) CreateAttendance(attendance *models.Attendance) error {
	query := `
		INSERT INTO spectrum.attendance 
		(training_id, student_id, attended, notes, recorded_by)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, recorded_at
	`
	return r.db.QueryRow(
		query,
		attendance.TrainingID,
		attendance.StudentID,
		attendance.Attended,
		attendance.Notes,
		attendance.RecordedBy,
	).Scan(&attendance.ID, &attendance.RecordedAt)
}

func (r *attendanceRepository) GetAttendanceByID(id int) (*models.Attendance, error) {
	query := `
		SELECT 
			a.id, a.training_id, a.student_id, a.attended, a.notes, 
			a.recorded_by, a.recorded_at,
			u.first_name || ' ' || u.last_name as student_name
		FROM spectrum.attendance a
		JOIN spectrum.students s ON a.student_id = s.id
		JOIN spectrum.users u ON s.user_id = u.id
		WHERE a.id = $1
	`

	attendance := &models.Attendance{}
	err := r.db.QueryRow(query, id).Scan(
		&attendance.ID, &attendance.TrainingID, &attendance.StudentID,
		&attendance.Attended, &attendance.Notes, &attendance.RecordedBy,
		&attendance.RecordedAt, &attendance.StudentName,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return attendance, nil
}

func (r *attendanceRepository) GetAttendanceByTraining(trainingID int) ([]models.Attendance, error) {
	query := `
		SELECT 
			a.id, a.training_id, a.student_id, a.attended, a.notes, 
			a.recorded_by, a.recorded_at,
			u.first_name || ' ' || u.last_name as student_name
		FROM spectrum.attendance a
		JOIN spectrum.students s ON a.student_id = s.id
		JOIN spectrum.users u ON s.user_id = u.id
		WHERE a.training_id = $1
		ORDER BY u.first_name, u.last_name
	`

	rows, err := r.db.Query(query, trainingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []models.Attendance
	for rows.Next() {
		var attendance models.Attendance
		err := rows.Scan(
			&attendance.ID, &attendance.TrainingID, &attendance.StudentID,
			&attendance.Attended, &attendance.Notes, &attendance.RecordedBy,
			&attendance.RecordedAt, &attendance.StudentName,
		)
		if err != nil {
			return nil, err
		}
		attendances = append(attendances, attendance)
	}

	return attendances, nil
}

func (r *attendanceRepository) GetAttendanceByStudent(studentID int, start, end time.Time) ([]models.Attendance, error) {
	query := `
		SELECT 
			a.id, a.training_id, a.student_id, a.attended, a.notes, 
			a.recorded_by, a.recorded_at,
			u.first_name || ' ' || u.last_name as student_name,
			ts.training_date, ts.start_time, ts.end_time,
			tg.name as group_name
		FROM spectrum.attendance a
		JOIN spectrum.students s ON a.student_id = s.id
		JOIN spectrum.users u ON s.user_id = u.id
		JOIN spectrum.training_schedule ts ON a.training_id = ts.id
		JOIN spectrum.training_groups tg ON ts.group_id = tg.id
		WHERE a.student_id = $1 AND ts.training_date BETWEEN $2 AND $3
		ORDER BY ts.training_date DESC, ts.start_time DESC
	`

	rows, err := r.db.Query(query, studentID, start.Format("2006-01-02"), end.Format("2006-01-02"))
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attendances []models.Attendance
	for rows.Next() {
		var attendance models.Attendance
		var trainingDate time.Time
		var startTime, endTime time.Time
		var groupName string

		err := rows.Scan(
			&attendance.ID, &attendance.TrainingID, &attendance.StudentID,
			&attendance.Attended, &attendance.Notes, &attendance.RecordedBy,
			&attendance.RecordedAt, &attendance.StudentName,
			&trainingDate, &startTime, &endTime, &groupName,
		)
		if err != nil {
			return nil, err
		}

		attendances = append(attendances, attendance)
	}

	return attendances, nil
}

// func (r *attendanceRepository) GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error) {
// 	query := `
// 		SELECT
// 			a.id, a.training_id, a.student_id, a.attended, a.notes,
// 			a.recorded_by, a.recorded_at
// 		FROM spectrum.attendance a
// 		WHERE a.student_id = $1 AND a.training_id = $2
// 	`

// 	attendance := &models.Attendance{}
// 	err := r.db.QueryRow(query, studentID, trainingID).Scan(
// 		&attendance.ID, &attendance.TrainingID, &attendance.StudentID,
// 		&attendance.Attended, &attendance.Notes, &attendance.RecordedBy,
// 		&attendance.RecordedAt, &attendance.Status, &attendance.CreatedAt, &attendance.UpdatedAt,
// 	)
// 	if err != nil {
// 		if err == sql.ErrNoRows {
// 			return nil, nil
// 		}
// 		return nil, err
// 	}
// 	return attendance, nil
// }

func (r *attendanceRepository) GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error) {
	query := `
        SELECT 
            id, training_id, student_id, attended, notes, 
            recorded_by, recorded_at, created_at, updated_at, status
        FROM spectrum.attendance 
        WHERE student_id = $1 AND training_id = $2
    `

	attendance := &models.Attendance{}

	// Используем sql.NullBool для правильной обработки NULL значений
	var attendedNull sql.NullBool

	err := r.db.QueryRow(query, studentID, trainingID).Scan(
		&attendance.ID,
		&attendance.TrainingID,
		&attendance.StudentID,
		&attendedNull,
		&attendance.Notes,
		&attendance.RecordedBy,
		&attendance.RecordedAt,
		&attendance.CreatedAt,
		&attendance.UpdatedAt,
		&attendance.Status,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		// Логируем ошибку для отладки
		fmt.Printf("Ошибка в GetStudentAttendanceForTraining: %v\n", err)
		fmt.Printf("studentID: %d, trainingID: %d\n", studentID, trainingID)
		return nil, err
	}

	// Преобразуем sql.NullBool в bool (если NULL, то false)
	attendance.Attended = attendedNull.Bool
	if !attendedNull.Valid {
		fmt.Printf("[GetStudentAttendanceForTraining] attended был NULL для id=%d, устанавливаем false\n", attendance.ID)
	}

	return attendance, nil
}

func (r *attendanceRepository) UpdateAttendance(attendance *models.Attendance) error {
	// Проверка, что ID установлен
	if attendance.ID == 0 {
		return fmt.Errorf("attendance ID не установлен (ID=0)")
	}

	// Логирование для отладки
	fmt.Printf("[UpdateAttendance] Обновление записи: id=%d, attended=%v, notes=%s, recorded_by=%v\n",
		attendance.ID, attendance.Attended, attendance.Notes, attendance.RecordedBy)

	// Сначала проверим текущее значение в БД
	var currentAttended sql.NullBool
	checkQuery := `SELECT attended FROM spectrum.attendance WHERE id = $1`
	err := r.db.QueryRow(checkQuery, attendance.ID).Scan(&currentAttended)
	if err != nil {
		fmt.Printf("[UpdateAttendance] Ошибка проверки текущего значения: %v\n", err)
	} else {
		fmt.Printf("[UpdateAttendance] Текущее значение в БД: attended=%v (valid=%v)\n",
			currentAttended.Bool, currentAttended.Valid)
	}

	// Используем Exec с проверкой количества обновленных строк
	query := `
		UPDATE spectrum.attendance 
		SET 
			attended = $1::boolean,
			notes = COALESCE(NULLIF($2, ''), notes),
			recorded_by = COALESCE($3, recorded_by),
			recorded_at = COALESCE($4, recorded_at),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $5
	`

	result, err := r.db.Exec(
		query,
		attendance.Attended,
		attendance.Notes,
		attendance.RecordedBy,
		attendance.RecordedAt,
		attendance.ID,
	)

	if err != nil {
		return fmt.Errorf("ошибка выполнения UPDATE запроса: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("ошибка получения количества обновленных строк: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("не найдена запись с id=%d для обновления", attendance.ID)
	}

	fmt.Printf("[UpdateAttendance] Обновлено строк: %d для id=%d\n", rowsAffected, attendance.ID)

	// Проверяем, что значение действительно обновилось, делая SELECT
	verifyQuery := `SELECT attended, recorded_at, updated_at FROM spectrum.attendance WHERE id = $1`
	var returnedAttendedNull sql.NullBool
	err = r.db.QueryRow(verifyQuery, attendance.ID).Scan(
		&returnedAttendedNull,
		&attendance.RecordedAt,
		&attendance.UpdatedAt,
	)

	if err != nil {
		return fmt.Errorf("ошибка проверки обновленного значения: %w", err)
	}

	// Преобразуем sql.NullBool в bool для сравнения
	returnedAttended := false
	if returnedAttendedNull.Valid {
		returnedAttended = returnedAttendedNull.Bool
	} else {
		// Если значение NULL, это ошибка - мы только что обновили его
		return fmt.Errorf("значение attended NULL после обновления для id=%d", attendance.ID)
	}

	// Проверка, что значение действительно обновилось
	if returnedAttended != attendance.Attended {
		return fmt.Errorf("значение attended не обновилось: ожидалось %v, получено %v (valid=%v)",
			attendance.Attended, returnedAttended, returnedAttendedNull.Valid)
	}

	fmt.Printf("[UpdateAttendance] Успешно обновлено: id=%d, attended в БД=%v (ожидалось %v, valid=%v)\n",
		attendance.ID, returnedAttended, attendance.Attended, returnedAttendedNull.Valid)

	return nil
}

func (r *attendanceRepository) DeleteAttendance(id int) error {
	query := `DELETE FROM spectrum.attendance WHERE id = $1`
	_, err := r.db.Exec(query, id)
	return err
}

func (r *attendanceRepository) GetTrainingAttendanceStats(trainingID int) (present, absent, total int, err error) {
	query := `
		SELECT 
			COUNT(*) as total,
			COUNT(CASE WHEN attended = true THEN 1 END) as present,
			COUNT(CASE WHEN attended = false THEN 1 END) as absent
		FROM spectrum.attendance 
		WHERE training_id = $1
	`

	err = r.db.QueryRow(query, trainingID).Scan(&total, &present, &absent)
	return present, absent, total, err
}

// Добавьте эти методы в существующую структуру
// todo://пофиксить добавить
func (r *attendanceRepository) GetParticipants(trainingID int) ([]models.AttendanceWithStudent, error) {
	query := `
        SELECT 
            a.id, 
            a.training_id, 
            a.student_id, 
            a.status,
            a.notes,
            a.created_at,
            a.updated_at,
            COALESCE(u.first_name || ' ' || u.last_name, 'Неизвестный') as student_name
        FROM spectrum.attendance a
        LEFT JOIN spectrum.students s ON a.student_id = s.id
        LEFT JOIN spectrum.users u ON s.user_id = u.id
        WHERE a.training_id = $1 AND a.status = 'registered'
        ORDER BY a.created_at ASC
    `

	rows, err := r.db.Query(query, trainingID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var participants []models.AttendanceWithStudent
	for rows.Next() {
		var participant models.AttendanceWithStudent
		err := rows.Scan(
			&participant.ID,
			&participant.TrainingID,
			&participant.StudentID,
			&participant.Status,
			&participant.Notes,
			&participant.CreatedAt,
			&participant.UpdatedAt,
			&participant.StudentName,
		)
		if err != nil {
			return nil, err
		}
		participants = append(participants, participant)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return participants, nil
}

func (r *attendanceRepository) GetStudentSchedule(studentID int, start, end time.Time) ([]models.AttendanceWithTraining, error) {
	query := `
        SELECT a.id, a.training_id, a.student_id, a.status, a.attended, 
               a.notes, a.recorded_by, a.recorded_at, a.created_at, a.updated_at,
               t.id, t.group_id, t.coach_id, t.training_date, t.start_time, t.end_time,
               t.description, t.max_participants, t.created_by, t.created_at, t.updated_at,
               g.name as group_name
        FROM spectrum.attendance a
        JOIN spectrum.training_schedule t ON a.training_id = t.id
        LEFT JOIN spectrum.training_groups g ON t.group_id = g.id
        WHERE a.student_id = $1 
          AND a.status = 'registered'
          AND t.training_date BETWEEN $2 AND $3
        ORDER BY t.training_date, t.start_time
    `

	rows, err := r.db.Query(query, studentID, start, end)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var schedule []models.AttendanceWithTraining
	for rows.Next() {
		var item models.AttendanceWithTraining
		err := rows.Scan(
			&item.ID, &item.TrainingID, &item.StudentID, &item.Status, &item.Attended,
			&item.Notes, &item.RecordedBy, &item.RecordedAt, &item.CreatedAt, &item.UpdatedAt,
			&item.Training.ID, &item.Training.GroupID, &item.Training.CoachID,
			&item.Training.TrainingDate, &item.Training.StartTime, &item.Training.EndTime,
			&item.Training.Description, &item.Training.MaxParticipants, &item.Training.CreatedBy,
			&item.Training.CreatedAt, &item.Training.UpdatedAt,
			&item.Training.GroupName,
		)
		if err != nil {
			return nil, err
		}
		schedule = append(schedule, item)
	}

	return schedule, nil
}

func (r *attendanceRepository) CreateAttendanceRecord(attendance models.Attendance) error {
	query := `
        INSERT INTO spectrum.attendance 
        (training_id, student_id, status, attended, notes, recorded_by, recorded_at)
        VALUES ($1, $2, $3, $4, $5, $6, $7)
        RETURNING id, created_at, updated_at
    `

	return r.db.QueryRow(
		query,
		attendance.TrainingID,
		attendance.StudentID,
		attendance.Status,
		attendance.Attended,
		attendance.Notes,
		attendance.RecordedBy,
		attendance.RecordedAt,
	).Scan(&attendance.ID, &attendance.CreatedAt, &attendance.UpdatedAt)
}

func (r *attendanceRepository) CancelAttendance(trainingID, studentID int) error {
	query := `
        UPDATE spectrum.attendance 
        SET status = 'cancelled', updated_at = NOW()
        WHERE training_id = $1 AND student_id = $2 AND status = 'registered'
    `

	_, err := r.db.Exec(query, trainingID, studentID)
	return err
}
