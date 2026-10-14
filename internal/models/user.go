package models

import "time"

const PDPolicyVersion = "2026-10-11"

type User struct {
	ID              int64      `db:"id" json:"id"`
	TelegramID      int64      `db:"telegram_id" json:"telegram_id"`
	FirstName       string     `db:"first_name" json:"first_name"`
	LastName        string     `db:"last_name" json:"last_name"`
	Username        string     `db:"username" json:"username"`
	PasswordHash    *string    `db:"password_hash" json:"-"`
	Role            string     `db:"role" json:"role"` // "student" или "coach"
	IsActive        bool       `db:"is_active" json:"is_active"`
	PDConsentAt     *time.Time `db:"pd_consent_at" json:"pd_consent_at,omitempty"`
	PDPolicyVersion *string    `db:"pd_policy_version" json:"pd_policy_version,omitempty"`
	RegisteredAt    time.Time  `db:"registered_at" json:"registered_at"`
	UpdatedAt       time.Time  `db:"updated_at" json:"updated_at"`
}

func (u *User) HasCurrentPDConsent() bool {
	return u != nil && u.PDConsentAt != nil && u.PDPolicyVersion != nil && *u.PDPolicyVersion == PDPolicyVersion
}
