package models

import "time"

type RefreshToken struct {
	ID        int64     `db:"id"`
	UserID    int64     `db:"user_id"`
	TokenHash []byte    `db:"token_hash"`
	ExpiresAt time.Time `db:"expires_at"`
	CreatedAt time.Time `db:"created_at"`
}
