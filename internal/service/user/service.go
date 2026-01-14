package user_service

import (
	"spectrum-club-bot/internal/models"
	"spectrum-club-bot/internal/repository"
	"spectrum-club-bot/internal/service"
	"time"
)

type userService struct {
	userRepo         repository.UserRepository
	studentRepo      repository.StudentRepository
	coachRepo        repository.CoachRepository
	subscriptionRepo repository.SubscriptionRepository
}

func NewUserService(
	userRepo repository.UserRepository,
	studentRepo repository.StudentRepository,
	coachRepo repository.CoachRepository,
	subscriptionRepo repository.SubscriptionRepository,
) service.UserService {
	return &userService{
		userRepo:         userRepo,
		studentRepo:      studentRepo,
		coachRepo:        coachRepo,
		subscriptionRepo: subscriptionRepo,
	}
}

func (s *userService) RegisterOrUpdate(telegramID int64, firstName, lastName, username string, role string) (*models.User, error) {
	user := &models.User{
		TelegramID:   telegramID,
		FirstName:    firstName,
		LastName:     lastName,
		Username:     username,
		Role:         role,
		RegisteredAt: time.Now(),
	}

	err := s.userRepo.CreateOrUpdate(user)
	if err != nil {
		return nil, err
	}

	existingStudent, err := s.studentRepo.GetByUserID(user.ID)
	if err != nil || existingStudent == nil {
		// Создаем нового студента
		student := &models.Student{
			UserID:    user.ID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		err = s.studentRepo.Create(student)
		if err != nil {
			return nil, err
		}

	}

	return user, nil
}

func (s *userService) GetUserProfile(telegramID int64) (*models.User, *models.Student, *models.Subscription, *models.Coach, error) {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return nil, nil, nil, nil, err
	}

	var student *models.Student
	var subscription *models.Subscription
	var coach *models.Coach

	if user.Role == "student" {
		student, err = s.studentRepo.GetByUserID(user.ID)
		if err != nil {
			return user, nil, nil, nil, nil
		}

		if student != nil {
			subscription, err = s.subscriptionRepo.GetActiveByStudentID(student.ID)
			if err != nil {
				return user, student, nil, nil, nil
			}
		}
	} else if user.Role == "coach" {
		coach, err = s.coachRepo.GetByUserID(user.ID)
		if err != nil {
			return user, nil, nil, nil, nil
		}
	}

	return user, student, subscription, coach, nil
}

func (s *userService) SetRole(telegramID int64, role string) error {
	return s.userRepo.UpdateRole(telegramID, role)
}

func (s *userService) RegisterAsCoach(telegramID int64, specialty, experience, description string) error {
	user, err := s.userRepo.GetByTelegramID(telegramID)
	if err != nil {
		return err
	}

	// Обновляем роль пользователя
	err = s.userRepo.UpdateRole(telegramID, "coach")
	if err != nil {
		return err
	}

	// Создаем запись тренера
	coach := &models.Coach{
		UserID:      user.ID,
		Specialty:   specialty,
		Experience:  experience,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
	return s.coachRepo.Create(coach)
}

// В user_service.go
func (s *userService) GetAllStudents() ([]*models.User, error) {
	return s.userRepo.GetAllStudents()
}

func (s *userService) GetByID(id int64) (*models.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *userService) GetByTelegramID(telegramID int64) (*models.User, error) {
	return s.userRepo.GetByTelegramID(telegramID)
}
