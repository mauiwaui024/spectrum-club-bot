package attendance

import (
	"database/sql"
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

func (r *attendanceRepository) GetStudentAttendanceForTraining(studentID, trainingID int) (*models.Attendance, error) {
	query := `
		SELECT 
			a.id, a.training_id, a.student_id, a.attended, a.notes, 
			a.recorded_by, a.recorded_at
		FROM spectrum.attendance a
		WHERE a.student_id = $1 AND a.training_id = $2
	`

	attendance := &models.Attendance{}
	err := r.db.QueryRow(query, studentID, trainingID).Scan(
		&attendance.ID, &attendance.TrainingID, &attendance.StudentID,
		&attendance.Attended, &attendance.Notes, &attendance.RecordedBy,
		&attendance.RecordedAt,
	)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil
		}
		return nil, err
	}
	return attendance, nil
}

func (r *attendanceRepository) UpdateAttendance(attendance *models.Attendance) error {
	query := `
		UPDATE spectrum.attendance 
		SET attended = $1, notes = $2, recorded_by = $3, recorded_at = CURRENT_TIMESTAMP
		WHERE id = $4
		RETURNING recorded_at
	`
	return r.db.QueryRow(
		query,
		attendance.Attended,
		attendance.Notes,
		attendance.RecordedBy,
		attendance.ID,
	).Scan(&attendance.RecordedAt)
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
