package session_service

import (
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
)

const (
	accessTTL  = 15 * time.Minute
	refreshTTL = 30 * 24 * time.Hour
)

var (
	ErrUnauthenticated = errors.New("user not authenticated")
	ErrInactiveUser    = errors.New("account deactivated")
)

type sessionService struct {
	refreshRepo repository.RefreshTokenRepository
	userRepo    repository.UserRepository
	jwtSecret   []byte
}

func NewSessionService(
	refreshRepo repository.RefreshTokenRepository,
	userRepo repository.UserRepository,
	jwtSecret string,
) service.SessionService {
	return &sessionService{
		refreshRepo: refreshRepo,
		userRepo:    userRepo,
		jwtSecret:   []byte(jwtSecret),
	}
}

func (s *sessionService) IssueSession(userID int64) (string, string, error) {
	access, err := s.signAccess(userID)
	if err != nil {
		return "", "", err
	}
	refresh, err := s.createRefresh(userID)
	if err != nil {
		return "", "", err
	}
	return access, refresh, nil
}

func (s *sessionService) ResolveUserID(accessToken, refreshToken string) (int64, string, string, bool, error) {
	accessToken = strings.TrimSpace(accessToken)
	refreshToken = strings.TrimSpace(refreshToken)

	if accessToken != "" {
		if userID, err := s.verifyAccess(accessToken); err == nil {
			if err := s.ensureActive(userID); err != nil {
				return 0, "", "", false, err
			}
			return userID, "", "", false, nil
		}
	}

	if refreshToken == "" {
		return 0, "", "", false, ErrUnauthenticated
	}

	hash := hashRefresh(refreshToken)
	row, err := s.refreshRepo.GetByHash(hash)
	if err != nil {
		return 0, "", "", false, err
	}
	if row == nil {
		// Possible stolen/replayed refresh: revoke nothing known; treat as unauthenticated.
		return 0, "", "", false, ErrUnauthenticated
	}
	if time.Now().After(row.ExpiresAt) {
		_ = s.refreshRepo.DeleteByHash(hash)
		return 0, "", "", false, ErrUnauthenticated
	}
	if err := s.ensureActive(row.UserID); err != nil {
		_ = s.refreshRepo.DeleteAllForUser(row.UserID)
		return 0, "", "", false, err
	}

	// Rotate refresh: delete old first. If create fails after delete, user must re-login.
	if err := s.refreshRepo.DeleteByHash(hash); err != nil {
		return 0, "", "", false, err
	}
	newRefresh, err := s.createRefresh(row.UserID)
	if err != nil {
		return 0, "", "", false, err
	}
	newAccess, err := s.signAccess(row.UserID)
	if err != nil {
		return 0, "", "", false, err
	}
	return row.UserID, newAccess, newRefresh, true, nil
}

func (s *sessionService) RevokeRefresh(rawRefresh string) error {
	rawRefresh = strings.TrimSpace(rawRefresh)
	if rawRefresh == "" {
		return nil
	}
	return s.refreshRepo.DeleteByHash(hashRefresh(rawRefresh))
}

func (s *sessionService) RevokeAllForUser(userID int64) error {
	return s.refreshRepo.DeleteAllForUser(userID)
}

func (s *sessionService) ensureActive(userID int64) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil || user == nil {
		return ErrUnauthenticated
	}
	if !user.IsActive {
		return ErrInactiveUser
	}
	return nil
}

func (s *sessionService) createRefresh(userID int64) (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	token := hex.EncodeToString(raw)
	if err := s.refreshRepo.Create(userID, hashRefresh(token), time.Now().Add(refreshTTL)); err != nil {
		return "", err
	}
	return token, nil
}

type accessClaims struct {
	UserID int64  `json:"user_id"`
	Exp    int64  `json:"exp"`
	Iat    int64  `json:"iat"`
	Typ    string `json:"typ"`
}

func (s *sessionService) signAccess(userID int64) (string, error) {
	now := time.Now()
	claims := accessClaims{
		UserID: userID,
		Exp:    now.Add(accessTTL).Unix(),
		Iat:    now.Unix(),
		Typ:    "access",
	}
	header := base64.RawURLEncoding.EncodeToString([]byte(`{"alg":"HS256","typ":"JWT"}`))
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", err
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	mac := hmac.New(sha256.New, s.jwtSecret)
	_, _ = mac.Write([]byte(signingInput))
	sig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
	return signingInput + "." + sig, nil
}

func (s *sessionService) verifyAccess(token string) (int64, error) {
	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return 0, fmt.Errorf("invalid token")
	}
	signingInput := parts[0] + "." + parts[1]
	sig, err := base64.RawURLEncoding.DecodeString(parts[2])
	if err != nil {
		return 0, fmt.Errorf("invalid token")
	}
	mac := hmac.New(sha256.New, s.jwtSecret)
	_, _ = mac.Write([]byte(signingInput))
	if !hmac.Equal(sig, mac.Sum(nil)) {
		return 0, fmt.Errorf("invalid token")
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return 0, fmt.Errorf("invalid token")
	}
	var claims accessClaims
	if err := json.Unmarshal(payload, &claims); err != nil || claims.Typ != "access" || claims.UserID <= 0 {
		return 0, fmt.Errorf("invalid token")
	}
	if time.Now().Unix() > claims.Exp {
		return 0, fmt.Errorf("expired")
	}
	return claims.UserID, nil
}

func hashRefresh(raw string) []byte {
	sum := sha256.Sum256([]byte(raw))
	return sum[:]
}
