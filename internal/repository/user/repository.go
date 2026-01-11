package user

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type userRepository struct {
	db *sqlx.DB
}

func NewUserRepository(db *sqlx.DB) repository.UserRepository {
	return &userRepository{db: db}
}

func (r *userRepository) CreateOrUpdate(user *models.User) error {
	query := `
		INSERT INTO spectrum.users (telegram_id, first_name, last_name, username, role, registered_at)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (telegram_id) 
		DO UPDATE SET 
			first_name = EXCLUDED.first_name,
			last_name = EXCLUDED.last_name,
			username = EXCLUDED.username,
			updated_at = CURRENT_TIMESTAMP
		RETURNING id
	`

	return r.db.QueryRow(
		query,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Role,
		user.RegisteredAt,
	).Scan(&user.ID)
}

func (r *userRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM spectrum.users WHERE telegram_id = $1`
	err := r.db.Get(&user, query, telegramID)
	return &user, err
}

func (r *userRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM spectrum.users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	return &user, err
}

func (r *userRepository) UpdateRole(telegramID int64, role string) error {
	query := `UPDATE spectrum.users SET role = $1 WHERE telegram_id = $2`
	_, err := r.db.Exec(query, role, telegramID)
	return err
}

func (r *userRepository) GetAllStudents() ([]*models.User, error) {
	var users []*models.User
	query := `SELECT * FROM spectrum.users WHERE role = 'student' ORDER BY first_name, last_name`

	err := r.db.Select(&users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}
