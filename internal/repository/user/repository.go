package user

import (
	"database/sql"
	"time"

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

func (r *userRepository) CreateWithPassword(user *models.User, passwordHash string) error {
	query := `
		INSERT INTO spectrum.users (telegram_id, first_name, last_name, username, role, password_hash, registered_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		RETURNING id
	`
	return r.db.QueryRow(
		query,
		user.TelegramID,
		user.FirstName,
		user.LastName,
		user.Username,
		user.Role,
		passwordHash,
		user.RegisteredAt,
	).Scan(&user.ID)
}

func (r *userRepository) UsernameExists(username string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM spectrum.users WHERE lower(username) = lower($1))`
	err := r.db.Get(&exists, query, username)
	return exists, err
}

func (r *userRepository) UpdateRoleByUserID(userID int64, role string) error {
	query := `UPDATE spectrum.users SET role = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2`
	_, err := r.db.Exec(query, role, userID)
	return err
}

func (r *userRepository) GetByTelegramID(telegramID int64) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM spectrum.users WHERE telegram_id = $1`
	err := r.db.Get(&user, query, telegramID)
	return &user, err
}

func (r *userRepository) GetByUsername(username string) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM spectrum.users WHERE username = $1`
	err := r.db.Get(&user, query, username)
	return &user, err
}

func (r *userRepository) GetByID(id int64) (*models.User, error) {
	var user models.User
	query := `SELECT * FROM spectrum.users WHERE id = $1`
	err := r.db.Get(&user, query, id)
	return &user, err
}

func (r *userRepository) RecordPDConsent(userID int64, policyVersion string, at time.Time) error {
	_, err := r.db.Exec(`
		UPDATE spectrum.users
		SET pd_consent_at = $1, pd_policy_version = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`, at, policyVersion, userID)
	return err
}

func (r *userRepository) SetPasswordHash(userID int64, passwordHash string) error {
	query := `
		UPDATE spectrum.users
		SET password_hash = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2
	`
	_, err := r.db.Exec(query, passwordHash, userID)
	return err
}

func (r *userRepository) UpdateRole(telegramID int64, role string) error {
	query := `UPDATE spectrum.users SET role = $1 WHERE telegram_id = $2`
	_, err := r.db.Exec(query, role, telegramID)
	return err
}

func (r *userRepository) UpdateUser(userID int64, firstName, lastName *string) error {
	query := `
		UPDATE spectrum.users 
		SET 
			first_name = COALESCE($1, first_name),
			last_name = COALESCE($2, last_name),
			updated_at = CURRENT_TIMESTAMP
		WHERE id = $3
	`
	_, err := r.db.Exec(query, firstName, lastName, userID)
	return err
}

func (r *userRepository) ListWithoutPassword() ([]*models.User, error) {
	var users []*models.User
	err := r.db.Select(&users, `
		SELECT * FROM spectrum.users
		WHERE is_active = true
		  AND (password_hash IS NULL OR btrim(password_hash) = '')
		ORDER BY last_name, first_name, id
	`)
	if err != nil {
		return nil, err
	}
	if users == nil {
		users = []*models.User{}
	}
	return users, nil
}

func (r *userRepository) SetUsernameIfEmpty(userID int64, username string) error {
	res, err := r.db.Exec(`
		UPDATE spectrum.users
		SET username = $1, updated_at = CURRENT_TIMESTAMP
		WHERE id = $2 AND (username IS NULL OR btrim(username) = '')
	`, username, userID)
	if err != nil {
		return err
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepository) GetAllStudents() ([]*models.User, error) {
	var users []*models.User
	query := `SELECT * FROM spectrum.users WHERE role = 'student' ORDER BY last_name, first_name`

	err := r.db.Select(&users, query)
	if err != nil {
		return nil, err
	}

	return users, nil
}

func (r *userRepository) Delete(userID int64) error {
	res, err := r.db.Exec(`DELETE FROM spectrum.users WHERE id = $1`, userID)
	if err != nil {
		return err
	}
	n, err := res.RowsAffected()
	if err != nil {
		return err
	}
	if n == 0 {
		return sql.ErrNoRows
	}
	return nil
}

func (r *userRepository) SetActive(userID int64, active bool) error {
	_, err := r.db.Exec(`
		UPDATE spectrum.users SET is_active = $1, updated_at = CURRENT_TIMESTAMP WHERE id = $2
	`, active, userID)
	return err
}
