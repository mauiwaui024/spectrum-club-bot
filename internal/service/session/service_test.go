package session_service

import (
	"errors"
	"testing"
	"time"

	"spectrum-club-bot/internal/models"
)

type memRefresh struct {
	byHash map[string]*models.RefreshToken
	seq    int64
}

func (m *memRefresh) Create(userID int64, tokenHash []byte, expiresAt time.Time) error {
	m.seq++
	key := string(tokenHash)
	m.byHash[key] = &models.RefreshToken{
		ID: m.seq, UserID: userID, TokenHash: append([]byte(nil), tokenHash...), ExpiresAt: expiresAt, CreatedAt: time.Now(),
	}
	return nil
}
func (m *memRefresh) GetByHash(tokenHash []byte) (*models.RefreshToken, error) {
	return m.byHash[string(tokenHash)], nil
}
func (m *memRefresh) DeleteByHash(tokenHash []byte) error {
	delete(m.byHash, string(tokenHash))
	return nil
}
func (m *memRefresh) DeleteAllForUser(userID int64) error {
	for k, v := range m.byHash {
		if v.UserID == userID {
			delete(m.byHash, k)
		}
	}
	return nil
}
func (m *memRefresh) DeleteExpired(time.Time) error { return nil }

type memUsers struct {
	byID map[int64]*models.User
}

func (m *memUsers) CreateOrUpdate(*models.User) error              { return nil }
func (m *memUsers) CreateWithPassword(*models.User, string) error  { return nil }
func (m *memUsers) GetByTelegramID(int64) (*models.User, error)    { return nil, errors.New("x") }
func (m *memUsers) GetByUsername(string) (*models.User, error)     { return nil, errors.New("x") }
func (m *memUsers) UsernameExists(string) (bool, error)            { return false, nil }
func (m *memUsers) GetByID(id int64) (*models.User, error)         { return m.byID[id], nil }
func (m *memUsers) SetPasswordHash(int64, string) error            { return nil }
func (m *memUsers) UpdateRole(int64, string) error                 { return nil }
func (m *memUsers) UpdateRoleByUserID(int64, string) error         { return nil }
func (m *memUsers) UpdateUser(int64, *string, *string) error       { return nil }
func (m *memUsers) SetActive(int64, bool) error                    { return nil }
func (m *memUsers) RecordPDConsent(int64, string, time.Time) error { return nil }
func (m *memUsers) ListWithoutPassword() ([]*models.User, error)   { return nil, nil }
func (m *memUsers) SetUsernameIfEmpty(int64, string) error { return nil }
func (m *memUsers) Delete(int64) error                     { return nil }
func (m *memUsers) GetAllStudents() ([]*models.User, error)        { return nil, nil }

func TestIssueAndResolveAccess(t *testing.T) {
	svc := NewSessionService(
		&memRefresh{byHash: map[string]*models.RefreshToken{}},
		&memUsers{byID: map[int64]*models.User{7: {ID: 7, IsActive: true}}},
		"test-secret",
	)
	access, refresh, err := svc.IssueSession(7)
	if err != nil {
		t.Fatal(err)
	}
	uid, _, _, rotated, err := svc.ResolveUserID(access, refresh)
	if err != nil || uid != 7 || rotated {
		t.Fatalf("uid=%d rotated=%v err=%v", uid, rotated, err)
	}
}

func TestSilentRefreshRotates(t *testing.T) {
	repo := &memRefresh{byHash: map[string]*models.RefreshToken{}}
	svc := NewSessionService(repo, &memUsers{byID: map[int64]*models.User{7: {ID: 7, IsActive: true}}}, "test-secret")
	_, refresh, err := svc.IssueSession(7)
	if err != nil {
		t.Fatal(err)
	}
	uid, newAccess, newRefresh, rotated, err := svc.ResolveUserID("bad.access.token", refresh)
	if err != nil || uid != 7 || !rotated || newAccess == "" || newRefresh == "" {
		t.Fatalf("uid=%d rotated=%v err=%v", uid, rotated, err)
	}
	if _, ok := repo.byHash[string(hashRefresh(refresh))]; ok {
		t.Fatal("old refresh should be deleted")
	}
	// Old refresh must not work again.
	if _, _, _, _, err := svc.ResolveUserID("", refresh); err == nil {
		t.Fatal("expected reuse failure")
	}
}

func TestInactiveUserRejected(t *testing.T) {
	svc := NewSessionService(
		&memRefresh{byHash: map[string]*models.RefreshToken{}},
		&memUsers{byID: map[int64]*models.User{7: {ID: 7, IsActive: false}}},
		"test-secret",
	)
	access, refresh, err := svc.IssueSession(7)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, _, _, err := svc.ResolveUserID(access, refresh); !errors.Is(err, ErrInactiveUser) {
		t.Fatalf("got %v", err)
	}
}
