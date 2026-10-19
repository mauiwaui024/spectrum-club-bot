package refresh

import (
	"database/sql"
	"time"

	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"

	"github.com/jmoiron/sqlx"
)

type refreshTokenRepository struct {
	db *sqlx.DB
}

func NewRefreshTokenRepository(db *sqlx.DB) repository.RefreshTokenRepository {
	return &refreshTokenRepository{db: db}
}

func (r *refreshTokenRepository) Create(userID int64, tokenHash []byte, expiresAt time.Time) error {
	_, err := r.db.Exec(`
		INSERT INTO spectrum.refresh_tokens (user_id, token_hash, expires_at)
		VALUES ($1, $2, $3)
	`, userID, tokenHash, expiresAt)
	return err
}

func (r *refreshTokenRepository) GetByHash(tokenHash []byte) (*models.RefreshToken, error) {
	var row models.RefreshToken
	err := r.db.Get(&row, `
		SELECT id, user_id, token_hash, expires_at, created_at
		FROM spectrum.refresh_tokens
		WHERE token_hash = $1
	`, tokenHash)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &row, nil
}

func (r *refreshTokenRepository) DeleteByHash(tokenHash []byte) error {
	_, err := r.db.Exec(`DELETE FROM spectrum.refresh_tokens WHERE token_hash = $1`, tokenHash)
	return err
}

func (r *refreshTokenRepository) DeleteAllForUser(userID int64) error {
	_, err := r.db.Exec(`DELETE FROM spectrum.refresh_tokens WHERE user_id = $1`, userID)
	return err
}

func (r *refreshTokenRepository) DeleteExpired(before time.Time) error {
	_, err := r.db.Exec(`DELETE FROM spectrum.refresh_tokens WHERE expires_at < $1`, before)
	return err
}
