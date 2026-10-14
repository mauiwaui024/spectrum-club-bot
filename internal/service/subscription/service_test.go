package subscription_service

import (
	"spectrum-club-bot/internal/models"
	"testing"
	"time"
)

type mockSubRepo struct {
	created []*models.Subscription
	byID    map[int64]*models.Subscription
}

func (m *mockSubRepo) Create(s *models.Subscription) error {
	s.ID = int64(len(m.created) + 1)
	cp := *s
	m.created = append(m.created, &cp)
	if m.byID == nil {
		m.byID = map[int64]*models.Subscription{}
	}
	m.byID[s.ID] = &cp
	return nil
}
func (m *mockSubRepo) GetByID(id int64) (*models.Subscription, error) {
	return m.byID[id], nil
}
func (m *mockSubRepo) GetActiveByStudentID(int64) (*models.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) GetHistoryByStudentID(int64) ([]*models.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) Update(*models.Subscription) error { return nil }
func (m *mockSubRepo) DecrementRemainingLessons(int64) error {
	return nil
}
func (m *mockSubRepo) AddLessons(id int64, count int) error {
	s := m.byID[id]
	s.RemainingLessons += count
	return nil
}
func (m *mockSubRepo) SetEndDate(id int64, endDate time.Time) error {
	s := m.byID[id]
	s.EndDate = endDate
	return nil
}
func (m *mockSubRepo) RemoveLessons(id int64, count int) error {
	s := m.byID[id]
	s.RemainingLessons -= count
	return nil
}
func (m *mockSubRepo) GetAll() ([]*models.Subscription, error) { return m.created, nil }
func (m *mockSubRepo) GetByStudentID(int64) ([]*models.Subscription, error) {
	return nil, nil
}
func (m *mockSubRepo) Delete(int64) error { return nil }

func TestMaxLessonsForSubscriptionType(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{1, 1}, {2, 2}, {3, 3}, {4, 4}, {8, 8}, {12, 12}, {16, 16},
		{5, 8}, {10, 12}, {20, 16},
	}
	for _, c := range cases {
		if got := MaxLessonsForSubscriptionType(c.in); got != c.want {
			t.Fatalf("MaxLessons(%d)=%d want %d", c.in, got, c.want)
		}
	}
}

func TestCreateWithEndDate_OK(t *testing.T) {
	repo := &mockSubRepo{}
	svc := NewSubscriptionService(repo)
	end := time.Now().AddDate(0, 0, 30)
	if err := svc.CreateWithEndDate(10, 8, end); err != nil {
		t.Fatal(err)
	}
	if len(repo.created) != 1 {
		t.Fatalf("created=%d", len(repo.created))
	}
	s := repo.created[0]
	if s.StudentID != 10 || s.TotalLessons != 8 || s.RemainingLessons != 8 {
		t.Fatalf("unexpected sub: %+v", s)
	}
}

func TestCreateWithEndDate_InvalidLessons(t *testing.T) {
	svc := NewSubscriptionService(&mockSubRepo{})
	err := svc.CreateWithEndDate(1, 7, time.Now().AddDate(0, 0, 1))
	if err == nil {
		t.Fatal("expected error for lessons=7")
	}
}

func TestCreateWithEndDate_PastDate(t *testing.T) {
	svc := NewSubscriptionService(&mockSubRepo{})
	err := svc.CreateWithEndDate(1, 4, time.Now().AddDate(0, 0, -2))
	if err == nil {
		t.Fatal("expected error for past end date")
	}
}

func TestAddLessons_CanRestoreUsed(t *testing.T) {
	repo := &mockSubRepo{byID: map[int64]*models.Subscription{
		1: {ID: 1, TotalLessons: 8, RemainingLessons: 5},
	}}
	svc := NewSubscriptionService(repo)
	if err := svc.AddLessons(1, 3); err != nil {
		t.Fatal(err)
	}
	if repo.byID[1].RemainingLessons != 8 {
		t.Fatalf("remaining = %d, want 8", repo.byID[1].RemainingLessons)
	}
	if repo.byID[1].TotalLessons != 8 {
		t.Fatalf("total = %d, want 8", repo.byID[1].TotalLessons)
	}
}

func TestSetEndDate_RejectsPast(t *testing.T) {
	repo := &mockSubRepo{byID: map[int64]*models.Subscription{
		1: {ID: 1, TotalLessons: 8, RemainingLessons: 8},
	}}
	svc := NewSubscriptionService(repo)
	err := svc.SetEndDate(1, time.Now().AddDate(0, 0, -2))
	if err == nil {
		t.Fatal("expected error for past end date")
	}
	if !repo.byID[1].EndDate.IsZero() {
		t.Fatal("end date should stay unchanged")
	}
}

func TestAddLessons_RespectsTypeLimit(t *testing.T) {
	repo := &mockSubRepo{byID: map[int64]*models.Subscription{
		1: {ID: 1, TotalLessons: 8, RemainingLessons: 8},
	}}
	svc := NewSubscriptionService(repo)
	if err := svc.AddLessons(1, 1); err == nil {
		t.Fatal("expected error when at type limit with no used lessons")
	}
}
