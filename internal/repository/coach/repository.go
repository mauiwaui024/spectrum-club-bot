package coach

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type coachRepository struct {
	db *sqlx.DB
}

func NewCoachRepository(db *sqlx.DB) repository.CoachRepository {
	return &coachRepository{db: db}
}

func (r *coachRepository) Create(coach *models.Coach) error {
	query := `
		INSERT INTO spectrum.coaches (user_id, specialty, experience, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		coach.UserID,
		coach.Specialty,
		coach.Experience,
		coach.Description,
		coach.CreatedAt,
		coach.UpdatedAt,
	).Scan(&coach.ID)
}

func (r *coachRepository) GetByUserID(userID int64) (*models.Coach, error) {
	var coach models.Coach
	query := `SELECT * FROM spectrum.coaches WHERE user_id = $1`
	err := r.db.Get(&coach, query, userID)
	return &coach, err
}

func (r *coachRepository) GetByCoachID(coachID int64) (*models.Coach, error) {
	var coach models.Coach
	query := `SELECT * FROM spectrum.coaches WHERE id = $1`
	err := r.db.Get(&coach, query, coachID)
	return &coach, err
}

func (r *coachRepository) Update(coach *models.Coach) error {
	query := `
		UPDATE spectrum.coaches 
		SET specialty = $1, experience = $2, description = $3, updated_at = $4
		WHERE id = $5
	`
	_, err := r.db.Exec(
		query,
		coach.Specialty,
		coach.Experience,
		coach.Description,
		coach.UpdatedAt,
		coach.ID,
	)
	return err
}

func (r *coachRepository) GetAll() ([]*models.Coach, error) {
	var coaches []*models.Coach
	query := `SELECT * FROM spectrum.coaches ORDER BY created_at DESC`
	err := r.db.Select(&coaches, query)
	return coaches, err
}

func (r *coachRepository) GetByID(id int64) (*models.Coach, error) {
	var coach models.Coach
	query := `SELECT * FROM spectrum.coaches WHERE id = $1`
	err := r.db.Get(&coach, query, id)
	return &coach, err
}
