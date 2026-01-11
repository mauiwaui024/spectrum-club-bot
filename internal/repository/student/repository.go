package student

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type studentRepository struct {
	db *sqlx.DB
}

func NewStudentRepository(db *sqlx.DB) repository.StudentRepository {
	return &studentRepository{db: db}
}

func (r *studentRepository) Create(student *models.Student) error {
	query := `
		INSERT INTO spectrum.students (user_id, athletic_title, created_at, updated_at)
		VALUES ($1, $2, $3, $4)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		student.UserID,
		student.AtleticTitle,
		student.CreatedAt,
		student.UpdatedAt,
	).Scan(&student.ID)
}

func (r *studentRepository) GetAll(userID int64) (*models.Student, error) {
	var student models.Student
	query := `SELECT * FROM spectrum.students WHERE user_id = $1`
	err := r.db.Get(&student, query, userID)
	return &student, err
}

func (r *studentRepository) GetByUserID(userID int64) (*models.Student, error) {
	var student models.Student
	query := `SELECT * FROM spectrum.students WHERE user_id = $1`
	err := r.db.Get(&student, query, userID)
	return &student, err
}

func (r *studentRepository) GetByID(id int64) (*models.Student, error) {
	var student models.Student
	query := `SELECT * FROM spectrum.students WHERE id = $1`
	err := r.db.Get(&student, query, id)
	return &student, err
}

func (r *studentRepository) Update(student *models.Student) error {
	query := `
		UPDATE spectrum.students 
		SET athletic_title = $1, updated_at = $2
		WHERE id = $3
	`
	_, err := r.db.Exec(
		query,
		student.AtleticTitle,
		student.UpdatedAt,
		student.ID,
	)
	return err
}
