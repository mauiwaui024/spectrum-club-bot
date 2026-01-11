package coach_service

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"time"
)

type coachService struct {
	coachRepo repository.CoachRepository
}

func NewCoachService(coachRepo repository.CoachRepository) service.CoachService {
	return &coachService{
		coachRepo: coachRepo,
	}
}

func (s *coachService) RegisterCoach(userID int64, specialty, experience, description string) error {
	coach := &models.Coach{
		UserID:      userID,
		Specialty:   specialty,
		Experience:  experience,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.coachRepo.Create(coach)
}

func (s *coachService) GetCoachByUserID(userID int64) (*models.Coach, error) {
	return s.coachRepo.GetByUserID(userID)
}

func (s *coachService) UpdateCoachProfile(coachID int64, specialty, experience, description string) error {
	coach := &models.Coach{
		ID:          coachID,
		Specialty:   specialty,
		Experience:  experience,
		Description: description,
		UpdatedAt:   time.Now(),
	}
	return s.coachRepo.Update(coach)
}

// func (s *coachService) GetAllCoaches() ([]*models.Coach, error) {
// 	return s.coachRepo.GetAll()
// }

func (s *coachService) GetByCoachID(coachID int64) (*models.Coach, error) {
	return s.coachRepo.GetByCoachID(coachID)
}
